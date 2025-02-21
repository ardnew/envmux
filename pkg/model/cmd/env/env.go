package env

import (
	"context"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
	"github.com/ardnew/groot/pkg/model/config"
)

const (
	ID        = "env"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "environment variables"
	longHelp  = `environment variables are user-definable key-value pairs ` +
		`that are accessible to a process and its children.`
)

// Command represents the env command.
type Command struct {
	model.Command
}

// Name returns the name of the command.
func (Command) Name() string { return ID }

// Syntax returns the syntax of the command.
func (Command) Syntax() string { return syntax }

// Help returns the short and long help descriptions of the command.
func (Command) Help() (short, long string) { return shortHelp, longHelp }

// Exec executes the command with the given context and arguments.
func (Command) Exec(context.Context, []string) error {
	return nil
}

// Make creates a new env Command with the given options.
func Make(opts ...pkg.Option[Command]) Command {
	withConfig := func(cfg config.Command) pkg.Option[Command] {
		return func(c Command) Command {
			// Configure default options
			// Configure command-line flags
			// Install command and subcommands
			c.Command = pkg.Make(model.WithConfig(cfg))
			return c
		}
	}
	// Ensure the [config.Command] is initialized before applying any options.
	return pkg.WithOptions(pkg.Make(withConfig(config.Make[Command]())), opts...)
}

// WithParent sets the parent command for the env Command.
func WithParent(ptr *model.Command) pkg.Option[Command] {
	return func(c Command) Command {
		c.Command = pkg.WithOptions(c.Command, model.WithParent(ptr))
		return c
	}
}
