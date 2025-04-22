package pkg

import (
	"errors"
)

var (
	// ErrInvalidConfig indicates that the configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrInvalidModel indicates that the model is invalid.
	ErrInvalidModel = errors.New("invalid model")
	// ErrInvalidOption indicates that the option is invalid.
	ErrInvalidOption = errors.New("invalid option")
	// ErrInvalidNamespace indicates that the namespace is invalid.
	ErrInvalidNamespace = errors.New("invalid namespace")
	// ErrInvalidEnvVarMap indicates that the variable mapping is invalid.
	ErrInvalidEnvVarMap = errors.New("invalid variable mapping")
	// ErrInvalidEnvVar indicates that the environment is invalid.
	ErrInvalidEnvVar = errors.New("invalid environment variable")
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
