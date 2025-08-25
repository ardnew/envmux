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
func (j Journal) AddToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, j)
}

// FromContext extracts a [Journal] from the provided [context.Context].
// It returns the [Journal] and a boolean indicating whether the value was
// present.
func FromContext(ctx context.Context) (Journal, bool) {
	j, ok := ctx.Value(contextKey{}).(Journal)

	return j, ok
}
