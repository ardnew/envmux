package env

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"runtime"
	"slices"
	"strconv"

	"github.com/carlmjohnson/flowmatic"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/file"

	"github.com/ardnew/envmux/config/env/vars"
	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
)

// Model is an environment model that can be used to evaluate namespaced
// environments with complex expressions from a custom file format.
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

	return pkg.Make(WithAST(ast)), nil
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
) (vars.Env[string], error) {
	s := slices.Collect(pkg.Filter(slices.Values(namespaces),
		func(ns string) bool { return ns != "" },
	))

	env, err := m.eval(ctx, s...)

	return env.eval, err
}

type subjectEnv struct {
	eval vars.Env[string]
	subs []string
}

func (m Model) eval(
	ctx context.Context, namespaces ...string,
) (subjectEnv, error) {
	if len(namespaces) == 0 {
		return subjectEnv{}, nil
	}

	if m.AST == nil {
		return subjectEnv{}, pkg.JoinErrors(
			pkg.ErrInvalidModel,
			pkg.ErrIncompleteParse,
		)
	}

	maxJobs := min(len(namespaces), runtime.NumCPU())
	if m.MaxParallelJobs <= 0 {
		m.MaxParallelJobs = maxJobs
	}

	maxJobs = min(m.MaxParallelJobs, maxJobs)

	env := subjectEnv{
		eval: vars.Env[string]{},
		subs: []string{},
	}

	// If we have only one job, run the evaluations serially in this goroutine
	// and don't fan out additional goroutines.
	if m.MaxParallelJobs == 1 {
		for _, ns := range namespaces {
			e, err := m.evalNamespace(ctx, ns)
			if err != nil {
				return subjectEnv{}, err
			}

			env = pkg.Wrap(env, export(e))
		}
	} else {
		// Evaluate all namespaces in parallel,
		// returning a slice of fully-evaluated environments.
		for group := range slices.Chunk(namespaces, maxJobs) {
			envs, err := flowmatic.Map(ctx, len(group), group, m.evalNamespace)
			if err != nil {
				return subjectEnv{}, err
			}

			env = pkg.Wrap(env, export(envs...))
		}
	}

	return env, nil
}

func export(sub ...subjectEnv) pkg.Option[subjectEnv] {
	return func(env subjectEnv) subjectEnv {
		for _, e := range sub {
			if e.eval == nil {
				continue
			}

			maps.Copy(env.eval, e.eval)
			env.subs = append(env.subs, e.subs...)
		}

		return env
	}
}

func (m Model) evalNamespace(
	ctx context.Context,
	namespace string,
) (subjectEnv, error) {
	match := func(filtered *parse.Namespace) bool {
		return isDefined(filtered) && filtered.Name == namespace
	}

	// This loop will iterate at maximum one time,
	// even if multiple namespaces with the given name exist.
	// The first namespace found will be used to evaluate the environment.
	// If the given namespace is not found, it will return an error.
	for space := range pkg.Filter(slices.Values(m.List), match) {
		// If the namespace is composed of other namespaces,
		// first add their evaluated environments to the current scope,
		// then add their mappings to the current environment.
		// This enables evaluation of nested namespaces with ancestor subjects.
		env := subjectEnv{
			eval: vars.Env[string]{},
			subs: []string{},
		}

		if len(space.Com) > 0 {
			var err error

			env, err = m.eval(ctx, slices.Collect(space.Compositions())...)
			if err != nil {
				return subjectEnv{}, err
			}
		}

		// Evaluate mappings
		subs := slices.Collect(space.Parameters())
		subs = append(subs, env.subs...)

		for _, dict := range space.Sta {
			v, err := m.evalMapping(ctx, dict, env.eval, subs)
			if err != nil {
				return subjectEnv{}, err
			}

			maps.Copy(env.eval, v)
		}

		return env, nil
	}

	if m.EvalRequiresDef {
		return subjectEnv{}, fmt.Errorf(
			"%w: %q is undefined",
			pkg.ErrInvalidNamespace,
			namespace,
		)
	}

	return subjectEnv{}, nil
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

// evalMapping evaluates a single mapping across all applicable subjects.
func (m Model) evalMapping(
	ctx context.Context,
	dict *parse.Statement,
	eval vars.Env[string],
	subs []string,
) (vars.Env[string], error) {
	env := pkg.Make(vars.WithContext(ctx), vars.WithExports(eval))

	env[vars.ParameterKey] = "" // placeholder for compiler

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
		return nil, exprError(dict.ID, err)
	}

	// Handle case with no subjects
	if len(subs) == 0 {
		subs = []string{""}

		delete(env, vars.ParameterKey)
	}

	// Process each subject
	for _, subj := range subs {
		if subj != "" {
			str := subj
			if val, err := strconv.Unquote(str); err == nil {
				str = val
			}

			env[vars.ParameterKey] = str
		}

		res, err := expr.Run(program, env.AsMap())
		if err != nil {
			return nil, err
		}

		str := fmt.Sprint(res)
		if val, err := strconv.Unquote(str); err == nil {
			str = val
		}

		env[dict.ID] = str
	}

	return collect(env).Export(), nil
}

func isDefined(ns *parse.Namespace) bool { return ns != nil }

func exprError(name string, err error) error {
	ferr := new(file.Error)
	if errors.As(err, &ferr) {
		err = ferr
	}

	return fmt.Errorf("%w(%s): %w", pkg.ErrInvalidExpression, name, err)
}
