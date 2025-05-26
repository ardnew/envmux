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
		return fmt.Errorf("%w (Model): %w", pkg.ErrInvalidJSON, err).Error()
	}
	return string(e)
}

func (m Model) IsZero() bool { return m.AST == nil }

func (m Model) Eval(
	ctx context.Context, namespaces ...string,
) (vars.Env[string], error) {
	if m.AST == nil {
		return nil, fmt.Errorf(
			"%w: %w", pkg.ErrInvalidModel, pkg.ErrIncompleteParse,
		)
	}

	// Evaluate all namespaces in parallel,
	// returning a slice of fully-evaluated environments.
	envs, err := flowmatic.Map(
		ctx, len(namespaces), namespaces, m.evalNamespace,
	)
	if err != nil {
		return nil, err
	}

	env := vars.Env[string]{}
	for _, ns := range envs {
		if ns == nil {
			continue
		}
		maps.Copy(env, ns)
	}
	return env, nil
}

func (m Model) evalNamespace(ctx context.Context, namespace string) (vars.Env[string], error) {
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
		eval := vars.Env[string]{}
		if len(spec.Coms) > 0 {
			var err error
			eval, err = m.Eval(ctx, slices.Collect(spec.Compositions())...)
			if err != nil {
				return nil, err
			}
		}

		// Evaluate mappings
		subs := slices.Collect(spec.Subjects())
		for _, dict := range spec.Maps {
			v, err := m.evalMapping(ctx, dict, eval, subs)
			if err != nil {
				return nil, err
			}
			maps.Copy(eval, v)
		}
		return eval, nil
	}
	return nil, pkg.ErrInvalidNamespace
}

func collect(e vars.Env[any]) vars.Env[any] {
	return maps.Collect(
		pkg.FilterKeys(vars.Cache().Complement(e),
			func(key string) bool {
				return key != vars.ContextKey && key != vars.SubjectKey
			},
		),
	)
}

// evalMapping evaluates a single mapping across all applicable subjects
func (m Model) evalMapping(
	ctx context.Context, dict *parse.Mapping, eval vars.Env[string], subs []string,
) (vars.Env[string], error) {
	env := pkg.Make(vars.WithContext(ctx), vars.WithExports(eval))

	currVal, defined := env[dict.Name]
	replace := len(dict.Prec) == 0

	if defined && !replace {
		return collect(env).Export(), nil
	}

	// Check for immediate assignment (unevaluated expression)
	if dict.Op[0] == ':' {
		if !defined || replace {
			env[dict.Name] = dict.Expr.Src
		}
		return collect(env).Export(), nil
	}

	env[vars.SubjectKey] = "" // placeholder for compiler

	// We have to pass the environment to both [expr.Compile] and [expr.Run].
	// The former builds type information for validating the latter.
	opt := []expr.Option{
		expr.Env(env.AsMap()),
		expr.WithContext(vars.ContextKey),
		expr.AllowUndefinedVariables(),
	}

	// Compile expression
	program, err := expr.Compile(dict.Expr.Src, opt...)
	if err != nil {
		return nil, exprError(dict.Name, err)
	}

	// Handle case with no subjects
	if len(subs) == 0 {
		subs = []string{""}
		delete(env, vars.SubjectKey)
	}

	// Process each subject
	for _, subj := range subs {
		if subj != "" {
			str := subj
			if val, err := strconv.Unquote(str); err == nil {
				str = val
			}
			env[vars.SubjectKey] = str
		}

		res, err := expr.Run(program, env.AsMap())
		if err != nil {
			return nil, err
		}

		str := fmt.Sprint(res)
		if val, err := strconv.Unquote(str); err == nil {
			str = val
		}

		if !defined || dict.Op[0] == '=' {
			currVal = str
		} else {
			switch dict.Op[0] {
			case '^':
				currVal = fmt.Sprintf("%s%s", str, currVal)
			case '+':
				currVal = fmt.Sprintf("%s%s", currVal, str)
			}
		}
		env[dict.Name] = currVal
		defined = true
	}

	return collect(env).Export(), nil
}

func isDefined(ns *parse.Namespace) bool { return ns != nil && ns.Spec != nil }

func exprError(name string, err error) error {
	ferr := new(file.Error)
	if errors.As(err, &ferr) {
		return fmt.Errorf("%w(%s): %w", pkg.ErrInvalidExpression, name, err)
	}
	return fmt.Errorf("%w(%s): %w", pkg.ErrInvalidExpression, name, err)
}
