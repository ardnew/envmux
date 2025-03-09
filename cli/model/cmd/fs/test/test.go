package test

import (
	"context"

	"github.com/ardnew/envmux/cli/model"
	"github.com/ardnew/envmux/cli/model/spec"
	"github.com/ardnew/envmux/pkg"
)

const (
	ID        = "test"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "subcommand test"
	longHelp  = `test is a subcommand for testing purposes.`
)

// Command represents the test command.
type Command struct {
	model.Command
}

func (Command) Name() string               { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }

// Exec executes the command with the given context and arguments.
func (Command) Exec(context.Context, []string) error {
	return nil
}

// Make creates a new test Command with the given options.
func Make(opts ...pkg.Option[Command]) (cmd Command) {
	// Ensure the [config.Command] is initialized before applying any options.
	cc := pkg.Make(withSpec(spec.Make(&cmd)))
	return pkg.Wrap(cc, opts...)
}

func withSpec(cs spec.Common) pkg.Option[Command] {
	return func(c Command) Command {
		// Configure default options
		// Configure command-line flags
		// Install command and subcommands
		c.Command = pkg.Make(model.WithSpec(cs))
		return c
	}
}

// WithParent sets the parent command for the test Command.
func WithParent(ptr *model.Command) pkg.Option[Command] {
	return func(c Command) Command {
		c.Command = pkg.Wrap(c.Command, model.WithParent(ptr))
		return c
	}
}
