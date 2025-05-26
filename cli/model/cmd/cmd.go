package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/cli/model"
	"github.com/ardnew/envmux/cli/model/cmd/fs"
	"github.com/ardnew/envmux/cli/model/cmd/ns"
	"github.com/ardnew/envmux/cli/model/proto"
	"github.com/ardnew/envmux/config"
	"github.com/ardnew/envmux/pkg"
)

const (
	ID        = "envmux"
	syntax    = ID + " [flags] [subcommand ...]"
	shortHelp = "virtual environments"
	longHelp  = ID + ` is a tool for managing virtual environments.`
)

// Command represents the root command for the application.
type Command struct {
	model.Command

	ID   string
	Args []string

	Stdout     io.Writer
	Stderr     io.Writer
	Config     string
	Namespaces string
	Verbose    bool

	config config.Model // configuration file AST
}

// ConfigPrefix returns the base prefix string used to construct the path to the
// configuration directory and the prefix for environment variable identifiers.
//
// ConfigPrefix is exported to allow prefix customization.
//
// By default, the prefix is the base name of the executable file
// unless it matches one of the following substitution rules:
//
//   - "__debug_bin" (default output of the dlv debugger): replaced with [ID]
var ConfigPrefix = func() string {
	id := os.Args[0]
	if exe, err := os.Executable(); err == nil {
		id = exe
	}
	id = filepath.Base(id)
	substitute := []struct {
		*regexp.Regexp
		string
	}{
		{regexp.MustCompile(`^__debug_bin\d+$`), ID}, // default output of the dlv debugger
	}
	for _, sub := range substitute {
		id = sub.Regexp.ReplaceAllString(id, sub.string)
	}
	return id
}

type defaultFlag[T any] struct {
	flag  string
	value T
}

//nolint:gochecknoglobals
var (
	defaultConfig     = defaultFlag[string]{flag: "config", value: "config"}
	defaultNamespaces = defaultFlag[string]{flag: "namespaces", value: "namespaces"}
)

func (Command) Name() string               { return ID }
func (Command) Syntax() string             { return syntax }
func (Command) Help() (short, long string) { return shortHelp, longHelp }

// Exec executes the command with the given context and arguments.
func (c Command) Exec(ctx context.Context, args []string) error {
	env, err := c.Eval(ctx, args...)
	if err != nil {
		return err
	}
	// fmt.Printf("%+v\n", env)
	s := pkg.Wrap(c.Command, model.WithEnv(env))
	if c.Verbose {
		_, _ = fmt.Fprintf(c.Stdout, "[%s] arg=%+v\ncfg=%+v\n", ID, args, s)
	}

	for _, e := range s.Environ() {
		_, _ = fmt.Fprintln(c.Stdout, e)
	}
	return nil
}

// Run parses and runs the command with the given context.
func (c Command) Run(ctx context.Context) error {
	if err := c.Parse(c.Args, getParseOptions(c.ID)...); err != nil {
		return err
	}
	read, err := pkg.ReaderFromFile(c.Namespaces)
	if err != nil {
		return fmt.Errorf("%w: %w: %s", pkg.ErrInvalidConfigFile, err, c.Namespaces)
	}
	err = c.Command.Run(ctx, pkg.Make(config.WithReader(read)), c.Args...)
	if err != nil {
		return err
	}
	return nil
}

// Make creates a new Command with the given options.
func Make(opts ...pkg.Option[Command]) (r Command) {
	// Ensure the [config.Command] is initialized before applying any options.
	c := pkg.Make(withProto(proto.Make(&r)))
	return pkg.Wrap(c, opts...)
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

// WithConfig sets the configuration file for the Command.
func WithConfig(path string) pkg.Option[Command] {
	return func(c Command) Command {
		c.Config = path
		return c
	}
}

// WithNamespace sets the namespace file for the Command.
func WithNamespace(path string) pkg.Option[Command] {
	return func(c Command) Command {
		c.Namespaces = path
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

func withProto(s proto.Type) pkg.Option[Command] {
	return func(c Command) Command {
		// Configure default options
		c.ID = ConfigPrefix()
		c.Args = os.Args[1:]
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Config = filepath.Join(getConfigDir(), defaultConfig.value)
		c.Namespaces = filepath.Join(getConfigDir(), defaultNamespaces.value)
		c.Verbose = false

		// Configure command-line flags
		s.BoolVar(&c.Verbose, 'v', "verbose", "log verbose output")
		s.StringVar(&c.Config, 'c', defaultConfig.flag, c.Config, "path to configuration file")
		s.StringVar(&c.Namespaces, 'f', defaultNamespaces.flag, c.Namespaces, "path to namespace file")

		// Install command and subcommands
		c.Command = pkg.Make(model.WithProto(s))
		_ = ns.Make(ns.WithParent(&c.Command))
		_ = fs.Make(fs.WithParent(&c.Command))

		return c
	}
}

// getConfigDir returns the configuration directory.
//
// The configuration directory is constructed by appending each given subdir
// to the root configuration directory.
//
// If no subdir arguments are given, the config ID is used (see [ConfigIDMap]).
//
// If the given subdir elements represents an absolute path (after joining),
// the absolute path is returned as-is.
//
// The root configuration directory is determined in the following order:
//
//  1. Environment variable XDG_CONFIG_HOME (if defined)
//  2. Environment variable HOME (if defined), with ".config" appended
//  3. Current working directory
func getConfigDir(subdir ...string) string {
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

	path := ConfigPrefix()
	if len(subdir) > 0 {
		path = filepath.Join(subdir...)
	}

	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(root, path)
}

// getParseOptions returns the options for parsing the command-line arguments.
func getParseOptions(envVarPrefix ...string) []ff.Option {
	if len(envVarPrefix) == 0 {
		envVarPrefix = []string{ConfigPrefix()}
	}

	return []ff.Option{
		ff.WithConfigFileFlag(defaultConfig.flag),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithConfigAllowMissingFile(),
		ff.WithEnvVarPrefix(pkg.FormatEnvVar(envVarPrefix...)),
	}
}
