package model

import (
	"context"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/cli/model/spec"
	"github.com/ardnew/envmux/pkg"
)

// Command represents the context and configuration of a command.
//
// It is not used as a command itself but as a container for actual commands.
// The types of actual commands are composed of this type via embedding.
//
// All Command methods are immutable.
// Any method that modifies its receiver will return a new Command.
//
// The With* functions returning a [pkg.Option] can be used with either
// [pkg.Wrap] to modify an existing Command or
// [pkg.Make] to create a new Command.
type Command struct {
	spec   spec.Common
	parent *Command
}

// Parse parses the command-line arguments.
func (c Command) Parse(args []string, opts ...ff.Option) error {
	return c.spec.Command.Parse(args, opts...)
}

// Run executes the command with the given context.
func (c Command) Run(ctx context.Context) error { return c.spec.Run(ctx) }

// IsZero checks if the Command is uninitialized.
func (c Command) IsZero() bool { return c.spec.IsZero() && c.parent == nil }

// Spec returns the Command specification.
func (c Command) Spec() spec.Common { return c.spec }

// Parent returns the parent Command.
//
// Nil is returned if the Command has no parent.
func (c Command) Parent() *Command { return c.parent }

// Flag returns the flag with the given name defined in the receiver
// or any of its ancestors.
//
// The second return value is true iff the flag is found.
func (c Command) Flag(name string) (ff.Flag, bool) {
	return c.spec.GetFlag(name)
}

// WithSpec sets the common fields specifying the Command.
func WithSpec(cs spec.Common) pkg.Option[Command] {
	return func(cmd Command) Command {
		cmd.spec = cs
		return cmd
	}
}

// WithParent sets the parent Command.
func WithParent(ptr *Command) pkg.Option[Command] {
	return func(cmd Command) Command {
		if ptr != nil {
			p, c := ptr.spec, cmd.spec
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
