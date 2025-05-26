package ns

import (
	"context"
	"fmt"

	"github.com/ardnew/envmux/cli/model"
	"github.com/ardnew/envmux/cli/model/cmd/ns/get"
	"github.com/ardnew/envmux/cli/model/proto"
	"github.com/ardnew/envmux/pkg"
)

const (
	ID        = "ns"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "namespace management"
	longHelp  = `namespaces are isolated sets of environment variables. ` +
		`composite namespaces can be generated dynamically.`
)

// Command represents the ns command.
type Command struct {
	model.Command
}

func (Command) Name() string               { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }

// Exec executes the command with the given context and arguments.
func (c Command) Exec(ctx context.Context, arg []string) error {
	verbose, _ := c.FlagAsBool("verbose")
	if verbose {
		_, err := fmt.Printf("[%s] arg=%+v\ncfg=%+v\n", ID, arg, c.Env())
		return err
	}
	return nil
}

// Make creates a new ns Command with the given options.
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
		_ = get.Make(get.WithParent(&c.Command))
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
