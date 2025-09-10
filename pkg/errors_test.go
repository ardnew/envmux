package pkg

import (
	"errors"
	"log/slog"
	"reflect"
	"testing"
	"unicode/utf8"
)

// unwrap helpers for attributed error types stored inside Error chains
func unwrapParse(e Error) ParseError { return e.Unwrap()[0].(ParseError) }
func unwrapEval(e Error) EvalError   { return e.Unwrap()[0].(EvalError) }

// helper standard library errors to wrap
var (
	errStdA = errors.New("std A")
	errStdB = errors.New("std B")
)

func TestMakeErrorAndErrorString(t *testing.T) {
	if MakeError().Error() != "" { // empty chain -> empty string
		t.Fatalf("expected empty string for empty error chain")
	}

	e := MakeError("alpha", "", "beta") // skips empty
	if got, want := e.Error(), "alpha: beta"; got != want {
		t.Fatalf("unexpected error string: got %q want %q", got, want)
	}
}

func TestWithErrorAndWithErrorMessage(t *testing.T) {
	base := Make(WithErrorMessage("one"))
	base = base.WrapMessage("two", "")         // ignore empty
	base = base.Wrap(errors.New("three"), nil) // ignore nil
	base = base.WrapMessage()                  // no-op
	if got, want := base.Error(), "one: two: three"; got != want {
		t.Fatalf("Wrap / WrapMessage mismatch: got %q want %q", got, want)
	}
}

func TestUnwrap(t *testing.T) {
	e := MakeError("x", "y")
	u := e.Unwrap()
	if len(u) != 2 || u[0].Error() != "x" || u[1].Error() != "y" {
		t.Fatalf("unexpected unwrap contents: %#v", u)
	}
}

func TestAttributesExcludesDetail(t *testing.T) {
	pe := unwrapParse(MakeParseError("line1", 0))
	attrs := Attributes(pe)
	foundDetail := false
	for _, a := range attrs {
		if a.Key == pe.DetailKey() { // should not appear
			foundDetail = true
		}
	}
	if foundDetail {
		t.Fatalf("detail key should be excluded from slog attrs")
	}

	// ensure returned attrs are usable with slog (non-zero length value string ok)
	for _, a := range attrs {
		_ = slog.Any(a.Key, a.Value.Any())
	}
}

func TestManifestErrorContextExtractionAndMarker(t *testing.T) {
	source := "first\nsecond\nthird"           // offsets: second line starts at 6
	ctx := makeManifestErrorContext(source, 8) // points into second line
	if ctx.Line != 1 || ctx.Column != 2 || ctx.Source != "second" {
		t.Fatalf("unexpected context: %+v", ctx)
	}
	if leaders := runeCount(ctx.Marker) - 1; leaders != ctx.Column { // column is zero based
		t.Fatalf("marker alignment mismatch: leaders=%d col=%d marker=%q", leaders, ctx.Column, ctx.Marker)
	}
	// offset past end
	ctx2 := makeManifestErrorContext("abc", 10)
	if ctx2.Source != "abc" || ctx2.Line != 0 || ctx2.Column != 3 {
		t.Fatalf("unexpected past-end context: %+v", ctx2)
	}
}

func TestManifestErrorContextEmptyLine(t *testing.T) {
	source := "line1\n\nline3"
	ctx := makeManifestErrorContext(source, 6) // points at empty second line
	if ctx.Line != 1 || ctx.Column != 0 || ctx.Source != "" {
		t.Fatalf("unexpected empty line context: %+v", ctx)
	}
}

func TestMakeMarker(t *testing.T) {
	cols := []int{-5, 0, 1, 3, 7}
	for _, col := range cols {
		m := makeMarker(col)
		r, _ := utf8.DecodeLastRuneInString(m)
		if r != '↑' {
			t.Fatalf("last rune must be arrow col=%d m=%q", col, m)
		}
		if col < 0 {
			if m != "↑" {
				t.Fatalf("negative col should yield base marker got=%q", m)
			}
			continue
		}
		if leaders := runeCount(m) - 1; leaders != col {
			t.Fatalf("leaders=%d want=%d marker=%q", leaders, col, m)
		}
	}
}

func TestParseErrorImplementsAttributed(t *testing.T) {
	perr := unwrapParse(MakeParseError("foo", 0))
	var a Attributed = perr
	if a.DetailKey() != "detail" {
		t.Fatalf("unexpected detail key")
	}
	if len(a.Details()) != 2 {
		t.Fatalf("expected 2 detail lines")
	}
	m := a.Attr()
	for _, key := range []string{"detail", "line", "column"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing attr key %q", key)
		}
	}
	if perr.Error() != "failed to parse manifest" {
		t.Fatalf("unexpected parse error string: %q", perr.Error())
	}
}

func TestEvalErrorAttributes(t *testing.T) {
	eerrWrapped := MakeEvalError("ns", "id", "foo\nbar", 5) // into second line
	eerr := unwrapEval(eerrWrapped)
	attrs := eerr.Attr()
	wantKeys := []string{"detail", "line", "column", "namespace", "ident"}
	for _, k := range wantKeys {
		if _, ok := attrs[k]; !ok {
			t.Fatalf("missing attr %q", k)
		}
	}
	if eerrWrapped.Error() != "failed to evaluate expression" {
		t.Fatalf("unexpected eval error string: %q", eerrWrapped.Error())
	}
	// ensure underlying context used
	if eerr.Line != 1 {
		t.Fatalf("expected line=1 got %d", eerr.Line)
	}
}

func TestPredeclaredSentinelErrors(t *testing.T) {
	// Verify non-empty messages
	set := map[string]string{}
	pairs := []struct {
		name string
		e    Error
	}{
		{"ErrUndefCommandExec", ErrUndefCommandExec},
		{"ErrUndefCommandFlagSet", ErrUndefCommandFlagSet},
		{"ErrUndefCommandUsage", ErrUndefCommandUsage},
		{"ErrInaccessibleManifest", ErrInaccessibleManifest},
		{"ErrUndefinedNamespace", ErrUndefinedNamespace},
		{"ErrInvalidIdentifier", ErrInvalidIdentifier},
		{"ErrInvalidJSON", ErrInvalidJSON},
	}
	for _, p := range pairs {
		if p.e.Error() == "" {
			t.Fatalf("%s has empty message", p.name)
		}
		set[p.name] = p.e.Error()
	}
	if len(set) != len(pairs) {
		t.Fatalf("expected %d sentinel messages, got %d", len(pairs), len(set))
	}
}

func TestWrapChainingOrder(t *testing.T) {
	chain := MakeError("a").WrapMessage("b").Wrap(errStdA).WrapMessage("c").Wrap(errStdB)
	if got, want := chain.Error(), "a: b: std A: c: std B"; got != want {
		t.Fatalf("unexpected chain order got=%q want=%q", got, want)
	}
	// verify unwrap preserves chronological order
	u := chain.Unwrap()
	gotSeq := make([]string, 0, len(u))
	for _, e := range u {
		gotSeq = append(gotSeq, e.Error())
	}
	wantSeq := []string{"a", "b", "std A", "c", "std B"}
	if !reflect.DeepEqual(gotSeq, wantSeq) {
		t.Fatalf("unwrap order mismatch got=%v want=%v", gotSeq, wantSeq)
	}
}

func TestAttributesHelperDeterministicKeys(t *testing.T) {
	eerr := unwrapEval(MakeEvalError("n", "x", "abc", 1))
	attrs := Attributes(eerr)
	// convert slice to map for quick presence check
	m := map[string]bool{}
	for _, a := range attrs {
		m[a.Key] = true
	}
	for _, want := range []string{"namespace", "ident", "line", "column"} {
		if !m[want] {
			t.Fatalf("missing slog attr %q", want)
		}
	}
}

// runeCount returns the number of runes in a string (copied from previous tests for marker alignment validation).
func runeCount(s string) int { return len([]rune(s)) }

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrUndefCommandExec", ErrUndefCommandExec},
		{"ErrUndefCommandFlagSet", ErrUndefCommandFlagSet},
		{"ErrUndefCommandUsage", ErrUndefCommandUsage},
		{"ErrInaccessibleManifest", ErrInaccessibleManifest},
		{"ErrUndefinedNamespace", ErrUndefinedNamespace},
		{"ErrInvalidIdentifier", ErrInvalidIdentifier},
		{"ErrInvalidJSON", ErrInvalidJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s should have a non-empty error message", tt.name)
			}
		})
	}
}

func TestMakeManifestErrorContext(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		offset         int
		expectedSource string
		expectedLine   int
		expectedColumn int
	}{
		{
			name:           "empty source",
			source:         "",
			offset:         0,
			expectedSource: "",
			expectedLine:   0,
			expectedColumn: 0,
		},
		{
			name:           "single line at beginning",
			source:         "hello world",
			offset:         0,
			expectedSource: "hello world",
			expectedLine:   0,
			expectedColumn: 0,
		},
		{
			name:           "single line at middle",
			source:         "hello world",
			offset:         5,
			expectedSource: "hello world",
			expectedLine:   0,
			expectedColumn: 5,
		},
		{
			name:           "multiline on first line",
			source:         "line1\nline2\nline3",
			offset:         3,
			expectedSource: "line1",
			expectedLine:   0,
			expectedColumn: 3,
		},
		{
			name:           "multiline on second line",
			source:         "line1\nline2\nline3",
			offset:         8,
			expectedSource: "line2",
			expectedLine:   1,
			expectedColumn: 2,
		},
		{
			name:           "multiline on third line",
			source:         "line1\nline2\nline3",
			offset:         13,
			expectedSource: "line3",
			expectedLine:   2,
			expectedColumn: 1,
		},
		{
			name:           "multiline with newline character",
			source:         "line1\nline2",
			offset:         5,
			expectedSource: "line1",
			expectedLine:   0,
			expectedColumn: 5,
		},
		{
			name:           "offset beyond source",
			source:         "hello",
			offset:         10,
			expectedSource: "hello",
			expectedLine:   0,
			expectedColumn: 5,
		},
		{
			name:           "empty line in middle",
			source:         "line1\n\nline3",
			offset:         6,
			expectedSource: "",
			expectedLine:   1,
			expectedColumn: 0,
		},
		{
			name:           "end of source",
			source:         "hello",
			offset:         5,
			expectedSource: "hello",
			expectedLine:   0,
			expectedColumn: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := makeManifestErrorContext(tt.source, tt.offset)

			if ctx.Source != tt.expectedSource {
				t.Errorf("Source = %q, want %q", ctx.Source, tt.expectedSource)
			}
			if ctx.Line != tt.expectedLine {
				t.Errorf("Line = %d, want %d", ctx.Line, tt.expectedLine)
			}
			if ctx.Column != tt.expectedColumn {
				t.Errorf("Column = %d, want %d", ctx.Column, tt.expectedColumn)
			}

			// Test that marker is created properly
			expectedMarker := makeMarker(tt.expectedColumn)
			if ctx.Marker != expectedMarker {
				t.Errorf("Marker = %q, want %q", ctx.Marker, expectedMarker)
			}
		})
	}
}

// (Removed duplicate TestMakeMarker variant)

func TestManifestErrorContext_Attr(t *testing.T) {
	ctx := makeManifestErrorContext("test source", 4)
	attr := ctx.Attr()

	expectedKeys := []string{"detail", "line", "column"}
	for _, key := range expectedKeys {
		if _, exists := attr[key]; !exists {
			t.Errorf("Attr() missing key %q", key)
		}
	}

	if detail, ok := attr["detail"].(map[string]any); ok {
		if source, exists := detail["source"]; !exists || source != "test source" {
			t.Errorf("Attr()[detail][source] = %v, want %q", source, "test source")
		}
		if _, exists := detail["marker"]; !exists {
			t.Errorf("Attr()[detail][marker] should exist")
		}
	} else {
		t.Errorf("Attr()[detail] should be map[string]any")
	}

	if line, exists := attr["line"]; !exists || line != 1 {
		t.Errorf("Attr()[line] = %v, want %d", line, 1)
	}
	if column, exists := attr["column"]; !exists || column != 5 {
		t.Errorf("Attr()[column] = %v, want %d", column, 5)
	}
}

func TestManifestErrorContext_DetailKey(t *testing.T) {
	ctx := makeManifestErrorContext("test", 0)
	if key := ctx.DetailKey(); key != "detail" {
		t.Errorf("DetailKey() = %q, want %q", key, "detail")
	}
}

func TestManifestErrorContext_Details(t *testing.T) {
	tests := []struct {
		name   string
		source string
		offset int
	}{
		{
			name:   "short source",
			source: "hello",
			offset: 2,
		},
		{
			name:   "longer source",
			source: "this is a longer source line",
			offset: 10,
		},
		{
			name:   "empty source",
			source: "",
			offset: 0,
		},
		{
			name:   "unicode characters",
			source: "测试unicode字符",
			offset: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := makeManifestErrorContext(tt.source, tt.offset)
			details := ctx.Details()

			// Should have exactly 2 lines (source, marker)
			if len(details) != 2 {
				t.Errorf("Details() returned %d lines, want 2", len(details))
			}

			// First line should equal the extracted source, second should equal marker
			if details[0] != ctx.Source {
				t.Errorf("Details()[0] = %q, want %q", details[0], ctx.Source)
			}
			if details[1] != ctx.Marker {
				t.Errorf("Details()[1] = %q, want %q", details[1], ctx.Marker)
			}

			// Verify the marker alignment width (leaders count) equals the column
			if leaders := runeCount(details[1]) - 1; leaders != ctx.Column {
				t.Errorf("marker leaders = %d, want %d", leaders, ctx.Column)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	w := MakeParseError("test source", 5)
	if w.Error() != "failed to parse manifest" {
		t.Errorf("ParseError.Error() = %q, want %q", w.Error(), "failed to parse manifest")
	}
	parseErr := unwrapParse(w)
	var _ Attributed = parseErr
	attr := parseErr.Attr()
	if _, exists := attr["detail"]; !exists {
		t.Errorf("ParseError should have detail attribute")
	}
	if _, exists := attr["line"]; !exists {
		t.Errorf("ParseError should have line attribute")
	}
	if _, exists := attr["column"]; !exists {
		t.Errorf("ParseError should have column attribute")
	}

	details := parseErr.Details()
	if len(details) != 2 {
		t.Errorf("ParseError.Details() should return 2 lines, got %d", len(details))
	}

	if key := parseErr.DetailKey(); key != "detail" {
		t.Errorf("ParseError.DetailKey() = %q, want %q", key, "detail")
	}
}

func TestEvalError(t *testing.T) {
	w := MakeEvalError("testns", "testident", "test source", 5)
	if w.Error() != "failed to evaluate expression" {
		t.Errorf("EvalError.Error() = %q, want %q", w.Error(), "failed to evaluate expression")
	}
	evalErr := unwrapEval(w)
	attr := evalErr.Attr()
	if namespace, exists := attr["namespace"]; !exists || namespace != "testns" {
		t.Errorf("EvalError.Attr()[namespace] = %v, want %q", namespace, "testns")
	}
	if ident, exists := attr["ident"]; !exists || ident != "testident" {
		t.Errorf("EvalError.Attr()[ident] = %v, want %q", ident, "testident")
	}

	// Test that it implements the Attributed interface
	var _ Attributed = evalErr

	// Test that the embedded context works
	if _, exists := attr["detail"]; !exists {
		t.Errorf("EvalError should have detail attribute")
	}
	if _, exists := attr["line"]; !exists {
		t.Errorf("EvalError should have line attribute")
	}
	if _, exists := attr["column"]; !exists {
		t.Errorf("EvalError should have column attribute")
	}

	details := evalErr.Details()
	if len(details) != 2 {
		t.Errorf("EvalError.Details() should return 2 lines, got %d", len(details))
	}

	if key := evalErr.DetailKey(); key != "detail" {
		t.Errorf("EvalError.DetailKey() = %q, want %q", key, "detail")
	}
}

func TestAttributes(t *testing.T) {
	tests := []struct {
		name     string
		attr     Attributed
		expected map[string]bool
	}{
		{
			name:     "ParseError",
			attr:     unwrapParse(MakeParseError("test source", 5)),
			expected: map[string]bool{"line": false, "column": false},
		},
		{
			name:     "EvalError",
			attr:     unwrapEval(MakeEvalError("testns", "testident", "test source", 5)),
			expected: map[string]bool{"namespace": false, "ident": false, "line": false, "column": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := Attributes(tt.attr)

			// Should exclude the detail key
			for _, attr := range attrs {
				if attr.Key == "detail" {
					t.Errorf("Attributes() should not include detail key")
				}
			}

			// Should include expected keys
			expectedKeys := tt.expected
			for _, attr := range attrs {
				if _, exists := expectedKeys[attr.Key]; exists {
					expectedKeys[attr.Key] = true
				}
			}

			for key, found := range expectedKeys {
				if !found {
					t.Errorf("Attributes() missing expected key %q", key)
				}
			}
		})
	}
}

func TestAttributesSlogAttrCreation(t *testing.T) {
	evalErr := unwrapEval(MakeEvalError("testns", "testident", "test source", 5))
	attrs := Attributes(evalErr)

	// Verify that each attribute is a valid slog.Attr
	for _, attr := range attrs {
		if attr.Key == "" {
			t.Errorf("Attribute key should not be empty")
		}
		// Verify the value is properly set (non-nil)
		if attr.Value.String() == "" && attr.Key != "column" && attr.Key != "line" {
			t.Errorf("Attribute %q value should not be empty", attr.Key)
		}
	}
}

// (removed duplicate runeCount)
