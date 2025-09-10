package pkg

import (
	"log/slog"
	"maps"
	"strings"
	"unicode/utf8"

	"github.com/ardnew/envmux/pkg/fn"
)

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
// expected to be logged line-by-line via [Attributed.Details].
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

// Error represents an error with an embedded message.
type Error struct{ string }

// JoinErrors returns a new error whose message is the concatenation of the
// non-nil input errors separated by ": ". If there are no non-nil inputs,
// JoinErrors returns nil.
func JoinErrors(err ...error) error {
	if len(err) == 0 {
		return nil
	}

	err = fn.FilterItems(err, fn.IsNonzero)

	if len(err) == 0 {
		return nil
	}

	var sb strings.Builder

	sb.WriteString(err[0].Error())

	for i := 1; i < len(err); i++ {
		sb.WriteString(": ")
		sb.WriteString(err[i].Error())
	}

	return Error{sb.String()}
}

// WithDetail returns a new error that wraps the receiver and appends any
// non-empty detail strings to its message using [JoinErrors].
func (e Error) WithDetail(str ...string) error {
	set := func(msg string) (error, bool) {
		return Error{msg}, msg != ""
	}

	return JoinErrors(append([]error{e}, fn.MapItems(str, set)...)...)
}

// Error returns the error message.
func (e Error) Error() string {
	if e.string == "" {
		return "<Error>"
	}

	return e.string
}

var (
	// ErrUndefCommandExec indicates that the command exec function is undefined.
	ErrUndefCommandExec = Error{"undefined exec function"}
	// ErrUndefCommandFlagSet indicates that the command flag set is undefined.
	ErrUndefCommandFlagSet = Error{"undefined flag set"}
	// ErrUndefCommandUsage indicates that the command name or usage is undefined.
	ErrUndefCommandUsage = Error{"undefined name or usage"}

	// ErrInaccessibleManifest indicates that the manifest cannot be accessed.
	ErrInaccessibleManifest = Error{"inaccessible manifest"}
	// ErrUndefinedNamespace indicates that the namespace is undefined.
	ErrUndefinedNamespace = Error{"undefined namespace"}

	// ErrIncompleteParse indicates that the parse is incomplete.
	ErrIncompleteParse = Error{"incomplete parse"}
	// ErrIncompleteEval indicates that the evaluation is incomplete.
	ErrIncompleteEval = Error{"incomplete evaluation"}

	// ErrUnexpectedToken indicates that an unexpected token was encountered.
	ErrUnexpectedToken = Error{"unexpected token"}
	// ErrInvalidIdentifier indicates that the identifier is invalid.
	ErrInvalidIdentifier = Error{"invalid identifier"}
	// ErrInvalidExpression indicates that an expression is invalid.
	ErrInvalidExpression = Error{"invalid expression"}

	// ErrInvalidJSON indicates that the JSON encoding is invalid.
	ErrInvalidJSON = Error{"invalid JSON encoding"}
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
func MakeParseError(source string, offset int) ParseError {
	return ParseError{
		manifestErrorContext: makeManifestErrorContext(source, offset),
	}
}

// Error implements the error interface.
func (e ParseError) Error() string {
	// return fmt.Sprintf("parse error: %s", e.Source)
	return "parse error"
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
func MakeEvalError(namespace, ident, source string, offset int) EvalError {
	return EvalError{
		manifestErrorContext: makeManifestErrorContext(source, offset),
		Namespace:            namespace,
		Ident:                ident,
	}
}

// Error implements the error interface.
func (e EvalError) Error() string {
	// return fmt.Sprintf("evaluation error: %s", e.Source)
	return "evaluation error"
}

// Attr returns structured attributes for the evaluation error, including the
// namespace and identifier where the error occurred.
func (e EvalError) Attr() map[string]any {
	a := e.manifestErrorContext.Attr()
	a["namespace"] = e.Namespace
	a["ident"] = e.Ident

	return a
}
