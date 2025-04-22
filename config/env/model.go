package env

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/carlmjohnson/flowmatic"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/file"

	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
)

type Model struct {
	*parse.Namespaces `json:"namespaces"`
}

// WithNamespaces is a functional [pkg.Option] that installs the namespace
// definitions parsed from a [parse.Model].
//
// It is a required option that must be applied prior to evaluating environment
// variables with [Model.Eval].
func WithNamespaces(ns *parse.Namespaces) pkg.Option[Model] {
	return func(m Model) Model {
		m.Namespaces = ns
		// m.env = make(map[string]string)
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

func isDefined(ns *parse.Namespace) bool { return ns != nil && ns.Spec != nil }

// func (m Model) domains(ns ...string) iter.Seq[domain] {
// 	return func(yield func(domain) bool) {
// 		for keep := range m.filter(func(filtered *parse.Namespace) bool {
// 			return isDefined(filtered) && slices.Contains(ns, filtered.Name)
// 		}) {
// 			for subj := range keep.Spec.Subjects() {
// 				dom := domain{
// 					path: []string{keep.Name},
// 					scheme: scheme{
// 						subject: subj,
// 						plan:    make(map[string]string),
// 					},
// 				}
// 				if !yield(dom) {
// 					return
// 				}
// 			}
// 		}
// 	}
// }

func exprError(name string, err error) error {
	ferr := new(file.Error)
	if errors.As(err, &ferr) {
		return fmt.Errorf("%w(%s): %w", pkg.ErrInvalidExpression, name, err)
	}
	return fmt.Errorf("%w(%s): %w", pkg.ErrInvalidExpression, name, err)
}

func (m Model) eval(ctx context.Context, namespace string, subject ...string) (plan, error) {
	match := func(filtered *parse.Namespace) bool {
		return isDefined(filtered) && filtered.Name == namespace
	}
	// This loop will iterate at maximum one time.
	// If the given namespace is not found, it will return an error.
	for parsed := range pkg.Filter(slices.Values(m.List), match) {
		var err error

		eval := make(plan)
		spec := parsed.Spec
		subs := slices.Collect(spec.Subjects(subject...))

		if len(spec.Coms) > 0 {
			eval, err = m.Eval(ctx, slices.Collect(spec.Compositions()), subs...)
			if err != nil {
				return nil, err
			}
		}

		vm := MakeVarMap(ctx).AddEnv(ReplaceMode, eval).
			WithSubject("__namespace(" + parsed.Name + ")")
		if err := vm.Err(); err != nil {
			return nil, err
		}

		for _, dict := range spec.Maps {
			replace := len(dict.Prec) == 0

			// We know the complete type of the context env that will be evaluated,
			// but the value of subject (string) will vary with each iteration below.
			//
			// So we use a dummy string as the subject value during compilation,
			// which aids the compiler in determining the type of the expression.
			//
			// The actual value will be passed in evalContext when running the
			// compiled expression below:
			//
			//  • Compile(dict.Expr.Src, typeContext)
			// 			—→ typeContext contains {SubjectKey: *new(string)}
			//
			//  • Run(dict.Expr.Src, evalContext)
			//      —→ evalContext contains {SubjectKey: each.Name}
			//
			opts := []expr.Option{
				expr.Env(vm),
				expr.WithContext(ContextKey),
			}
			prog, err := expr.Compile(dict.Expr.Src, opts...)
			if err != nil {
				return nil, exprError(dict.Name, err)
			}

			if len(subs) == 0 {
				vm = vm.Del(SubjectKey)
				if err := vm.Err(); err != nil {
					return nil, err
				}
				if _, ok := vm[dict.Name]; ok && !replace {
					continue
				}
				val, err := expr.Run(prog, map[string]any(vm))
				if err != nil {
					return nil, exprError(dict.Name, err)
				}
				strval := fmt.Sprintf("%v", val)
				eval[dict.Name] = strval
				vm = vm.Add(varEditModeWithReplace(replace), dict.Name, strval)
				if err := vm.Err(); err != nil {
					return nil, err
				}
				continue
			}

			for _, each := range subs {
				vm = vm.WithSubject(each)
				if err := vm.Err(); err != nil {
					return nil, err
				}
				if _, ok := vm[dict.Name]; ok && !replace {
					continue
				}
				val, err := expr.Run(prog, map[string]any(vm))
				if err != nil {
					return nil, exprError(dict.Name, err)
				}
				strval := fmt.Sprintf("%v", val)
				eval[dict.Name] = strval
				vm = vm.Add(varEditModeWithReplace(replace), dict.Name, strval)
				if err := vm.Err(); err != nil {
					return nil, err
				}
			}
		}
		return eval, nil
	}
	return nil, pkg.ErrInvalidNamespace
}

func (m Model) Eval(ctx context.Context, namespace []string, subject ...string) (map[string]string, error) {
	if m.Namespaces == nil {
		return nil, fmt.Errorf("%w: %w", pkg.ErrInvalidModel, pkg.ErrIncompleteParse)
	}
	task := func(ctx context.Context, ns string) (plan, error) {
		return m.eval(ctx, ns, subject...)
	}
	env, err := flowmatic.Map(ctx, len(namespace), namespace, task)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", pkg.ErrIncompleteEval, err)
	}
	merged := make(map[string]string)
	for _, plan := range env {
		maps.Copy(merged, plan)
	}
	return merged, nil
}
