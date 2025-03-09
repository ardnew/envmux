package pkg

// Option functions return their argument with modifications applied.
type Option[T any] func(T) T

// Make returns a new object of type T with the given options applied.
func Make[T any](opts ...Option[T]) (t T) { return Wrap(t, opts...) }

// Wrap returns t after applying the given options.
func Wrap[T any](t T, opts ...Option[T]) T {
	for _, o := range opts {
		t = o(t)
	}
	return t
}
