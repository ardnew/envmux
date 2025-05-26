package pkg

import (
	"errors"

	trace "github.com/pkg/errors"
)

var (
	// MaxStackFrames is the max number of frames printed from errors.
	MaxStackFrames = 4
	// SkipStackFrames is the number of top-of-stack frames omitted from errors.
	SkipStackFrames = 1
)

var (
	// ErrInvalidFile indicates that the configuration file is invalid.
	ErrInvalidConfigFile = errors.New("invalid configuration file")
	// ErrInvalidConfig indicates that the configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrInvalidModel indicates that the model is invalid.
	ErrInvalidModel = errors.New("invalid model")
	// ErrInvalidOption indicates that the option is invalid.
	ErrInvalidOption = errors.New("invalid option")
	// ErrInvalidNamespaceFile indicates that the namespace file is invalid.
	ErrInvalidNamespaceFile = errors.New("invalid namespace file")
	// ErrInvalidNamespace indicates that the namespace is invalid.
	ErrInvalidNamespace = errors.New("invalid namespace")
	// ErrInvalidEnvVar indicates that the environment is invalid.
	ErrInvalidEnvVar = errors.New("invalid environment variable")
	// ErrInvalidFlag indicates that the flag is invalid.
	ErrInvalidFlag = errors.New("invalid flag")
	// ErrInvalidComposition indicates that the composition is invalid.
	ErrInvalidComposition = errors.New("invalid composition")
	// ErrInvalidSubject indicates that the subject is invalid.
	ErrInvalidSubject = errors.New("invalid subject")
	// ErrInvalidIdent indicates that an identifier is invalid.
	ErrInvalidIdent = errors.New("invalid identifier")
	// ErrInvalidExpression indicates that an expression is invalid.
	ErrInvalidExpression = errors.New("invalid expression")

	// ErrInvalidJSON indicates that the JSON encoding is invalid.
	ErrInvalidJSON = errors.New("invalid JSON encoding")

	// ErrIncompleteParse indicates that the parse is incomplete.
	ErrIncompleteParse = errors.New("incomplete parse")
	// ErrIncompleteEval indicates that the evaluation is incomplete.
	ErrIncompleteEval = errors.New("incomplete evaluation")
)

type stackTracer interface{ StackTrace() trace.StackTrace }

// type TracedError struct{ error }

// func Trace(err error, message ...string) TracedError {
// 	return TracedError{error: trace.Wrap(err, strings.Join(message, " "))}
// }

// func Tracef(err error, format string, args ...any) TracedError {
// 	return TracedError{error: trace.Wrapf(err, format, args...)}
// }

// func Error(message string) TracedError {
// 	return TracedError{error: trace.New(message)}
// }

// func Errorf(format string, args ...any) TracedError {
// 	return TracedError{error: trace.Errorf(format, args...)}
// }

// func (e TracedError) Unwrap() error { return e.error }
// func (e TracedError) Cause() error  { return trace.Cause(e.error) }

// func (e TracedError) Error() string {
// 	return fmt.Sprintf("%+v", e.StackTrace(SkipStackFrames, MaxStackFrames))
// }

// StackTrace returns a slice of count stack frames,
// starting at 0-based index offset.
//
// If offset < 0 or len(stack) <= offset or count == 0, return empty.
// If count < 0 or offset+count >= len(stack), stack[offset:] is returned.
// func (e TracedError) StackTrace(offset, count int) trace.StackTrace {
// 	if err, ok := e.error.(stackTracer); ok {
// 		t := err.StackTrace()
// 		if offset < 0 || len(t) <= offset {
// 			return nil
// 		}
// 		if count < 0 || offset+count >= len(t) {
// 			count = len(t) - offset
// 		}
// 		return t[offset : offset+count]
// 	}
// 	panic(
// 		trace.Wrapf(e, "stack trace unavailable (original error: %T)", e.Unwrap()),
// 	)
// }
