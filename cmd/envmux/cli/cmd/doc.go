// Package cmd defines abstractions and helpers for CLI subcommands.
// It provides a small framework on top of [ff] to build command trees.
//
// The primary abstraction is [Node], which wraps an [ff.Command] and its
// [ff.FlagSet] along with composable options to build CLI trees.
//
// [ff]: https://github.com/peterbourgon/ff
package cmd
