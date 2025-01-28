package pkg

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/peterbourgon/ff/v4"
)

type Config struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Verbose bool

	Flags   *ff.FlagSet
	Command *ff.Command
}

func MakeConfig(name string, opt ...Option[Config]) Config {
	if len(opt) == 0 {
		opt = append(opt, WithDefaults(name))
	}
	return WithOptions(Config{}, opt...)
}

func WithDefaults(name string) Option[Config] {
	return func(Config) Config {
		c := Config{}
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Flags = ff.NewFlagSet(name)
		c.Command = &ff.Command{
			Name:      c.Flags.GetName(),
			ShortHelp: "generate interactive environments",
			Usage:     c.Flags.GetName() + " [flags] command [...]",
			Flags:     c.Flags,
			Exec: func(_ context.Context, arg []string) error {
				_, err := fmt.Fprintf(c.Stdout, "[%s] arg=%+v\n", name, arg)
				return err
			},
		}

		c.Flags.BoolVar(&c.Verbose, 'v', "verbose", "log verbose output")
		return c
	}
}

func WithOutput(stdout, stderr io.Writer) Option[Config] {
	return func(c Config) Config {
		c.Stdout = stdout
		c.Stderr = stderr
		return c
	}
}

func WithVerbose(verbose bool) Option[Config] {
	return func(c Config) Config {
		c.Verbose = verbose
		return c
	}
}

func WithFlags(flags *ff.FlagSet) Option[Config] {
	return func(c Config) Config {
		c.Flags = flags
		return c
	}
}

func WithCommand(cmd *ff.Command) Option[Config] {
	return func(c Config) Config {
		c.Command = cmd
		return c
	}
}
