package config

import (
	"testing"

	"github.com/alecthomas/repr"
	"github.com/stretchr/testify/require"
)

func TestMakeParser(t *testing.T) {
	parser := MakeParser()

	ast, err := parser.Parser.ParseString("", `
default {
	region = "us-west-2"
	start_date = "2025-03-02T14:02:48,793956264-06:00"
	end-date = "14.7h 2s"
	shell =  << .User >>
	bucket = backups
	HOME = << .User.HomeDir >>
} [
	/usr
	/usr/local
]

env1: foo, extra {
  PATH = "/bin"
	MANPATH = "/share/man"
	INFOPATH = "/share/info"
	LD_LIBRARY_PATH = /lib
	PKG_CONFIG_PATH = /lib/pkgconfig
}

extra { EXTRA = << "data" >> }

packs [ "/opt/something", "/opt/another" ]

// special: default, other {
//   LD_LIBRARY_PATH = "/lib/x86_64-linux-gnu"
// } [
// 	"/opt"
// ]

`,
	// participle.Trace(os.Stderr),
	)
	repr.Println(ast)
	require.NoError(t, err)
}
