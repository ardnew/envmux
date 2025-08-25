// Package errs provides error types and helpers for envmux.
package pkg

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/expr-lang/expr/file"

	"github.com/ardnew/envmux/manifest/config"
)

// Error represents an error with an embedded message.
type Error struct{ string }

func JoinErrors(err ...error) error {
	if len(err) == 0 {
		return nil
	}

	var msg strings.Builder

	for i, e := range err {
		if e != nil {
			if i > 0 {
				msg.WriteString(": ")
			}

			msg.WriteString(e.Error())
		}
	}

	return Error{msg.String()}
}

// Error returns the error message.
func (e Error) Error() string {
	if e.string == "" {
		return "<Error>"
	}

	return e.string
}

var (
	// ErrInvalidConfig indicates that the configuration is invalid.
	ErrInvalidConfig = Error{"invalid configuration"}

	// ErrInvalidCommand indicates that the command is invalid.
	ErrInvalidCommand = Error{"invalid command"}
	// ErrInvalidFlagSet indicates that the flag set is invalid.
	ErrInvalidFlagSet = Error{"invalid flag set"}
	// ErrInvalidInterface indicates that the interface is invalid.
	ErrInvalidInterface = Error{"invalid interface"}
	// ErrInvalidModel indicates that the model is invalid.
	ErrInvalidModel = Error{"invalid model"}

	// ErrInvalidDefinitions indicates that the definitions file is invalid.
	ErrInvalidDefinitions = Error{"invalid namespace definitions"}
	// ErrInvalidNamespace indicates that the namespace is invalid.
	ErrInvalidNamespace = Error{"invalid namespace"}
	// ErrInvalidComposite indicates that the composite is invalid.
	ErrInvalidComposite = Error{"invalid composite"}
	// ErrInvalidParameter indicates that the parameter is invalid.
	ErrInvalidParameter = Error{"invalid parameter"}
	// ErrInvalidStatement indicates that the statement is invalid.
	ErrInvalidStatement = Error{"invalid statement"}
	// ErrInvalidEnvVar indicates that the environment is invalid.
	ErrInvalidEnvVar = Error{"invalid environment variable"}
	// ErrInvalidExpression indicates that an expression is invalid.
	ErrInvalidExpression = Error{"invalid expression"}

	// ErrInvalidJSON indicates that the JSON encoding is invalid.
	ErrInvalidJSON = Error{"invalid JSON encoding"}

	// ErrIncompleteParse indicates that the parse is incomplete.
	ErrIncompleteParse = Error{"incomplete parse"}
	// ErrIncompleteEval indicates that the evaluation is incomplete.
	ErrIncompleteEval = Error{"incomplete evaluation"}

	// ErrClosedStream indicates that the stream is closed.
	ErrClosedStream = Error{"closed stream"}
	// ErrUnacceptableStream indicates that the stream is unacceptable.
	ErrUnacceptableStream = Error{"unacceptable stream"}

	// ErrUnexpectedToken indicates that an unexpected token was encountered.
	ErrUnexpectedToken = Error{"unexpected token"}

	// ErrUnexpectedEOF indicates that an unexpected EOF was encountered.
	ErrUnexpectedEOF = Error{"unexpected end of file"}
	// ErrEOF indicates that the end of file was reached.
	ErrEOF = Error{"end of file"}
)

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

	var (
		n *NamespaceError
		x *ExpressionError
		u *UnexpectedTokenError
	)

	var msg, pos string

	switch {
	case errors.As(e.Err, &x):
		if e.Lvl > 0 {
			pos = fmt.Sprintf(" at %s%s", ref, x.position())
		}

		msg = fmt.Sprintf("%s: %v", pos, x)

	case errors.As(e.Err, &n):
		if e.Lvl > 0 {
			pos = fmt.Sprintf(" at %s%s", ref, n.position())
		}

		msg = fmt.Sprintf("%s: %v", pos, n)

	case errors.As(e.Err, &u):
		if e.Lvl > 0 {
			pos = fmt.Sprintf(" at %s[%d:%d]", ref, u.Line, u.Column)
		}

		msg = fmt.Sprintf("%s: %v", pos, u)

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

type NamespaceError struct {
	ID  string
	Err error
}

func (e NamespaceError) Error() string {
	var id string
	if e.ID != "" {
		id = fmt.Sprintf(" (in namespace %q)", e.ID)
	}

	ee, ue := e.Err, new(UnexpectedTokenError)
	if errors.As(e.Err, &ue) {
		ee = ue
	}

	return fmt.Sprintf("%v%s", ee, id)
}

func (e NamespaceError) position() string {
	ue := new(UnexpectedTokenError)
	if errors.As(e.Err, &ue) {
		return fmt.Sprintf("[%d:%d]", ue.Line, ue.Column)
	}

	return ""
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

	ee, ue := e.Err, new(file.Error)
	if errors.As(e.Err, &ue) {
		ee = fmt.Errorf("%w: %s", ErrInvalidExpression, ue.Message)
		ap = "\t" + strings.ReplaceAll(ue.Snippet, "\n", "\n\t")
	}

	return fmt.Sprintf("%v%s%s", ee, id, ap)
}

func (e ExpressionError) position() string {
	ue := new(file.Error)
	if errors.As(e.Err, &ue) {
		return fmt.Sprintf("[%d:%d]", ue.Line, ue.Column)
	}

	return ""
}

type UnexpectedTokenError struct {
	Offset       int
	Line, Column int
	Tok          string
	Msg          []string
}

func (e UnexpectedTokenError) Error() string {
	var sb strings.Builder

	switch {
	case e.Tok != "":
		sb.WriteRune(' ')
		sb.WriteString(strconv.Quote(e.Tok))
	}

	for _, n := range e.Msg {
		sb.WriteString(": ")
		sb.WriteString(n)
	}

	return fmt.Sprintf("%v%s", ErrUnexpectedToken, sb.String())
}
