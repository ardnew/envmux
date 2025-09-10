package cli

import (
	"errors"
	"testing"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root"
)

func TestRunError_Error(t *testing.T) {
	tests := []struct {
		name   string
		result RunError
		want   string
	}{
		{
			name:   "no error",
			result: RunError{},
			want:   "",
		},
		{
			name:   "with error",
			result: RunError{Err: errors.New("test error")},
			want:   "test error",
		},
		{
			name:   "with help and error",
			result: RunError{Err: errors.New("test error"), Help: "help text"},
			want:   "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Error(); got != tt.want {
				t.Errorf("RunError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrRunOK(t *testing.T) {
	// Test that ErrRunOK is the zero value
	if ErrRunOK.Err != nil {
		t.Error("ErrRunOK.Err should be nil")
	}

	if ErrRunOK.Help != "" {
		t.Error("ErrRunOK.Help should be empty")
	}

	if ErrRunOK.Code != 0 {
		t.Error("ErrRunOK.Code should be 0")
	}

	if ErrRunOK.Error() != "" {
		t.Error("ErrRunOK.Error() should return empty string")
	}
}

func TestMakeResult(t *testing.T) {
	// Create a mock node for testing
	node := root.Init()

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantErr  bool
	}{
		{
			name:     "no error",
			err:      nil,
			wantCode: 0,
			wantErr:  false,
		},
		{
			name:     "help error",
			err:      ff.ErrHelp,
			wantCode: 0,
			wantErr:  false,
		},
		{
			name:     "no exec error",
			err:      ff.ErrNoExec,
			wantCode: 0,
			wantErr:  false,
		},
		{
			name:     "other error",
			err:      errors.New("test error"),
			wantCode: 1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeResult(node, tt.err)

			if result.Code != tt.wantCode {
				t.Errorf("MakeResult().Code = %v, want %v", result.Code, tt.wantCode)
			}

			hasErr := result.Err != nil
			if hasErr != tt.wantErr {
				t.Errorf("MakeResult().Err = %v, want error: %v", result.Err, tt.wantErr)
			}

			// Help should be set for ff.ErrHelp and ff.ErrNoExec
			if tt.err == ff.ErrHelp || tt.err == ff.ErrNoExec {
				if result.Help == "" {
					t.Error("MakeResult().Help should not be empty for help/noexec errors")
				}
			}
		})
	}
}

func TestRunError_Fields(t *testing.T) {
	// Test that RunError fields are accessible
	result := RunError{
		Err:  errors.New("test"),
		Help: "help message",
		Code: 42,
	}

	if result.Err.Error() != "test" {
		t.Errorf("Expected Err message 'test', got %q", result.Err.Error())
	}

	if result.Help != "help message" {
		t.Errorf("Expected Help 'help message', got %q", result.Help)
	}

	if result.Code != 42 {
		t.Errorf("Expected Code 42, got %d", result.Code)
	}
}
