package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/repr"

	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
)

var (
	getenvDebugOnce = sync.OnceValue(getenvDebug)

	source = map[string]string{
		"test": `
default <defs, pkgs> (
	"/usr",
	"/usr/local"
) {
	region = "us-west-2";
	start_date = "2025-03-02T14:02:48,793956264-06:00";
	end_date = "14.7h 2s";
	SHELL =  user.Username;
	bucket := 12;
	HOME? := "$HOME";
}

env1 <default, defs, pkgs> {
	PATH ?+= /* foo */ "/bin";
	MANPATH = {
		"/share/man" + "foo"     + { "qwq" }
	} | upper();
	INFOPATH ?^= "/share/info";
	LD_LIBRARY_PATH ^= "/lib";
	PKG_CONFIG_PATH += "/lib/pkgconfig";
}

env2 <default> ( "/root/.local" ) {	PATH = "/sbin"; }

defs { defs = "data"; }

pkgs ( "/opt/something", "/opt/another" )

// special (default, other) {
//   LD_LIBRARY_PATH = "/lib/x86_64-linux-gnu"
// } [
// 	"/opt"
// ]

`,
	}
)

func getenvDebug() (debug bool) {
	if truth, err := strconv.ParseBool(os.Getenv("DEBUG")); err == nil {
		debug = truth
	}
	return debug
}

type testTrace bool

func (t testTrace) Write(p []byte) (n int, err error) {
	if t {
		return os.Stderr.Write(p)
	}
	return len(p), nil
}

func TestParser(t *testing.T) {
	trace := testTrace(getenvDebugOnce())
	parse.ParseOptions = append(parse.ParseOptions, participle.Trace(trace))

	m := pkg.Make(WithReader(bufio.NewReader(strings.NewReader(source["test"]))))
	if m.Err() != nil {
		t.Fatalf("parse: %v", m.Err())
	}

	result, err := m.Env().Eval(t.Context(), []string{"env2"})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	t.Logf("Eval: %s", repr.String(result, repr.Indent(" ")))

	if trace {
		t.Logf("Parsed: %s", repr.String(m.Env(), repr.Indent(" ")))
	}
}
