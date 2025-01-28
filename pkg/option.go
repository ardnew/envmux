package pkg

// Option functions return their argument with modifications applied.
type Option[T any] func(T) T

// With unwraps and returns the receiver's object
// after applying the given options to it.
func WithOptions[T any](s T, opts ...Option[T]) T {
	for _, opt := range opts {
		s = opt(s)
	}
	return s
}
