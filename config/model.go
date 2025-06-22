package config

import (
	"context"
	"io"

	"github.com/ardnew/envmux/config/env"
	"github.com/ardnew/envmux/config/env/vars"
	"github.com/ardnew/envmux/config/parse"
	"github.com/ardnew/envmux/pkg"
)

type Model struct {
	env env.Model
	err error
}

func WithReader(r io.Reader) pkg.Option[Model] {
	return func(m Model) Model {
		var ns *parse.AST

		ns, m.err = parse.Load(r)
		if m.err != nil {
			return m
		}

		m.env = pkg.Make(env.WithAST(ns))

		return m
	}
}

func (m Model) Env() env.Model { return m.env }
func (m Model) Err() error     { return m.err }

func (m Model) Eval(
	ctx context.Context,
	namespace ...string,
) (vars.Env[string], error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.env.Eval(ctx, namespace...)
}
