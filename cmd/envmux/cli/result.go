package cli

import (
	"errors"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/pkg/fn"
)

// RunError represents the result of executing a command.
type RunError struct {
	Err  error
	Help string
	Code int
}

// ErrRunOK is the default successful result.
var ErrRunOK = RunError{} //nolint:exhaustruct

func (r RunError) Error() string {
	if r.Err != nil {
		return r.Err.Error()
	}

	return ""
}

// resultErr sets the error and code for the Result.
func resultErr(err error, code int) fn.Option[RunError] {
	return func(r RunError) RunError {
		r.Err = err
		r.Code = code

		return r
	}
}

// resultHelp sets the help message for the Result.
func resultHelp(help string) fn.Option[RunError] {
	return func(r RunError) RunError {
		r.Help = help

		return r
	}
}

// MakeResult creates a Result based on the given command and error.
func MakeResult(node cmd.Node, err error) RunError {
	resultUsage := resultHelp(ffhelp.Command(node.Command()).String())

	switch {
	case err == nil:
		return ErrRunOK
	case errors.Is(err, ff.ErrHelp):
		return fn.Make(resultUsage)
	case errors.Is(err, ff.ErrNoExec):
		return fn.Make(resultUsage)
	default:
		return fn.Make(resultErr(err, 1))
	}
}
