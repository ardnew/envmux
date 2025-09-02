// Package errs provides error types and helpers for envmux.
package pkg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/expr-lang/expr/file"

	"github.com/ardnew/envmux/manifest/config"
	"github.com/ardnew/envmux/pkg/fn"
)

// Error represents an error with an embedded message.
type Error struct{ string }

// JoinErrors returns a new error that concatenates the messages of the
// non-nil errors in err with ": " as the separator.
// If err is empty or contains only nil values, nil is returned.
func JoinErrors(err ...error) error {
	if len(err) == 0 {
		return nil
	}

	return Error{strings.Join(fn.MapItems(err, getMessage), ": ")}
}

// WithDetail returns a new error that wraps the receiver
// and appends detail to its error message.
func (e Error) WithDetail(str ...string) error {
	return JoinErrors(append([]error{e}, fn.MapItems(str, putMessage)...)...)
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

	// ErrInvalidIdentifier indicates that the identifier is invalid.
	ErrInvalidIdentifier = Error{"invalid identifier"}
	// ErrInvalidExpression indicates that an expression is invalid.
	ErrInvalidExpression = Error{"invalid expression"}

	// ErrInvalidJSON indicates that the JSON encoding is invalid.
	ErrInvalidJSON = Error{"invalid JSON encoding"}

	// ErrIncompleteParse indicates that the parse is incomplete.
	ErrIncompleteParse = Error{"incomplete parse"}
	// ErrIncompleteEval indicates that the evaluation is incomplete.
	ErrIncompleteEval = Error{"incomplete evaluation"}

	// ErrUnexpectedToken indicates that an unexpected token was encountered.
	ErrUnexpectedToken = Error{"unexpected token"}
)

// GetMessage returns the non-nil [error]'s non-empty [error.Error] and true.
//
// If err is nil or its error empty, the empty string and false are returned.
func getMessage(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	s := err.Error()

	return s, s != ""
}

// PutMessage returns a new [Error] with the given non-empty error and true.
//
// If the error is empty, the [Error] zero value and false are returned.
func putMessage(err string) (error, bool) { //nolint:revive
	return Error{err}, err != ""
}

type IncompleteParseError struct {
	Err error
	Def []string
	Lvl int // Verbose level for error reporting
}

func (e IncompleteParseError) Error() string {
	var def strings.Builder

	for i, s := range e.Def {
		if i > 0 {
			def.WriteString(",")
		}

		if s = strings.TrimSpace(s); s == "" {
			continue // skip empty definitions
		}

		if s == config.StdinManifestPath {
			s = "STDIN"
		}

		def.WriteString(s)
	}

	ref := def.String()

	var x *ExpressionError

	var msg, pos string

	switch {
	case errors.As(e.Err, &x):
		if e.Lvl > 0 {
			pos = fmt.Sprintf(" at %s%s", ref, x.position())
		}

		msg = fmt.Sprintf("%s: %v", pos, x)

	case def.Len() > 0:
		if e.Lvl > 0 {
			pos = " at " + ref
		}

		msg = fmt.Sprintf("%s: %v", pos, e.Err)

	default:
		msg = fmt.Sprintf(": %v", e.Err)
	}

	return fmt.Sprintf("%v%s", ErrIncompleteParse, msg)
}

type ExpressionError struct {
	Namespace string
	Statement string
	Err       error
}

func (e ExpressionError) Error() string {
	var id, ap string
	if e.Namespace != "" {
		id = fmt.Sprintf(
			" (expression %q in namespace %q)",
			e.Statement,
			e.Namespace,
		)
	}

	ee, fe := e.Err, new(file.Error)
	if errors.As(e.Err, &fe) {
		ee = fmt.Errorf("%w: %s", ErrInvalidExpression, fe.Message)
		ap = "\t" + strings.ReplaceAll(fe.Snippet, "\n", "\n\t")
	}

	return fmt.Sprintf("%v%s%s", ee, id, ap)
}

func (e ExpressionError) position() string {
	fe := new(file.Error)
	if errors.As(e.Err, &fe) {
		return fmt.Sprintf("[%d:%d]", fe.Line, fe.Column)
	}

	return ""
}
