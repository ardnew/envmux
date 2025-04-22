package env

import (
	"bufio"
	"context"
	"fmt"
	"maps"
	"os"
	"os/user"
	"runtime"
	"strings"
	"sync"

	"github.com/ardnew/envmux/pkg"
)

var (
	SubjectKey = `@`
	ContextKey = `ctx`
)

type VarMap map[string]any

type VarEditMode int

const (
	InsertMode VarEditMode = 1 << iota
	ReplaceMode
)

func varEditModeWithReplace(replace bool) VarEditMode {
	if replace {
		return ReplaceMode
	}
	return InsertMode
}

type Target struct {
	OS   string
	Arch string
}

var (
	// We cache makeCachedVars to avoid repeatedly evaluating the environment.
	//
	// The returned Context is a reference, however, so callers can modify the
	// value returned from subsequent calls to makeCachedVars. This would be bad.
	//
	// Instead, use [makeContext] to get a safely-mutable copy of makeCachedVars.
	makeCachedVars = sync.OnceValue(func() VarMap {
		return VarMap{
			"error":    error(nil),
			"target":   getTarget(),
			"platform": getPlatform(),
			"hostname": getHostname(),
			"user":     getUser(),
			"shell":    getShell(),
		}
	})
	// cloneCachedVars returns all content of the cached VarMap.
	//
	// Shallow-copying the cached VarMap is less expensive than evaluating a new
	// instance each time.
	cloneCachedVars = func() VarMap { return maps.Clone(makeCachedVars()) }
)

func MakeVarMap(ctx context.Context, opt ...pkg.Option[VarMap]) VarMap {
	return pkg.Wrap(cloneCachedVars().WithContext(ctx), opt...)
}

func (v VarMap) Err() error {
	if err, ok := v["error"]; ok && err != nil {
		if err, ok := err.(error); ok {
			return err
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

func varMapError(format string, args ...any) func(VarMap) VarMap {
	err := fmt.Errorf(format, args...)
	return func(VarMap) VarMap {
		return VarMap{
			"error": fmt.Errorf("%w: %w", pkg.ErrInvalidEnvVarMap, err),
		}
	}
}

func (v VarMap) set(mode VarEditMode, key string, value any, readonly ...VarMap) VarMap {
	key = strings.TrimSpace(key)
	if key == "" {
		return varMapError("empty symbol=%q", key)(v)
	}
	if mode != ReplaceMode {
		if _, ok := v[key]; ok {
			return varMapError("duplicate symbol=%q", key)(v)
		}
	}
	for _, ro := range readonly {
		if ro == nil {
			continue
		}
		if _, ok := ro[key]; ok {
			return varMapError("read-only symbol=%q", key)(v)
		}
	}
	v[key] = value
	return v
}

func (v VarMap) WithSubject(subject string) VarMap {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return varMapError("empty subject")(v)
	}
	return v.set(ReplaceMode, SubjectKey, subject)
}

func (v VarMap) WithContext(ctx context.Context) VarMap {
	if ctx == nil {
		return varMapError("nil context.Context")(v)
	}
	return v.set(ReplaceMode, ContextKey, ctx)
}

func (v VarMap) Add(mode VarEditMode, key string, value any) VarMap {
	// Check base context only; we can overwrite user-defined symbols.
	return v.set(mode, key, value, cloneCachedVars())
}

func (v VarMap) AddEnv(mode VarEditMode, env ...map[string]string) VarMap {
	for _, e := range env {
		for key, val := range e {
			v = v.Add(mode, key, val)
		}
	}
	return v
}

func (v VarMap) Del(key ...string) VarMap {
	for _, k := range key {
		k = strings.TrimSpace(k)
		if k != "" {
			delete(v, k)
		}
	}
	return v
}

func getTarget() Target {
	return Target{
		OS:   getTargetOS(),
		Arch: getTargetArch(),
	}
}

func getPlatform() Target {
	return Target{
		OS:   getPlatformOS(),
		Arch: getPlatformArch(),
	}
}

func getPlatformOS() string {
	return runtime.GOOS
}

func getPlatformArch() string {
	return runtime.GOARCH
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func getUser() *user.User {
	user, err := user.Current()
	if err != nil {
		return nil
	}
	return user
}

func getShell() string {
	shell, ok := os.LookupEnv("SHELL")
	if ok {
		return shell
	}
	u := getUser()
	if u == nil || u.Username == "" {
		return ""
	}
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := s.Text()
		e := strings.Split(l, ":")
		if len(e) > 6 && e[0] == u.Username {
			return e[6]
		}
	}
	return ""
}

func getTargetOS() string {
	return getPlatformOS()
}

func getTargetArch() string {
	arch := getPlatformArch()
	switch arch {
	case "386":
		return "i386"
	case "amd64":
		return "x86_64"
	case "arm":
		arm, ok := os.LookupEnv("GOARM")
		if ok {
			arm, _, _ = strings.Cut(arm, ",")
			switch strings.TrimSpace(arm) {
			case "5", "6", "7":
				return "armv" + arm
			}
		}
	case "arm64":
		if getPlatformOS() != "darwin" {
			return "aarch64"
		}
	case "mipsle":
		return "mipsel"
	}
	return arch
}
