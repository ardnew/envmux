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
func Run(ctx context.Context) RunError {
	node := root.Init()

	return MakeResult(node, node.Command().ParseAndRun(
		ctx, os.Args[1:], cmd.FlagOptions()...,
	))
}
