// Package fs defines the "fs" subcommand used for file system management.
package fs

import (
	"context"
	"fmt"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/pkg"
)

var _ = cmd.Node(Node{}) //nolint:exhaustruct

func Init() Node { return new(Node).Init().(Node) } //nolint:forcetypeassert

//go:generate sed -i -E "s/(const ID = )\"[^\"]+\"/\\1\"$GOPACKAGE\"/" "$GOFILE"
const ID = "fs"

const (
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "file system management"
	longHelp  = `configure and modify file system content`
)

type Node struct {
	cmd.Config
}

func (n Node) Init() cmd.Node { //nolint:ireturn
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
