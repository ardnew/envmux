package pkg

// Option functions return their argument with modifications applied.
type Option[T any] func(T) T

// WithOptions unwraps and returns the receiver's object
// after applying the given options to it.
func WithOptions[T any](t T, opts ...Option[T]) T {
	for _, o := range opts {
		t = o(t)
	}
	return t
}

// Make returns a new object of type T with the given options applied.
func Make[T any](opts ...Option[T]) (t T) { return WithOptions(t, opts...) }
