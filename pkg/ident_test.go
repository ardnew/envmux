package pkg

import (
	"slices"
	"testing"
)

func TestName(t *testing.T) {
	expected := "envmux"
	if Name != expected {
		t.Errorf("Expected Name to be %q, got %q", expected, Name)
	}
}

func TestDescription(t *testing.T) {
	expected := "static environment compositor"
	if Description != expected {
		t.Errorf("Expected Description to be %q, got %q", expected, Description)
	}
}

func TestVersion(t *testing.T) {
	// Version is embedded from VERSION file, so it should not be empty
	// We can't test the exact value since it may change
	if Version == "" {
		t.Error("Expected Version to be non-empty")
	}
}

func TestAuthor(t *testing.T) {
	if len(Author) == 0 {
		t.Error("Expected Author to have at least one entry")
	}

	// Test if a known author is present
	if len(Author) > 0 {
		expectedName := "ardnew"
		expectedEmail := "andrew@ardnew.com"

		if !slices.ContainsFunc(Author, func(a AuthorInfo) bool {
			return a.Name == expectedName && a.Email == expectedEmail
		}) {
			t.Errorf("Expected Author to contain %q, %q", expectedName, expectedEmail)
		}
	}
}

func TestAuthorStruct(t *testing.T) {
	// Test that Author slice has the expected structure
	for i, author := range Author {
		if author.Name == "" && author.Email == "" {
			t.Errorf("Author[%d] must define at least Name or Email", i)
		}
	}
}
