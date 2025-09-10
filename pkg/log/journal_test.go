package log

import (
	"context"
	"log/slog"
	"testing"
)

// Test Make with default option path (no opts) and WithJotter behavior
func TestJournalMake_DefaultAndWithJotter(t *testing.T) {
	// default path should use MakeJotter and produce a usable logger
	j := Make()
	if j.Logger == nil {
		t.Fatalf("expected non-nil slog.Logger from Make()")
	}

	// ensure logger can log without panic by emitting a debug record
	// using a discard handler to avoid output
	// Build a custom jotter and enforce handler wiring
	jot := MakeJotter(WithDiscard(), WithLeveler(slog.LevelDebug))
	j2 := Make(WithJotter(jot))
	if j2.Logger == nil {
		t.Fatalf("expected non-nil slog.Logger from Make(WithJotter)")
	}

	// Emit a record; no error path to assert, just ensure no panic
	j2.Debug("debug message", slog.String("k", "v"))
}

func TestJournalContextHelpers(t *testing.T) {
	j := Make(WithJotter(MakeJotter(WithDiscard())))

	// AddToContext and FromContext hit
	ctx := j.AddToContext(context.Background())
	if got, ok := FromContext(ctx); !ok || got.Logger == nil {
		t.Fatalf("expected journal in context; ok=%v loggerNil=%v", ok, got.Logger == nil)
	}

	// FromContext miss path
	if _, ok := FromContext(context.Background()); ok {
		t.Fatalf("expected no journal in empty context")
	}

	// MustFromContext success
	_ = MustFromContext(ctx)

	// MustFromContext panic path
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic when journal missing from context")
			}
		}()
		_ = MustFromContext(context.Background())
	}()
}

func TestJournalAddToContextCancelVariants(t *testing.T) {
	j := Make(WithJotter(MakeJotter(WithDiscard())))

	// AddToContextCancel
	ctx, cancel := j.AddToContextCancel(context.Background())
	if _, ok := FromContext(ctx); !ok {
		t.Fatalf("expected journal in context from AddToContextCancel")
	}
	cancel()

	// AddToContextCancelCause
	ctx2, cancelCause := j.AddToContextCancelCause(context.Background())
	if _, ok := FromContext(ctx2); !ok {
		t.Fatalf("expected journal in context from AddToContextCancelCause")
	}
	cancelCause(nil)
}
