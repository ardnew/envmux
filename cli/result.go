package cli

import (
	"errors"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

// Result represents the result of executing a command.
type Result struct {
	Err  error
	Help string
	Code int
}

// ResultOK is the default successful result.
var ResultOK = Result{}

// resultErr sets the error and code for the Result.
func resultErr(err error, code int) pkg.Option[Result] {
	return func(r Result) Result {
		r.Err = err
		r.Code = code
		return r
	}
}

// resultHelp sets the help message for the Result.
func resultHelp(help string) pkg.Option[Result] {
	return func(r Result) Result {
		r.Help = help
		return r
	}
}

// MakeResult creates a Result based on the given command and error.
func MakeResult(mod model.Command, err error) Result {
	resultUsage := resultHelp(ffhelp.Command(mod.Config().Command).String())
	switch {
	case err == nil:
		return ResultOK
	case errors.Is(err, ff.ErrHelp):
		return pkg.WithOptions(Result{}, resultUsage)
	case errors.Is(err, ff.ErrNoExec):
		return pkg.WithOptions(Result{}, resultUsage)
	default:
		return pkg.WithOptions(Result{}, resultUsage, resultErr(err, 1))
	}
}
