package config

import (
	"io"

	"github.com/ardnew/envmux/config/env"
	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
)

type Model struct {
	env env.Model
	err error
}

func WithReader(r io.Reader) pkg.Option[Model] {
	return func(m Model) Model {
		var ns *parse.Namespaces
		ns, m.err = parse.Build(r)
		if m.err != nil {
			return m
		}
		m.env = pkg.Make(env.WithNamespaces(ns))
		return m
	}
}

func (m Model) Env() env.Model { return m.env }
func (m Model) Err() error     { return m.err }
