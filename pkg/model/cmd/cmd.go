package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/model"
	"github.com/ardnew/groot/pkg/model/cmd/env"
	"github.com/ardnew/groot/pkg/model/cmd/fs"
	"github.com/ardnew/groot/pkg/model/spec"
)

const (
	ID        = "groot"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "virtual environments"
	longHelp  = ID + ` is a tool for managing virtual environments.`
)

type defaultFlag[T any] struct {
	flag  string
	value T
}

//nolint:gochecknoglobals
var defaultConfigFile = defaultFlag[string]{flag: "config", value: "config"}

// Command represents the root command for the application.
type Command struct {
	model.Command

	ID   string
	Args []string

	Stdout  io.Writer
	Stderr  io.Writer
	File    string
	Verbose bool
}

func (Command) Name() string               { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }

// Exec executes the command with the given context and arguments.
func (Command) Exec(context.Context, []string) error {
	// _, err := fmt.Fprintf(c.Stdout, "[%s] arg=%+v\n", ID, arg)
	return nil
}

// Run parses and runs the command with the given context.
func (c Command) Run(ctx context.Context) error {
	if err := c.Command.Parse(c.Args, getParseOptions(c.ID)...); err != nil {
		return err
	}
	if err := c.Command.Run(ctx); err != nil {
		return err
	}
	return nil
}

// Make creates a new Command with the given options.
func Make(opts ...pkg.Option[Command]) Command {
	withSpec := func(cs spec.Common) pkg.Option[Command] {
		return func(c Command) Command {
			// Configure default options
			c.ID = getConfigID()
			c.Args = os.Args[1:]
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.File = filepath.Join(getConfigDir(), defaultConfigFile.value)
			c.Verbose = false
			// Configure command-line flags
			cs.BoolVar(&c.Verbose, 'v', "verbose", "log verbose output")
			cs.StringVar(&c.File, 'c', defaultConfigFile.flag, c.File, "path to configuration file")
			// Install command and subcommands
			c.Command = pkg.Make(model.WithSpec(cs))
			_ = env.Make(env.WithParent(&c.Command))
			_ = fs.Make(fs.WithParent(&c.Command))
			return c
		}
	}
	// Ensure the [config.Command] is initialized before applying any options.
	return pkg.WithOptions(pkg.Make(withSpec(spec.Make[Command]())), opts...)
}

// WithArgs sets the arguments for the Command.
func WithArgs(args ...string) pkg.Option[Command] {
	return func(c Command) Command {
		if len(args) > 0 {
			c.ID = filepath.Base(args[0])
			c.Args = args[1:] // empty slice if len(args) == 1
		}
		return c
	}
}

// WithOutput sets the output writers for the Command.
func WithOutput(stdout, stderr io.Writer) pkg.Option[Command] {
	return func(c Command) Command {
		c.Stdout = stdout
		c.Stderr = stderr
		return c
	}
}

// WithFile sets the configuration file for the Command.
func WithFile(file string) pkg.Option[Command] {
	return func(c Command) Command {
		c.File = file
		return c
	}
}

// WithVerbose sets the verbose flag for the Command.
func WithVerbose(verbose bool) pkg.Option[Command] {
	return func(c Command) Command {
		c.Verbose = verbose
		return c
	}
}

// getConfigID returns the ID of the configuration.
func getConfigID() string {
	id := os.Args[0]
	if exe, err := os.Executable(); err == nil {
		id = exe
	}
	if strings.Contains(id, "__debug_bin") {
		id = ID
	}
	return filepath.Base(id)
}

// getConfigDir returns the configuration directory.
func getConfigDir(relPath ...string) string {
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
	path := getConfigID()
	if len(relPath) > 0 {
		path = filepath.Join(relPath...)
	}
	return filepath.Join(root, path)
}

// getParseOptions returns the options for parsing the command-line arguments.
func getParseOptions(configFile string, envVarPrefix ...string) []ff.Option {
	if len(envVarPrefix) == 0 {
		envVarPrefix = []string{getConfigID()}
	}
	return []ff.Option{
		ff.WithConfigFileFlag(configFile),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithConfigAllowMissingFile(),
		ff.WithEnvVarPrefix(pkg.FormatEnvVar(envVarPrefix...)),
	}
}
