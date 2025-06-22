package env

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/carlmjohnson/flowmatic"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/file"

	"github.com/ardnew/envmux/config/env/vars"
	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
)

type Model struct {
	*parse.AST `json:"namespaces"`
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
	env, err := m.eval(ctx, namespaces...)

	return env.eval, err
}

type subjectEnv struct {
	eval vars.Env[string]
	subs []string
}

func (m Model) eval(
	ctx context.Context, namespaces ...string,
) (subjectEnv, error) {
	if m.AST == nil {
		return subjectEnv{}, pkg.JoinErrors(
			pkg.ErrInvalidModel,
			pkg.ErrIncompleteParse,
		)
	}

	// Evaluate all namespaces in parallel,
	// returning a slice of fully-evaluated environments.
	envs, err := flowmatic.Map(
		ctx, len(namespaces), namespaces, m.evalNamespace,
	)
	if err != nil {
		return subjectEnv{}, err
	}

	env := subjectEnv{
		eval: vars.Env[string]{},
		subs: make([]string, 0, len(envs)),
	}

	for _, ns := range envs {
		if ns.eval == nil {
			continue
		}

		maps.Copy(env.eval, ns.eval)
		env.subs = append(env.subs, ns.subs...)
	}

	return env, nil
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
	for parsed := range pkg.Filter(slices.Values(m.List), match) {
		spec := parsed.Spec

		// If the namespace is composed of other namespaces,
		// first add their evaluated environments to the current scope,
		// then add their mappings to the current environment.
		// This enables evaluation of nested namespaces with ancestor subjects.
		env := subjectEnv{
			eval: vars.Env[string]{},
			subs: []string{},
		}

		if len(spec.Com) > 0 {
			var err error

			env, err = m.eval(ctx, slices.Collect(spec.Compositions())...)
			if err != nil {
				return subjectEnv{}, err
			}
		}

		// Evaluate mappings
		subs := slices.Collect(spec.Parameters())
		subs = append(subs, env.subs...)

		for _, dict := range spec.Sta {
			v, err := m.evalMapping(ctx, dict, env.eval, subs)
			if err != nil {
				return subjectEnv{}, err
			}

			maps.Copy(env.eval, v)
		}

		return env, nil
	}

	return subjectEnv{}, pkg.ErrInvalidNamespace
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

func isDefined(ns *parse.Namespace) bool { return ns != nil && ns.Spec != nil }

func exprError(name string, err error) error {
	ferr := new(file.Error)
	if errors.As(err, &ferr) {
		err = ferr
	}

	return fmt.Errorf("%w(%s): %w", pkg.ErrInvalidExpression, name, err)
}
