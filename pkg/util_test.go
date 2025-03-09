package pkg

import (
	"errors"
	"iter"
	"slices"
	"testing"
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
		result := FormatEnvVar(test.input...)
		if result != test.expected {
			t.Errorf("FormatEnvVar(%v) = %v; want %v", test.input, result, test.expected)
		}
		failsafe := EnvVarOption{}.FormatEnvVar(test.input...)
		if failsafe != test.expected {
			t.Errorf("FormatEnvVar(%v) = %v; want %v", test.input, failsafe, test.expected)
		}
	}

	var validPanic bool
	defer func(valid *bool) {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && errors.Is(err, ErrDefaultEnvVarOption) {
				*valid = true
			}
		}
	}(&validPanic)

	DefaultEnvVarOption = EnvVarOption{}
	FormatEnvVar("foo")

	if !validPanic {
		t.Errorf("FormatEnvVar: expected panic with ErrDefaultEnvVarOption")
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		input     []int
		transform func(int) int
		expected  []int
	}{
		{nil, nil, nil},
		{nil, func(x int) int { return x + 1 }, nil},

		{[]int{}, func(x int) int { return x + 1 }, []int{}},
		{[]int{}, nil, []int{}},

		{[]int{0}, func(x int) int { return x + 1 }, []int{1}},
		{[]int{0}, nil, []int{0}},

		{[]int{1, 2, 3}, func(x int) int { return x + 1 }, []int{2, 3, 4}},
		{[]int{1, 2, 3}, func(x int) int { return x }, []int{1, 2, 3}},
		{[]int{1, 2, 3}, nil, []int{1, 2, 3}},
	}

	for _, test := range tests {

		var inputSeq iter.Seq[int]
		if test.input != nil {
			inputSeq = slices.Values(test.input)
		}

		resultSeq := Map(inputSeq, test.transform)

		var result []int
		if resultSeq != nil {
			result = slices.Collect(resultSeq)
		}

		if !equal(result, test.expected) {
			t.Errorf("Map(%v) = %v; want %v", test.input, result, test.expected)
		}
	}

	var i int
	Map(slices.Values([]int{1, 2, 3}), func(v int) int { return v + 100 })(func(v int) bool {
		i = v
		return false
	})
	if i != 101 {
		t.Errorf("Map: expected early stop after first element")
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
