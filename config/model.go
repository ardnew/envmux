package config

import (
	"context"
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
		var ns *parse.AST
		ns, m.err = parse.Build(r)
		if m.err != nil {
			return m
		}
		m.env = pkg.Make(env.WithAST(ns))
		return m
	}
}

func (m Model) Env() env.Model { return m.env }
func (m Model) Err() error     { return m.err }

func (m Model) Eval(ctx context.Context, namespace ...string) (map[string]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.env.Eval(ctx, namespace...)
}

// contextKey is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type contextKey int

// key is the key for user.User values in Contexts. It is
// unexported; clients use user.NewContext and user.FromContext
// instead of using this key directly.
var key contextKey

func (m Model) AsContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

func FromContext(ctx context.Context) (Model, bool) {
	m, ok := ctx.Value(key).(Model)
	return m, ok
}
