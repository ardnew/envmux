package cli

import (
	"errors"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/pkg"
)

// RunError represents the result of executing a command.
// A zero value indicates success.
type RunError struct {
	Err  error
	Help string
	Code int
}

// ErrRunOK is the default successful result.
// ErrRunOK is a sentinel success value.
var ErrRunOK = RunError{} //nolint:exhaustruct

// Error implements the error interface for [RunError]. It returns the wrapped
// error text, or an empty string for success.
func (r RunError) Error() string {
	if r.Err != nil {
		return r.Err.Error()
	}

	return ""
}

// resultErr sets the error and code for the Result.
func resultErr(err error, code int) pkg.Option[RunError] {
	return func(r RunError) RunError {
		r.Err = err
		r.Code = code

		return r
	}
}

// resultHelp sets the help message for the Result.
func resultHelp(help string) pkg.Option[RunError] {
	return func(r RunError) RunError {
		r.Help = help

		return r
	}
}

// MakeResult creates a [RunError] based on the given command and error.
func MakeResult(node cmd.Node, err error) RunError {
	resultUsage := resultHelp(ffhelp.Command(node.Command()).String())

	switch {
	case err == nil:
		return ErrRunOK
	case errors.Is(err, ff.ErrHelp):
		return pkg.Make(resultUsage)
	case errors.Is(err, ff.ErrNoExec):
		return pkg.Make(resultUsage)
	default:
		return pkg.Make(resultErr(err, 1))
	}
}
