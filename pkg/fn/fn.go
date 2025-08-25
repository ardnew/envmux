package fn

import "iter"

// OK returns its first argument unchanged.
//
// It is useful when composing a function expecting an argument of type T
// with a function returning multiple values.
//
//nolint:ireturn
func OK[T, R any](v T, _ ...R) T { return v }

// Apply returns a sequence that yields in-order elements of s transformed by f.
// The types of input and output sequence elements are the same, T.
//
// If f is nil, the identity function is used to yield the original sequence.
// If s is nil, nil is returned.
func Apply[T any](
	s iter.Seq[T],
	f func(T) (T, bool),
) iter.Seq[T] {
	return Map(s, f)
}

// Map returns a sequence that yields in-order elements of s transformed by f.
// The types of input and output sequence elements may differ, T and R.
//
// If f is nil, the identity function is used to yield the original sequence.
// If s is nil, nil is returned.
func Map[T, R any](s iter.Seq[T], f func(T) (R, bool)) iter.Seq[R] {
	if s == nil {
		return nil
	}

	if f == nil {
		f = func(x T) (R, bool) { return any(x).(R), true } //nolint:forcetypeassert
	}

	return func(yield func(R) bool) {
		for item := range s {
			item, keep := f(item)
			if keep && !yield(item) {
				return
			}
		}
	}
}

// Filter returns a sequence that yields in-order elements of s
// that satisfy the predicate keep.
// If s is nil, nil is returned.
// If keep is nil, all elements are yielded.
//
// See [FilterItems] for the analogue that operates on slices.
func Filter[T any](s iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	if s == nil {
		return nil
	}

	if keep == nil {
		keep = func(T) bool { return true }
	}

	return func(yield func(T) bool) {
		for item := range s {
			if keep(item) && !yield(item) {
				return
			}
		}
	}
}

// FilterItems returns a sequence that yields in-order elements of s
// that satisfy the predicate keep.
// If s is nil, nil is returned.
// If keep is nil, all elements are yielded.
//
// See [Filter] for the analogue that operates on sequences.
func FilterItems[T any](s []T, keep func(T) bool) iter.Seq[T] {
	if s == nil {
		return nil
	}

	if keep == nil {
		keep = func(T) bool { return true }
	}

	return func(yield func(T) bool) {
		for _, item := range s {
			if keep(item) && !yield(item) {
				return
			}
		}
	}
}

// FilterKeys returns a sequence that yields in-order key-value pairs of s
// for which the key satisfies the predicate keep.
// If s is nil, nil is returned.
// If keep is nil, all key-value pairs are yielded.
func FilterKeys[K comparable, V any](
	s iter.Seq2[K, V],
	keep func(K) bool,
) iter.Seq2[K, V] {
	if s == nil {
		return nil
	}

	if keep == nil {
		keep = func(_ K) bool { return true }
	}

	return func(yield func(K, V) bool) {
		for key, val := range s {
			if keep(key) && !yield(key, val) {
				return
			}
		}
	}
}

// Unique is a set of unique values of comparable type T.
// It is implemented as a map from T to an empty struct,
// since the empty struct is zero-sized and requires no memory.
//
// The zero value of Unique is an empty set and is safe to use.
//
// Test for set membership with [Unique.Has].
// Use [Unique.Set] to test for membership before also adding an element.
type Unique[T comparable] map[T]struct{}

// Has returns whether the receiver contains the given value.
func (u Unique[T]) Has(v T) bool {
	_, ok := u[v]

	return ok
}

// Add adds the given value to the receiver.
//
// Use [Unique.Set] to determine whether the value was already present.
func (u Unique[T]) Add(v T) {
	u[v] = struct{}{}
}

// Set adds the given value to the receiver if it is not already present
// and returns whether the value was added.
//
// Use [Unique.Add] to add the value unconditionally.
func (u Unique[T]) Set(v T) bool {
	if u.Has(v) {
		return false
	}

	u[v] = struct{}{}

	return true
}
