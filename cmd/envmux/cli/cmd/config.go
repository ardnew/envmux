package cmd

import (
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffval"

	"github.com/ardnew/envmux/pkg"
)

// ID is the identifier of the command-line application.
// It is used to construct the configuration prefix and as the default
// identifier for the environment variable prefix.
const ID = pkg.Name

//nolint:gochecknoglobals
var (
	// ConfigFlag is the flag name used to specify the configuration file.
	ConfigFlag = "config"

	// FlagOptions returns the options for parsing the command-line arguments.
	FlagOptions = func() []ff.Option {
		return []ff.Option{
			ff.WithConfigFileFlag(ConfigFlag),
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithConfigAllowMissingFile(),
			ff.WithEnvVarPrefix(pkg.FormatEnvVar(pkg.ConfigPrefix(ID))),
			// ff.WithEnvIgnoreShortVarNames(),
		}
	}
)

func WithFlagConfig[T ffval.ValueType](ptr *T) pkg.Option[ff.FlagConfig] {
	return func(cfg ff.FlagConfig) ff.FlagConfig {
		cfg.Value = ffval.NewValueDefault(ptr, *ptr)

		return cfg
	}
}

// WithIncFlagConfig is similar to WithFlagConfig, but it enables counting
// the number of times the flag is set.
func WithIncFlagConfig[T ffval.ValueType](
	ptr *T,
	counter *int,
) pkg.Option[ff.FlagConfig] {
	return func(cfg ff.FlagConfig) ff.FlagConfig {
		val := ffval.NewValueDefault(ptr, *ptr)

		// Capture the original ParseFunc initialized in NewValueDefault above.
		parse := val.ParseFunc

		// Wrap the original ParseFunc with a counter increment.
		val.ParseFunc = func(s string) (T, error) {
			if counter != nil {
				*counter++
			}

			return parse(s) // Call the original ParseFunc to parse the value.
		}

		// Return a FlagConfig with the modified Value.
		cfg.Value = val

		return cfg
	}
}

func WithRepFlagConfig[T ffval.ValueType](
	slice *[]T,
) pkg.Option[ff.FlagConfig] {
	return func(cfg ff.FlagConfig) ff.FlagConfig {
		ptr := new(T)
		val := ffval.NewValue(ptr)

		// Capture the original ParseFunc initialized in NewValueDefault above.
		parse := val.ParseFunc

		// Wrap the original ParseFunc with a counter increment.
		val.ParseFunc = func(s string) (T, error) {
			// Call the original ParseFunc to parse the value.
			v, err := parse(s)
			if err == nil {
				*slice = append(*slice, v)
			}

			return v, err
		}

		// Return a FlagConfig with the modified Value.
		cfg.Value = val

		return cfg
	}
}
