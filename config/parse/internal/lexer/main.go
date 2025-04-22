package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ardnew/envmux/config/parse"
)

func main() {
	gen := parse.LexerGenerator()
	b, err := gen.MarshalJSON()
	if err != nil {
		panic(err)
	}
	if len(os.Args) < 2 {
		panic("missing argument: output file (JSON)")
	}
	arg, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	cmd := exec.CommandContext(context.Background(), "jq")
	cmd.Stdin = bytes.NewReader(b)
	out, err := os.Create(arg)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
