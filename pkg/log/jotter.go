package log

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/ardnew/envmux/pkg"
)

//nolint:gochecknoglobals
var (
	// DefaultOutput is the default output destination for log messages.
	DefaultOutput = os.Stdout

	// DefaultOptions are the default handler options for slog handlers.
	DefaultOptions = &slog.HandlerOptions{} //nolint:exhaustruct

	// DefaultLevel is the default logging level for a [Jotter].
	DefaultLevel = slog.LevelInfo

	// DefaultHandler is the default slog handler used by [Jotter],
	// initialized with [DefaultOutput] and [DefaultOptions].
	DefaultHandler = slog.NewTextHandler(DefaultOutput, DefaultOptions)
)

// Jotter is a structured logger that wraps [slog.Leveler] and [slog.Handler].
// It provides composable options for configuring logging behavior.
type Jotter struct {
	leveler slog.Leveler
	handler slog.Handler
}

// MakeJotter constructs a new [Jotter] using the provided options.
// If no options are provided, [DefaultLevel] and [DefaultHandler] are used.
// Each option modifies the [Jotter] and can override previous settings.
func MakeJotter(opts ...pkg.Option[Jotter]) Jotter {
	if len(opts) == 0 {
		opts = append(opts, WithLeveler(DefaultLevel), WithHandler(DefaultHandler))
	}

	return pkg.Make(opts...)
}

// WithLeveler returns an option that sets the [slog.Leveler] for a [Jotter].
// If the provided leveler is already a [Jotter], chaining is prevented.
func WithLeveler(leveler slog.Leveler) pkg.Option[Jotter] {
	return func(jotter Jotter) Jotter {
		if j, ok := leveler.(Jotter); ok {
			return j // prevent self-referential chains
		}

		jotter.leveler = leveler

		return jotter
	}
}

// WithHandler returns an option that sets the [slog.Handler] for a [Jotter].
// If the provided handler is already a [Jotter], chaining is prevented.
func WithHandler(handler slog.Handler) pkg.Option[Jotter] {
	return func(jotter Jotter) Jotter {
		if j, ok := handler.(Jotter); ok {
			return j // prevent self-referential chains
		}

		jotter.handler = handler

		return jotter
	}
}

// WithText returns an option that sets a text [slog.Handler] for a [Jotter].
// The handler writes to the provided [io.Writer] with the given options.
func WithText(w io.Writer, opts *slog.HandlerOptions) pkg.Option[Jotter] {
	return WithHandler(slog.NewTextHandler(w, opts))
}

// WithJSON returns an option that sets a JSON [slog.Handler] for a [Jotter].
// The handler writes to the provided [io.Writer] with the given options.
func WithJSON(w io.Writer, opts *slog.HandlerOptions) pkg.Option[Jotter] {
	return WithHandler(slog.NewJSONHandler(w, opts))
}

// WithDiscard returns an option that sets a discard [slog.Handler] for a
// [Jotter].
// The handler discards all log output.
func WithDiscard() pkg.Option[Jotter] {
	return WithHandler(slog.DiscardHandler)
}

// Level returns the current logging level of the [Jotter].
// Implements [slog.Leveler.Level].
func (j Jotter) Level() slog.Level {
	return j.leveler.Level()
}

// Enabled reports whether logging is enabled for the given level.
// Implements [slog.Handler.Enabled].
func (j Jotter) Enabled(_ context.Context, level slog.Level) bool {
	return level >= j.leveler.Level()
}

// Handle processes a log record using the [Jotter]'s handler.
// Implements [slog.Handler.Handle].
func (j Jotter) Handle(ctx context.Context, r slog.Record) error {
	return j.handler.Handle(ctx, r)
}

// WithAttrs returns a new [slog.Handler] with the provided attributes added.
// Implements [slog.Handler.WithAttrs].
func (j Jotter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return pkg.Wrap(j, WithHandler(j.handler.WithAttrs(attrs)))
}

// WithGroup returns a new [slog.Handler] with the provided group name.
// Implements [slog.Handler.WithGroup].
func (j Jotter) WithGroup(name string) slog.Handler {
	return pkg.Wrap(j, WithHandler(j.handler.WithGroup(name)))
}
