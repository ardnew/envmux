// Package env defines an environment model that can parse and evaluate
// namespaced variables defined with complex expressions.
package env

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"runtime"
	"slices"

	"github.com/carlmjohnson/flowmatic"
	"github.com/expr-lang/expr"

	"github.com/ardnew/envmux/config/env/vars"
	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
)

// Model evaluates namespaced environments with expression support.
// Parses and evaluates environment definitions from a custom file format.
type Model struct {
	*parse.AST `json:"namespaces"`

	// Maximum number of jobs that may be run simultaneously.
	MaxParallelJobs int `json:"jobs,omitempty"`

	// Whether the model requires all namespaces be defined for evaluation.
	EvalRequiresDef bool `json:"requires,omitempty"`
}

// Parse reads a namespace definition from the given [io.Reader] and returns a
// [Model] that can be used to evaluate constructed environments.
func (m Model) Parse(ctx context.Context, r io.Reader) (Model, error) {
	ast, err := parse.Make(ctx, r)
	if err != nil {
		return Model{}, err
	}

	return pkg.Wrap(m, WithAST(ast)), nil
}

// WithAST is a functional [pkg.Option] that installs the namespace
// definitions parsed from a [parse.AST].
//
// It is a required option that must be applied prior to evaluating environment
// variables with [Model.Eval].
func WithAST(ast *parse.AST) pkg.Option[Model] {
	return func(m Model) Model {
		m.AST = ast

		return m
	}
}

// WithMaxParallelJobs is a functional [pkg.Option] that sets the maximum
// number of parallel jobs to run when evaluating the environment.
//
// The default number of jobs is equal to the number of CPU cores available.
// A value of 0 means to use the default number of jobs.
func WithMaxParallelJobs(n int) pkg.Option[Model] {
	return func(m Model) Model {
		m.MaxParallelJobs = n

		return m
	}
}

// WithEvalRequiresDef is a functional [pkg.Option] that sets whether the
// model requires all namespaces to be defined for evaluation.
func WithEvalRequiresDef(b bool) pkg.Option[Model] {
	return func(m Model) Model {
		m.EvalRequiresDef = b

		return m
	}
}

func (m Model) String() string {
	e, err := json.Marshal(m)
	if err != nil {
		return pkg.JoinErrors(pkg.ErrInvalidJSON, err).Error()
	}

	return string(e)
}

func (m Model) IsZero() bool { return m.AST == nil }

// func (m Model) Eval(
// 	ctx context.Context, namespaces ...string,
// ) (vars.Env[any], error) {
// 	return nil, nil
// }

func (m Model) Eval(
	ctx context.Context, namespaces ...string,
) (vars.Env[any], error) {
	s := slices.Collect(pkg.Filter(slices.Values(namespaces),
		func(ns string) bool { return ns != "" },
	))

	list := make([]parse.Composite, len(s))
	for i, id := range s {
		list[i] = parse.Composite{Ident: id}
	}

	env, err := m.eval(ctx, list...)

	return env.eval, err
}

type parameterEnv struct {
	eval vars.Env[any]
	pars []any
}

func (m Model) eval(
	ctx context.Context, composites ...parse.Composite,
) (parameterEnv, error) {
	if len(composites) == 0 {
		return parameterEnv{}, nil //nolint:exhaustruct
	}

	maxJobs := min(len(composites), runtime.NumCPU())
	if m.MaxParallelJobs <= 0 {
		m.MaxParallelJobs = maxJobs
	}

	maxJobs = min(m.MaxParallelJobs, maxJobs)

	env := parameterEnv{
		eval: vars.Env[any]{},
		pars: []any{},
	}

	// If we have only one job, run the evaluations serially in this goroutine
	// and don't fan out additional goroutines.
	if m.MaxParallelJobs == 1 {
		for _, co := range composites {
			e, err := m.evalComposition(ctx, co)
			if err != nil {
				return parameterEnv{}, err
			}

			env = pkg.Wrap(env, export(e))
		}
	} else {
		// Evaluate all namespaces in parallel,
		// returning a slice of fully-evaluated environments.
		for group := range slices.Chunk(composites, maxJobs) {
			envs, err := flowmatic.Map(ctx, len(group), group, m.evalComposition)
			if err != nil {
				return parameterEnv{}, err
			}

			env = pkg.Wrap(env, export(envs...))
		}
	}

	return env, nil
}

func export(sub ...parameterEnv) pkg.Option[parameterEnv] {
	return func(env parameterEnv) parameterEnv {
		for _, e := range sub {
			if e.eval == nil {
				continue
			}

			maps.Copy(env.eval, e.eval)
			env.pars = append(env.pars, e.pars...)
		}

		return env
	}
}

// FindDuplicateNamespaces is a debug option that panics on the
// detection of duplicate namespace definitions in the model,
//
// These most likely occur when a namespace is composed of itself
// indirectly or when the same source is included multiple times.
const findDuplicateNamespaces = false

func (m Model) evalComposition(
	ctx context.Context,
	composite parse.Composite,
) (parameterEnv, error) {
	matchIdent := func(ns parse.Namespace) bool {
		return ns.Ident == composite.Ident
	}

	// Locate the first namespace in the receiver whose identifier
	// matches the given composite namespace identifier.
	//
	// This protects against duplicate namespace definitions
	// and recursive namespace compositions,
	// but it doesn't currently notify the user of such conflicts.
	idx := slices.IndexFunc(m.Namespaces, matchIdent)

	// Verify that the namespace exists in the model.
	if idx < 0 {
		// Throw an error if we have enabled the option that
		// requires a definition for each namespace evaluated.
		if m.EvalRequiresDef {
			return parameterEnv{}, fmt.Errorf(
				"%w: %q is undefined",
				pkg.ErrInvalidNamespace,
				composite,
			)
		}

		// Otherwise, we silently ignore unknown namespaces.
		return parameterEnv{}, nil //nolint:exhaustruct
	}

	// This is the real namespace def resolved
	// from the given composite namespace identifier.
	def := m.Namespaces[idx]

	if findDuplicateNamespaces {
		// Now that it has been resolved, we can search for duplicate definitions
		// (and not just duplicate identifiers).
		matchDups := func(s string) func(ns parse.Namespace) bool {
			return func(ns parse.Namespace) bool {
				return ns.String() == s
			}
		}(def.String())

		pos, dup := idx, make([]int, 0, len(m.Namespaces)-idx-1)

		for 0 <= pos && pos < len(m.Namespaces)-1 {
			off := pos + 1
			if pos = slices.IndexFunc(m.Namespaces[off:], matchDups); pos >= 0 {
				dup = append(dup, off+pos)
			}
		}

		if numDuplicates := len(dup); numDuplicates > 0 {
			panic(fmt.Sprintf(
				"found %d duplicate definitions for namespace:\n\t%s\n",
				numDuplicates,
				def.String(),
			))
		}
	}

	env := parameterEnv{
		eval: vars.Env[any]{},
		pars: []any{},
	}

	// Recursively evaluate and collect the environments of
	// all composite namespaces declared by the current namespace.
	if len(def.Composites) > 0 {
		var err error

		env, err = m.eval(ctx, def.Composites...)
		if err != nil {
			return parameterEnv{}, err
		}
	}

	// Collect all parameters:
	par := slices.Concat(
		slices.Collect(def.Arguments()),       // namespace definition
		env.pars,                              // composite definitions
		slices.Collect(composite.Arguments()), // composite inline parameters
	)

	// Evaluate mappings
	for _, sta := range def.Statements {
		v, err := m.evalMapping(ctx, def, sta, env.eval, par)
		if err != nil {
			return parameterEnv{}, err
		}

		maps.Copy(env.eval, v)
	}

	return env, nil
}

func collect(e vars.Env[any]) vars.Env[any] {
	return maps.Collect(
		pkg.FilterKeys(vars.Cache().Complement(e),
			func(key string) bool {
				return key != vars.ContextKey && key != vars.ParameterKey
			},
		),
	)
}

// evalMapping evaluates a single mapping across all applicable parameters.
func (m Model) evalMapping(
	ctx context.Context,
	space parse.Namespace,
	dict parse.Statement,
	eval vars.Env[any],
	pars []any,
) (vars.Env[any], error) {
	env := pkg.Make(vars.WithContext(ctx), vars.WithExports(eval))

	env[vars.ParameterKey] = any(nil) // placeholder for compiler

	// We have to pass the environment to both [expr.Compile] and [expr.Run].
	// The former builds type information for validating the latter.
	opt := []expr.Option{
		expr.Env(env.AsMap()),
		expr.WithContext(vars.ContextKey),
		expr.AllowUndefinedVariables(),
	}

	// Compile expression
	program, err := expr.Compile(dict.Expression.Src, opt...)
	if err != nil {
		return nil, pkg.ExpressionError{
			Namespace: space.Ident, Statement: dict.Ident, Err: err,
		}
	}

	// Handle case with no parameters
	if len(pars) == 0 {
		pars = []any{nil}

		delete(env, vars.ParameterKey)
	}

	// Process each parameter
	for _, par := range pars {
		if par != nil {
			// If the parameter is a string, unquote it to make a literal value.
			env[vars.ParameterKey] = pkg.Unquote(par)
		}

		res, err := expr.Run(program, env.AsMap())
		if err != nil {
			return nil, err
		}

		env[dict.Ident] = pkg.Unquote(res)
	}

	return collect(env), nil
}
