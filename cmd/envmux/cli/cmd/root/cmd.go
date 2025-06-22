// Package root defines the root command executed when no subcommands are given.
package root

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root/fs"
	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root/ns"
	"github.com/ardnew/envmux/config"
	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
)

func Init() cmd.Node { return Root{}.Init() }

const ID = cmd.ID

const (
	syntax    = ID + ` [flags] [subcommand ...]`
	shortHelp = `namespaced environments`
	longHelp  = ID + ` constructs and evaluates static namespaced environments.`
)

//nolint:gochecknoglobals
var (
	versionFlag = ff.FlagConfig{
		ShortName:     'V',
		LongName:      `version`,
		Usage:         `show semantic version`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	verboseFlag = ff.FlagConfig{
		ShortName:     'v',
		LongName:      `verbose`,
		Usage:         `enable verbose output`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	configFlag = ff.FlagConfig{
		ShortName:     'c',
		LongName:      `config`,
		Usage:         `default command-line options`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	sourceFlag = ff.FlagConfig{
		ShortName:     's',
		LongName:      `source-file`,
		Usage:         `namespace source file ("-" is stdin)`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	defineFlag = ff.FlagConfig{
		ShortName:     'S',
		LongName:      `source`,
		Usage:         `namespace source definitions`,
		Placeholder:   `DEF`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
)

type Root struct {
	cmd.Config

	Version    bool
	Verbose    bool
	ConfigPath string
	SourcePath string
	SourceDef  []string

	verboseLevel int
}

func (r Root) Init() cmd.Node {
	r.Version = false
	r.Verbose = false
	r.ConfigPath = filepath.Join(cmd.ConfigDir(), configFlag.LongName)
	r.SourcePath = filepath.Join(cmd.ConfigDir(), `default`)
	r.SourceDef = []string{}

	r.Config = pkg.Wrap(
		r.Config,
		cmd.WithUsage(
			cmd.Usage{
				Name:      cmd.ConfigPrefix(),
				Syntax:    syntax,
				ShortHelp: shortHelp,
				LongHelp:  longHelp,
			},
			func(ctx context.Context, args []string) error {
				if r.Version {
					fmt.Println(ID, "version", pkg.Version)

					return nil
				}

				var (
					read io.Reader
					err  error
				)

				if len(r.SourceDef) > 0 {
					read, err = definitionsFromString(strings.Join(r.SourceDef, "\n"))
				} else {
					read, err = definitionsFromPath(r.SourcePath)
				}

				if err != nil {
					return fmt.Errorf("%w: %w", pkg.ErrInvalidDefinitions, err)
				}

				if r.VerboseLevel() > 0 {
					parse.ParseOptions = parse.TraceOptions(os.Stderr)
				}

				cfg := pkg.Make(config.WithReader(read))
				if cfg.Err() != nil {
					return fmt.Errorf("%w: %w", pkg.ErrInvalidConfig, cfg.Err())
				}

				if len(args) == 0 {
					args = pkg.DefaultNamespace
				}

				env, err := cfg.Eval(ctx, args...)
				if err != nil {
					return fmt.Errorf("%w: %w", pkg.ErrIncompleteEval, err)
				}

				for _, v := range env.Environ() {
					fmt.Println(v)
				}

				return nil
			},
		),
		cmd.WithFlags(
			pkg.Wrap(versionFlag, cmd.WithFlagConfig(&r.Version)),
			pkg.Wrap(verboseFlag, cmd.WithIncFlagConfig(&r.Verbose, &r.verboseLevel)),
			pkg.Wrap(configFlag, cmd.WithFlagConfig(&r.ConfigPath)),
			pkg.Wrap(sourceFlag, cmd.WithFlagConfig(&r.SourcePath)),
			pkg.Wrap(defineFlag, cmd.WithRepFlagConfig(&r.SourceDef)),
		),
		cmd.WithSubcommands(
			fs.Init(),
			ns.Init(),
		),
	)

	return r
}

func (r Root) VerboseLevel() int {
	if r.Verbose {
		return r.verboseLevel
	}

	return 0
}

func definitionsFromPath(path string) (io.Reader, error) {
	// Handle special cases for path to definitions file:
	//  1. If "-" given as flag argument, use stdin
	//  2. If flag argument is a relative path, use the first existing:
	//     a. relative to CWD
	//     b. relative to the config directory
	if path == "-" {
		// Read from stdin
		return os.Stdin, nil
	}

	var (
		r   io.Reader
		err error
	)

	// During the first iteration, control will either:
	//
	//  1. successfully construct a reader (r != nil, break loop), or
	//  2. fail to construct a reader (r == nil) with error (err != nil), and:
	//     A. the path is absolute, error persists (err != nil, break loop), or
	//     B. the path is relative to CWD, try to set path as an absolute path
	//        relative to the config directory, and:
	//       a. absolute path failed, overwrite error (err != nil, break loop), or
	//       b. absolute path succeeded, clear error (err == nil, continue to 1).
	//         - The second iteration will terminate at either 1 or 2A
	//           because condition 2B (the "recursive" step) will be false,
	//           since path is now guaranteed to be absolute.
	for r == nil && err == nil {
		if r, err = pkg.ReaderFromFile(path); err != nil && !filepath.IsAbs(path) {
			path, err = filepath.Abs(filepath.Join(cmd.ConfigDir(), path))
		}
	}

	return r, err
}

func definitionsFromString(def string) (io.Reader, error) {
	if def == "" {
		return nil, fmt.Errorf("%w: empty definition", pkg.ErrInvalidDefinitions)
	}

	// Use a string reader to read the definitions from a string
	return io.NopCloser(strings.NewReader(def)), nil
}
