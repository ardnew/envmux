package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ardnew/envmux/pkg"
)

func setup() error {
	out, err := exec.Command(
		"go", "clean", "-i", "-r", "-x", "-testcache",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setup failed: %w: %s", err, out)
	}

	return nil
}

func run(t *testing.T, test func(*testing.T, *exec.Cmd), args ...string) {
	bin, err := filepath.Abs(pkg.Name)
	if err != nil {
		t.Fatalf("failed to get absolute path for binary: %v", err)
	}

	// Remove any existing binary from the working directory
	if err = setup(); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	// Call `go build` to generate a new binary
	out, err := exec.Command(
		"go", "build", "-o", bin,
	).CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed with %v: %s", err, out)
	}

	// Check if the binary was created
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		t.Fatal("envmux binary was not created")
	}

	// Clean up the binary after the test
	defer os.Remove(bin)

	// Construct a command and call the test function
	test(t, exec.Command(bin, args...))
}

func fail(t *testing.T, cmd *exec.Cmd, args ...any) {
	var sb strings.Builder
	var eq bool

	for _, arg := range args {
		// Separate each argument with ": " by default
		sep, str := ": ", fmt.Sprintf("%v", arg)

		// If previous argument ended with '=', quote and append directly
		if eq {
			sep, str = "", strconv.Quote(str)
		}
		eq = strings.HasSuffix(str, "=")

		fmt.Fprintf(&sb, "%s%s", sep, str)
	}

	cc := "failure"
	if cmd != nil {
		cc = fmt.Sprintf("%s %+v", filepath.Base(cmd.Path), cmd.Args)
	}

	t.Errorf("%v%s", cc, sb.String())
}

func die(status int, t *testing.T, cmd *exec.Cmd, args ...any) {
	fail(t, cmd, args...)
	os.Exit(status)
}

func TestMainBasic(t *testing.T) {
	// Run the command with no arguments to test help output
	run(t, func(t *testing.T, cmd *exec.Cmd) {
		out, err := cmd.CombinedOutput()
		if err != nil {
			die(1, t, cmd, err, string(out))
		}

		// Check that the output contains expected help text
		if !strings.Contains(string(out), "Usage") {
			fail(t, cmd, "expected=", "*Usage*", "got=", string(out))
		}
	})
}

func TestVersionFlag(t *testing.T) {
	run(t, func(t *testing.T, cmd *exec.Cmd) {
		out, err := cmd.CombinedOutput()
		if err != nil {
			die(1, t, cmd, err, string(out))
		}

		if !strings.Contains(string(out), "version") {
			fail(t, cmd, "expected=", "*version*", "got=", string(out))
		}
	}, "--version")
}

func TestConfigParsing(t *testing.T) {
	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "envmux-test")
	if err != nil {
		die(1, t, nil, "create temp dir", err)
	}

	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := []byte(`
verbose true
`)

	if err = os.WriteFile(configPath, configContent, 0o644); err != nil {
		die(1, t, nil, "write config file", err)
	}

	run(t, func(t *testing.T, cmd *exec.Cmd) {
		out, err := cmd.CombinedOutput()
		if err != nil {
			die(1, t, cmd, err, string(out))
		}

		if !strings.Contains(string(out), "test") {
			fail(t, cmd, "expected=", "*test*", "got=", string(out))
		}
	}, "--config", configPath)
}
