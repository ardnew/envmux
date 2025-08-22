package fn

import (
	"iter"
	"slices"
	"testing"
)

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

		resultSeq := Apply(inputSeq, test.transform)

		var result []int
		if resultSeq != nil {
			result = slices.Collect(resultSeq)
		}

		if !equal(result, test.expected) {
			t.Errorf("Map(%v) = %v; want %v", test.input, result, test.expected)
		}
	}

	var i int
	Apply(slices.Values([]int{1, 2, 3}), func(v int) int { return v + 100 })(func(v int) bool {
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
