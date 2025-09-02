package fn_test

import (
	"slices"
	"testing"

	"github.com/ardnew/envmux/pkg/fn"
)

func TestOK(t *testing.T) {
	// Test with various types
	t.Run("int", func(t *testing.T) {
		value := 42
		result := fn.OK(value, 123, 456)
		if result != value {
			t.Errorf("OK(%d, ...) = %d, want %d", value, result, value)
		}
	})
	
	t.Run("string", func(t *testing.T) {
		value := "test"
		result := fn.OK(value, "ignored", "also ignored")
		if result != value {
			t.Errorf("OK(%q, ...) = %q, want %q", value, result, value)
		}
	})
	
	t.Run("struct", func(t *testing.T) {
		type testStruct struct{ Name string }
		value := testStruct{Name: "test"}
		result := fn.OK(value, testStruct{Name: "ignored"})
		if result != value {
			t.Errorf("OK(%+v, ...) = %+v, want %+v", value, result, value)
		}
	})
}

func TestIsEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"equal ints", 5, 5, true},
		{"unequal ints", 5, 3, false},
		{"equal strings", "hello", "hello", true},
		{"unequal strings", "hello", "world", false},
		{"equal bools", true, true, true},
		{"unequal bools", true, false, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch a := tt.a.(type) {
			case int:
				b := tt.b.(int)
				if got := fn.IsEqual(a, b); got != tt.want {
					t.Errorf("IsEqual(%v, %v) = %v, want %v", a, b, got, tt.want)
				}
			case string:
				b := tt.b.(string)
				if got := fn.IsEqual(a, b); got != tt.want {
					t.Errorf("IsEqual(%v, %v) = %v, want %v", a, b, got, tt.want)
				}
			case bool:
				b := tt.b.(bool)
				if got := fn.IsEqual(a, b); got != tt.want {
					t.Errorf("IsEqual(%v, %v) = %v, want %v", a, b, got, tt.want)
				}
			}
		})
	}
}

func TestIsUnequal(t *testing.T) {
	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"equal ints", 5, 5, false},
		{"unequal ints", 5, 3, true},
		{"equal strings", "hello", "hello", false},
		{"unequal strings", "hello", "world", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch a := tt.a.(type) {
			case int:
				b := tt.b.(int)
				if got := fn.IsUnequal(a, b); got != tt.want {
					t.Errorf("IsUnequal(%v, %v) = %v, want %v", a, b, got, tt.want)
				}
			case string:
				b := tt.b.(string)
				if got := fn.IsUnequal(a, b); got != tt.want {
					t.Errorf("IsUnequal(%v, %v) = %v, want %v", a, b, got, tt.want)
				}
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		if !fn.IsZero(0) {
			t.Error("IsZero(0) should be true")
		}
		if fn.IsZero(1) {
			t.Error("IsZero(1) should be false")
		}
	})
	
	t.Run("string", func(t *testing.T) {
		if !fn.IsZero("") {
			t.Error("IsZero(\"\") should be true")
		}
		if fn.IsZero("test") {
			t.Error("IsZero(\"test\") should be false")
		}
	})
	
	t.Run("bool", func(t *testing.T) {
		if !fn.IsZero(false) {
			t.Error("IsZero(false) should be true")
		}
		if fn.IsZero(true) {
			t.Error("IsZero(true) should be false")
		}
	})
}

func TestIsNonzero(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		if fn.IsNonzero(0) {
			t.Error("IsNonzero(0) should be false")
		}
		if !fn.IsNonzero(1) {
			t.Error("IsNonzero(1) should be true")
		}
	})
	
	t.Run("string", func(t *testing.T) {
		if fn.IsNonzero("") {
			t.Error("IsNonzero(\"\") should be false")
		}
		if !fn.IsNonzero("test") {
			t.Error("IsNonzero(\"test\") should be true")
		}
	})
}

func TestApply(t *testing.T) {
	// Test Apply with nil sequence
	result := fn.Apply(nil, func(x int) (int, bool) { return x * 2, true })
	if result != nil {
		t.Error("Apply(nil, f) should return nil")
	}
	
	// Test Apply with sequence and transformation
	seq := slices.Values([]int{1, 2, 3, 4, 5})
	double := func(x int) (int, bool) { return x * 2, true }
	result = fn.Apply(seq, double)
	
	var collected []int
	for v := range result {
		collected = append(collected, v)
	}
	
	expected := []int{2, 4, 6, 8, 10}
	if !slices.Equal(collected, expected) {
		t.Errorf("Apply result = %v, want %v", collected, expected)
	}
	
	// Test Apply with filtering transformation
	evenOnly := func(x int) (int, bool) { return x, x%2 == 0 }
	result = fn.Apply(slices.Values([]int{1, 2, 3, 4, 5}), evenOnly)
	
	collected = nil
	for v := range result {
		collected = append(collected, v)
	}
	
	expected = []int{2, 4}
	if !slices.Equal(collected, expected) {
		t.Errorf("Apply with filter result = %v, want %v", collected, expected)
	}
}

func TestMap(t *testing.T) {
	// Test Map with nil sequence
	result := fn.Map(nil, func(x int) (string, bool) { return "", true })
	if result != nil {
		t.Error("Map(nil, f) should return nil")
	}
	
	// Test Map with nil function
	seq := slices.Values([]int{1, 2, 3})
	result2 := fn.Map(seq, func(int) (int, bool) { return 0, false })
	
	var collected []int
	for v := range result2 {
		collected = append(collected, v)
	}
	
	if len(collected) != 0 {
		t.Errorf("Map with false predicate should yield no items, got %v", collected)
	}
	
	// Test Map with type transformation
	intToString := func(x int) (string, bool) {
		if x > 0 {
			return string(rune('a' + x - 1)), true
		}
		return "", false
	}
	
	result3 := fn.Map(slices.Values([]int{1, 2, 3}), intToString)
	var stringResults []string
	for v := range result3 {
		stringResults = append(stringResults, v)
	}
	
	expected := []string{"a", "b", "c"}
	if !slices.Equal(stringResults, expected) {
		t.Errorf("Map type transformation result = %v, want %v", stringResults, expected)
	}
}

func TestMapItems(t *testing.T) {
	// Test MapItems with nil slice
	result := fn.MapItems(nil, func(x int) (string, bool) { return "", true })
	if result != nil {
		t.Error("MapItems(nil, f) should return nil")
	}
	
	// Test MapItems with transformation
	input := []int{1, 2, 3, 4, 5}
	double := func(x int) (int, bool) { return x * 2, true }
	result2 := fn.MapItems(input, double)
	
	expected := []int{2, 4, 6, 8, 10}
	if !slices.Equal(result2, expected) {
		t.Errorf("MapItems result = %v, want %v", result2, expected)
	}
	
	// Test MapItems with filtering
	evenOnly := func(x int) (int, bool) { return x, x%2 == 0 }
	result3 := fn.MapItems([]int{1, 2, 3, 4, 5}, evenOnly)
	
	expected2 := []int{2, 4}
	if !slices.Equal(result3, expected2) {
		t.Errorf("MapItems with filter result = %v, want %v", result3, expected2)
	}
}

func TestFilter(t *testing.T) {
	// Test Filter with nil sequence
	result := fn.Filter(nil, func(x int) bool { return true })
	if result != nil {
		t.Error("Filter(nil, f) should return nil")
	}
	
	// Test Filter with nil predicate
	seq := slices.Values([]int{1, 2, 3})
	result2 := fn.Filter(seq, nil)
	
	var collected []int
	for v := range result2 {
		collected = append(collected, v)
	}
	
	expected := []int{1, 2, 3}
	if !slices.Equal(collected, expected) {
		t.Errorf("Filter(seq, nil) should return original sequence, got %v, want %v", collected, expected)
	}
	
	// Test Filter with predicate
	isEven := func(x int) bool { return x%2 == 0 }
	result3 := fn.Filter(slices.Values([]int{1, 2, 3, 4, 5}), isEven)
	
	collected = nil
	for v := range result3 {
		collected = append(collected, v)
	}
	
	expected2 := []int{2, 4}
	if !slices.Equal(collected, expected2) {
		t.Errorf("Filter result = %v, want %v", collected, expected2)
	}
}

func TestFilterItems(t *testing.T) {
	// Test FilterItems with nil slice
	result := fn.FilterItems(nil, func(x int) bool { return true })
	if result != nil {
		t.Error("FilterItems(nil, f) should return nil")
	}
	
	// Test FilterItems with nil predicate
	input := []int{1, 2, 3}
	result2 := fn.FilterItems(input, nil)
	
	if !slices.Equal(result2, input) {
		t.Errorf("FilterItems(slice, nil) should return original slice, got %v, want %v", result2, input)
	}
	
	// Test FilterItems with predicate
	isOdd := func(x int) bool { return x%2 == 1 }
	result3 := fn.FilterItems([]int{1, 2, 3, 4, 5}, isOdd)
	
	expected := []int{1, 3, 5}
	if !slices.Equal(result3, expected) {
		t.Errorf("FilterItems result = %v, want %v", result3, expected)
	}
}

func TestFilterKeys(t *testing.T) {
	// Create a simple key-value sequence
	kvMap := map[string]int{"a": 1, "b": 2, "c": 3}
	seq := func(yield func(string, int) bool) {
		for k, v := range kvMap {
			if !yield(k, v) {
				return
			}
		}
	}
	
	// Test FilterKeys with nil sequence  
	var nilSeq func(func(string, int) bool)
	result := fn.FilterKeys(nilSeq, func(k string) bool { return true })
	if result != nil {
		t.Error("FilterKeys(nil, f) should return nil")
	}
	
	// Test FilterKeys with nil predicate
	result2 := fn.FilterKeys(seq, nil)
	collected := make(map[string]int)
	for k, v := range result2 {
		collected[k] = v
	}
	
	// Should collect all items
	if len(collected) != len(kvMap) {
		t.Errorf("FilterKeys(seq, nil) should return all items, got %d items, want %d", len(collected), len(kvMap))
	}
	
	// Test FilterKeys with predicate
	startsWithA := func(k string) bool { return k == "a" }
	result3 := fn.FilterKeys(seq, startsWithA)
	
	collected = make(map[string]int)
	for k, v := range result3 {
		collected[k] = v
	}
	
	if len(collected) != 1 {
		t.Errorf("FilterKeys should return 1 item, got %d", len(collected))
	}
	
	if val, exists := collected["a"]; !exists || val != 1 {
		t.Errorf("FilterKeys should contain 'a': 1, got %v", collected)
	}
}

func TestUniqueHas(t *testing.T) {
	u := make(fn.Unique[string])
	
	// Test empty set
	if u.Has("test") {
		t.Error("Empty Unique should not contain 'test'")
	}
	
	// Add item and test
	u.Add("test")
	if !u.Has("test") {
		t.Error("Unique should contain 'test' after adding")
	}
	
	if u.Has("other") {
		t.Error("Unique should not contain 'other'")
	}
}

func TestUniqueAdd(t *testing.T) {
	u := make(fn.Unique[int])
	
	// Add items
	u.Add(1)
	u.Add(2)
	u.Add(1) // duplicate
	
	// Check all items exist
	if !u.Has(1) {
		t.Error("Unique should contain 1")
	}
	if !u.Has(2) {
		t.Error("Unique should contain 2")
	}
	
	// Check length (Go maps with struct{} values)
	if len(u) != 2 {
		t.Errorf("Unique should have 2 items, got %d", len(u))
	}
}

func TestUniqueSet(t *testing.T) {
	u := make(fn.Unique[string])
	
	// Test setting new item
	added := u.Set("new")
	if !added {
		t.Error("Set should return true for new item")
	}
	
	if !u.Has("new") {
		t.Error("Unique should contain 'new' after Set")
	}
	
	// Test setting existing item
	added = u.Set("new")
	if added {
		t.Error("Set should return false for existing item")
	}
	
	// Length should still be 1
	if len(u) != 1 {
		t.Errorf("Unique should have 1 item, got %d", len(u))
	}
}