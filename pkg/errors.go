package pkg

import (
	"errors"
	"log/slog"
	"maps"
	"strings"
	"unicode/utf8"

	"github.com/ardnew/envmux/pkg/fn"
)

// Error represents a chain of errors.
type Error struct{ err []error }

// MakeError constructs an Error from non-empty messages.
func MakeError(msgs ...string) Error {
	return Make(WithErrorMessage(msgs...))
}

// WithError returns an option that appends non-nil errors to an Error chain.
func WithError(errs ...error) Option[Error] {
	return func(e Error) Error {
		e.err = append(e.err, fn.FilterItems(errs, fn.IsNonZero)...)

		return e
	}
}

// WithErrorMessage returns an option that appends non-empty messages as errors
// to an Error chain.
func WithErrorMessage(msgs ...string) Option[Error] {
	return WithError(fn.MapItems(msgs, func(m string) (error, bool) {
		if m == "" {
			return nil, false
		}

		return errors.New(m), true
	})...)
}

// Wrap appends additional non-nil errors to the chain and returns a new Error.
func (e Error) Wrap(errs ...error) Error {
	return Wrap(e, WithError(errs...))
}

// WrapMessage appends additional non-empty messages as errors to the chain and
// returns a new Error.
func (e Error) WrapMessage(msgs ...string) Error {
	return Wrap(e, WithErrorMessage(msgs...))
}

// Unwrap returns the chain of errors.
func (e Error) Unwrap() []error {
	return e.err
}

// Error returns the chain of error messages separated by a colon.
func (e Error) Error() string {
	if len(e.err) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, err := range fn.FilterItems(e.err, fn.IsNonZero) {
		if i > 0 {
			sb.WriteString(": ")
		}

		sb.WriteString(err.Error())
	}

	return sb.String()
}

// Attributed is implemented by errors that expose structured attributes for
// logging and presentation. Implementations should return a map of key-value
// pairs via Attr, a key name used for multi-line details via DetailKey, and a
// set of formatted detail lines via Details.
type Attributed interface {
	Attr() map[string]any
	DetailKey() string
	Details() []string
}

// Attributes converts the structured fields of an [Attributed] value into a
// slice of [slog.Attr] suitable for structured logging with [slog]. The entry
// whose key equals [Attributed.DetailKey] is omitted, since the details are
// expected to be handled line-by-line via [Attributed.Details].
func Attributes(attr Attributed) []slog.Attr {
	a := fn.FilterKeys(maps.All(attr.Attr()), func(k string) bool {
		return k != attr.DetailKey()
	})

	s := []slog.Attr{}
	for key, value := range a {
		s = append(s, slog.Attr{Key: key, Value: slog.AnyValue(value)})
	}

	return s
}

var (
	// ErrUndefCommandExec indicates that the command exec function is undefined.
	ErrUndefCommandExec = MakeError("undefined exec function")
	// ErrUndefCommandFlagSet indicates that the command flag set is undefined.
	ErrUndefCommandFlagSet = MakeError("undefined flag set")
	// ErrUndefCommandUsage indicates that the command name or usage is undefined.
	ErrUndefCommandUsage = MakeError("undefined name or usage")

	// ErrInaccessibleManifest indicates that the manifest cannot be accessed.
	ErrInaccessibleManifest = MakeError("inaccessible manifest")
	// ErrUndefinedNamespace indicates that the namespace is undefined.
	ErrUndefinedNamespace = MakeError("undefined namespace")
	// ErrInvalidIdentifier indicates that the identifier is invalid.
	ErrInvalidIdentifier = MakeError("invalid identifier")

	// ErrInvalidJSON indicates that the JSON encoding is invalid.
	ErrInvalidJSON = MakeError("invalid JSON encoding")
)

// manifestErrorContext captures a source excerpt and position information used
// to enrich parse and evaluation errors originating from manifest content.
type manifestErrorContext struct {
	Source string
	Marker string
	Line   int
	Column int
}

const (
	markerSymbol = "↑"
	markerLeader = "…"
)

// makeManifestErrorContext computes line/column information and a visual
// marker for a byte offset into source. It extracts the line of text
// containing the offset and generates a caret-like marker aligned to the
// column.
func makeManifestErrorContext(source string, offset int) manifestErrorContext {
	var c manifestErrorContext

	var begin, end int

	for i, r := range source[:min(len(source), offset)] {
		c.Column++
		if r == '\n' {
			c.Line++
			c.Column = 0
			begin = i + 1
		}
	}

	c.Marker = makeMarker(c.Column)
	end = begin + strings.IndexByte(source[begin:], '\n')

	switch {
	case begin > end:
		end = len(source)
	case begin == end:
		return c
	}

	c.Source = source[begin:end]

	return c
}

// makeMarker returns a fixed-width ASCII marker string ending with markerSymbol
// and padded on the left by repetitions of markerLeader to align under the
// specified column. Column is zero-based.
func makeMarker(column int) string {
	// Build a fixed-width (in runes) marker that ends with markerSymbol.
	// markerLeader fills as much as possible before markerSymbol.
	width := column + 1
	if width <= 0 {
		return markerSymbol
	}

	var sb strings.Builder

	n := max(width-max(utf8.RuneCountInString(markerSymbol), 1), 0)

	sb.Grow(n*len(markerLeader) + len(markerSymbol))

	for range n {
		sb.WriteString(markerLeader)
	}

	sb.WriteString(markerSymbol)

	return sb.String()
}

// Attr implements [Attributed] by returning a map of fields suitable for
// structured logging, including a nested value under DetailKey with "source"
// and "marker", and top-level 1-based line and column numbers.
func (c manifestErrorContext) Attr() map[string]any {
	return map[string]any{
		c.DetailKey(): map[string]any{
			"source": c.Source,
			"marker": c.Marker,
		},
		"line":   c.Line + 1,
		"column": c.Column + 1,
	}
}

// DetailKey implements [Attributed] and returns the attribute key under which
// multi-line details are grouped in Attr.
func (c manifestErrorContext) DetailKey() string {
	return "detail"
}

// Details implements [Attributed] and returns a boxed, multi-line rendering of
// the source excerpt and marker suitable for human-readable logs.
func (c manifestErrorContext) Details() []string {
	// Return the raw source line and the aligned marker. When printed with a
	// newline between them, the marker points to the problematic rune.
	return []string{c.Source, c.Marker}
}

// ParseError represents an error that occurred while parsing a manifest.
type ParseError struct {
	manifestErrorContext
}

// MakeParseError constructs a [ParseError] with contextual information derived
// from source at the given byte offset.
func MakeParseError(source string, offset int) Error {
	return Make(WithError(ParseError{
		manifestErrorContext: makeManifestErrorContext(source, offset),
	}))
}

// Error implements the error interface.
func (e ParseError) Error() string {
	return "failed to parse manifest"
}

// EvalError represents an error that occurred while evaluating an expression
// in a specific namespace and identifier.
type EvalError struct {
	manifestErrorContext

	Namespace string
	Ident     string
}

// MakeEvalError constructs an [EvalError] for the specified namespace,
// identifier, and location in the expression source.
func MakeEvalError(namespace, ident, source string, offset int) Error {
	return Make(WithError(EvalError{
		manifestErrorContext: makeManifestErrorContext(source, offset),
		Namespace:            namespace,
		Ident:                ident,
	}))
}

// Error implements the error interface.
func (e EvalError) Error() string {
	return "failed to evaluate expression"
}

// Attr returns structured attributes for the evaluation error, including the
// namespace and identifier where the error occurred.
func (e EvalError) Attr() map[string]any {
	a := e.manifestErrorContext.Attr()
	a["namespace"] = e.Namespace
	a["ident"] = e.Ident

	return a
}
