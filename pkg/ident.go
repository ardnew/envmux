//nolint:gochecknoglobals
package pkg

import (
	_ "embed"
)

//go:embed VERSION
var Version string

const (
	Name        = "envmux"
	Description = Name + " is a tool for managing environments."
)

var Author = []struct {
	Name, Email string
}{
	{"ardnew", "andrew@ardnew.com"},
}
