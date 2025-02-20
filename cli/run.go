package cli

import (
	"context"

	"github.com/ardnew/groot/pkg/model/cmd"
)

func Run() Result {
	b := cmd.Make()

	return MakeResult(b.Command, b.Run(context.Background()))
}
