package fs

import (
	"context"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
	"github.com/ardnew/groot/pkg/model/cmd/fs/test"
	"github.com/ardnew/groot/pkg/model/config"
)

const (
	ID        = "fs"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "file system operations"
	longHelp  = `file system operations for managing files and directories.`
)

// Command represents the fs command.
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

// Make creates a new fs Command with the given options.
func Make(opts ...pkg.Option[Command]) Command {
	withConfig := func(cfg config.Command) pkg.Option[Command] {
		return func(c Command) Command {
			// Configure default options
			// Configure command-line flags
			// Install command and subcommands
			c.Command = pkg.Make(model.WithConfig(cfg))
			_ = test.Make(test.WithParent(&c.Command))
			return c
		}
	}
	// Ensure the [config.Command] is initialized before applying any options.
	return pkg.WithOptions(pkg.Make(withConfig(config.Make[Command]())), opts...)
}

// WithParent sets the parent command for the fs Command.
func WithParent(ptr *model.Command) pkg.Option[Command] {
	return func(c Command) Command {
		c.Command = pkg.WithOptions(c.Command, model.WithParent(ptr))
		return c
	}
}
