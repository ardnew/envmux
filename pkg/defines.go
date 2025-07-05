package pkg

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//nolint:gochecknoglobals
var (

	// InlineSourcePrefix is used as a prefix for command-line arguments to the
	// `--source` flag to indicate that the definition(s) are provided inline from
	// the argument itself and not read from a file.
	//
	// The flag can be used multiple times, but this prefix must be used for each
	// instance that contains inline definitions. Both inline and file-based
	// definitions can be mixed in the same command-line invocation.
	InlineSourcePrefix = "="

	// StdinSourcePath is the special path used to indicate that the source
	// definitions should be read from standard input (stdin).
	StdinSourcePath = "-"
)

// SourceFile is the default file(s) containing namespace definitions.
//
//nolint:gochecknoglobals
var SourceFile = func(cmd string) []string {
	return []string{filepath.Join(ConfigDir(cmd), "default")}
}

// Namespace is the default namespace(s) used for evaluation.
//
//nolint:gochecknoglobals
var Namespace = func() []string {
	return []string{`default`}
}

// ConfigPrefix returns the base prefix string used to construct the path to the
// configuration directory and the prefix for environment variable identifiers.
//
// By default, the prefix is the base name of the executable file
// unless it matches one of the following substitution rules:
//
//   - "__debug_bin" (default output of the dlv debugger): replaced with [ID]
//   - "^\.+" (dot-prefixed names): remove the dot prefix
//
//nolint:gochecknoglobals
var ConfigPrefix = func(cmd string) string {
	id := os.Args[0]
	if exe, err := os.Executable(); err == nil {
		id = exe
	}

	id = strings.TrimSuffix(filepath.Base(id), ".test")

	substitute := []struct {
		*regexp.Regexp
		string
	}{
		{regexp.MustCompile(`^__debug_bin\d+$`), cmd}, // default output from dlv
		{regexp.MustCompile(`^\.+`), ""},              // remove leading dots
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
//
//nolint:gochecknoglobals
var ConfigDir = func(cmd string) string {
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

	return filepath.Join(root, ConfigPrefix(cmd))
}
