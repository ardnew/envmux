package pkg_test

import (
	"testing"

	"github.com/ardnew/envmux/pkg"
)

func TestName(t *testing.T) {
	expected := "envmux"
	if pkg.Name != expected {
		t.Errorf("Expected Name to be %q, got %q", expected, pkg.Name)
	}
}

func TestDescription(t *testing.T) {
	expected := "envmux is a tool for managing environments."
	if pkg.Description != expected {
		t.Errorf("Expected Description to be %q, got %q", expected, pkg.Description)
	}
}

func TestVersion(t *testing.T) {
	// Version is embedded from VERSION file, so it should not be empty
	// We can't test the exact value since it may change
	if pkg.Version == "" {
		t.Error("Expected Version to be non-empty")
	}
}

func TestAuthor(t *testing.T) {
	if len(pkg.Author) == 0 {
		t.Error("Expected Author to have at least one entry")
	}
	
	// Test the first author entry
	if len(pkg.Author) > 0 {
		expectedName := "ardnew"
		expectedEmail := "andrew@ardnew.com"
		
		if pkg.Author[0].Name != expectedName {
			t.Errorf("Expected first author name to be %q, got %q", expectedName, pkg.Author[0].Name)
		}
		
		if pkg.Author[0].Email != expectedEmail {
			t.Errorf("Expected first author email to be %q, got %q", expectedEmail, pkg.Author[0].Email)
		}
	}
}

func TestAuthorStruct(t *testing.T) {
	// Test that Author slice has the expected structure
	for i, author := range pkg.Author {
		if author.Name == "" {
			t.Errorf("Author[%d].Name should not be empty", i)
		}
		if author.Email == "" {
			t.Errorf("Author[%d].Email should not be empty", i)
		}
	}
}