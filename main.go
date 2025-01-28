package main

import (
	"os"

	"github.com/ardnew/groot/cli"
)

func main() {
	result := cli.Run()
	if result.Help != "" {
		println(result.Help)
	}
	if result.Err != nil {
		println("error:", result.Err.Error())
	}
	os.Exit(result.Code)
}
