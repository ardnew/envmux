//go:build pprof

package prof

import (
	"maps"
	"slices"
	"strings"

	"github.com/pkg/profile"

	"github.com/ardnew/envmux/pkg"
)

func Modes() []string {
	return slices.Collect(pkg.Filter(
		maps.Keys(mode),
		func(k string) bool { return k != "quiet" },
	))
}

var (
	defaultMode = profile.TraceProfile

	mode = map[string]func(*profile.Profile){
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
		"quiet":     profile.Quiet,
	}
)

type control struct {
	mode []func(*profile.Profile)
	path string
}

func (c control) start(args ...string) interface{ Stop() } {
	opt := []pkg.Option[control]{}
	arg := []string{}

	for _, p := range args {
		mode, marg, cut := strings.Cut(p, `=`)

		if arg = append(arg, mode); cut {
			opt = append(opt, withhPath(marg))
		}
	}

	c = pkg.Wrap(c, append(opt, withhMode(arg...))...)

	if c.path != "" {
		c.mode = append(c.mode, profile.ProfilePath(c.path))
	}

	return profile.Start(c.mode...)
}

func withhMode(modes ...string) pkg.Option[control] {
	return func(c control) control {
		seen := pkg.Unique[string]{}

		for _, m := range modes {
			if fn, ok := mode[m]; ok && seen.Set(m) {
				c.mode = append(c.mode, fn)
			}
		}

		if len(seen) == 0 && len(modes) > 0 {
			c.mode = append(c.mode, defaultMode)
		}

		return c
	}
}

func withhPath(path string) pkg.Option[control] {
	return func(c control) control {
		c.path += path
		return c
	}
}
