package fs

import (
	"context"

	"github.com/ardnew/envmux/cli/model"
	"github.com/ardnew/envmux/cli/model/cmd/fs/test"
	"github.com/ardnew/envmux/cli/model/spec"
	"github.com/ardnew/envmux/pkg"
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

func (Command) Name() string               { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }

// Exec executes the command with the given context and arguments.
func (c Command) Exec(context.Context, []string) error {
	return nil
}

// Make creates a new fs Command with the given options.
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
		_ = test.Make(test.WithParent(&c.Command))
		return c
	}
}

// WithParent sets the parent command for the fs Command.
func WithParent(ptr *model.Command) pkg.Option[Command] {
	return func(c Command) Command {
		c.Command = pkg.Wrap(c.Command, model.WithParent(ptr))
		return c
	}
}
