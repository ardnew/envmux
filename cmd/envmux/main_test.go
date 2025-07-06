package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMainBasic(t *testing.T) {
	stdout := expected{contains: []string{
		`COMMAND
  envmux -- namespaced environments

USAGE
  envmux [flags] [subcommand ...]`,
	}}

	stderr := expected{equals: []string{}}

	run(t, stdout, stderr, `--help`)
}

func TestVersionFlag(t *testing.T) {
	stdout := expected{contains: []string{
		`envmux version`,
	}}

	stderr := expected{equals: nil}

	run(t, stdout, stderr, `--version`)
}

func TestConfigParsing(t *testing.T) {
	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "envmux-test")
	if err != nil {
		t.Fatalf("(%d) %s: %s", exitUnexpectedError, "create temp dir", err)
	}

	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := []byte(`
verbose true
`)

	if err = os.WriteFile(configPath, configContent, 0o644); err != nil {
		t.Fatalf("(%d) %s: %s", exitUnexpectedError, "write config file", err)
	}

	stdout := expected{equals: nil}

	stderr := expected{contains: []string{
		`"<EOF>" AST
  "<EOF>" capture{}
    "<EOF>" parse.Namespaces`,
	}}

	run(t, stdout, stderr, "--config", configPath, "-i")
}

func TestNamespaceDefinitions(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		input    string
		out      expected
		err      expected
		expectOK bool
	}{
		{
			name:     "basic_default_namespace",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `default{foo=1+2;}`,
			out:      expected{contains: []string{`foo=3`}},
			expectOK: true,
		},
		{
			name:     "default_with_spaces",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `default { foo = 1+2; }`,
			out:      expected{contains: []string{`foo=3`}},
			expectOK: true,
		},
		{
			name:     "default_with_empty_parens",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `default<>(){ foo = 1+2; }`,
			out:      expected{contains: []string{`foo=3`}},
			expectOK: true,
		},
		{
			name:     "default_with_no_output_comment",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `default()<>{ foo = 1+2; } /**** NO OUTPUT ****/`,
			out:      expected{equals: []string{}},
			err:      expected{contains: []string{`unexpected token "<": composites "<…>" must be declared before parameters "(…)"`}},
			expectOK: false,
		},
		{
			name:     "default_with_trailing_parens",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `default{ foo = 1+2; }()<>`,
			out:      expected{equals: []string{}},
			err:      expected{contains: []string{`unexpected token "(": parameters "(…)" must be declared before statements "{…}"`}},
			expectOK: false,
		},
		{
			name:     "default_with_block_comment",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `default <> () { foo = 1+2; } /* comment */`,
			out:      expected{contains: []string{`foo=3`}},
			expectOK: true,
		},
		{
			name:     "default_with_line_comment",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `default <> () { foo = 1+2; } // comment`,
			out:      expected{contains: []string{`foo=3`}},
			expectOK: true,
		},
		{
			name:     "commented_out_default",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `// default <> () { foo = 1+2; }`,
			out:      expected{equals: []string{}},
			expectOK: true,
		},
		{
			name:     "commented_out_custom",
			args:     []string{"-i", "-j", "1", "-s", "-", "custom"},
			input:    `// custom { foo = 1+2; }`,
			out:      expected{equals: []string{}},
			expectOK: true,
		},
		{
			name:     "default_with_inline_comment",
			args:     []string{"-i", "-j", "1", "-s", "-"},
			input:    `default // <> () { foo = 1+2; }`,
			out:      expected{equals: []string{}},
			expectOK: true,
		},
		{
			name:     "custom_with_inline_comment",
			args:     []string{"-i", "-j", "1", "-s", "-", "custom"},
			input:    `custom // { foo = 1+2; }`,
			out:      expected{equals: []string{}},
			expectOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up stdin for the test
			origStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write the test input
			go func() {
				w.Write([]byte(tc.input))
				w.Close()
			}()

			if tc.err == nil {
				if tc.expectOK {
					tc.err = expected{equals: nil}
				} else {
					tc.err = expected{contains: []string{"error"}}
				}
			}

			run(t, tc.out, tc.err, tc.args...)

			// Restore stdin
			os.Stdin = origStdin
		})
	}
}

func TestMultipleNamespaces(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		input    string
		out      expected
		err      expected
		expectOK bool
	}{
		{
			name: "default_and_custom_with_comments",
			args: []string{"-i", "-j", "1", "-s", "-"},
			input: `default { foo /* comment */ = "abc"; }
custom { /* comment */ foo = 1+2; }`,
			out:      expected{contains: []string{`foo="abc"`}},
			expectOK: true,
		},
		{
			name: "custom_and_default_with_comments",
			args: []string{"-i", "-j", "1", "-s", "-", "custom"},
			input: `default /* comment */ { foo = "abc"; }
/* comment */ custom { foo = 1+2; }`,
			out:      expected{contains: []string{`foo=3`}},
			expectOK: true,
		},
		{
			name: "default_with_custom_dependency",
			args: []string{"-i", "-j", "1", "-s", "-"},
			input: `default <custom> { foo = /* comment */ "abc"; }
custom { foo = 1+2 /* comment */; }`,
			out:      expected{contains: []string{`foo="abc"`}},
			expectOK: true,
		},
		{
			name: "custom_with_default_dependency",
			args: []string{"-i", "-j", "1", "-s", "-", "custom"},
			input: `default { foo = "abc"; }
custom <default> { foo = 1+2; /* comment */ }`,
			out:      expected{contains: []string{`foo=3`}},
			expectOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up stdin for the test
			origStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write the test input
			go func() {
				w.Write([]byte(tc.input))
				w.Close()
			}()

			if tc.err == nil {
				if tc.expectOK {
					tc.err = expected{equals: nil}
				} else {
					tc.err = expected{contains: []string{"error"}}
				}
			}

			run(t, tc.out, tc.err, tc.args...)

			// Restore stdin
			os.Stdin = origStdin
		})
	}
}

func TestVariableInheritance(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		input    string
		out      expected
		err      expected
		expectOK bool
	}{
		{
			name: "default_inherits_custom_single_var",
			args: []string{"-i", "-j", "1", "-s", "-"},
			input: `default <custom> { foo = "abc"; }
custom { foo = 1+2; bar = "xyz"; }`,
			out:      expected{contains: []string{`foo="abc"`, `bar="xyz"`}},
			expectOK: true,
		},
		{
			name: "custom_inherits_default_single_var",
			args: []string{"-i", "-j", "1", "-s", "-", "custom"},
			input: `default { foo = "abc"; }
custom <default> { foo = 1+2; bar = "xyz"; }`,
			out:      expected{contains: []string{`foo=3`, `bar="xyz"`}},
			expectOK: true,
		},
		{
			name: "default_inherits_custom_multiple_vars",
			args: []string{"-i", "-j", "1", "-s", "-"},
			input: `default <custom> { foo = "abc"; bar = "xyz"; }
custom { foo = 1+2; }`,
			out:      expected{contains: []string{`foo="abc"`, `bar="xyz"`}},
			expectOK: true,
		},
		{
			name: "custom_inherits_default_multiple_vars",
			args: []string{"-i", "-j", "1", "-s", "-", "custom"},
			input: `default { foo = "abc"; bar = "xyz"; }
custom <default> { foo = 1+2; }`,
			out:      expected{contains: []string{`foo=3`, `bar="xyz"`}},
			expectOK: true,
		},
		{
			name: "default_inherits_custom_mixed_types",
			args: []string{"-i", "-j", "1", "-s", "-"},
			input: `default <custom> { foo = "abc"; bar = 2+3; }
custom { foo = 1+2; bar = "xyz"; }`,
			out:      expected{contains: []string{`foo="abc"`, `bar=5`}},
			expectOK: true,
		},
		{
			name: "custom_inherits_default_mixed_types",
			args: []string{"-i", "-j", "1", "-s", "-", "custom"},
			input: `default { foo = "abc"; bar = 2+3; }
custom <default> { foo = 1+2; bar = "xyz"; }`,
			out:      expected{contains: []string{`foo=3`, `bar="xyz"`}},
			expectOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up stdin for the test
			origStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write the test input
			go func() {
				w.Write([]byte(tc.input))
				w.Close()
			}()

			if tc.err == nil {
				if tc.expectOK {
					tc.err = expected{equals: nil}
				} else {
					tc.err = expected{contains: []string{"error"}}
				}
			}

			run(t, tc.out, tc.err, tc.args...)

			// Restore stdin
			os.Stdin = origStdin
		})
	}
}
