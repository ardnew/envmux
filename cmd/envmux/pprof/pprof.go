//go:build pprof

package pprof

import (
	"maps"
	_ "net/http/pprof"
	"slices"

	"github.com/pkg/profile"

	"github.com/ardnew/envmux/pkg"
	"github.com/ardnew/envmux/pkg/fn"
)

// Modes returns the list of supported profiling modes when built with the
// pprof build tag. The special mode "quiet" is omitted from the list.
func Modes() []string {
	return slices.Collect(fn.Filter(
		maps.Keys(mode),
		func(k string) bool { return k != "quiet" },
	))
}

var mode = map[string]func(*profile.Profile){
	"block":     profile.BlockProfile,
	"cpu":       profile.CPUProfile,
	"clock":     profile.ClockProfile,
	"goroutine": profile.GoroutineProfile,
	"mem":       profile.MemProfile,
	"allocs":    profile.MemProfileAllocs,
	"heap":      profile.MemProfileHeap,
	"mutex":     profile.MutexProfile,
	"thread":    profile.ThreadcreationProfile,
	"trace":     profile.TraceProfile,
	// "quiet":     profile.Quiet,
}

type control struct {
	mode []func(*profile.Profile)
}

func start(mode, path string, quiet bool) interface{ Stop() } {
	c := pkg.Make(withMode(mode))

	if len(c.mode) == 0 {
		return ignore{}
	}

	return profile.Start(
		pkg.Wrap(c, withPath(path), withQuiet(quiet)).mode...,
	)
}

func withMode(m string) pkg.Option[control] {
	return func(c control) control {
		if fn, ok := mode[m]; ok {
			c.mode = append(c.mode, fn)
		}

		return c
	}
}

func withPath(p string) pkg.Option[control] {
	return func(c control) control {
		if p != "" {
			c.mode = append(c.mode, profile.ProfilePath(p))
		}
		return c
	}
}

func withQuiet(v bool) pkg.Option[control] {
	return func(c control) control {
		if v {
			c.mode = append(c.mode, profile.Quiet)
		}
		return c
	}
}
