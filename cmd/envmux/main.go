package main

import (
	"context"
	"os"

	"github.com/ardnew/envmux/cmd/envmux/cli"
)

// exit provides indirection for testing purposes.
//
//nolint:gochecknoglobals
var exit = os.Exit

// main is the entry point for the envmux application.
func main() {
	exit(cli.Run(context.Background()))
}
