package fs

import (
	"context"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
	"github.com/ardnew/groot/pkg/model/fs/test"
)

const (
	ID        = "fs"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "file system operations"
	longHelp  = `file system operations for managing files and directories.`
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
			WithModel(
				model.WithInterface(c),
			),
			func(g Command) Command {
				_ = test.Make(test.WithModel(model.WithParent(&g.Command)))
				return g
			},
		)
	}
}

func WithModel(opts ...pkg.Option[model.Command]) pkg.Option[Command] {
	return func(c Command) Command {
		c.Command = pkg.WithOptions(c.Command, opts...)
		return c
	}
}
