// Package log provides logging utilities built on top of [slog] for
// structured logging.
package log

import (
	"context"
	"log/slog"
	"unique"

	"github.com/ardnew/envmux/pkg"
)

// Journal wraps [slog.Logger] to provide structured logging capabilities.
// It is configured using composable options.
type Journal struct {
	*slog.Logger
}

// Make constructs a new [Journal] using the provided options.
// If no options are provided, it uses the default created with [MakeJotter].
// Each option modifies the [Journal] and can override previous settings.
//
// Users should add a [Journal] to contexts that require consistent logging,
// e.g., contexts that share grouped attributes or log handlers, so that the
// same logger is always implicitly available via [FromContext].
// For example:
//
//	ctx, cancel := log.Make().AddToContextCancel(context.Background())
//	defer func() {
//		log.MustFromContext(ctx).Debug("exiting")
//		cancel()
//	})()
func Make(opts ...pkg.Option[Journal]) Journal {
	if len(opts) == 0 {
		opts = append(opts, WithJotter(MakeJotter()))
	}

	return unique.Make(pkg.Make(opts...)).Value()
}

// WithJotter returns an option that sets the [Jotter] as the [slog.Handler] for
// the [Journal].
// The resulting [Journal] will use the provided [Jotter] for logging.
func WithJotter(jotter Jotter) pkg.Option[Journal] {
	return func(journal Journal) Journal {
		journal.Logger = slog.New(jotter)

		return journal
	}
}

// contextKey is an unexported type used as the key for storing [Journal]
// values in a [context.Context]. This prevents collisions with other context
// keys.
type contextKey struct{}

// AddToContext returns a [context.Context] that wraps the [Journal] value.
// Use [FromContext] to retrieve the [Journal] from the context.
//
// Users should add the same [Journal] to all contexts so that a consistent
// logger is implicitly available via [FromContext] throughout the application.
func (j Journal) AddToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, j)
}

// AddToContextCancel behaves like [Journal.AddToContext] but also returns a
// [context.CancelFunc] that can be used to cancel the context.
//
// Users should add the same [Journal] to all contexts so that a consistent
// logger is implicitly available via [FromContext] throughout the application.
func (j Journal) AddToContextCancel(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	return j.AddToContext(ctx), cancel
}

// AddToContextCancelCause behaves like [Journal.AddToContext] but also returns
// a [context.CancelCauseFunc] that can be used to cancel the context with a
// causal error.
//
// Users should add the same [Journal] to all contexts so that a consistent
// logger is implicitly available via [FromContext] throughout the application.
func (j Journal) AddToContextCancelCause(
	ctx context.Context,
) (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(ctx)

	return j.AddToContext(ctx), cancel
}

// FromContext extracts a [Journal] from the provided [context.Context].
// It returns the [Journal] and a boolean indicating whether the value was
// present.
//
// A [Journal] must have been added previously to the given context using one of
// the [Journal.AddToContext] methods.
// If no [Journal] is present, the bool returned is false.
func FromContext(ctx context.Context) (Journal, bool) {
	j, ok := ctx.Value(contextKey{}).(Journal)

	return j, ok
}

// MustFromContext extracts and returns a [Journal] from the provided
// [context.Context]. It panics if no [Journal] is present.
//
// A [Journal] must have been added previously to the given context using one of
// the [Journal.AddToContext] methods.
//
//nolint:funcorder
func MustFromContext(ctx context.Context) Journal {
	j, ok := FromContext(ctx)
	if !ok {
		panic("no Journal in context")
	}

	return j
}
