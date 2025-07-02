// Package cli provides the entry point and configuration
// for the main command-line application.
package cli

import (
	"context"
	"os"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root"
)

// Run executes the root command and returns the result.
// Entry point for the CLI application.
func Run() RunError {
	node := root.Init()

	return MakeResult(node, node.Command().ParseAndRun(
		context.Background(), os.Args[1:], cmd.FlagOptions()...,
	))
}
