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
func (m Model) Parse(r io.Reader) (Model, error) {
	ast, err := parse.Make(r)()
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

func (m Model) Eval(
	ctx context.Context, namespaces ...string,
) (vars.Env[any], error) {
	s := slices.Collect(pkg.Filter(slices.Values(namespaces),
		func(ns string) bool { return ns != "" },
	))

	list := make([]parse.Composite, len(s))
	for i, id := range s {
		list[i] = parse.Composite{ID: id}
	}

	env, err := m.eval(ctx, list...)

	return env.eval, err
}

type parameterEnv struct {
	eval vars.Env[any]
	pars []any
}

func (m Model) eval(
	ctx context.Context, namespaces ...parse.Composite,
) (parameterEnv, error) {
	if len(namespaces) == 0 {
		return parameterEnv{}, nil //nolint:exhaustruct
	}

	maxJobs := min(len(namespaces), runtime.NumCPU())
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
		for _, ns := range namespaces {
			e, err := m.evalNamespace(ctx, ns)
			if err != nil {
				return parameterEnv{}, err
			}

			env = pkg.Wrap(env, export(e))
		}
	} else {
		// Evaluate all namespaces in parallel,
		// returning a slice of fully-evaluated environments.
		for group := range slices.Chunk(namespaces, maxJobs) {
			envs, err := flowmatic.Map(ctx, len(group), group, m.evalNamespace)
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

func (m Model) evalNamespace(
	ctx context.Context,
	namespace parse.Composite,
) (parameterEnv, error) {
	match := func(filtered parse.Namespace) bool {
		return filtered.ID == namespace.ID
	}

	// This loop will iterate at maximum one time,
	// even if multiple namespaces with the given name exist.
	// The first namespace found will be used to evaluate the environment.
	for space := range pkg.Filter(m.Defs.Seq(), match) {
		// If the namespace is composed of other namespaces,
		// first add their evaluated environments to the current scope,
		// then add their mappings to the current environment.
		// This enables evaluation of nested namespaces with ancestor parameters.
		env := parameterEnv{
			eval: vars.Env[any]{},
			pars: []any{},
		}

		if space.Com.Len() > 0 {
			var err error

			env, err = m.eval(ctx, slices.Collect(space.Com.Seq())...)
			if err != nil {
				return parameterEnv{}, err
			}
		}

		// Collect the parameters from the namespace and the environment.
		pars := slices.Concat(slices.Collect(space.Par.Values()), env.pars,
			slices.Collect(namespace.Params.Values()))

		// Evaluate mappings
		for dict := range space.Sta.Seq() {
			v, err := m.evalMapping(ctx, space, dict, env.eval, pars)
			if err != nil {
				return parameterEnv{}, err
			}

			maps.Copy(env.eval, v)
		}

		return env, nil
	}

	if m.EvalRequiresDef {
		return parameterEnv{}, fmt.Errorf(
			"%w: %q is undefined",
			pkg.ErrInvalidNamespace,
			namespace,
		)
	}

	return parameterEnv{}, nil //nolint:exhaustruct
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
	program, err := expr.Compile(dict.Ex.Src, opt...)
	if err != nil {
		return nil, &pkg.ExpressionError{
			NS: space.ID, Var: dict.ID, Err: err,
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

		env[dict.ID] = pkg.Unquote(res)
	}

	return collect(env), nil
}
