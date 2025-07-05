package main

import (
	"context"
	"os"

	"github.com/ardnew/envmux/cmd/envmux/cli"
)

// main is the entry point for the envmux application.
func main() {
	var result cli.RunError

	ctx := context.Background()
	defer exit(ctx, &result)

	result = cli.Run(ctx)

	if result.Help != "" {
		println(result.Help)
	}

	if result.Err != nil {
		println("error:", result.Err.Error())
	}
}

func exit(ctx context.Context, runErr *cli.RunError) {
	ctx.Done()
	os.Exit(runErr.Code)
}
