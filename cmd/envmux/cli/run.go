package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/ardnew/envmux/cmd/envmux/cli/cmd"
	"github.com/ardnew/envmux/cmd/envmux/cli/cmd/root"
	"github.com/ardnew/envmux/pkg"
	"github.com/ardnew/envmux/pkg/log"
)

// Run executes the root command and returns the resulting [RunError]. It uses
// the provided [context.Context] for cancellation and deadlines.
func Run(ctx context.Context) int {
	ctx, cancel := log.Make().AddToContextCancelCause(ctx)

	node := root.Init()
	stat := node.Command().ParseAndRun(
		ctx, os.Args[1:], cmd.FlagOptions()...,
	)

	cancel(MakeResult(node, stat))

	return yield(ctx)
}

func yield(ctx context.Context) int {
	<-ctx.Done()

	var runErr RunError
	if errors.As(context.Cause(ctx), &runErr) {
		switch {
		case runErr.Help != "":
			println(runErr.Help)

		case runErr.Err != nil:
			if jot, ok := log.FromContext(ctx); ok {
				if attrErr, ok := runErr.Err.(pkg.Attributed); ok {
					jot.LogAttrs(ctx, slog.LevelError, runErr.Error(),
						pkg.Attributes(attrErr)...)

					for _, s := range attrErr.Details() {
						fmt.Fprintf(log.DefaultOutput, "|\t%s\n", s)
					}
				} else {
					jot.LogAttrs(ctx, slog.LevelError, "unhandled error",
						slog.Attr{Key: "error", Value: slog.StringValue(runErr.Error())},
					)
				}
			}
		}

		return runErr.Code
	}

	return -1
}
