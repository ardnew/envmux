package pkg_test

import (
	"testing"

	"github.com/ardnew/envmux/pkg"
)

// TestPackageImport ensures the package can be imported successfully
func TestPackageImport(t *testing.T) {
	// This test verifies that the package can be imported and used
	// Since pkg/doc.go only contains package documentation,
	// we test that we can reference the package constants
	if pkg.Name == "" {
		t.Error("Expected package name to be non-empty")
	}
}