package spec

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

// Common defines common content of command-line (sub)commands.
type Common struct {
	*ff.Command
	*ff.FlagSet
}

// IsZero checks if the Common is uninitialized.
func (c Common) IsZero() bool { return c.Command == nil && c.FlagSet == nil }

// Make returns a new Common initialized with the given options.
//
// The Common passed to each Option is fully-initialized
// according to the Interface type parameter.
func Make[I Interface](opts ...pkg.Option[Common]) Common {
	// This Option must always be the first applied to a Common.
	withInterface := func(impl Interface) pkg.Option[Common] {
		return func(c Common) Common {
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
	// Ensure the [Common] is initialized before applying any options.
	return pkg.WithOptions(pkg.Make(withInterface(*new(I))), opts...)
}
