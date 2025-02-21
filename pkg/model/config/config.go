package config

import (
	"context"

	"github.com/ardnew/groot/pkg"
	"github.com/peterbourgon/ff/v4"
)

// Interface is the interface used by all command-line (sub)commands.
type Interface interface {
	Name() string
	Syntax() string
	Help() (short, long string)
	Exec(ctx context.Context, args []string) error
}

// Command defines a command-line (sub)command and its associated flags.
type Command struct {
	*ff.Command
	*ff.FlagSet
}

// IsZero checks if the Command is uninitialized.
func (c Command) IsZero() bool { return c.Command == nil && c.FlagSet == nil }

// Make returns a new Command initialized with the given options.
//
// The Command passed to each Option is fully-initialized
// according to the Interface type parameter.
func Make[I Interface](opts ...pkg.Option[Command]) Command {
	// This Option must always be the first applied to a Command.
	withInterface := func(impl Interface) pkg.Option[Command] {
		return func(c Command) Command {
			// Configure default options
			shortHelp, longHelp := impl.Help()
			c.FlagSet = ff.NewFlagSet(impl.Name())
			c.Command = &ff.Command{
				Name:      impl.Name(),
				ShortHelp: shortHelp,
				LongHelp:  longHelp,
				Usage:     impl.Syntax(),
				Flags:     c.FlagSet,
				Exec:      impl.Exec,
			}
			return c
		}
	}
	// Ensure the [Command] is initialized before applying any options.
	return pkg.WithOptions(pkg.Make(withInterface(*new(I))), opts...)
}
