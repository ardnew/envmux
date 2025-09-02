//go:build !pprof

package pprof

func Modes() []string { return nil }

type control = ignore

func start(string, string, bool) interface{ Stop() } { return control{} }
