package model

import (
	"context"
	"strconv"
	"strings"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/cli/model/proto"
	"github.com/ardnew/envmux/config"
	"github.com/ardnew/envmux/pkg"
)

// Command represents the context and configuration of a command.
//
// It is not used as a command itself but as a container for actual commands.
// The types of actual commands are composed of this type via embedding.
//
// All Command methods are immutable.
// Any method that modifies its receiver will return a new Command.
//
// The With* functions returning a [pkg.Option] can be used with either
// [pkg.Wrap] to modify an existing Command or
// [pkg.Make] to create a new Command.
type Command struct {
	proto  proto.Type
	parent *Command
}

// Parse parses the command-line arguments.
func (c Command) Parse(args []string, opts ...ff.Option) error {
	return c.proto.Command.Parse(args, opts...)
}

// Run executes the command with the given context and configuration.
func (c Command) Run(ctx context.Context, cfg config.Model, args ...string) error {
	return c.proto.Run(cfg.AsContext(ctx))
}

func (c Command) Eval(ctx context.Context, args ...string) (map[string]string, error) {
	return c.proto.Eval(ctx, args...)
}

// IsZero checks if the Command is uninitialized.
func (c Command) IsZero() bool { return c.proto.IsZero() && c.parent == nil }

// Parent returns the parent Command.
//
// Nil is returned if the Command has no parent.
func (c Command) Parent() *Command { return c.parent }

// Config returns the parsed input configuration model.
// Use [config.Model.Err] to check for errors parsing the configuration.
func (c Command) Config() config.Model { return c.proto.Config() }

// Env returns the resolved environment variables.
func (c Command) Env() map[string]string { return c.proto.Env() }

// Environ returns the environment variables as a slice of strings
// in the form "key=value" suitable for use with os.Exec or os.Environ.
// If the environment is nil, an empty slice is returned.
func (c Command) Environ() []string {
	env := c.proto.Env()
	if env == nil {
		return []string{}
	}
	// Convert the map to a slice of strings in the form "key=value"
	result := make([]string, 0, len(env))
	for k, v := range env {
		var sb strings.Builder
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(strconv.Quote(v))
		result = append(result, sb.String())
	}
	return result
}

// Definition returns the command's underlying Command.
func (c Command) Definition() *ff.Command { return c.proto.Command }

// FlagSet returns the command's underlying FlagSet.
func (c Command) FlagSet() *ff.FlagSet { return c.proto.FlagSet }

// Flag returns the flag with the given name defined in the receiver
// or any of its ancestors.
//
// The second return value is true iff the flag is found.
func (c Command) Flag(name string) (ff.Flag, bool) {
	return c.proto.GetFlag(name)
}

func (c Command) FlagAsBool(name string) (bool, error) {
	f, ok := c.Flag(name)
	if !ok {
		return false, pkg.ErrInvalidFlag
	}
	b, err := strconv.ParseBool(f.GetValue())
	if err != nil {
		d, derr := strconv.ParseBool(f.GetDefault())
		if derr != nil {
			return false, pkg.ErrInvalidFlag
		}
		return d, nil
	}
	return b, nil
}

func (c Command) FlagAsString(name string) (string, error) {
	f, ok := c.Flag(name)
	if !ok {
		return "", pkg.ErrInvalidFlag
	}
	return f.GetValue(), nil
}

func (c Command) FlagAsInt(name string) (int, error) {
	f, ok := c.Flag(name)
	if !ok {
		return 0, pkg.ErrInvalidFlag
	}
	i, err := strconv.ParseInt(f.GetValue(), 0, 0)
	if err != nil {
		d, derr := strconv.ParseInt(f.GetDefault(), 0, 0)
		if derr != nil {
			return 0, pkg.ErrInvalidFlag
		}
		return int(d), nil
	}
	return int(i), nil
}

func (c Command) FlagAsInt64(name string) (int64, error) {
	f, ok := c.Flag(name)
	if !ok {
		return 0, pkg.ErrInvalidFlag
	}
	i, err := strconv.ParseInt(f.GetValue(), 0, 64)
	if err != nil {
		d, derr := strconv.ParseInt(f.GetDefault(), 0, 64)
		if derr != nil {
			return 0, pkg.ErrInvalidFlag
		}
		return d, nil
	}
	return i, nil
}

func (c Command) FlagAsUint(name string) (uint, error) {
	f, ok := c.Flag(name)
	if !ok {
		return 0, pkg.ErrInvalidFlag
	}
	u, err := strconv.ParseUint(f.GetValue(), 0, 0)
	if err != nil {
		d, derr := strconv.ParseUint(f.GetDefault(), 0, 0)
		if derr != nil {
			return 0, pkg.ErrInvalidFlag
		}
		return uint(d), nil
	}
	return uint(u), nil
}

func (c Command) FlagAsUint64(name string) (uint64, error) {
	f, ok := c.Flag(name)
	if !ok {
		return 0, pkg.ErrInvalidFlag
	}
	u, err := strconv.ParseUint(f.GetValue(), 0, 64)
	if err != nil {
		d, derr := strconv.ParseUint(f.GetDefault(), 0, 64)
		if derr != nil {
			return 0, pkg.ErrInvalidFlag
		}
		return d, nil
	}
	return u, nil
}

func (c Command) FlagAsFloat64(name string) (float64, error) {
	f, ok := c.Flag(name)
	if !ok {
		return 0, pkg.ErrInvalidFlag
	}
	f64, err := strconv.ParseFloat(f.GetValue(), 64)
	if err != nil {
		d, derr := strconv.ParseFloat(f.GetDefault(), 64)
		if derr != nil {
			return 0, pkg.ErrInvalidFlag
		}
		return d, nil
	}
	return f64, nil
}

// WithProto sets the common fields specifying the Command.
func WithProto(t proto.Type) pkg.Option[Command] {
	return func(c Command) Command {
		c.proto = t
		return c
	}
}

func WithConfig(cfg config.Model) pkg.Option[Command] {
	return func(c Command) Command {
		c.proto = pkg.Wrap(c.proto, proto.WithConfig(cfg))
		return c
	}
}

func WithEnv(env map[string]string) pkg.Option[Command] {
	return func(c Command) Command {
		c.proto = pkg.Wrap(c.proto, proto.WithEnv(env))
		return c
	}
}

// WithParent sets the parent Command.
func WithParent(ptr *Command) pkg.Option[Command] {
	return func(cmd Command) Command {
		if ptr != nil {
			p, c := ptr.proto, cmd.proto
			if p.Command != nil && c.Command != nil {
				p.Subcommands = append(p.Subcommands, c.Command)
			}
			if p.FlagSet != nil && c.FlagSet != nil {
				c.SetParent(p.FlagSet)
			}
		}
		cmd.parent = ptr
		return cmd
	}
}
