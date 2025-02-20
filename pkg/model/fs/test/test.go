package test

import (
	"context"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
)

const (
	ID        = "test"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "subcommand test"
	longHelp  = `test is a subcommand for testing purposes.`
)

type Command struct {
	model.Command
}

func (Command) Name() string               { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }
func (Command) Exec(context.Context, []string) error {
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
