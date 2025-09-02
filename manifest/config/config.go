package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Prefix returns the base prefix string used to construct the path to the
// configuration directory and the prefix for environment variable identifiers.
//
// By default, the prefix is the base name of the executable file.
// unless it matches one of the following substitution rules:
//
//   - "__debug_bin" (default output of the dlv debugger): replaced with [ID]
//   - "^\.+" (dot-prefixed names): remove the dot prefix
//
//nolint:gochecknoglobals
var Prefix = func(cmd string) string {
	id := os.Args[0]
	if exe, err := os.Executable(); err == nil {
		id = exe
	}

	id = strings.TrimSuffix(filepath.Base(id), ".test")

	for rex, rep := range map[*regexp.Regexp]string{
		regexp.MustCompile(`^__debug_bin\d+$`): cmd, // default output from dlv
		regexp.MustCompile(`^\.+`):             "",  // remove leading dot(s)
	} {
		id = rex.ReplaceAllString(id, rep)
	}

	return id
}

// Dir returns the configuration directory path.
//
// The directory path is constructed by appending [Prefix]
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
var Dir = func(cmd string) string {
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

	return filepath.Join(root, Prefix(cmd))
}

var Cache = func(cmd string) string {
	root, ok := os.LookupEnv("XDG_CACHE_HOME")
	if !ok {
		if root, ok = os.LookupEnv("HOME"); ok {
			root = filepath.Join(root, ".cache")
		} else {
			var err error
			if root, err = os.Getwd(); err != nil {
				root = "."
			}
		}
	}

	return filepath.Join(root, Prefix(cmd))
}

// StdinManifestPath is the special path used to indicate that the namespace
// definitions should be read from standard input (stdin).
//
//nolint:gochecknoglobals
var StdinManifestPath = "-"

// DefaultManifestPath is the default file(s) containing namespace definitions.
//
//nolint:gochecknoglobals
var DefaultManifestPath = func(cmd string) []string {
	return []string{filepath.Join(Dir(cmd), "default")}
}

// DefaultNamespace is the default namespace(s) used for evaluation.
//
//nolint:gochecknoglobals
var DefaultNamespace = func() []string { return []string{`default`} }
