package cli

import (
	"context"

	"github.com/ardnew/groot/pkg/model/cmd"
)

// Run executes the root command and returns the result.
func Run() Result {
	b := cmd.Make()

	return MakeResult(b.Command, b.Run(context.Background()))
}
