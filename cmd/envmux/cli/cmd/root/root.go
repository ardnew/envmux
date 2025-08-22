// Package root defines the root command executed when no subcommands are given.
package root

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root/fs"
	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root/ns"
	"github.com/ardnew/envmux/pkg"
	"github.com/ardnew/envmux/pkg/errs"
	"github.com/ardnew/envmux/pkg/fn"
	"github.com/ardnew/envmux/pkg/prof"
	"github.com/ardnew/envmux/pkg/run"
	"github.com/ardnew/envmux/spec/env"
)

var _ = cmd.Node(Node{}) //nolint:exhaustruct

func Init() Node { return new(Node).Init().(Node) } //nolint:forcetypeassert

const ID = pkg.Name

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
		Usage:         `treat undefined namespaces as errors`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	parallelLimitFlag = ff.FlagConfig{
		ShortName:     'j',
		LongName:      `jobs`,
		Usage:         `maximum number of parallel tasks used during evaluation`,
		Placeholder:   `N`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	bufferSizeFlag = ff.FlagConfig{
		ShortName:     'b',
		LongName:      `buffer-size`,
		Usage:         `size of parse buffer in bytes (or given SI units)`,
		Placeholder:   `N`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	configFlag = ff.FlagConfig{
		ShortName:     'c',
		LongName:      cmd.ConfigFlag,
		Usage:         `config file containing default command-line flags with options`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	manifestFlag = ff.FlagConfig{
		ShortName:     'm',
		LongName:      `manifest`,
		Usage:         `manifest file containing namespace definitions ("-" is stdin)`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	definesFlag = ff.FlagConfig{
		ShortName:     'd',
		LongName:      `define`,
		Usage:         `inline namespace definitions to append to manifest`,
		Placeholder:   `SOURCE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	ignoreDefaultFlag = ff.FlagConfig{
		ShortName:     'i',
		LongName:      `ignore-default`,
		Usage:         `ignore default manifest file`,
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

type Node struct {
	cmd.Config

	Version  bool
	Verbose  bool
	ReqDef   bool
	MaxJobs  int
	BufSize  int
	ConfPath []string
	Manifest []string
	Defines  []string
	IgnDef   bool
	Profile  []string

	verboseLevel int
}

func (r Node) Init(args ...any) cmd.Node { //nolint:ireturn
	r = Node{ //nolint:exhaustruct
		Version:  false,
		Verbose:  false,
		ReqDef:   false,
		MaxJobs:  runtime.NumCPU(),
		BufSize:  1 << 15, //nolint:mnd // 32 KiB
		ConfPath: []string{filepath.Join(run.ConfigDir(ID), configFlag.LongName)},
	}

	flags := []ff.FlagConfig{
		fn.Wrap(versionFlag, cmd.WithFlagConfig(&r.Version)),
		fn.Wrap(verboseFlag, cmd.WithIncFlagConfig(&r.Verbose, &r.verboseLevel)),
		fn.Wrap(reqDefFlag, cmd.WithFlagConfig(&r.ReqDef)),
		fn.Wrap(parallelLimitFlag, cmd.WithFlagConfig(&r.MaxJobs)),
		fn.Wrap(bufferSizeFlag, cmd.WithFlagConfig(&r.BufSize)),
		fn.Wrap(configFlag, cmd.WithRepFlagConfig(&r.ConfPath)),
		fn.Wrap(manifestFlag, cmd.WithRepFlagConfig(&r.Manifest)),
		fn.Wrap(definesFlag, cmd.WithRepFlagConfig(&r.Defines)),
		fn.Wrap(ignoreDefaultFlag, cmd.WithFlagConfig(&r.IgnDef)),
	}

	// If compiled with build tag pprof, add profiling flags.
	if len(prof.Modes()) > 0 {
		flags = append(
			flags,
			fn.Wrap(profileFlag, cmd.WithRepFlagConfig(&r.Profile)),
		)
	}

	// This must be postponed until after the command-line is parsed.
	defaultManifest := []string{filepath.Join(run.ConfigDir(ID), `default`)}

	r.Config = fn.Wrap(
		r.Config,
		cmd.WithUsage(
			cmd.Usage{
				Name:      run.ConfigPrefix(ID),
				Syntax:    syntax,
				ShortHelp: shortHelp,
				LongHelp:  longHelp,
			},
			func(ctx context.Context, args []string) error {
				// Initialize the profiler with all profiling modes provided
				// via the command-line flags.
				// If no profiling modes are provided, the profiler is not started.
				//
				// Note that profiling is only available when built with "-tags pprof".
				// Otherwise, this is a no-op.
				defer prof.Init(r.Profile...).Stop()

				if r.Version {
					fmt.Println(ID, "version", pkg.Version)

					return nil
				}

				// Only set the default manifest file
				// if flag --ignore-default is unset.
				if !r.IgnDef {
					r.Manifest = append(r.Manifest, defaultManifest...)
				}

				var (
					mod env.Model
					err error
				)

				mod, err = env.Make(
					ctx,
					r.Manifest,
					r.Defines,
					env.WithMaxParallelJobs(r.MaxJobs),
					env.WithEvalRequiresDef(r.ReqDef),
				)
				if err != nil {
					return fmt.Errorf("%w: %w", errs.ErrInvalidDefinitions, err)
				}

				mod, err = mod.Parse()
				if err != nil {
					return errs.IncompleteParseError{
						Err: err, Def: r.Manifest, Lvl: r.VerboseLevel(),
					}
				}

				if len(args) == 0 {
					args = run.Namespace()
				}

				env, err := mod.Eval(ctx, args...)
				if err != nil {
					return fmt.Errorf("%w: %w", errs.ErrIncompleteEval, err)
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

func (r Node) VerboseLevel() int {
	if r.Verbose {
		return r.verboseLevel
	}

	return 0
}
