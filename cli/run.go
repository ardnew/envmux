package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ardnew/groot/pkg"
	"github.com/ardnew/groot/pkg/env"
	"github.com/ardnew/groot/pkg/fs"
)

func Run() Result {
	var (
		ctx = context.Background()
		cfg = pkg.MakeConfig(filepath.Base(exe()))
		arg = os.Args[1:]
	)
	_ = env.New(&cfg)
	_ = fs.New(&cfg)

	return MakeResult(cfg, run(ctx, cfg, arg))
}

func run(ctx context.Context, cfg pkg.Config, arg []string) error {
	if err := cfg.Command.Parse(arg); err != nil {
		return err
	}
	if err := cfg.Command.Run(ctx); err != nil {
		return err
	}
	return nil
}

func exe() string {
	p, err := os.Executable()
	if err != nil {
		return os.Args[0]
	}
	return p
}
