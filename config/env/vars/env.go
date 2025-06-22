package vars

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/ardnew/envmux/pkg"
)

// Env maps identifiers to objects for evaluating expressions.
type Env[T any] map[string]T

// ParameterKey is the identifier used in expressions to refer to the implicit
// parameter of the current expression evaluation.
var ParameterKey = `_`

// Cache returns a new copy of the current process environment.
//
// The values in the returned map contain structured data,
// as opposed to simple strings like conventional environment variables.
//
// Users can access structured map data by key (or, for structs: field name)
// as identifiers within any parsed expression.
//
// Nested data is accessed using either map or struct notation
// (e.g., `user.Name` == `user["Name"]`).
//
// The process environment is only accessible during expression evaluation.
// It is not exported to the final environment variables of a namespace.
//
// The environment content is constructed only once
// and written to a volatile, in-memory cache.
//
// The cached process environment is immutable
// and is independent of all user/namespace configurations.
func Cache() Env[any] {
	// We only want to construct the environment once because it is expensive.
	env := sync.OnceValue(func() Env[any] {
		return Env[any]{
			"target":   getTarget(),
			"platform": getPlatform(),
			"hostname": getHostname(),
			"user":     getUser(),
			"shell":    getShell(),
		}
	})

	// But since maps are reference types, we always return a clone (deep copy).
	// This prevents the caller modifying the cached singleton,
	// and it avoids the cost of re-evaluating the environment.
	return maps.Clone(env())
}

// ContextKey is the identifier used by [github.com/expr-lang/expr] internally
// to manage the evaluation [context.Context] of expressions.
//
// For example, if the given context is canceled (due to interrupt, timeout,
// etc.), the goroutine evaluating the expression will be terminated
// automatically.
const ContextKey = `ctx`

// WithContext is a functional [pkg.Option] that installs the [context.Context]
// used when evaluating expressions.
//
// The environment is lazy-loaded via [Cache] if it is uninitialized.
// If the context is nil, then [ContextKey] is removed from the environment.
func WithContext(ctx context.Context) pkg.Option[Env[any]] {
	return func(v Env[any]) Env[any] {
		if v.IsZero() {
			v = Cache() // lazy-initialize the cache
		} else if ctx == nil {
			delete(v, ContextKey)

			return v
		}

		v[ContextKey] = ctx

		return v
	}
}

func WithExports(env ...map[string]string) pkg.Option[Env[any]] {
	add := map[string]any{}

	for _, e := range env {
		for key, val := range e {
			add[key] = val
		}
	}

	return func(v Env[any]) Env[any] {
		if v.IsZero() {
			v = Cache() // lazy-initialize the cache
		}

		maps.Insert(v, v.Complement(add))

		return v
	}
}

func WithEach(seq iter.Seq2[string, any]) pkg.Option[Env[any]] {
	return func(v Env[any]) Env[any] {
		if v.IsZero() {
			v = Cache() // lazy-initialize the cache
		}

		maps.Insert(v, seq)

		return v
	}
}

// IsZero returns whether the receiver is nil or empty.
func (e Env[T]) IsZero() bool { return len(e) == 0 }

func (e Env[T]) AsMap() map[string]T { return map[string]T(e) }

// Export returns a new environment with all values converted to strings.
//
// If a format verb is provided, the first verb is used to format each value.
// The format verb is passed to [fmt.Sprintf].
func (e Env[T]) Export(verb ...string) Env[string] {
	ss := make(Env[string], len(e))
	// "Fast"-path for a pre-defined format.
	if len(verb) > 0 {
		for key, val := range e {
			ss[key] = fmt.Sprintf(verb[0], val)
		}

		return ss
	}

	for key, val := range e {
		switch v := any(val).(type) {
		case string:
			ss[key] = v
		case []byte:
			ss[key] = string(v)
		case fmt.Formatter:
			ss[key] = fmt.Sprint(v)
		case fmt.Stringer:
			ss[key] = v.String()
		case error:
			ss[key] = v.Error()
		case fmt.GoStringer:
			ss[key] = v.GoString()
		default:
			ss[key] = fmt.Sprint(v)
		}
	}

	return ss
}

// Environ returns a slice of strings for each element in the environment
// in the format "key=value".
func (e Env[T]) Environ() []string {
	ss := make([]string, 0, len(e))

	for key, val := range e.Export() {
		var sb strings.Builder

		sb.Grow(len(key) + len(val) + 1)
		sb.WriteString(key)
		sb.WriteRune('=')
		sb.WriteString(val)
		ss = append(ss, sb.String())
	}

	return ss
}

// Complement returns a sequence of all key-value pairs from the given universe
// for which the key is not already defined in the receiver environment.
//
// Unlike conventional set operations,
// if a key exists in multiple environments from the given universe,
// they will be yielded in the order they are defined.
//
// Complement is useful for merging multiple environments
// without overwriting a set of reserved keys.
func (e Env[T]) Complement(universe ...map[string]T) iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		for _, u := range universe {
			for key, val := range u {
				if _, reserved := e[key]; reserved {
					continue
				}

				if !yield(key, val) {
					return
				}
			}
		}
	}
}

func (e Env[T]) Omit(keys ...string) iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		for key, val := range e {
			if slices.Contains(keys, key) {
				continue
			}

			if !yield(key, val) {
				return
			}
		}
	}
}
