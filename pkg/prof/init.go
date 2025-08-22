// Package prof provides profiling capabilities for the module.
package prof

// Init initializes the profiler and returns an interface for stopping it.
//
// If build tag pprof is unset, starting and stopping the profiler is a no-op.
func Init(arg ...string) interface{ Stop() } {
	return control{}.start(arg...)
}
