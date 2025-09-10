package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

var exitWriter io.Writer = os.Stderr

//nolint:gochecknoinits
func init() {
	exit = func(code int) { fmt.Fprintf(exitWriter, "exit %d\n", code) }
}

// captureExit captures output sent to exitWriter during the execution of run.
func captureExit(run func()) string {
	origWriter := exitWriter
	defer func() { exitWriter = origWriter }()

	var sb strings.Builder
	exitWriter = &sb

	run()

	return sb.String()
}

func TestMain_HelpExitsZero(t *testing.T) {
	t.Helper()

	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"envmux", "--help"}

	out := captureExit(main)

	if !bytes.Contains([]byte(out), []byte("exit 0")) {
		t.Fatalf("main with --help should print 'exit 0', got: %q", out)
	}
}

func TestMain_UnknownFlagExitsNonZero(t *testing.T) {
	t.Helper()

	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"envmux", "--definitely-not-a-flag"}

	out := captureExit(main)

	if !bytes.Contains([]byte(out), []byte("exit 1")) {
		t.Fatalf("main with unknown flag should print 'exit 1', got: %q", out)
	}
}
