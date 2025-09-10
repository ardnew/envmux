package runtime

import (
	"os"
	"runtime"
	"testing"
)

func TestTarget(t *testing.T) {
	// Test Target struct initialization
	target := Target{
		OS:   "linux",
		Arch: "amd64",
	}

	if target.OS != "linux" {
		t.Errorf("Expected OS to be 'linux', got %s", target.OS)
	}

	if target.Arch != "amd64" {
		t.Errorf("Expected Arch to be 'amd64', got %s", target.Arch)
	}
}

func TestGetPlatform(t *testing.T) {
	// Save original env vars
	origGOOS := os.Getenv("GOOS")
	origGOARCH := os.Getenv("GOARCH")
	origGOHOSTOS := os.Getenv("GOHOSTOS")
	origGOHOSTARCH := os.Getenv("GOHOSTARCH")

	defer func() {
		// Restore original env vars
		os.Setenv("GOOS", origGOOS)
		os.Setenv("GOARCH", origGOARCH)
		os.Setenv("GOHOSTOS", origGOHOSTOS)
		os.Setenv("GOHOSTARCH", origGOHOSTARCH)
	}()

	t.Run("default runtime values", func(t *testing.T) {
		// Clear all env vars
		os.Unsetenv("GOOS")
		os.Unsetenv("GOARCH")
		os.Unsetenv("GOHOSTOS")
		os.Unsetenv("GOHOSTARCH")

		target := GetPlatform()

		// Should use runtime.GOOS and runtime.GOARCH
		if target.OS != runtime.GOOS {
			t.Errorf("Expected OS to be %s, got %s", runtime.GOOS, target.OS)
		}

		if target.Arch != runtime.GOARCH {
			t.Errorf("Expected Arch to be %s, got %s", runtime.GOARCH, target.Arch)
		}
	})

	t.Run("GOOS and GOARCH env vars", func(t *testing.T) {
		// Clear host vars, set build vars
		os.Unsetenv("GOHOSTOS")
		os.Unsetenv("GOHOSTARCH")
		os.Setenv("GOOS", "freebsd")
		os.Setenv("GOARCH", "arm64")

		target := GetPlatform()

		if target.OS != "freebsd" {
			t.Errorf("Expected OS to be 'freebsd', got %s", target.OS)
		}

		if target.Arch != "arm64" {
			t.Errorf("Expected Arch to be 'arm64', got %s", target.Arch)
		}
	})

	t.Run("GOHOST env vars take precedence", func(t *testing.T) {
		// Set both host and build vars
		os.Setenv("GOOS", "linux")
		os.Setenv("GOARCH", "amd64")
		os.Setenv("GOHOSTOS", "darwin")
		os.Setenv("GOHOSTARCH", "arm64")

		target := GetPlatform()

		// Host vars should take precedence
		if target.OS != "darwin" {
			t.Errorf("Expected OS to be 'darwin', got %s", target.OS)
		}

		if target.Arch != "arm64" {
			t.Errorf("Expected Arch to be 'arm64', got %s", target.Arch)
		}
	})
}

func TestGetTarget(t *testing.T) {
	// Save original env vars
	origGOOS := os.Getenv("GOOS")
	origGOARCH := os.Getenv("GOARCH")
	origGOHOSTOS := os.Getenv("GOHOSTOS")
	origGOHOSTARCH := os.Getenv("GOHOSTARCH")
	origGOARM := os.Getenv("GOARM")

	defer func() {
		// Restore original env vars
		os.Setenv("GOOS", origGOOS)
		os.Setenv("GOARCH", origGOARCH)
		os.Setenv("GOHOSTOS", origGOHOSTOS)
		os.Setenv("GOHOSTARCH", origGOHOSTARCH)
		os.Setenv("GOARM", origGOARM)
	}()

	tests := []struct {
		name         string
		inputOS      string
		inputArch    string
		inputGOARM   string
		expectedOS   string
		expectedArch string
	}{
		{
			name:         "386 to i386",
			inputOS:      "linux",
			inputArch:    "386",
			expectedOS:   "linux",
			expectedArch: "i386",
		},
		{
			name:         "amd64 to x86_64",
			inputOS:      "linux",
			inputArch:    "amd64",
			expectedOS:   "linux",
			expectedArch: "x86_64",
		},
		{
			name:         "arm with GOARM=5",
			inputOS:      "linux",
			inputArch:    "arm",
			inputGOARM:   "5",
			expectedOS:   "linux",
			expectedArch: "armv5",
		},
		{
			name:         "arm with GOARM=7",
			inputOS:      "linux",
			inputArch:    "arm",
			inputGOARM:   "7",
			expectedOS:   "linux",
			expectedArch: "armv7",
		},
		{
			name:         "arm with GOARM=8 (invalid)",
			inputOS:      "linux",
			inputArch:    "arm",
			inputGOARM:   "8",
			expectedOS:   "linux",
			expectedArch: "arm", // should remain unchanged
		},
		{
			name:         "arm64 on linux to aarch64",
			inputOS:      "linux",
			inputArch:    "arm64",
			expectedOS:   "linux",
			expectedArch: "aarch64",
		},
		{
			name:         "arm64 on darwin remains arm64",
			inputOS:      "darwin",
			inputArch:    "arm64",
			expectedOS:   "darwin",
			expectedArch: "arm64",
		},
		{
			name:         "mipsle to mipsel",
			inputOS:      "linux",
			inputArch:    "mipsle",
			expectedOS:   "linux",
			expectedArch: "mipsel",
		},
		{
			name:         "unchanged arch",
			inputOS:      "windows",
			inputArch:    "amd64",
			expectedOS:   "windows",
			expectedArch: "x86_64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up env for this test
			os.Unsetenv("GOOS")
			os.Unsetenv("GOARCH")
			os.Setenv("GOHOSTOS", tt.inputOS)
			os.Setenv("GOHOSTARCH", tt.inputArch)

			if tt.inputGOARM != "" {
				os.Setenv("GOARM", tt.inputGOARM)
			} else {
				os.Unsetenv("GOARM")
			}

			target := GetTarget()

			if target.OS != tt.expectedOS {
				t.Errorf("Expected OS to be %s, got %s", tt.expectedOS, target.OS)
			}

			if target.Arch != tt.expectedArch {
				t.Errorf("Expected Arch to be %s, got %s", tt.expectedArch, target.Arch)
			}
		})
	}
}

func TestGetTarget_GOARMWithComma(t *testing.T) {
	// Test GOARM value with comma (e.g., "7,softfloat")
	origEnvs := map[string]string{
		"GOOS":       os.Getenv("GOOS"),
		"GOARCH":     os.Getenv("GOARCH"),
		"GOHOSTOS":   os.Getenv("GOHOSTOS"),
		"GOHOSTARCH": os.Getenv("GOHOSTARCH"),
		"GOARM":      os.Getenv("GOARM"),
	}

	defer func() {
		for k, v := range origEnvs {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Set up test environment
	os.Unsetenv("GOOS")
	os.Unsetenv("GOARCH")
	os.Setenv("GOHOSTOS", "linux")
	os.Setenv("GOHOSTARCH", "arm")
	os.Setenv("GOARM", "7,softfloat")

	target := GetTarget()

	if target.OS != "linux" {
		t.Errorf("Expected OS to be 'linux', got %s", target.OS)
	}

	if target.Arch != "armv7" {
		t.Errorf("Expected Arch to be 'armv7', got %s", target.Arch)
	}
}
