package cli

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/ardnew/envmux/pkg/log"
)

// Test the public Run API using controllable flags so Exec doesn't run.
func TestRun_HelpExitsZero(t *testing.T) {
	t.Helper()

	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"envmux", "--help"}

	code := Run(context.Background())
	if code != 0 {
		t.Fatalf("Run(--help) = %d, want 0", code)
	}
}

func TestRun_UnknownFlagExitsNonZero(t *testing.T) {
	t.Helper()

	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"envmux", "--definitely-not-a-flag"}

	code := Run(context.Background())
	if code == 0 {
		t.Fatalf("Run(unknown flag) = %d, want non-zero", code)
	}
}

// attributed implements pkg.Attributed.
type attributed struct{ msg string }

func (m attributed) Error() string      { return m.msg }
func (attributed) Attr() map[string]any { return map[string]any{"k": "v"} }
func (attributed) DetailKey() string    { return "details" }
func (attributed) Details() []string    { return []string{"line1", "line2"} }

func TestYield_WithAttributedError(t *testing.T) {
	t.Helper()

	j := log.MakeJotter() // use defaults to ensure leveler/handler set
	jot := log.Make(log.WithJotter(j))

	ctx, cancel := jot.AddToContextCancelCause(context.Background())
	defer cancel(nil)

	// Inject an attributed error wrapped in RunError.
	cancel(RunError{Err: attributed{"boom"}, Code: 7})

	if got := yield(ctx); got != 7 {
		t.Fatalf("yield(ctx) = %d, want 7", got)
	}
}

func TestYield_WithHelp(t *testing.T) {
	t.Helper()

	j := log.MakeJotter(log.WithLeveler(log.DefaultLevel), log.WithDiscard())
	jot := log.Make(log.WithJotter(j))
	ctx, cancel := jot.AddToContextCancelCause(context.Background())
	defer cancel(nil)

	cancel(RunError{Help: "usage"})

	if got := yield(ctx); got != 0 {
		t.Fatalf("yield(ctx) = %d, want 0", got)
	}
}

func TestYield_WithPlainError(t *testing.T) {
	t.Helper()

	j := log.MakeJotter(log.WithLeveler(log.DefaultLevel), log.WithDiscard())
	jot := log.Make(log.WithJotter(j))
	ctx, cancel := jot.AddToContextCancelCause(context.Background())
	defer cancel(nil)

	cancel(RunError{Err: errors.New("plain"), Code: 3})

	if got := yield(ctx); got != 3 {
		t.Fatalf("yield(ctx) = %d, want 3", got)
	}
}

func TestYield_NoRunError(t *testing.T) {
	t.Helper()

	j := log.MakeJotter(log.WithLeveler(log.DefaultLevel), log.WithDiscard())
	jot := log.Make(log.WithJotter(j))
	ctx, cancel := jot.AddToContextCancelCause(context.Background())
	defer cancel(nil)

	cancel(errors.New("not a RunError"))

	if got := yield(ctx); got != -1 {
		t.Fatalf("yield(ctx) = %d, want -1", got)
	}
}
