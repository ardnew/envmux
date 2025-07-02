package pkg

import (
	"bufio"
	"io"
	"iter"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// OK returns its first argument unchanged.
//
// It is useful when composing a function expecting an argument of type T
// with a function returning multiple values.
//
//nolint:ireturn
func OK[T, R any](v T, _ ...R) T { return v }

// ReaderFromFile returns a buffered reader from the given file name.
func ReaderFromFile(filename string) (io.Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return bufio.NewReader(f), nil
}

// EnvVarOption configures environment variable identifier formatting.
type EnvVarOption struct {
	Case    cases.Caser // Letter case transformation.
	Break   []byte      // String to insert between runs.
	Unicode bool        // Accept Unicode code points as valid glyphs.
}

// DefaultEnvVarOption is the default formatting for variable identifiers.
//
// It replaces all runs of invalid glyphs with a single underscore and converts
// all valid glyphs to uppercase.
//
//nolint:gochecknoglobals
var DefaultEnvVarOption = EnvVarOption{
	Case:    cases.Upper(language.Und), // Use all-uppercase glyphs.
	Break:   []byte{'_'},               // Separate valid runs with an underscore.
	Unicode: false,                     // Do not accept Unicode (ASCII-only).
}

// FormatEnvVar formats a run of strings as an environment variable identifier
// using [DefaultEnvVarOption].
func FormatEnvVar(run ...string) string {
	return DefaultEnvVarOption.FormatEnvVar(run...)
}

// FormatEnvVar formats a run of strings as an environment variable identifier.
func (o EnvVarOption) FormatEnvVar(run ...string) string {
	o = o.asValid()

	var sb strings.Builder

	brk := false
	for i, s := range run {
		if i > 0 && !brk {
			brk = true

			sb.Write(o.Break)
		}

		brk = o.formatEnvVarWord(&sb, s, i == 0, brk)
	}

	return sb.String()
}

// formatEnvVarWord formats a single word as an environment variable identifier
// and appends it to the current identifier constructed so far in sb.
func (o EnvVarOption) formatEnvVarWord(
	sb *strings.Builder,
	s string,
	isFirstRun, brk bool,
) bool {
	r := []rune(strings.TrimSpace(s))
	t := []rune(o.Case.String(string(r)))

	for j := range r {
		isLetter := isASCIILetter(r[j])
		isDigit := isASCIIDigit(r[j])

		switch {
		case isFirstRun && j == 0 && isDigit:
			brk = false

			sb.Write(o.Break)
			sb.WriteRune(t[j])
		case isLetter || isDigit:
			brk = false

			sb.WriteRune(t[j])
		default:
			if !brk {
				brk = true

				sb.Write(o.Break)
			}
		}
	}

	return brk
}

func isASCIILetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func (o EnvVarOption) isValid() bool {
	return o.Case != (cases.Caser{}) && o.Break != nil
}

func (o EnvVarOption) asValid() EnvVarOption {
	if o.isValid() {
		return o
	}

	if !DefaultEnvVarOption.isValid() {
		panic(ErrInvalidEnvVar)
	}

	if o.Case == (cases.Caser{}) {
		o.Case = DefaultEnvVarOption.Case
	}

	if o.Break == nil {
		o.Break = DefaultEnvVarOption.Break
	}

	return o
}

// Map returns a sequence that yields in-order elements of s transformed by f.
// If f is nil, the identity function is used to yield the original sequence.
// If s is nil, nil is returned.
func Map[T any](s iter.Seq[T], f func(T) T) iter.Seq[T] {
	if s == nil {
		return nil
	}

	if f == nil {
		f = func(x T) T { return x }
	}

	return func(yield func(T) bool) {
		for item := range s {
			if !yield(f(item)) {
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
// Filter is to slices as [FilterKeys] is to maps.
func Filter[T any](s iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	if s == nil {
		return nil
	}

	if keep == nil {
		keep = func(_ T) bool { return true }
	}

	return func(yield func(T) bool) {
		for item := range s {
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
//
// FilterKeys is to maps as [Filter] is to slices.
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
// Use [Unique.Set] to simultaneously test for membership and add an element.
type Unique[T comparable] map[T]struct{}

// Has returns whether the receiver contains the given value.
func (u Unique[T]) Has(v T) bool {
	_, ok := u[v]

	return ok
}

// Add adds the given value to the receiver.
//
// Use [Unique.Set] to determine whether the value was added or was already
// present.
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
