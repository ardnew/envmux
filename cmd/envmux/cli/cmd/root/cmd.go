// Package root defines the root command executed when no subcommands are given.
package root

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root/fs"
	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root/ns"
	"github.com/ardnew/envmux/config/env"
	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
	"github.com/ardnew/envmux/pkg/prof"
)

var _ = cmd.Node(Root{}) //nolint:exhaustruct

func Init() Root { return new(Root).Init().(Root) } //nolint:forcetypeassert

const ID = cmd.ID

const (
	syntax    = ID + ` [flags] [subcommand ...]`
	shortHelp = `namespaced environments`
	longHelp  = ID + ` constructs and evaluates static namespaced environments.`
)

//nolint:gochecknoglobals,exhaustruct
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
	reqDefFlag = ff.FlagConfig{
		ShortName:     'u',
		LongName:      `require-definitions`,
		Usage:         `eval undef namespaces as runtime errors`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	maxJobsFlag = ff.FlagConfig{
		ShortName:     'j',
		LongName:      `jobs`,
		Usage:         `maximum number of parallel eval jobs`,
		Placeholder:   `N`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	confPathFlag = ff.FlagConfig{
		ShortName:     'c',
		LongName:      `config`,
		Usage:         `default command-line options`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	sourceFlag = ff.FlagConfig{
		ShortName:     's',
		LongName:      `source`,
		Usage:         `namespace definitions ("-" is stdin)`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	ignDefFlag = ff.FlagConfig{
		ShortName:     'i',
		LongName:      `ignore-default`,
		Usage:         `ignore default namespace definitions`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	profileFlag = ff.FlagConfig{
		ShortName: 'p',
		LongName:  `profile`,
		Usage: `enable profiling` + fmt.Sprintf(
			` (%s)`,
			strings.Join(prof.Modes(), `|`),
		),
		Placeholder:   `TYPE`,
		NoPlaceholder: false,
		NoDefault:     true,
	}
)

type Root struct {
	cmd.Config

	Version  bool
	Verbose  bool
	ReqDef   bool
	MaxJobs  int
	ConfPath []string
	Source   []string
	IgnDef   bool
	Profile  []string

	verboseLevel int
}

func (r Root) Init() cmd.Node { //nolint:ireturn
	r = Root{ //nolint:exhaustruct
		Version:  false,
		Verbose:  false,
		ReqDef:   false,
		MaxJobs:  runtime.NumCPU(),
		ConfPath: []string{filepath.Join(pkg.ConfigDir(ID), confPathFlag.LongName)},
	}

	flags := []ff.FlagConfig{
		pkg.Wrap(versionFlag, cmd.WithFlagConfig(&r.Version)),
		pkg.Wrap(verboseFlag, cmd.WithIncFlagConfig(&r.Verbose, &r.verboseLevel)),
		pkg.Wrap(reqDefFlag, cmd.WithFlagConfig(&r.ReqDef)),
		pkg.Wrap(maxJobsFlag, cmd.WithFlagConfig(&r.MaxJobs)),
		pkg.Wrap(confPathFlag, cmd.WithRepFlagConfig(&r.ConfPath)),
		pkg.Wrap(sourceFlag, cmd.WithRepFlagConfig(&r.Source)),
		pkg.Wrap(ignDefFlag, cmd.WithFlagConfig(&r.IgnDef)),
	}

	if len(prof.Modes()) > 0 {
		flags = append(
			flags,
			pkg.Wrap(profileFlag, cmd.WithRepFlagConfig(&r.Profile)),
		)
	}

	// Only set the default source when no source is provided
	// and the --ignore-default(-i) flag is not set.
	// Thus, this must be postponed until after the command-line is parsed.
	defaultSource := []string{filepath.Join(pkg.ConfigDir(ID), `default`)}

	r.Config = pkg.Wrap(
		r.Config,
		cmd.WithUsage(
			cmd.Usage{
				Name:      pkg.ConfigPrefix(ID),
				Syntax:    syntax,
				ShortHelp: shortHelp,
				LongHelp:  longHelp,
			},
			func(ctx context.Context, args []string) error {
				defer prof.Init(r.Profile...).Stop()

				if r.Version {
					fmt.Println(ID, "version", pkg.Version)

					return nil
				}

				if !r.IgnDef {
					r.Source = append(r.Source, defaultSource...)
				}

				var src []io.Reader

				for _, def := range r.Source {
					if def = strings.TrimSpace(def); def == "" {
						continue // skip empty definitions
					}

					var (
						uno io.Reader
						err error
					)

					if strings.HasPrefix(def, pkg.InlineSourcePrefix) {
						// Inline definitions provided as command-line argument
						uno, err = definitionsFromString(
							strings.TrimPrefix(def, pkg.InlineSourcePrefix),
						)
					} else {
						// Definitions read from file
						uno, err = definitionsFromPath(def)
					}

					if err != nil {
						return fmt.Errorf("%w: %w", pkg.ErrInvalidDefinitions, err)
					}

					src = append(src, uno)
				}

				if r.VerboseLevel() > 0 {
					parse.ParseOptions = parse.TraceOptions(os.Stderr)
				}

				e, err := pkg.Make(
					env.WithMaxParallelJobs(r.MaxJobs),
					env.WithEvalRequiresDef(r.ReqDef),
				).Parse(io.MultiReader(src...))
				if err != nil {
					return &pkg.IncompleteParseError{Err: err, Src: r.Source}
				}

				if len(args) == 0 {
					args = pkg.Namespace()
				}

				env, err := e.Eval(ctx, args...)
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
			flags...,
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
	//  1. If [pkg.StdinSourcePath] given as flag argument, use stdin
	//  2. If flag argument is a relative path, use the first existing:
	//     a. relative to CWD
	//     b. relative to the config directory
	if path == pkg.StdinSourcePath {
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
			path, err = filepath.Abs(filepath.Join(pkg.ConfigDir(ID), path))
		}
	}

	return r, err
}

func definitionsFromString(def string) (io.Reader, error) {
	// Use a string reader to read the definitions from a string
	return io.NopCloser(strings.NewReader(def)), nil
}
