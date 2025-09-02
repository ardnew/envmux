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
	"github.com/ardnew/envmux/cmd/envmux/pprof"
	"github.com/ardnew/envmux/manifest"
	"github.com/ardnew/envmux/manifest/config"
	"github.com/ardnew/envmux/pkg"
	"github.com/ardnew/envmux/pkg/fn"
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
	verboseLevelFlag = ff.FlagConfig{
		ShortName:     'v',
		LongName:      `verbose`,
		Usage:         `enable verbose output`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	noDefaultDefinitionFlag = ff.FlagConfig{
		ShortName:     'i',
		LongName:      `ignore-default`,
		Usage:         `ignore default manifest file`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	strictDefinitionsFlag = ff.FlagConfig{
		ShortName:     's',
		LongName:      `strict-definitions`,
		Usage:         `treat undefined namespaces as errors`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	parallelEvalLimitFlag = ff.FlagConfig{
		ShortName:     'j',
		LongName:      `jobs`,
		Usage:         `maximum number of parallel tasks used during evaluation`,
		Placeholder:   `N`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	configurationPathFlag = ff.FlagConfig{
		ShortName:     'c',
		LongName:      cmd.ConfigFlag,
		Usage:         `config file containing default command-line flags with options`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	manifestPathFlag = ff.FlagConfig{
		ShortName:     'm',
		LongName:      `manifest`,
		Usage:         `manifest file containing namespace definitions ("-" is stdin)`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	inlineDefinitionFlag = ff.FlagConfig{
		ShortName:     'd',
		LongName:      `define`,
		Usage:         `inline namespace definitions to append to manifest`,
		Placeholder:   `SOURCE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	profileFlag = ff.FlagConfig{
		ShortName: 'p',
		LongName:  `profile`,
		Usage: `write pprof profile to file ` + fmt.Sprintf(
			` (%s)`,
			strings.Join(pprof.Modes(), `|`),
		),
		Placeholder:   `TYPE[=DIR]`,
		NoPlaceholder: false,
		NoDefault:     true,
	}
)

type Node struct {
	cmd.Config

	Version             bool
	Verbose             bool
	NoDefaultDefinition bool
	StrictDefinitions   bool
	ParallelEvalLimit   int
	ConfigurationPath   []string
	ManifestPath        []string
	InlineDefinition    []string
	Profile             string

	verboseLevel int
}

func (r Node) Init(args ...any) cmd.Node { //nolint:ireturn
	r = Node{ //nolint:exhaustruct
		Version:           false,
		Verbose:           false,
		StrictDefinitions: false,
		ParallelEvalLimit: runtime.NumCPU(),
		ConfigurationPath: []string{
			filepath.Join(config.Dir(ID), configurationPathFlag.LongName),
		},
	}

	flags := fn.FilterItems(
		// If compiled with build tag pprof, add profiling flags.
		[]ff.FlagConfig{
			pkg.Wrap(profileFlag, cmd.WithFlagConfig(&r.Profile)),
		},
		func(ff.FlagConfig) bool { return len(pprof.Modes()) > 0 },
	)

	// Remaining flags are all added unconditionally.
	flags = append(flags,
		pkg.Wrap(
			versionFlag,
			cmd.WithFlagConfig(&r.Version),
		),
		pkg.Wrap(
			verboseLevelFlag,
			cmd.WithIncFlagConfig(&r.Verbose, &r.verboseLevel),
		),
		pkg.Wrap(
			noDefaultDefinitionFlag,
			cmd.WithFlagConfig(&r.NoDefaultDefinition),
		),
		pkg.Wrap(
			strictDefinitionsFlag,
			cmd.WithFlagConfig(&r.StrictDefinitions),
		),
		pkg.Wrap(
			parallelEvalLimitFlag,
			cmd.WithFlagConfig(&r.ParallelEvalLimit),
		),
		pkg.Wrap(
			configurationPathFlag,
			cmd.WithRepFlagConfig(&r.ConfigurationPath),
		),
		pkg.Wrap(
			manifestPathFlag,
			cmd.WithRepFlagConfig(&r.ManifestPath),
		),
		pkg.Wrap(
			inlineDefinitionFlag,
			cmd.WithRepFlagConfig(&r.InlineDefinition),
		),
	)

	// This must be postponed until after the command-line is parsed.
	defaultManifest := []string{filepath.Join(config.Dir(ID), `default`)}

	r.Config = pkg.Wrap(
		r.Config,
		cmd.WithUsage(
			cmd.Usage{
				Name:      config.Prefix(ID),
				Syntax:    syntax,
				ShortHelp: shortHelp,
				LongHelp:  longHelp,
			},
			func(ctx context.Context, args []string) error {
				// Initialize the profiler with profiling mode and path provided
				// via the command-line flag.
				prof := pprof.Profiler{
					Mode:  r.Profile,
					Path:  filepath.Join(config.Cache(ID), profileFlag.LongName),
					Quiet: !r.Verbose,
				}

				// Note that profiling is only available when built with "-tags pprof".
				// Otherwise, this is a no-op.
				defer prof.Start().Stop()

				if r.Version {
					fmt.Println(ID, "version", pkg.Version)

					return nil
				}

				// Only set the default manifest file
				// if flag --ignore-default is unset.
				if !r.NoDefaultDefinition {
					r.ManifestPath = append(r.ManifestPath, defaultManifest...)
				}

				var (
					man manifest.Model
					err error
				)

				man, err = manifest.Make(
					ctx,
					r.ManifestPath,
					r.InlineDefinition,
					manifest.WithParallelEvalLimit(r.ParallelEvalLimit),
					manifest.WithStrictDefinitions(r.StrictDefinitions),
				)
				if err != nil {
					return fmt.Errorf("%w: %w", pkg.ErrInaccessibleManifest, err)
				}

				man, err = man.Parse()
				if err != nil {
					return pkg.IncompleteParseError{
						Err: err, Def: r.ManifestPath, Lvl: r.VerboseLevel(),
					}
				}

				if len(args) == 0 {
					args = config.DefaultNamespace()
				}

				env, err := man.Eval(ctx, args...)
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

func (r Node) VerboseLevel() int {
	if r.Verbose {
		return r.verboseLevel
	}

	return 0
}
