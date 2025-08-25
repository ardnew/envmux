// Package cmd defines the interface for command-line (sub)commands.
// Subcommands are represented using type [Node].
package cmd

import (
	"context"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/pkg"
)

type Node interface {
	Command() *ff.Command
	FlagSet() *ff.FlagSet
	Init(args ...any) Node
}

type Usage struct {
	Name      string
	Syntax    string
	ShortHelp string
	LongHelp  string
}

type Exec func(ctx context.Context, args []string) error

type Config struct {
	cmd *ff.Command
	set *ff.FlagSet
}

func (c Config) Command() *ff.Command { return c.cmd }
func (c Config) FlagSet() *ff.FlagSet { return c.set }

func WithUsage(usage Usage, exec Exec) pkg.Option[Config] {
	return func(c Config) Config {
		c.set = ff.NewFlagSet(usage.Name)

		c.cmd = &ff.Command{
			Name:        usage.Name,
			Usage:       usage.Syntax,
			ShortHelp:   usage.ShortHelp,
			LongHelp:    usage.LongHelp,
			Exec:        exec,
			Flags:       c.set,
			Subcommands: []*ff.Command{},
		}

		return c
	}
}

func WithFlags(cfgs ...ff.FlagConfig) pkg.Option[Config] {
	return func(c Config) Config {
		err := Validate(c.Command(), c.FlagSet())
		if err != nil {
			return c // Invalid node, return as-is
		}

		for _, cfg := range cfgs {
			if _, err := c.set.AddFlag(cfg); err != nil {
				continue // Skip invalid flags
			}
		}

		c.cmd.Flags = c.set

		return c
	}
}

func WithSubcommands(subs ...Node) pkg.Option[Config] {
	return func(c Config) Config {
		err := Validate(c.Command(), c.FlagSet())
		if err != nil {
			return c // Invalid node, return as-is
		}

		for _, sub := range subs {
			err := Validate(sub.Command(), sub.FlagSet())
			if err != nil {
				continue
			}

			c.cmd.Subcommands = append(c.cmd.Subcommands, sub.Command())
			sub.FlagSet().SetParent(c.FlagSet())
		}

		return c
	}
}

func Validate(cmd *ff.Command, set ff.Flags) (err error) {
	switch {
	case cmd == nil, cmd.Exec == nil:
		return pkg.ErrInvalidCommand

	case set == nil, cmd.Flags == nil:
		return pkg.ErrInvalidFlagSet

	case cmd.Name == "",
		cmd.Usage == "",
		cmd.ShortHelp == "",
		cmd.LongHelp == "":
		return pkg.ErrInvalidInterface

	default:
		return nil
	}
}
