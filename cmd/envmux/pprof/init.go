// Package pprof provides profiling capabilities for the module.
package pprof

import (
	"strings"
)

// Profiler configures and initializes the profiler.
type Profiler struct {
	Mode  string
	Path  string
	Quiet bool
}

// Start initializes the profiler and returns an interface for stopping it.
//
// Mode specifies the profiler mode to use, and path specifies the default
// output directory where profiling data will be written.
// If mode contains a delimiter "=", then the LHS is used as profiling mode,
// and the RHS overrides the output directory.
//
// If build tag pprof is unset, starting and stopping the profiler is a no-op.
func (p Profiler) Start() interface{ Stop() } {
	// If no mode is specified, do nothing.
	if p.Mode == "" {
		return ignore{}
	}

	// If mode contains a delimiter "=", then the LHS is used as profiling mode,
	// and the RHS overrides the output directory.
	if mode, path, ok := strings.Cut(p.Mode, `=`); ok {
		p.Mode, p.Path = mode, path
	}

	return start(p.Mode, p.Path, p.Quiet)
}

type ignore struct{}

func (ignore) Stop() {}
