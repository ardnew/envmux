package main

import (
	"os"

	"github.com/ardnew/groot/cli"
)

// main is the entry point for the groot application.
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
