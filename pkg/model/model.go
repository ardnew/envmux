package model

import (
	"context"

	"github.com/ardnew/groot/pkg"
	"github.com/peterbourgon/ff/v4"
)

// Interface describes a command and its syntax.
type Interface interface {
	Name() string
	Syntax() string
	Help() (short, long string)
	Exec(ctx context.Context, args []string) error
}

// Command defines a single command-line argument.
type Command struct {
	config Config
	parent *Command
}

func (m Command) IsZero() bool   { return m.config.IsZero() && m.parent == nil }
func (m Command) Config() Config { return m.config }
func (m Command) Parent() Command {
	if m.parent != nil {
		return *m.parent
	}
	return Command{}
}

func (m Command) Parse(args []string, opts ...ff.Option) error {
	return m.config.Command.Parse(args, opts...)
}

func (m Command) Run(ctx context.Context) error {
	return m.config.Command.Run(ctx)
}

func WithInterface(i Interface) pkg.Option[Command] {
	return WithConfig(
		func(c Config) Config {
			shortHelp, longHelp := i.Help()
			c.FlagSet = ff.NewFlagSet(i.Name())
			c.Command = &ff.Command{
				Name:      i.Name(),
				ShortHelp: shortHelp,
				LongHelp:  longHelp,
				Usage:     i.Syntax(),
				Flags:     c.FlagSet,
				Exec:      i.Exec,
			}
			return c
		},
	)
}

func WithConfig(opts ...pkg.Option[Config]) pkg.Option[Command] {
	return func(m Command) Command {
		m.config = pkg.WithOptions(m.config, opts...)
		return m
	}
}

func WithParent(p *Command) pkg.Option[Command] {
	return func(m Command) Command {
		if p != nil {
			pcmd, mcmd := p.config.Command, m.config.Command
			if pcmd != nil && mcmd != nil {
				pcmd.Subcommands = append(pcmd.Subcommands, mcmd)
			}
			pfs, mfs := p.config.FlagSet, m.config.FlagSet
			if pfs != nil && mfs != nil {
				mfs.SetParent(pfs)
			}
		}
		m.parent = p
		return m
	}
}
