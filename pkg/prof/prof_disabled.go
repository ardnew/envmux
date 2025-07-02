//go:build !pprof

package prof

func Modes() []string { return []string{} }

type control struct{}

func (control) Stop() {}

func (c control) start(...string) interface{ Stop() } { return c }
