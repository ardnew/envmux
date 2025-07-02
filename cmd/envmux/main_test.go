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
		cc = fmt.Sprintf("%+v", cmd.Args)
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

func TestNamespaceDefinitions(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		input    string
		expectOK bool
	}{
		{
			name:     "basic_default_namespace",
			args:     []string{"-s", "-"},
			input:    "default{foo=1+2;}",
			expectOK: true,
		},
		{
			name:     "default_with_spaces",
			args:     []string{"-s", "-"},
			input:    "default { foo = 1+2; }",
			expectOK: true,
		},
		{
			name:     "default_with_empty_parens",
			args:     []string{"-s", "-"},
			input:    "default<>(){ foo = 1+2; }",
			expectOK: true,
		},
		{
			name:     "default_with_no_output_comment",
			args:     []string{"-s", "-"},
			input:    "default()<>{ foo = 1+2; } /**** NO OUTPUT ****/",
			expectOK: true,
		},
		{
			name:     "default_with_trailing_parens",
			args:     []string{"-s", "-"},
			input:    "default{ foo = 1+2; }()<>",
			expectOK: true,
		},
		{
			name:     "default_with_block_comment",
			args:     []string{"-s", "-"},
			input:    "default <> () { foo = 1+2; } /* comment */",
			expectOK: true,
		},
		{
			name:     "default_with_line_comment",
			args:     []string{"-s", "-"},
			input:    "default <> () { foo = 1+2; } // comment",
			expectOK: true,
		},
		{
			name:     "custom_namespace_with_hash_comment",
			args:     []string{"-s", "-", "custom"},
			input:    "custom { foo = 1+2; } # comment",
			expectOK: true,
		},
		{
			name:     "commented_out_default",
			args:     []string{"-s", "-"},
			input:    "// default <> () { foo = 1+2; }",
			expectOK: true,
		},
		{
			name:     "commented_out_custom",
			args:     []string{"-s", "-", "custom"},
			input:    "# custom { foo = 1+2; }",
			expectOK: true,
		},
		{
			name:     "default_with_inline_comment",
			args:     []string{"-s", "-"},
			input:    "default // <> () { foo = 1+2; }",
			expectOK: true,
		},
		{
			name:     "custom_with_inline_comment",
			args:     []string{"-s", "-", "custom"},
			input:    "custom # { foo = 1+2; }",
			expectOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, func(t *testing.T, cmd *exec.Cmd) {
				cmd.Args = append(cmd.Args, tc.args...)
				cmd.Stdin = strings.NewReader(tc.input)

				out, err := cmd.CombinedOutput()

				if tc.expectOK && err != nil {
					fail(t, cmd, "unexpected error", err, "output=", string(out))
				} else if !tc.expectOK && err == nil {
					fail(t, cmd, "expected error but got none", "output=", string(out))
				}
			})
		})
	}
}

func TestMultipleNamespaces(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		input    string
		expectOK bool
	}{
		{
			name: "default_and_custom_with_comments",
			args: []string{"-s", "-"},
			input: `default { foo /* comment */ = "abc"; }
custom { /* comment */ foo = 1+2; }`,
			expectOK: true,
		},
		{
			name: "custom_and_default_with_comments",
			args: []string{"-s", "-", "custom"},
			input: `default /* comment */ { foo = "abc"; }
/* comment */ custom { foo = 1+2; }`,
			expectOK: true,
		},
		{
			name: "default_with_custom_dependency",
			args: []string{"-s", "-"},
			input: `default <custom> { foo = /* comment */ "abc"; }
custom { foo = 1+2 /* comment */; }`,
			expectOK: true,
		},
		{
			name: "custom_with_default_dependency",
			args: []string{"-s", "-", "custom"},
			input: `default { foo = "abc"; }
custom <default> { foo = 1+2; /* comment */ }`,
			expectOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, func(t *testing.T, cmd *exec.Cmd) {
				cmd.Args = append(cmd.Args, tc.args...)
				cmd.Stdin = strings.NewReader(tc.input)

				out, err := cmd.CombinedOutput()

				if tc.expectOK && err != nil {
					fail(t, cmd, "unexpected error", err, "output=", string(out))
				} else if !tc.expectOK && err == nil {
					fail(t, cmd, "expected error but got none", "output=", string(out))
				}
			})
		})
	}
}

func TestVariableInheritance(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		input    string
		expectOK bool
	}{
		{
			name: "default_inherits_custom_single_var",
			args: []string{"-s", "-"},
			input: `default <custom> { foo = "abc"; }
custom { foo = 1+2; bar = "xyz"; }`,
			expectOK: true,
		},
		{
			name: "custom_inherits_default_single_var",
			args: []string{"-s", "-", "custom"},
			input: `default { foo = "abc"; }
custom <default> { foo = 1+2; bar = "xyz"; }`,
			expectOK: true,
		},
		{
			name: "default_inherits_custom_multiple_vars",
			args: []string{"-s", "-"},
			input: `default <custom> { foo = "abc"; bar = "xyz"; }
custom { foo = 1+2; }`,
			expectOK: true,
		},
		{
			name: "custom_inherits_default_multiple_vars",
			args: []string{"-s", "-", "custom"},
			input: `default { foo = "abc"; bar = "xyz"; }
custom <default> { foo = 1+2; }`,
			expectOK: true,
		},
		{
			name: "default_inherits_custom_mixed_types",
			args: []string{"-s", "-"},
			input: `default <custom> { foo = "abc"; bar = 2+3; }
custom { foo = 1+2; bar = "xyz"; }`,
			expectOK: true,
		},
		{
			name: "custom_inherits_default_mixed_types",
			args: []string{"-s", "-", "custom"},
			input: `default { foo = "abc"; bar = 2+3; }
custom <default> { foo = 1+2; bar = "xyz"; }`,
			expectOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, func(t *testing.T, cmd *exec.Cmd) {
				cmd.Args = append(cmd.Args, tc.args...)
				cmd.Stdin = strings.NewReader(tc.input)

				out, err := cmd.CombinedOutput()

				if tc.expectOK && err != nil {
					fail(t, cmd, "unexpected error", err, "output=", string(out))
				} else if !tc.expectOK && err == nil {
					fail(t, cmd, "expected error but got none", "output=", string(out))
				}
			})
		})
	}
}

func TestJSONOutput(t *testing.T) {
	run(t, func(t *testing.T, cmd *exec.Cmd) {
		input := "default { foo = 1+2; bar = \"test\"; }"
		cmd.Args = append(cmd.Args, "-j", "1", "-s", "-")
		cmd.Stdin = strings.NewReader(input)

		out, err := cmd.CombinedOutput()
		if err != nil {
			fail(t, cmd, "unexpected error", err, "output=", string(out))
		}

		// Basic check that output looks like JSON
		outStr := strings.TrimSpace(string(out))
		if !strings.HasPrefix(outStr, "{") || !strings.HasSuffix(outStr, "}") {
			fail(t, cmd, "expected JSON output", "got=", outStr)
		}
	})
}
