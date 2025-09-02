package shell

import (
	"errors"
	"testing"

	"github.com/ardnew/envmux/pkg"
)

func TestFormatEnvVar(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"foo", "bar"}, "FOO_BAR"},
		{[]string{"foo", "_bar_", "baz"}, "FOO_BAR_BAZ"},
		{[]string{"123", "leading", "digits"}, "_123_LEADING_DIGITS"},
		{[]string{"trailing", "digits", "123"}, "TRAILING_DIGITS_123"},
		{[]string{"a!b@c#d$"}, "A_B_C_D_"},
		{[]string{"  with  spaces  "}, "WITH_SPACES"},
		{[]string{"&^%special_chars!@#"}, "_SPECIAL_CHARS_"},
		{[]string{"__consecutive__underscores__"}, "_CONSECUTIVE_UNDERSCORES_"},
		{[]string{"mixed__CASE__123"}, "MIXED_CASE_123"},
		{[]string{"__"}, "_"},
		{[]string{"", ""}, "_"},
		{[]string{}, ""},
		{[]string{"_"}, "_"},
		{[]string{"_a", "b_"}, "_A_B_"},
		{[]string{"a", "_b"}, "A_B"},
		{[]string{"a", "b_"}, "A_B_"},
		{[]string{"a", "b", "_"}, "A_B_"},
		{[]string{"a", "_", "b"}, "A_B"},
		{[]string{"_", "a", "b"}, "_A_B"},
		{[]string{"_", "a", "_", "b"}, "_A_B"},
		{[]string{"_", "_", "_", "_"}, "_"},
		{[]string{"!@#$%^&*()"}, "_"},
		{[]string{"new!!input@", "test"}, "NEW_INPUT_TEST"},
	}

	for _, test := range tests {
		result := MakeIdent(test.input...)
		if result != test.expected {
			t.Errorf("FormatEnvVar(%v) = %v; want %v", test.input, result, test.expected)
		}
		failsafe := IdentFormat{}.makeIdent(test.input...)
		if failsafe != test.expected {
			t.Errorf("FormatEnvVar(%v) = %v; want %v", test.input, failsafe, test.expected)
		}
	}

	var validPanic bool
	defer func(valid *bool) {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && errors.Is(err, pkg.ErrInvalidIdentifier) {
				*valid = true
			}
		}
	}(&validPanic)

	DefaultIdentFormat = IdentFormat{}
	MakeIdent("foo")

	if !validPanic {
		t.Errorf("FormatEnvVar: expected panic with ErrInvalidEnvVar")
	}
}
