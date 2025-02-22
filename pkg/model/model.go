package model

import (
	"context"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model/spec"
)

// Command represents the context and configuration of a command.
// It is not used as a command itself but as a container for actual commands.
// The types of actual commands are composed of this type via embedding.
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
func (c Command) Parent() Command {
	if c.parent != nil {
		return *c.parent
	}
	return Command{}
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
