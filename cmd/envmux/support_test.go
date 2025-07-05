package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/ardnew/envmux/cmd/envmux/cli"
)

const (
	exitUnexpectedError = 1 << iota
	exitReadStdout
	exitReadStderr
	exitFailStdout
	exitFailStderr
)

type (
	comparison byte
	expected   map[comparison][]string

	failure struct {
		mesg []any
		code int
	}
	failures []failure
)

const (
	irrelevant comparison = iota
	equals
	contains
	matches
)

func (c comparison) String() string {
	prefix := ""

	if int8(c) < 0 {
		c = ^c
		prefix = "!"
	}

	switch c {
	case irrelevant:
		return ""
	case equals:
		return prefix + "equals"
	case contains:
		return prefix + "contains"
	case matches:
		return prefix + "matches"
	default:
		return fmt.Sprintf("unknown(%s%d)", prefix, c)
	}
}

func (e expected) test(s string) (res bool, exp, act string, cmp comparison) {
	for cmp, text := range e {
		ver, negate := cmp, int8(cmp) < 0

		if negate {
			cmp = ^cmp
		}

		switch {
		case len(text) == 0:
			switch cmp {
			case irrelevant:
				continue
			case equals, contains, matches:
				if negate != (s == "") {
					continue
				}
				return false, "", s, ver
			}

		case s == "":
			switch cmp {
			case irrelevant:
				continue
			case equals, contains, matches:
				if negate != (len(text) == 0) {
					continue
				}
				return false, strings.Join(text, "\n-- and --\n"), "", ver
			}
		}

		for _, t := range text {
			var r bool

			switch cmp {
			case irrelevant:
				continue
			case equals:
				r = s == t
			case contains:
				r = strings.Contains(s, t)
			case matches:
				r = regexp.MustCompile(t).MatchString(s)
			default:
				r = false
			}

			if r == negate {
				return false, t, s, ver
			}
		}
	}

	return true, "", "", irrelevant
}

// run executes the Main function with the given args and calls the test
// function
// with the captured output streams that can be read asynchronously.
func run(t *testing.T, expOut, expErr expected, args ...string) {
	// Call the Main function which returns stdout and stderr readers
	stdout, stderr := mainAsync(t.Context(), args...)

	// Ensure streams are closed after test completes
	defer stdout.Close()
	defer stderr.Close()

	outBytes, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("(%d): %s", exitReadStdout, err)
	}

	errBytes, err := io.ReadAll(stderr)
	if err != nil {
		t.Fatalf("(%d): %s", exitReadStderr, err)
	}

	sout := string(outBytes)
	serr := string(errBytes)

	logResult := func(e expected, s, w string, code int) int {
		if res, exp, act, cmp := e.test(s); !res {
			t.Logf("exp. %s (%s):", w, cmp)
			t.Logf("\n%s\n", exp)
			t.Logf("act. %s:", w)
			t.Logf("\n%s\n", act)
			t.Errorf("(%d|%08b) unexpected output on %s", code, code, w)
			return code
		}
		return 0
	}

	code := 0
	code |= logResult(expOut, sout, "stdout", exitFailStdout)
	code |= logResult(expErr, serr, "stderr", exitFailStderr)

	if code != 0 {
		t.Fatalf("failed (%d|%08b)", code, code)
	}
}

// mainAsync runs the envmux application with the provided arguments and returns
// readers for stdout and stderr that stream output asynchronously.
//
// This is a convenience function primarily used for testing.
func mainAsync(
	ctx context.Context,
	args ...string,
) (stdout, stderr io.ReadCloser) {
	// Create pipes for stdout and stderr
	rout, wout, _ := os.Pipe()
	rerr, werr, _ := os.Pipe()

	// Save original stdout, stderr, and args
	sysStdout := os.Stdout
	sysStderr := os.Stderr
	sysArgs := os.Args

	// Replace stdout and stderr with our pipes
	os.Stdout = wout
	os.Stderr = werr

	// Replace args for the duration of the function
	os.Args = append([]string{"envmux"}, args...)

	g, ctx := errgroup.WithContext(ctx)
	defer ctx.Done()

	// Run the application in a goroutine
	g.Go(func() (err error) {
		defer func() {
			errWait := g.Wait()
			if err == nil && errWait != nil {
				err = errWait
			}
		}()
		defer func() {
			// Restore original stdout, stderr, and args
			os.Stdout = sysStdout
			os.Stderr = sysStderr
			os.Args = sysArgs

			// Close writers to signal EOF to readers
			wout.Close()
			werr.Close()
		}()

		// Run the application
		result := cli.Run(ctx)

		// Write help text to stdout if present
		if result.Help != "" {
			_, err := wout.WriteString(result.Help + "\n")
			if err != nil {
				return fmt.Errorf("%w: failed to write help to stdout", err)
			}
		}

		// Write error to stderr if present
		if result.Err != nil {
			_, err := werr.WriteString("error: " + result.Err.Error() + "\n")
			if err != nil {
				return fmt.Errorf("%w: failed to write error to stderr", err)
			}
		}

		return nil
	})

	// Return readers for the test to consume asynchronously
	return rout, rerr
}
