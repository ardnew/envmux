package builtin

import (
	"os"
	"testing"
)

// Test the system functions indirectly through the Cache

func TestSystemFunctionsInCache(t *testing.T) {
	cache := Cache()

	// Test that system-related values are present in cache
	systemKeys := []string{"target", "platform", "hostname", "user", "shell"}

	for _, key := range systemKeys {
		if _, exists := cache[key]; !exists {
			t.Errorf("Cache should contain system key %q", key)
		}
	}
}

func TestGetTargetViaCache(t *testing.T) {
	cache := Cache()

	target, ok := cache["target"]
	if !ok {
		t.Fatal("Cache should contain 'target' key")
	}

	// The target should be a runtime.Target struct
	// We can't easily test the exact type without importing the runtime package here
	// But we can verify it's not nil
	if target == nil {
		t.Error("target should not be nil")
	}
}

func TestGetPlatformViaCache(t *testing.T) {
	cache := Cache()

	platform, ok := cache["platform"]
	if !ok {
		t.Fatal("Cache should contain 'platform' key")
	}

	if platform == nil {
		t.Error("platform should not be nil")
	}
}

func TestGetHostnameViaCache(t *testing.T) {
	cache := Cache()

	hostname, ok := cache["hostname"]
	if !ok {
		t.Fatal("Cache should contain 'hostname' key")
	}

	hostnameStr, isString := hostname.(string)
	if !isString {
		t.Errorf("hostname should be string, got %T", hostname)
	}

	// Hostname might be empty on some systems, but should not be nil
	_ = hostnameStr // Use the value to avoid unused variable warning
}

func TestGetUserViaCache(t *testing.T) {
	cache := Cache()

	user, ok := cache["user"]
	if !ok {
		t.Fatal("Cache should contain 'user' key")
	}

	// User might be nil on some systems where user.Current() fails
	// If not nil, it should be a *user.User
	if user != nil {
		// We can't easily test the exact type here
		// Just verify it's something meaningful
		if user == "" {
			t.Error("user should not be empty string if not nil")
		}
	}
}

func TestGetShellViaCache(t *testing.T) {
	cache := Cache()

	shell, ok := cache["shell"]
	if !ok {
		t.Fatal("Cache should contain 'shell' key")
	}

	shellStr, isString := shell.(string)
	if !isString {
		t.Errorf("shell should be string, got %T", shell)
	}

	// Shell might be empty on some systems but should be a string
	_ = shellStr // Use the value
}

func TestGetShellWithSHELLEnv(t *testing.T) {
	// Save original SHELL env var
	originalShell := os.Getenv("SHELL")

	defer func() {
		if originalShell == "" {
			os.Unsetenv("SHELL")
		} else {
			os.Setenv("SHELL", originalShell)
		}
	}()

	// Set SHELL env var
	testShell := "/bin/bash"
	os.Setenv("SHELL", testShell)

	// Since the cache is likely already initialized, we can't easily test
	// the dynamic behavior without clearing the cache. This test documents
	// the expected behavior but may not actually verify it due to caching.

	// The shell value in cache should reflect the SHELL env var if it was set
	// when the cache was first initialized
	cache := Cache()
	shell, ok := cache["shell"]

	if !ok {
		t.Fatal("Cache should contain 'shell' key")
	}

	if _, isString := shell.(string); !isString {
		t.Errorf("shell should be string, got %T", shell)
	}
}

// Test system function behavior indirectly
func TestSystemFunctionsBehavior(t *testing.T) {
	cache := Cache()

	// Test that target and platform are different types/values
	target := cache["target"]
	platform := cache["platform"]

	if target == nil || platform == nil {
		t.Fatal("Both target and platform should be present")
	}

	// They should both exist but may have different values
	// (target uses GNU/LLVM conventions, platform uses Go conventions)
}

func TestCacheSystemConsistency(t *testing.T) {
	// Test that multiple calls to Cache() return consistent system values
	cache1 := Cache()
	cache2 := Cache()

	systemKeys := []string{"target", "platform", "hostname", "user", "shell"}

	for _, key := range systemKeys {
		val1, ok1 := cache1[key]
		val2, ok2 := cache2[key]

		if ok1 != ok2 {
			t.Errorf("Key %q presence should be consistent across cache calls", key)
		}

		if ok1 && ok2 {
			// For system values, they should be identical between cache calls
			// (though we can't do deep equality easily here)
			if (val1 == nil) != (val2 == nil) {
				t.Errorf("Key %q nullability should be consistent across cache calls", key)
			}
		}
	}
}
