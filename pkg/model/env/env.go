package env

import (
	"context"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
)

const (
	ID        = "env"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "environment variables"
	longHelp  = `environment variables are user-definable key-value pairs ` +
		`that are accessible to a process and its children.`
)

type Command struct {
	model.Command
}

func (Command) Name() string               { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }
func (Command) Exec(context.Context, []string) error {
	// _, err := fmt.Fprintf(c.Stdout, "[%s] arg=%+v\n", ID, arg)
	return nil
}

func Make(opts ...pkg.Option[Command]) (cfg Command) {
	return pkg.WithOptions(cfg,
		append(
			[]pkg.Option[Command]{WithDefaults()},
			opts...,
		)...,
	)
}

func WithDefaults() pkg.Option[Command] {
	return func(c Command) Command {
		return pkg.WithOptions(c,
			WithModel(model.WithInterface(c)),
		)
	}
}

func WithModel(opts ...pkg.Option[model.Command]) pkg.Option[Command] {
	return func(c Command) Command {
		c.Command = pkg.WithOptions(c.Command, opts...)
		return c
	}
}
