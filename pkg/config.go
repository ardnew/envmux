package pkg

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v4"
)

const (
	fileFlag = "config"
)

type Config struct {
	Stdout    io.Writer
	Stderr    io.Writer
	File      string
	VarPrefix string
	Verbose   bool

	Flags   *ff.FlagSet
	Command *ff.Command
}

func MakeConfig(name string, opt ...Option[Config]) Config {
	if len(opt) == 0 {
		opt = append(opt, WithDefaults(name))
	}
	return WithOptions(Config{}, opt...)
}

func defaultConfigDir(name string) string {
	root, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		if root, ok = os.LookupEnv("HOME"); ok {
			root = filepath.Join(root, ".config")
		} else {
			var err error
			root, err = os.Getwd()
			if err != nil {
				root = "."
			}
		}
	}
	return filepath.Join(root, name)
}

func (c Config) GetParseOptions() []ff.Option {
	opt := []ff.Option{
		ff.WithConfigFileFlag(fileFlag),
		ff.WithConfigAllowMissingFile(),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix(c.VarPrefix),
	}
	return opt
}

func WithDefaults(name string) Option[Config] {
	12
	c := Config{
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		File:      filepath.Join(defaultConfigDir(name), "config"),
		VarPrefix: strings.ToUpper(name),
		Verbose:   false,
		Flags:     ff.NewFlagSet(name),
	}

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
	c.Flags.StringVar(&c.File, 'c', fileFlag, c.File, "path to configuration file")
	return c
}

func WithOutput(stdout, stderr io.Writer) Option[Config] {
	return func(c Config) Config {
		c.Stdout = stdout
		c.Stderr = stderr
		return c
	}
}

func WithFile(file string) Option[Config] {
	return func(c Config) Config {
		c.File = file
		return c
	}
}

func WithEnvVarPrefix(prefix string) Option[Config] {
	return func(c Config) Config {
		c.VarPrefix = prefix
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
