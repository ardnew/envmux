// Package vars provides definitions for environment variables derived from
// the process environment and other system information.
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

				// Functions
				"cwd": cwd,
				"file": map[string]any{
					"exists":    fileExists,
					"isDir":     fileIsDir,
					"isRegular": fileIsRegular,
					"isSymlink": fileIsSymlink,
					"perms":     filePerm,
					"stat":      fileStat,
				},
				"path": map[string]any{
					"abs": pathAbs,
					"cat": pathCat,
					"rel": pathRel,
				},
				"mung": map[string]any{
					"prefix":   mungPrefix,
					"prefixif": mungPrefixIf,
				},
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

func WithExports(env ...map[string]any) pkg.Option[Env[any]] {
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

func Export(key string, value any) string {
	var sb strings.Builder

	sb.WriteString(strings.TrimSpace(key))
	sb.WriteRune('=')
	sb.WriteString(format(value))

	return sb.String()
}

func formatSlice[T any](
	slice []T,
	lhs, rhs, delim string,
	format func(T) string,
) string {
	var sb strings.Builder

	sb.WriteString(lhs)

	for i, item := range slice {
		if i > 0 {
			sb.WriteString(delim)
		}

		sb.WriteString(format(item))
	}

	sb.WriteString(rhs)

	return sb.String()
}

func format(value any) string {
	switch v := any(value).(type) {
	case nil:
		return `""`

	case fmt.Formatter:
		return fmt.Sprint(v)

	case fmt.Stringer:
		return v.String()

	case fmt.GoStringer:
		return v.GoString()

	case error:
		return v.Error()

	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		complex64, complex128:
		return fmt.Sprint(v)

	case string:
		return strconv.Quote(v)

	case []byte:
		if str, ok := pkg.ParseASCII(v); ok {
			return strconv.Quote(str)
		} else {
			return fmt.Sprintf("%+q", v)
		}

	case []bool, []int, []int8, []int16, []int32, []int64,
		[]uint, []uint16, []uint32, []uint64,
		[]float32, []float64,
		[]complex64, []complex128:
		return formatSlice(v.([]any), "[ ", " ]", ", ", format)

	case []string:
		return formatSlice(v, "[ ", " ]", ", ", strconv.Quote)

	default:
		return fmt.Sprintf("%+v", v)
	}
}

// IsZero returns whether the receiver is nil or empty.
func (e Env[T]) IsZero() bool { return len(e) == 0 }

func (e Env[T]) AsMap() map[string]T { return map[string]T(e) }

// Environ returns a slice of strings for each element in the environment
// in the format "key=value".
func (e Env[T]) Environ() []string {
	ss := make([]string, 0, len(e))

	for key, val := range e {
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
