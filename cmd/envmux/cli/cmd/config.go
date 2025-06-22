package cmd

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/peterbourgon/ff/v4"

	"github.com/ardnew/envmux/pkg"
)

// ID is the identifier of the command-line application.
// It is used to construct the configuration prefix and as the default
// identifier for the environment variable prefix.
const ID = pkg.Name

// ConfigFlag is the flag name used to specify the configuration file.
var ConfigFlag = "config"

// FlagOptions returns the options for parsing the command-line arguments.
var FlagOptions = func() []ff.Option {
	return []ff.Option{
		ff.WithConfigFileFlag(ConfigFlag),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithConfigAllowMissingFile(),
		ff.WithEnvVarPrefix(pkg.FormatEnvVar(ConfigPrefix())),
		// ff.WithEnvIgnoreShortVarNames(),
	}
}

// ConfigPrefix returns the base prefix string used to construct the path to the
// configuration directory and the prefix for environment variable identifiers.
//
// By default, the prefix is the base name of the executable file
// unless it matches one of the following substitution rules:
//
//   - "__debug_bin" (default output of the dlv debugger): replaced with [ID]
//   - "^\.+" (dot-prefixed names): remove the dot prefix
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
		{regexp.MustCompile(`^__debug_bin\d+$`), ID}, // default output from dlv
		{regexp.MustCompile(`^\.+`), ""},             // remove leading dots
	}
	for _, sub := range substitute {
		id = sub.ReplaceAllString(id, sub.string)
	}

	return id
}

// ConfigDir returns the configuration directory path.
//
// The directory path is constructed by appending [ConfigPrefix]
// to the user's default configuration directory.
//
// The user's default configuration directory is the first directory found
// in the following order:
//
//  1. Environment variable XDG_CONFIG_HOME
//  2. Environment variable HOME, with ".config" appended
//  3. Current working directory
//
// Otherwise, none of these directories can be determined,
// and `filepath.Join(".", ConfigPrefix())` is returned.
var ConfigDir = func() string {
	root, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		if root, ok = os.LookupEnv("HOME"); ok {
			root = filepath.Join(root, ".config")
		} else {
			var err error
			if root, err = os.Getwd(); err != nil {
				root = "."
			}
		}
	}

	return filepath.Join(root, ConfigPrefix())
}
