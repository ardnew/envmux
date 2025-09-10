package builtin

import (
	"testing"
)

func TestCache(t *testing.T) {
	cache := Cache()

	// Test that cache is not nil and contains expected keys
	if cache == nil {
		t.Fatal("Cache() should not return nil")
	}

	expectedKeys := []string{"target", "platform", "hostname", "user", "shell", "cwd", "file", "path", "mung"}
	for _, key := range expectedKeys {
		if _, exists := cache[key]; !exists {
			t.Errorf("Cache() should contain key %q", key)
		}
	}
}

func TestCacheTarget(t *testing.T) {
	cache := Cache()

	target, ok := cache["target"]
	if !ok {
		t.Fatal("Cache should contain 'target' key")
	}

	// The target should not be nil
	if target == nil {
		t.Error("target should not be nil")
	}
}

func TestCachePlatform(t *testing.T) {
	cache := Cache()

	platform, ok := cache["platform"]
	if !ok {
		t.Fatal("Cache should contain 'platform' key")
	}

	// The platform should not be nil
	if platform == nil {
		t.Error("platform should not be nil")
	}
}

func TestCacheHostname(t *testing.T) {
	cache := Cache()

	hostname, ok := cache["hostname"]
	if !ok {
		t.Fatal("Cache should contain 'hostname' key")
	}

	// The hostname might be empty string on some systems, but should be a string
	if _, isString := hostname.(string); !isString {
		t.Errorf("hostname should be string, got %T", hostname)
	}
}

func TestCacheUser(t *testing.T) {
	cache := Cache()

	user, ok := cache["user"]
	if !ok {
		t.Fatal("Cache should contain 'user' key")
	}

	// User might be nil on some systems, but if not nil should be *user.User
	if user != nil {
		// We can't easily check the type without importing os/user
		// Just verify it's not empty
		if user == "" {
			t.Error("user should not be empty string if not nil")
		}
	}
}

func TestCacheShell(t *testing.T) {
	cache := Cache()

	shell, ok := cache["shell"]
	if !ok {
		t.Fatal("Cache should contain 'shell' key")
	}

	// Shell should be a string (might be empty on some systems)
	if _, isString := shell.(string); !isString {
		t.Errorf("shell should be string, got %T", shell)
	}
}

func TestCacheCwd(t *testing.T) {
	cache := Cache()

	cwdFunc, ok := cache["cwd"]
	if !ok {
		t.Fatal("Cache should contain 'cwd' key")
	}

	// cwd should be a function
	if cwdFunc == nil {
		t.Error("cwd should not be nil")
	}
}

func TestCacheFile(t *testing.T) {
	cache := Cache()

	file, ok := cache["file"]
	if !ok {
		t.Fatal("Cache should contain 'file' key")
	}

	fileMap, ok := file.(map[string]any)
	if !ok {
		t.Fatalf("file should be map[string]any, got %T", file)
	}

	expectedFileFuncs := []string{"exists", "isDir", "isRegular", "isSymlink", "perms", "stat"}
	for _, funcName := range expectedFileFuncs {
		if _, exists := fileMap[funcName]; !exists {
			t.Errorf("file map should contain function %q", funcName)
		}
	}
}

func TestCachePath(t *testing.T) {
	cache := Cache()

	path, ok := cache["path"]
	if !ok {
		t.Fatal("Cache should contain 'path' key")
	}

	pathMap, ok := path.(map[string]any)
	if !ok {
		t.Fatalf("path should be map[string]any, got %T", path)
	}

	expectedPathFuncs := []string{"abs", "cat", "rel"}
	for _, funcName := range expectedPathFuncs {
		if _, exists := pathMap[funcName]; !exists {
			t.Errorf("path map should contain function %q", funcName)
		}
	}
}

func TestCacheMung(t *testing.T) {
	cache := Cache()

	mung, ok := cache["mung"]
	if !ok {
		t.Fatal("Cache should contain 'mung' key")
	}

	mungMap, ok := mung.(map[string]any)
	if !ok {
		t.Fatalf("mung should be map[string]any, got %T", mung)
	}

	expectedMungFuncs := []string{"prefix", "prefixif"}
	for _, funcName := range expectedMungFuncs {
		if _, exists := mungMap[funcName]; !exists {
			t.Errorf("mung map should contain function %q", funcName)
		}
	}
}

func TestCacheClone(t *testing.T) {
	// Test that Cache() returns a clone, not the original
	cache1 := Cache()
	cache2 := Cache()

	// Modify one cache
	cache1["test"] = "value"

	// The other cache should not be affected
	if _, exists := cache2["test"]; exists {
		t.Error("Cache() should return a clone, not the original map")
	}
}

// Test some functions indirectly through their presence in cache
func TestFunctionTypes(t *testing.T) {
	cache := Cache()

	// Test file functions are callable
	fileMap := cache["file"].(map[string]any)

	// We can't call the functions directly due to type assertions,
	// but we can verify they exist and are not nil
	for name, fn := range fileMap {
		if fn == nil {
			t.Errorf("file.%s function should not be nil", name)
		}
	}

	// Test path functions
	pathMap := cache["path"].(map[string]any)
	for name, fn := range pathMap {
		if fn == nil {
			t.Errorf("path.%s function should not be nil", name)
		}
	}

	// Test mung functions
	mungMap := cache["mung"].(map[string]any)
	for name, fn := range mungMap {
		if fn == nil {
			t.Errorf("mung.%s function should not be nil", name)
		}
	}
}
