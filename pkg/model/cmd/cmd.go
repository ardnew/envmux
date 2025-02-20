package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
	"github.com/ardnew/groot/pkg/model/env"
	"github.com/ardnew/groot/pkg/model/fs"
	"github.com/peterbourgon/ff/v4"
)

const (
	ID        = "groot"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "virtual environments"
	longHelp  = ID + ` is a tool for managing virtual environments.`
)

const (
	fileFlag = "config"
)

type Command struct {
	model.Command

	ID   string
	Args []string

	Stdout  io.Writer
	Stderr  io.Writer
	File    string
	Verbose bool
}

func (Command) Name() string               { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }
func (Command) Exec(context.Context, []string) error {
	// _, err := fmt.Fprintf(c.Stdout, "[%s] arg=%+v\n", ID, arg)
	return nil
}

func (c Command) Run(ctx context.Context) error {
	if err := c.Command.Parse(c.Args, getParseOptions(c.ID)...); err != nil {
		return err
	}
	if err := c.Command.Run(ctx); err != nil {
		return err
	}
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
		c.ID = filepath.Base(os.Args[0])
		c.Args = os.Args[1:]

		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		c.File = filepath.Join(getConfigDir(c.ID), "config")

		c.Verbose = false

		return pkg.WithOptions(c,
			WithModel(
				model.WithInterface(c),
				model.WithConfig(
					func(n model.Config) model.Config {
						n.BoolVar(&c.Verbose, 'v', "verbose", "log verbose output")
						n.StringVar(&c.File, 'c', "config", c.File, "path to configuration file")
						return n
					},
				),
			),
			func(g Command) Command {
				_ = env.Make(env.WithModel(model.WithParent(&g.Command)))
				_ = fs.Make(fs.WithModel(model.WithParent(&g.Command)))
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

func WithArgs(args ...string) pkg.Option[Command] {
	return func(c Command) Command {
		if len(args) > 0 {
			c.ID = filepath.Base(args[0])
			c.Args = args[1:] // empty slice if len(args) == 1
		}
		return c
	}
}

func WithOutput(stdout, stderr io.Writer) pkg.Option[Command] {
	return func(c Command) Command {
		c.Stdout = stdout
		c.Stderr = stderr
		return c
	}
}

func WithFile(file string) pkg.Option[Command] {
	return func(c Command) Command {
		c.File = file
		return c
	}
}

func WithVerbose(verbose bool) pkg.Option[Command] {
	return func(c Command) Command {
		c.Verbose = verbose
		return c
	}
}

// func getConfigID() string {
// 	id := os.Args[0]
// 	if exe, err := os.Executable(); err == nil {
// 		id = exe
// 	}
// 	return filepath.Base(id)
// }

func getConfigDir(id string) string {
	root, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		if root, ok = os.LookupEnv("HOME"); ok {
			root = filepath.Join(root, ".config")
		} else {
			var err error
			root, err = os.Getwd()
			if err != nil {
				root = "."
			}
		}
	}
	return filepath.Join(root, id)
}

func getParseOptions(id string) []ff.Option {
	return []ff.Option{
		ff.WithConfigFileFlag(fileFlag),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithConfigAllowMissingFile(),
		ff.WithEnvVarPrefix(strings.ToUpper(id)),
	}
}
