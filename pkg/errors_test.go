package pkg

import (
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      Error
		expected string
	}{
		{
			name:     "empty error",
			err:      Error{},
			expected: "<Error>",
		},
		{
			name:     "error with message",
			err:      Error{"test error"},
			expected: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestError_WithDetail(t *testing.T) {
	tests := []struct {
		name     string
		err      Error
		details  []string
		expected string
	}{
		{
			name:     "no details",
			err:      Error{"base error"},
			details:  []string{},
			expected: "base error",
		},
		{
			name:     "empty detail",
			err:      Error{"base error"},
			details:  []string{""},
			expected: "base error",
		},
		{
			name:     "single detail",
			err:      Error{"base error"},
			details:  []string{"detail"},
			expected: "base error: detail",
		},
		{
			name:     "multiple details",
			err:      Error{"base error"},
			details:  []string{"detail1", "detail2"},
			expected: "base error: detail1: detail2",
		},
		{
			name:     "mixed empty and nonempty details",
			err:      Error{"base error"},
			details:  []string{"", "detail1", "", "detail2"},
			expected: "base error: detail1: detail2",
		},
		{
			name:     "empty base error with details",
			err:      Error{""},
			details:  []string{"detail1", "detail2"},
			expected: "<Error>: detail1: detail2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.WithDetail(tt.details...)
			if result.Error() != tt.expected {
				t.Errorf("Error.WithDetail() = %v, want %v", result.Error(), tt.expected)
			}
		})
	}
}

func TestJoinErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected string
		isNil    bool
	}{
		{
			name:   "no errors",
			errors: []error{},
			isNil:  true,
		},
		{
			name:   "only nil errors",
			errors: []error{nil, nil},
			isNil:  true,
		},
		{
			name:     "single error",
			errors:   []error{Error{"error1"}},
			expected: "error1",
		},
		{
			name:     "multiple errors",
			errors:   []error{Error{"error1"}, Error{"error2"}},
			expected: "error1: error2",
		},
		{
			name:     "mixed nil and nonnil errors",
			errors:   []error{nil, Error{"error1"}, nil, Error{"error2"}},
			expected: "error1: error2",
		},
		{
			name:     "three errors",
			errors:   []error{Error{"error1"}, Error{"error2"}, Error{"error3"}},
			expected: "error1: error2: error3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinErrors(tt.errors...)
			if tt.isNil {
				if result != nil {
					t.Errorf("JoinErrors() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("JoinErrors() = nil, want %v", tt.expected)
				} else if result.Error() != tt.expected {
					t.Errorf("JoinErrors() = %v, want %v", result.Error(), tt.expected)
				}
			}
		})
	}
}

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
		{"ErrIncompleteParse", ErrIncompleteParse},
		{"ErrIncompleteEval", ErrIncompleteEval},
		{"ErrUnexpectedToken", ErrUnexpectedToken},
		{"ErrInvalidIdentifier", ErrInvalidIdentifier},
		{"ErrInvalidExpression", ErrInvalidExpression},
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

func TestMakeMarker(t *testing.T) {
	tests := []struct {
		name     string
		column   int
		expected string
	}{
		{
			name:     "zero column",
			column:   0,
			expected: "↑",
		},
		{
			name:     "negative column",
			column:   -1,
			expected: "↑",
		},
		{
			name:     "column 1",
			column:   1,
			expected: "…↑",
		},
		{
			name:     "column 3",
			column:   3,
			expected: "………↑",
		},
		{
			name:     "column 10",
			column:   10,
			expected: "…………………………↑",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeMarker(tt.column)
			if result != tt.expected {
				t.Errorf("makeMarker(%d) = %q, want %q", tt.column, result, tt.expected)
			}
		})
	}
}

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
	parseErr := MakeParseError("test source", 5)

	if parseErr.Error() != "parse error" {
		t.Errorf("ParseError.Error() = %q, want %q", parseErr.Error(), "parse error")
	}

	// Test that it implements the Attributed interface
	var _ Attributed = parseErr

	// Test that the embedded context works
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
	evalErr := MakeEvalError("testns", "testident", "test source", 5)

	if evalErr.Error() != "evaluation error" {
		t.Errorf("EvalError.Error() = %q, want %q", evalErr.Error(), "evaluation error")
	}

	// Test Attr method includes namespace and ident
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
			attr:     MakeParseError("test source", 5),
			expected: map[string]bool{"line": false, "column": false},
		},
		{
			name:     "EvalError",
			attr:     MakeEvalError("testns", "testident", "test source", 5),
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
	evalErr := MakeEvalError("testns", "testident", "test source", 5)
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

// runeCount returns the number of runes in a string.
func runeCount(s string) int { return len([]rune(s)) }
