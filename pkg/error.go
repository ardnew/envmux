package pkg

import (
	"strings"
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
)

// ErrInvalidConfigFile indicates that the configuration file is invalid.
// ErrInvalidConfigFile = Error{"invalid configuration file"}
// ErrInvalidFlag indicates that the flag is invalid.
// ErrInvalidFlag = Error{"invalid flag"}
// ErrInvalidComposition indicates that the composition is invalid.
// ErrInvalidComposition = Error{"invalid composition"}
// ErrInvalidParameter indicates that the subject is invalid.
// ErrInvalidParameter = Error{"invalid subject"}
// ErrInvalidIdent indicates that an identifier is invalid.
// ErrInvalidIdent = Error{"invalid identifier"}
// ErrIncompleteInit indicates that the initialization is incomplete.
// ErrIncompleteInit = Error{"incomplete initialization"}
