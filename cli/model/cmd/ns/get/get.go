package get

import (
	"context"
	"fmt"

	"github.com/ardnew/envmux/cli/model"
	"github.com/ardnew/envmux/cli/model/proto"
	"github.com/ardnew/envmux/pkg"
)

const (
	ID        = "get"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "evaluate environment variable from namespace"
	longHelp  = `evaluate the expression assigned to an environment variable ` +
		`in the context of a given namespace.`
)

// Command represents the get command.
type Command struct {
	model.Command
}

func (c Command) Name() string             { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }

// Exec executes the command with the given context and arguments.
func (c Command) Exec(ctx context.Context, arg []string) error {
	_, err := fmt.Printf("[%s] arg=%+v\ncfg=%+v\n", ID, arg, c.Env())
	return err
}

// Make creates a new env Command with the given options.
func Make(opts ...pkg.Option[Command]) (cmd Command) {
	// Ensure the [config.Command] is initialized before applying any options.
	cc := pkg.Make(withSpec(proto.Make(&cmd)))
	return pkg.Wrap(cc, opts...)
}

func withSpec(s proto.Type) pkg.Option[Command] {
	return func(c Command) Command {
		// Configure default options
		// Configure command-line flags
		// Install command and subcommands
		c.Command = pkg.Make(model.WithProto(s))
		return c
	}
}

// WithParent sets the parent command for the env Command.
func WithParent(ptr *model.Command) pkg.Option[Command] {
	return func(c Command) Command {
		c.Command = pkg.Wrap(c.Command, model.WithParent(ptr))
		return c
	}
}
