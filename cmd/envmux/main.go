package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/ardnew/envmux/cmd/envmux/cli"
	"github.com/ardnew/envmux/pkg/log"
)

// main is the entry point for the envmux application.
func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	ctx = log.Make().AddToContext(ctx)

	cancel(cli.Run(ctx))

	exit(ctx)
}

func exit(ctx context.Context) {
	<-ctx.Done()

	var runErr cli.RunError
	if errors.As(context.Cause(ctx), &runErr) {
		switch {
		case runErr.Help != "":
			println(runErr.Help)

		case runErr.Err != nil:
			if jot, ok := log.FromContext(ctx); ok {
				jot.LogAttrs(ctx, slog.LevelError, "unhandled error",
					slog.Attr{Key: "error", Value: slog.StringValue(runErr.Err.Error())},
				)
			}
		}

		os.Exit(runErr.Code)
	}

	os.Exit(-1)
}
