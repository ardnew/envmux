package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPrefix(t *testing.T) {
	// Note: Prefix function uses os.Executable() which cannot be easily mocked
	// in tests. We'll test the basic functionality and regex patterns indirectly.

	t.Run("basic functionality", func(t *testing.T) {
		result := Prefix("testcmd")

		// Result should not be empty
		if result == "" {
			t.Error("Prefix should not return empty string")
		}

		// Should be a valid filename (no path separators)
		if strings.Contains(result, "/") || strings.Contains(result, "\\") {
			t.Error("Prefix should return just the base name, not a path")
		}
	})

	// Test individual regex patterns by simulating the behavior
	t.Run("debug binary pattern", func(t *testing.T) {
		// Test that the regex would match debug binaries
		rex := regexp.MustCompile(`^__debug_bin\d+$`)

		if !rex.MatchString("__debug_bin123") {
			t.Error("Debug binary regex should match __debug_bin123")
		}

		if !rex.MatchString("__debug_bin1") {
			t.Error("Debug binary regex should match __debug_bin1")
		}

		if rex.MatchString("__debug_bin") {
			t.Error("Debug binary regex should not match __debug_bin without digits")
		}
	})

	t.Run("dot prefix pattern", func(t *testing.T) {
		rex := regexp.MustCompile(`^\.+`)

		if !rex.MatchString(".envmux") {
			t.Error("Dot prefix regex should match .envmux")
		}

		if !rex.MatchString("...envmux") {
			t.Error("Dot prefix regex should match ...envmux")
		}

		if rex.MatchString("envmux") {
			t.Error("Dot prefix regex should not match envmux")
		}
	})
}

func TestDir(t *testing.T) {
	// Save original environment variables
	origXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	origHome := os.Getenv("HOME")

	defer func() {
		restoreEnv("XDG_CONFIG_HOME", origXDGConfigHome)
		restoreEnv("HOME", origHome)
	}()

	tests := []struct {
		name           string
		xdgConfigHome  string
		home           string
		cmd            string
		expectedSuffix string
	}{
		{
			name:           "XDG_CONFIG_HOME set",
			xdgConfigHome:  "/custom/config",
			home:           "/home/user",
			cmd:            "testcmd",
			expectedSuffix: "testcmd",
		},
		{
			name:           "HOME set, no XDG",
			xdgConfigHome:  "",
			home:           "/home/user",
			cmd:            "testcmd",
			expectedSuffix: filepath.Join(".config", "testcmd"),
		},
		{
			name:           "neither set",
			xdgConfigHome:  "",
			home:           "",
			cmd:            "testcmd",
			expectedSuffix: "testcmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Unsetenv("XDG_CONFIG_HOME")
			os.Unsetenv("HOME")

			if tt.xdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			}
			if tt.home != "" {
				os.Setenv("HOME", tt.home)
			}

			result := Dir(tt.cmd)

			// Since Prefix uses the actual executable name, we need to get what Prefix returns
			expectedPrefix := Prefix(tt.cmd)

			if tt.xdgConfigHome != "" {
				expected := filepath.Join(tt.xdgConfigHome, expectedPrefix)
				if result != expected {
					t.Errorf("Dir(%q) = %q, want %q", tt.cmd, result, expected)
				}
			} else if tt.home != "" {
				expected := filepath.Join(tt.home, ".config", expectedPrefix)
				if result != expected {
					t.Errorf("Dir(%q) = %q, want %q", tt.cmd, result, expected)
				}
			} else {
				// When neither is set, should use current working directory + prefix
				if !strings.HasSuffix(result, expectedPrefix) {
					t.Errorf("Dir(%q) = %q, should end with %q", tt.cmd, result, expectedPrefix)
				}
			}
		})
	}
}

func TestCache(t *testing.T) {
	// Save original environment variables
	origXDGCacheHome := os.Getenv("XDG_CACHE_HOME")
	origHome := os.Getenv("HOME")

	defer func() {
		restoreEnv("XDG_CACHE_HOME", origXDGCacheHome)
		restoreEnv("HOME", origHome)
	}()

	tests := []struct {
		name           string
		xdgCacheHome   string
		home           string
		cmd            string
		expectedSuffix string
	}{
		{
			name:           "XDG_CACHE_HOME set",
			xdgCacheHome:   "/custom/cache",
			home:           "/home/user",
			cmd:            "testcmd",
			expectedSuffix: "testcmd",
		},
		{
			name:           "HOME set, no XDG",
			xdgCacheHome:   "",
			home:           "/home/user",
			cmd:            "testcmd",
			expectedSuffix: filepath.Join(".cache", "testcmd"),
		},
		{
			name:           "neither set",
			xdgCacheHome:   "",
			home:           "",
			cmd:            "testcmd",
			expectedSuffix: "testcmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Unsetenv("XDG_CACHE_HOME")
			os.Unsetenv("HOME")

			if tt.xdgCacheHome != "" {
				os.Setenv("XDG_CACHE_HOME", tt.xdgCacheHome)
			}
			if tt.home != "" {
				os.Setenv("HOME", tt.home)
			}

			result := Cache(tt.cmd)

			// Since Prefix uses the actual executable name, we need to get what Prefix returns
			expectedPrefix := Prefix(tt.cmd)

			if tt.xdgCacheHome != "" {
				expected := filepath.Join(tt.xdgCacheHome, expectedPrefix)
				if result != expected {
					t.Errorf("Cache(%q) = %q, want %q", tt.cmd, result, expected)
				}
			} else if tt.home != "" {
				expected := filepath.Join(tt.home, ".cache", expectedPrefix)
				if result != expected {
					t.Errorf("Cache(%q) = %q, want %q", tt.cmd, result, expected)
				}
			} else {
				// When neither is set, should use current working directory + prefix
				if !strings.HasSuffix(result, expectedPrefix) {
					t.Errorf("Cache(%q) = %q, should end with %q", tt.cmd, result, expectedPrefix)
				}
			}
		})
	}
}

func TestStdinManifestPath(t *testing.T) {
	expected := "-"
	if StdinManifestPath != expected {
		t.Errorf("StdinManifestPath = %q, want %q", StdinManifestPath, expected)
	}
}

func TestDefaultManifestPath(t *testing.T) {
	cmd := "testcmd"
	result := DefaultManifestPath(cmd)

	if len(result) == 0 {
		t.Error("DefaultManifestPath should return at least one path")
	}

	// Should contain the config dir + "default"
	expected := filepath.Join(Dir(cmd), "default")
	if len(result) > 0 && result[0] != expected {
		t.Errorf("DefaultManifestPath(%q)[0] = %q, want %q", cmd, result[0], expected)
	}
}

func TestDefaultNamespace(t *testing.T) {
	result := DefaultNamespace()

	if len(result) == 0 {
		t.Error("DefaultNamespace should return at least one namespace")
	}

	expected := "default"
	if len(result) > 0 && result[0] != expected {
		t.Errorf("DefaultNamespace()[0] = %q, want %q", result[0], expected)
	}
}

func TestConfigIntegration(t *testing.T) {
	// Test that Dir and DefaultManifestPath work together consistently
	cmd := "integration_test"

	dir := Dir(cmd)
	manifestPaths := DefaultManifestPath(cmd)

	if len(manifestPaths) == 0 {
		t.Fatal("DefaultManifestPath should return at least one path")
	}

	// The first manifest path should be within the config directory
	expectedPath := filepath.Join(dir, "default")
	if manifestPaths[0] != expectedPath {
		t.Errorf("Expected manifest path %q, got %q", expectedPath, manifestPaths[0])
	}
}

// Helper function to restore environment variables
func restoreEnv(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}
