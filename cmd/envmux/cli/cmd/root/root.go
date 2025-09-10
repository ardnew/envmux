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

// Init constructs and returns the root command node.
func Init() Node { return new(Node).Init().(Node) } //nolint:forcetypeassert

// ID is the canonical root command name.
const ID = pkg.Name

const (
	syntax    = ID + ` [flags] [subcommand ...]`
	shortHelp = `namespaced environments`
	longHelp  = ID + ` constructs and evaluates static namespaced environments.`
)

//nolint:gochecknoglobals,exhaustruct
var (
	versionFlag = ff.FlagConfig{
		// ShortName:     'V',
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
	isolateDefinitionsFlag = ff.FlagConfig{
		ShortName:     'i',
		LongName:      `isolate`,
		Usage:         `omit default global namespace definitions`,
		NoPlaceholder: true,
		NoDefault:     true,
	}
	strictDefinitionsFlag = ff.FlagConfig{
		ShortName:     's',
		LongName:      `strict`,
		Usage:         `evaluate undefined namespaces as runtime errors`,
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
		Usage:         `read default command-line flags from newline-delimited file`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	manifestPathFlag = ff.FlagConfig{
		ShortName:     'm',
		LongName:      `manifest`,
		Usage:         `read namespace definitions from manifest file ("-" is stdin)`,
		Placeholder:   `FILE`,
		NoPlaceholder: false,
		NoDefault:     false,
	}
	inlineDefinitionFlag = ff.FlagConfig{
		ShortName:     'd',
		LongName:      `define`,
		Usage:         `append inline namespace definition(s) to manifest`,
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

	Version            bool
	Verbose            bool
	IsolateDefinitions bool
	StrictDefinitions  bool
	ParallelEvalLimit  int
	ConfigurationPath  []string
	ManifestPath       []string
	InlineDefinition   []string
	Profile            string

	verboseLevel int
}

func (r Node) Init(...any) cmd.Node { //nolint:ireturn
	r = Node{ //nolint:exhaustruct
		Version:           false,
		Verbose:           false,
		StrictDefinitions: false,
		ParallelEvalLimit: runtime.NumCPU(),
		ConfigurationPath: []string{
			filepath.Join(config.Dir(ID), configurationPathFlag.LongName),
		},
	}

	flags := append(
		fn.FilterItems(
			[]ff.FlagConfig{
				pkg.Wrap(profileFlag, cmd.WithFlagConfig(&r.Profile)),
			},
			func(ff.FlagConfig) bool { return len(pprof.Modes()) > 0 },
		),
		pkg.Wrap(
			versionFlag,
			cmd.WithFlagConfig(&r.Version),
		),
		pkg.Wrap(
			verboseLevelFlag,
			cmd.WithIncFlagConfig(&r.Verbose, &r.verboseLevel),
		),
		pkg.Wrap(
			isolateDefinitionsFlag,
			cmd.WithFlagConfig(&r.IsolateDefinitions),
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
				//
				// Note that profiling is only available when built with "-tags pprof".
				// Otherwise, Start and Stop are both no-op (safe to call).
				defer pprof.Profiler{
					Mode:  r.Profile,
					Path:  filepath.Join(config.Cache(ID), profileFlag.LongName),
					Quiet: !r.Verbose,
				}.Start().Stop()

				if r.Version {
					fmt.Println(ID, "version", pkg.Version)

					return nil
				}

				// Always add the default manifest file unless flag --isolate is set.
				if !r.IsolateDefinitions {
					r.ManifestPath = append(
						r.ManifestPath,
						config.DefaultManifestPath(ID)...,
					)
				}

				if len(args) == 0 {
					args = config.DefaultNamespace()
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
					return pkg.JoinErrors(pkg.ErrInaccessibleManifest, err)
				}

				man, err = man.Parse()
				if err != nil {
					return err
				}

				env, err := man.Eval(ctx, args...)
				if err != nil {
					return err
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

// VerboseLevel returns the number of -v flags specified on the command line.
func (r Node) VerboseLevel() int {
	if r.Verbose {
		return r.verboseLevel
	}

	return 0
}
