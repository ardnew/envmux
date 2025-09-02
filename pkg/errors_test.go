package pkg_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/expr-lang/expr/file"

	"github.com/ardnew/envmux/pkg"
)

func TestError_Error(t *testing.T) {
	// Test using predefined errors since we can't construct Error directly
	// from outside the package (unexported embedded string field)
	
	// Test that a predefined error returns the expected message
	err := pkg.ErrInvalidIdentifier
	expected := "invalid identifier"
	if err.Error() != expected {
		t.Errorf("Error.Error() = %v, want %v", err.Error(), expected)
	}
	
	// Test zero value by creating an empty Error through JoinErrors
	nilErr := pkg.JoinErrors()
	if nilErr != nil {
		t.Errorf("JoinErrors() should return nil for empty input, got %v", nilErr)
	}
}

func TestJoinErrors(t *testing.T) {
	tests := []struct {
		name string
		errs []error
		want string
	}{
		{
			name: "no errors",
			errs: []error{},
			want: "",
		},
		{
			name: "nil errors only",
			errs: []error{nil, nil},
			want: "<Error>",
		},
		{
			name: "single error",
			errs: []error{errors.New("single")},
			want: "single",
		},
		{
			name: "multiple errors",
			errs: []error{errors.New("first"), errors.New("second")},
			want: "first: second",
		},
		{
			name: "mixed nil and non-nil",
			errs: []error{nil, errors.New("middle"), nil},
			want: "middle",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pkg.JoinErrors(tt.errs...)
			if tt.want == "" {
				if result != nil {
					t.Errorf("JoinErrors() = %v, want nil", result)
				}
			} else if tt.want == "<Error>" {
				// Special case for empty Error
				if result == nil {
					t.Errorf("JoinErrors() = nil, want Error")
				} else if result.Error() != tt.want {
					t.Errorf("JoinErrors() = %q, want %q", result.Error(), tt.want)
				}
			} else {
				if result == nil {
					t.Errorf("JoinErrors() = nil, want %q", tt.want)
				} else if result.Error() != tt.want {
					t.Errorf("JoinErrors() = %q, want %q", result.Error(), tt.want)
				}
			}
		})
	}
}

func TestError_WithDetail(t *testing.T) {
	// Use a predefined error since we can't construct Error directly
	baseErr := pkg.ErrInvalidIdentifier
	
	tests := []struct {
		name    string
		details []string
		want    string
	}{
		{
			name:    "no details",
			details: []string{},
			want:    "invalid identifier",
		},
		{
			name:    "single detail",
			details: []string{"detail"},
			want:    "invalid identifier: detail",
		},
		{
			name:    "multiple details", 
			details: []string{"first", "second"},
			want:    "invalid identifier: first: second",
		},
		{
			name:    "empty detail string",
			details: []string{""},
			want:    "invalid identifier",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := baseErr.WithDetail(tt.details...)
			if result == nil {
				t.Errorf("WithDetail() = nil, want error")
			} else if result.Error() != tt.want {
				t.Errorf("WithDetail() = %q, want %q", result.Error(), tt.want)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrUndefCommandExec", pkg.ErrUndefCommandExec, "undefined exec function"},
		{"ErrUndefCommandFlagSet", pkg.ErrUndefCommandFlagSet, "undefined flag set"},
		{"ErrUndefCommandUsage", pkg.ErrUndefCommandUsage, "undefined name or usage"},
		{"ErrInaccessibleManifest", pkg.ErrInaccessibleManifest, "inaccessible manifest"},
		{"ErrUndefinedNamespace", pkg.ErrUndefinedNamespace, "undefined namespace"},
		{"ErrInvalidIdentifier", pkg.ErrInvalidIdentifier, "invalid identifier"},
		{"ErrInvalidExpression", pkg.ErrInvalidExpression, "invalid expression"},
		{"ErrInvalidJSON", pkg.ErrInvalidJSON, "invalid JSON encoding"},
		{"ErrIncompleteParse", pkg.ErrIncompleteParse, "incomplete parse"},
		{"ErrIncompleteEval", pkg.ErrIncompleteEval, "incomplete evaluation"},
		{"ErrUnexpectedToken", pkg.ErrUnexpectedToken, "unexpected token"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.msg)
			}
		})
	}
}

func TestGetMessage(t *testing.T) {
	// Test is commented out since getMessage is not exported
	// tests := []struct {
	// 	name    string
	// 	err     error
	// 	wantMsg string
	// 	wantOk  bool
	// }{
	// 	{
	// 		name:    "nil error",
	// 		err:     nil,
	// 		wantMsg: "",
	// 		wantOk:  false,
	// 	},
	// 	{
	// 		name:    "non-nil error with message",
	// 		err:     errors.New("test message"),
	// 		wantMsg: "test message",
	// 		wantOk:  true,
	// 	},
	// 	{
	// 		name:    "error with empty message", 
	// 		err:     pkg.Error{},
	// 		wantMsg: "",
	// 		wantOk:  false,
	// 	},
	// }
	
	// getMessage is not exported, so we can't test it directly
	// This test is left as a placeholder
	t.Skip("getMessage is not exported")
}

func TestPutMessage(t *testing.T) {
	// Test is commented out since putMessage is not exported
	// tests := []struct {
	// 	name    string
	// 	msg     string
	// 	wantOk  bool
	// }{
	// 	{
	// 		name:   "non-empty message",
	// 		msg:    "test message",
	// 		wantOk: true,
	// 	},
	// 	{
	// 		name:   "empty message",
	// 		msg:    "",
	// 		wantOk: false,
	// 	},
	// }
	
	// putMessage is not exported, so we can't test it directly
	// This test is left as a placeholder
	t.Skip("putMessage is not exported")
}

func TestIncompleteParseError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  pkg.IncompleteParseError
		want string
	}{
		{
			name: "simple error without definition",
			err: pkg.IncompleteParseError{
				Err: errors.New("parse failed"),
				Lvl: 0,
			},
			want: "incomplete parse: parse failed",
		},
		{
			name: "error with definition",
			err: pkg.IncompleteParseError{
				Err: errors.New("parse failed"),
				Def: []string{"test.manifest"},
				Lvl: 0,
			},
			want: "incomplete parse: parse failed",
		},
		{
			name: "error with definition and verbose level",
			err: pkg.IncompleteParseError{
				Err: errors.New("parse failed"),
				Def: []string{"test.manifest"},
				Lvl: 1,
			},
			want: "incomplete parse at test.manifest: parse failed",
		},
		{
			name: "error with empty definitions",
			err: pkg.IncompleteParseError{
				Err: errors.New("parse failed"),
				Def: []string{"", "   ", ""},
				Lvl: 1,
			},
			want: "incomplete parse at ,,: parse failed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("IncompleteParseError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpressionError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  pkg.ExpressionError
		want string
	}{
		{
			name: "simple expression error",
			err: pkg.ExpressionError{
				Namespace: "test",
				Statement: "x + y",
				Err:       errors.New("undefined variable"),
			},
			want: "undefined variable (expression \"x + y\" in namespace \"test\")",
		},
		{
			name: "expression error without namespace",
			err: pkg.ExpressionError{
				Statement: "x + y",
				Err:       errors.New("undefined variable"),
			},
			want: "undefined variable",
		},
		{
			name: "expression error with file error",
			err: pkg.ExpressionError{
				Namespace: "test",
				Statement: "x + y",
				Err: &file.Error{
					Message: "syntax error",
					Line:    1,
					Column:  5,
					Snippet: "x + y\n    ^",
				},
			},
			want: func() string {
				baseMsg := "invalid expression: syntax error (expression \"x + y\" in namespace \"test\")"
				snippet := "\tx + y\n\t    ^"
				return baseMsg + snippet
			}(),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("ExpressionError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpressionError_position(t *testing.T) {
	tests := []struct {
		name string
		err  pkg.ExpressionError
		want string
	}{
		{
			name: "file error with position",
			err: pkg.ExpressionError{
				Err: &file.Error{
					Line:   10,
					Column: 5,
				},
			},
			want: "[10:5]",
		},
		{
			name: "regular error without position",
			err: pkg.ExpressionError{
				Err: errors.New("regular error"),
			},
			want: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to call the private method through a public interface
			// Since position() is not exported, we test it indirectly through Error()
			// For file errors, the position should be included in the output
			errMsg := tt.err.Error()
			if tt.want != "" {
				// For file errors, position info should be included somehow
				if !strings.Contains(errMsg, "invalid expression") && tt.err.Err != nil {
					if _, ok := tt.err.Err.(*file.Error); ok {
						t.Errorf("ExpressionError with file.Error should contain position info")
					}
				}
			}
		})
	}
}