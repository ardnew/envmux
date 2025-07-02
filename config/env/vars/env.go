package vars

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/ardnew/envmux/pkg"
)

// Env maps identifiers to objects for evaluating expressions.
type Env[T any] map[string]T

// ParameterKey is the identifier used in expressions to refer to the implicit
// parameter of the current expression evaluation.
var ParameterKey = `_` //nolint:gochecknoglobals

// Cache returns the process environment cache.
//
// Returns a copy to prevent modification of the singleton.
func Cache() Env[any] {
	// Use sync.Once with a private variable to store the singleton
	// instead of recreating the map on each call
	return maps.Clone(envCache())
}

// Private singleton cache.
//
//nolint:gochecknoglobals
var (
	envCacheOnce sync.Once
	envCacheVal  Env[any]
	envCache     = func() Env[any] {
		envCacheOnce.Do(func() {
			envCacheVal = Env[any]{
				"target":   getTarget(),
				"platform": getPlatform(),
				"hostname": getHostname(),
				"user":     getUser(),
				"shell":    getShell(),
			}
		})

		return envCacheVal
	}
)

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

func Export(keyval ...string) string {
	if len(keyval) > 0 {
		keyval[0] = strings.TrimSpace(keyval[0])
	}

	//nolint:gomnd,mnd
	switch len(keyval) {
	case 0:
		return ""
	case 1:
		return keyval[0] + "="
	case 2:
		return keyval[0] + "=" + strconv.Quote(keyval[1])
	default:
		elem := pkg.Map(slices.Values(keyval[1:]), strconv.Quote)

		return keyval[0] + "=( " + strings.Join(slices.Collect(elem), " ") + " )"
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
		ss = append(ss, Export(key, val))
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
// See [Env.Omit] for a similar operation that yields key-value pairs from the
// receiver [Env] instead of the given arguments, which results in a sequence
// of key-value pairs with unique keys.
func (e Env[T]) Complement(universe ...Env[T]) iter.Seq2[string, T] {
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

// Omit returns a sequence of all key-value pairs from the receiver environment
// for which the key is not in the given list.
//
// See [Env.Complement] for a similar operation that yields key-value pairs
// from the given arguments instead of the receiver [Env], which results in a
// sequence of key-value pairs that allows for duplicate keys.
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
