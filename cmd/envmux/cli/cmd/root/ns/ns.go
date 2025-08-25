// Package ns defines the "ns" subcommand used for namespace operations.
package ns

import (
	"context"
	"fmt"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/pkg"
)

var _ = cmd.Node(Node{}) //nolint:exhaustruct

func Init() Node { return new(Node).Init().(Node) } //nolint:forcetypeassert

//go:generate sed -i -E "s/(const ID = )\"[^\"]+\"/\\1\"$GOPACKAGE\"/" "$GOFILE"
const ID = "ns"

const (
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "namespace operations"
	longHelp  = `compose and evaluate environmental namespaces`
)

type Node struct {
	cmd.Config
}

func (n Node) Init(...any) cmd.Node { //nolint:ireturn
	n.Config = pkg.Wrap(
		n.Config,
		cmd.WithUsage(
			cmd.Usage{
				Name:      ID,
				Syntax:    syntax,
				ShortHelp: shortHelp,
				LongHelp:  longHelp,
			},
			func(_ context.Context, _ []string) error {
				fmt.Println("Parent: ", n.Command().GetParent().Name)

				return nil
			},
		),
		cmd.WithFlags(),
		cmd.WithSubcommands(),
	)

	return n
}
