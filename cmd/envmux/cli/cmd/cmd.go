package cmd

import (
	"context"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/pkg"
)

// Node describes a CLI node with an [ff.Command] and its [ff.FlagSet].
// Implementations should be instantiable via a zero value followed by Init.
type Node interface {
	Command() *ff.Command
	FlagSet() *ff.FlagSet
	Init(args ...any) Node
}

// Usage describes basic CLI usage text for a command.
type Usage struct {
	Name      string
	Syntax    string
	ShortHelp string
	LongHelp  string
}

// Exec is the function signature executed by a command.
type Exec func(ctx context.Context, args []string) error

// Config bundles an [ff.Command] and its [ff.FlagSet] for a [Node].
type Config struct {
	cmd *ff.Command
	set *ff.FlagSet
}

// Command returns the underlying [ff.Command].
func (c Config) Command() *ff.Command { return c.cmd }

// FlagSet returns the underlying [ff.FlagSet].
func (c Config) FlagSet() *ff.FlagSet { return c.set }

// WithUsage constructs the underlying [ff.Command] from [Usage] and [Exec].
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

// WithFlags adds flag configurations to the node's [ff.FlagSet] and wires
// them into the underlying [ff.Command]. Invalid flags are skipped.
func WithFlags(cfgs ...ff.FlagConfig) pkg.Option[Config] {
	return func(c Config) Config {
		err := Validate(c.Command(), c.FlagSet())
		if err != nil {
			return c // Invalid node, return as-is
		}

		for _, cfg := range cfgs {
			_, err := c.set.AddFlag(cfg)
			if err != nil {
				continue // Skip invalid flags
			}
		}

		c.cmd.Flags = c.set

		return c
	}
}

// WithSubcommands appends validated subcommands to the node and sets their
// parent flag set appropriately.
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

// Validate returns an error if a command or its flag set is incomplete.
func Validate(cmd *ff.Command, set ff.Flags) (err error) {
	switch {
	case cmd == nil, cmd.Exec == nil:
		return pkg.ErrUndefCommandExec

	case set == nil, cmd.Flags == nil:
		return pkg.ErrUndefCommandFlagSet

	case cmd.Name == "",
		cmd.Usage == "",
		cmd.ShortHelp == "",
		cmd.LongHelp == "":
		return pkg.ErrUndefCommandUsage

	default:
		return nil
	}
}
