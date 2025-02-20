package cli

import (
	"errors"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

type Result struct {
	Err  error
	Help string
	Code int
}

var ResultOK = Result{}

func resultErr(err error, code int) pkg.Option[Result] {
	return func(r Result) Result {
		r.Err = err
		r.Code = code
		return r
	}
}

func resultHelp(help string) pkg.Option[Result] {
	return func(r Result) Result {
		r.Help = help
		return r
	}
}

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
