package env

import (
	"context"
	"fmt"
	"strings"

	"github.com/ardnew/groot/pkg"
	"github.com/peterbourgon/ff/v4"
)

const ID = "env"

type Config struct {
	*pkg.Config
	Command *ff.Command
	Flags   *ff.FlagSet
}

func New(parent *pkg.Config) *Config {
	path := []string{parent.Command.Name, ID}
	c := Config{
		Config: parent,
		Flags:  ff.NewFlagSet(strings.Join(path, " ")).SetParent(parent.Flags),
	}

	c.Command = &ff.Command{
		Name:      ID,
		ShortHelp: "environment variables",
		Usage:     c.Flags.GetName() + " [flags] [command] [...]",
		Flags:     c.Flags,
		Exec: func(_ context.Context, arg []string) error {
			_, err := fmt.Fprintf(c.Stdout, "[%s] arg=%+v\n", ID, arg)
			return err
		},
	}

	c.Config.Command.Subcommands = append(c.Config.Command.Subcommands, c.Command)
	return &c
}
