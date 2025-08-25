package runtime

import (
	"os"
	"runtime"
	"strings"
)

// Target contains string identifiers for a target operating system and
// instruction set architecture.
//
// Leaving the conventions unspecified allows this type to be used
// in a variety of contexts.
type Target struct {
	OS   string
	Arch string
}

// GetTarget returns the [Target] with GNU GCC/LLVM conventions.
func GetTarget() Target {
	t := GetPlatform()
	switch t.Arch {
	case "386":
		t.Arch = "i386"
	case "amd64":
		t.Arch = "x86_64"
	case "arm":
		arm, ok := os.LookupEnv("GOARM")
		if ok {
			arm, _, _ = strings.Cut(arm, ",")
			switch strings.TrimSpace(arm) {
			case "5", "6", "7":
				t.Arch = "armv" + arm
			}
		}
	case "arm64":
		if t.OS != "darwin" {
			t.Arch = "aarch64"
		}
	case "mipsle":
		t.Arch = "mipsel"
	}

	return t
}

// GetPlatform returns the [Target] with [Go conventions].
//
// [Go conventions]:
// https://cs.opensource.google/go/go/+/master:src/cmd/dist/build.go
func GetPlatform() Target {
	var (
		o, a string
		ok   bool
	)

	if o, ok = os.LookupEnv("GOHOSTOS"); !ok {
		if o, ok = os.LookupEnv("GOOS"); !ok {
			o = runtime.GOOS
		}
	}

	if a, ok = os.LookupEnv("GOHOSTARCH"); !ok {
		if a, ok = os.LookupEnv("GOARCH"); !ok {
			a = runtime.GOARCH
		}
	}

	return Target{
		OS:   o,
		Arch: a,
	}
}
