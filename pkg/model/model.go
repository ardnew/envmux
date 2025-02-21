package model

import (
	"context"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model/config"
	"github.com/peterbourgon/ff/v4"
)

type Config = config.Command

// Command defines a single command-line argument.
type Command struct {
	config Config
	parent *Command
}

func (m Command) Parse(args []string, opts ...ff.Option) error {
	return m.config.Command.Parse(args, opts...)
}

func (m Command) Run(ctx context.Context) error { return m.config.Run(ctx) }

func (m Command) IsZero() bool   { return m.config.IsZero() && m.parent == nil }
func (m Command) Config() Config { return m.config }
func (m Command) Parent() Command {
	if m.parent != nil {
		return *m.parent
	}
	return Command{}
}

func WithConfig(config Config) pkg.Option[Command] {
	return func(cmd Command) Command {
		cmd.config = config
		return cmd
	}
}

func WithParent(ptr *Command) pkg.Option[Command] {
	return func(cmd Command) Command {
		if ptr != nil {
			p, c := ptr.config, cmd.config
			if p.Command != nil && c.Command != nil {
				p.Subcommands = append(p.Subcommands, c.Command)
			}
			if p.FlagSet != nil && c.FlagSet != nil {
				c.SetParent(p.FlagSet)
			}
		}
		cmd.parent = ptr
		return cmd
	}
}
