// Package prof provides profiling capabilities for the module.
package prof

func Init(arg ...string) interface{ Stop() } { return control{}.start(arg...) }
