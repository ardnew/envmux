package pkg

// Option functions return their argument with modifications applied.
type Option[T any] func(T) T

// WithOptions unwraps and returns the receiver's object
// after applying the given options to it.
// nolint: ireturn
func WithOptions[T any](t T, opts ...Option[T]) T {
	for _, o := range opts {
		t = o(t)
	}
	return t
}
