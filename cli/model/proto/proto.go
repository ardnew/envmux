package proto

import (
	"context"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/config"
	"github.com/ardnew/envmux/pkg"
)

// Interface is the interface used by all command-line (sub)commands.
type Interface interface {
	Name() string
	Syntax() string
	Help() (short, long string)
	Exec(ctx context.Context, args []string) error
}

// Type defines common content shared by all command-line (sub)commands.
type Type struct {
	*ff.Command
	*ff.FlagSet

	config   config.Model
	defaults []string
	env      map[string]string
}

// IsZero checks if the receiver is uninitialized.
func (t Type) IsZero() bool { return t.Command == nil && t.FlagSet == nil }

// Config returns the parsed input configuration model.
// Use [config.Model.Err] to check for errors parsing the configuration.
func (t Type) Config() config.Model { return t.config }

// Env returns the resolved environment variables.
func (t Type) Env() map[string]string { return t.env }

func (t Type) Eval(ctx context.Context, args ...string) (map[string]string, error) {
	cfg, ok := config.FromContext(ctx)
	if !ok {
		return nil, pkg.ErrInvalidConfigFile
	}
	t = pkg.Wrap(t, WithConfig(cfg))
	if t.config.Env().IsZero() {
		return nil, pkg.ErrInvalidConfig
	}
	if t.config.Err() != nil {
		return nil, t.config.Err()
	}
	if len(args) == 0 {
		args = t.defaults
	}
	return t.config.Eval(ctx, args...)
}

// Make returns a new Type initialized with the given options.
//
// The Type passed to each Option is fully-initialized
// according to the Interface type parameter.
func Make(impl Interface, opts ...pkg.Option[Type]) Type {
	// This Option must always be the first applied to a Type.
	withInterface := func(impl Interface) pkg.Option[Type] {
		return func(t Type) Type {
			// Configure default options
			shortHelp, longHelp := impl.Help()
			t.FlagSet = ff.NewFlagSet(impl.Name())
			t.Command = &ff.Command{
				Name:      impl.Name(),
				ShortHelp: shortHelp,
				LongHelp:  longHelp,
				Usage:     impl.Syntax(),
				Flags:     t.FlagSet,
				Exec:      impl.Exec,
			}
			return t
		}
	}
	// Ensure the [Type] is initialized before applying any options.
	t := pkg.Make(
		withInterface(impl),
		WithDefaults(pkg.DefaultNamespace...),
	)
	return pkg.Wrap(t, opts...)
}

// WithConfig sets the parsed input configuration model.
func WithConfig(cfg config.Model) pkg.Option[Type] {
	return func(t Type) Type {
		t.config = cfg
		return t
	}
}

// WithDefaults sets the default namespaces.
func WithDefaults(defaults ...string) pkg.Option[Type] {
	return func(t Type) Type {
		t.defaults = defaults
		return t
	}
}

// WithEnv sets the resolved environment variables.
func WithEnv(env map[string]string) pkg.Option[Type] {
	return func(t Type) Type {
		t.env = env
		return t
	}
}
