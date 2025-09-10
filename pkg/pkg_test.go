// Package pkg_test tests package [pkg] as an imported package.
package pkg_test

import (
	"testing"

	"github.com/ardnew/envmux/pkg"
)

// TestPackageImport ensures the package can be imported successfully.
// This is a basic sanity check.
//
// This test must be in a package external to 'pkg'.
func TestPackageImport(t *testing.T) {
	if pkg.Name == "" {
		t.Error("Expected package name to be non-empty")
	}
}
