package pprof_test

import (
	"strings"
	"testing"

	"github.com/ardnew/envmux/cmd/envmux/pprof"
)

func TestProfiler_Start(t *testing.T) {
	tests := []struct {
		name string
		prof pprof.Profiler
	}{
		{
			name: "empty mode",
			prof: pprof.Profiler{Mode: ""},
		},
		{
			name: "with mode",
			prof: pprof.Profiler{Mode: "cpu", Path: "/tmp", Quiet: true},
		},
		{
			name: "mode with path delimiter",
			prof: pprof.Profiler{Mode: "cpu=/custom/path"},
		},
		{
			name: "quiet mode",
			prof: pprof.Profiler{Mode: "mem", Quiet: true},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start should not panic
			stopper := tt.prof.Start()
			
			// Should return something with Stop method
			if stopper == nil {
				t.Error("Start() should return non-nil stopper")
			}
			
			// Stop should not panic
			stopper.Stop()
		})
	}
}

func TestProfiler_StartWithDelimiter(t *testing.T) {
	// Test that mode with "=" delimiter is handled
	prof := pprof.Profiler{
		Mode: "cpu=/custom/output/path",
		Path: "/default/path",
	}
	
	stopper := prof.Start()
	if stopper == nil {
		t.Error("Start() should return non-nil stopper")
	}
	
	// Should not panic
	stopper.Stop()
}

func TestProfiler_Fields(t *testing.T) {
	prof := pprof.Profiler{
		Mode:  "cpu",
		Path:  "/tmp/profile",
		Quiet: true,
	}
	
	if prof.Mode != "cpu" {
		t.Errorf("Expected Mode 'cpu', got %q", prof.Mode)
	}
	
	if prof.Path != "/tmp/profile" {
		t.Errorf("Expected Path '/tmp/profile', got %q", prof.Path)
	}
	
	if !prof.Quiet {
		t.Error("Expected Quiet to be true")
	}
}

func TestModes(t *testing.T) {
	// Test that Modes() returns a slice (may be nil if pprof build tag not set)
	modes := pprof.Modes()
	
	// In this build environment, we likely don't have the pprof tag,
	// so modes might be nil
	_ = modes // Just verify it doesn't panic
	
	// If modes is not nil, verify it contains strings
	if modes != nil {
		for i, mode := range modes {
			if mode == "" {
				t.Errorf("Mode at index %d should not be empty", i)
			}
		}
	}
}

func TestIgnoreType(t *testing.T) {
	// Test the ignore type behavior indirectly
	prof := pprof.Profiler{Mode: ""} // Empty mode uses ignore
	
	stopper := prof.Start()
	if stopper == nil {
		t.Error("Start() should return non-nil stopper even for empty mode")
	}
	
	// Should not panic
	stopper.Stop()
}

func TestProfiler_ModeDelimiterParsing(t *testing.T) {
	// Test various delimiter scenarios
	tests := []struct {
		name         string
		mode         string
		expectedMode string
		expectedPath string
	}{
		{
			name:         "no delimiter",
			mode:         "cpu",
			expectedMode: "cpu",
			expectedPath: "", // Original path should be preserved
		},
		{
			name:         "with delimiter",
			mode:         "cpu=/custom/path",
			expectedMode: "cpu",
			expectedPath: "/custom/path",
		},
		{
			name:         "empty after delimiter",
			mode:         "mem=",
			expectedMode: "mem",
			expectedPath: "",
		},
		{
			name:         "multiple delimiters",
			mode:         "trace=/path/with=equals",
			expectedMode: "trace",
			expectedPath: "/path/with=equals",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the internal parsing without exposing it,
			// but we can test that various mode formats don't cause panics
			
			prof := pprof.Profiler{
				Mode: tt.mode,
				Path: "/default/path",
			}
			
			stopper := prof.Start()
			if stopper == nil {
				t.Error("Start() should return non-nil stopper")
			}
			
			stopper.Stop()
		})
	}
}

func TestProfiler_StringCutBehavior(t *testing.T) {
	// Test that strings.Cut works as expected (used in profiler)
	tests := []struct {
		s        string
		sep      string
		before   string
		after    string
		found    bool
	}{
		{"cpu=/tmp", "=", "cpu", "/tmp", true},
		{"mem", "=", "mem", "", false},
		{"trace=/path/with/slashes", "=", "trace", "/path/with/slashes", true},
		{"", "=", "", "", false},
	}
	
	for _, tt := range tests {
		before, after, found := strings.Cut(tt.s, tt.sep)
		if before != tt.before {
			t.Errorf("Cut(%q, %q): before = %q, want %q", tt.s, tt.sep, before, tt.before)
		}
		if after != tt.after {
			t.Errorf("Cut(%q, %q): after = %q, want %q", tt.s, tt.sep, after, tt.after)
		}
		if found != tt.found {
			t.Errorf("Cut(%q, %q): found = %v, want %v", tt.s, tt.sep, found, tt.found)
		}
	}
}