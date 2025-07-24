package parse

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/ardnew/envmux/config/parse/stream"
	"github.com/ardnew/envmux/pkg"
)

const StringerAlwaysShowsMeta = true

// Namespace associates a composition of environment variable definitions with
// a Namespace identifier.
//
// Variable definitions are expressed entirely with the [expr-lang] grammar.
//
// [expr-lang]: https://github.com/expr-lang/expr
type Namespace struct {
	Ident string

	Composites []Composite
	Parameters []Parameter
	Statements []Statement
}

func (n Namespace) String() string {
	if n.Ident == "" {
		return ""
	}

	var com, par, sta string

	if len(n.Composites) > 0 {
		coms := make([]string, len(n.Composites))
		for i, c := range n.Composites {
			coms[i] = c.String()
		}

		com = strings.Join(coms, FS)
		if !StringerAlwaysShowsMeta {
			com = fmt.Sprintf("%s%s%s", co, com, cc)
		}
	}

	if len(n.Parameters) > 0 {
		pars := make([]string, len(n.Parameters))
		for i, p := range n.Parameters {
			pars[i] = p.String()
		}

		par = strings.Join(pars, FS)
		if !StringerAlwaysShowsMeta {
			par = fmt.Sprintf("%s%s%s", po, par, pc)
		}
	}

	if len(n.Statements) > 0 {
		stas := make([]string, len(n.Statements))
		for i, s := range n.Statements {
			stas[i] = s.String()
		}

		sta = strings.Join(stas, RS)
		if !StringerAlwaysShowsMeta {
			sta = fmt.Sprintf("%s%s%s", so, sta, sc)
		}
	}

	if StringerAlwaysShowsMeta {
		com = fmt.Sprintf("%s%s%s", co, com, cc)
		par = fmt.Sprintf("%s%s%s", po, par, pc)
		sta = fmt.Sprintf("%s%s%s", so, sta, sc)
	}

	return fmt.Sprintf("%s%s%s%s", n.Ident, com, par, sta)
}

// Arguments returns a [Parameter.Value] sequence of all [Namespace.Parameter]s.
func (n Namespace) Arguments() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, p := range n.Parameters {
			if !yield(p.Value) {
				return
			}
		}
	}
}

// func (n Namespace) Composites() iter.Seq[Composite] { return n.Com.Seq() }
// func (n Namespace) Parameters() iter.Seq[Parameter] { return n.Par.Seq() }
// func (n Namespace) Statements() iter.Seq[Statement] { return n.Sta.Seq() }

func namespaces(
	ctx context.Context,
	sg *stream.Group[stream.Token],
) stream.Group[Namespace] {
	var getType stream.TypeResolver = tokenType()

	var stage stream.Stage[Namespace] = func() (Namespace, error) {
		tok, err := sg.Accept(getType.Predicate(`NS`))

		switch {
		case errors.Is(err, pkg.ErrClosedStream):
			return Namespace{}, pkg.ErrEOF
		case errors.Is(err, pkg.ErrUnacceptableStream):
			return Namespace{}, pkg.UnexpectedTokenError{
				Tok: tok.Lexeme(),
				Msg: []string{`expected namespace identifier`},
			}
		}

		ns := Namespace{Ident: tok.Value} //nolint:exhaustruct

		for c := range composites(ctx, sg).Channel {
			ns.Composites = append(ns.Composites, c)
		}

		for p := range parameters(ctx, sg).Channel {
			ns.Parameters = append(ns.Parameters, p)
		}

		for s := range statements(ctx, ns.Ident, sg).Channel {
			ns.Statements = append(ns.Statements, s)
		}

		return ns, nil
	}

	return pkg.Make(stage.Pipe(ctx))
}

// tok, ok := <-sg.Chan
// if !ok {
// 	return namespace{}, pkg.ErrEOF
// }

// switch tok.Type {
// case getType(`NS`):
// 	ns := namespace{id: tok.Value}
// 	return ns, nil

// default:
// 	return namespace{}, nil
// }

// 	advance := consume(lex, `XX`)

// 	tok := lex.Next()
// 	if tok.Type != symbol()(`NS`) {
// 		return pkg.UnexpectedTokenError{
// 			Tok: tok,
// 			Msg: []string{`expected namespace identifier`},
// 		}
// 	}

// 	n.ID = tok.Value

// 	if !advance() {
// 		return nil
// 	}

// 	if lex.Peek().Type == symbol()(`CO`) {
// 		if err := n.Com.Parse(lex); err != nil {
// 			return err
// 		}

// 		if !advance() {
// 			return nil
// 		}
// 	}

// 	if !advance() {
// 		return nil
// 	}

// 	if lex.Peek().Type == symbol()(`PO`) {
// 		if err := n.Par.Parse(lex); err != nil {
// 			return err
// 		}

// 		if !advance() {
// 			return nil
// 		}

// 		if peek := lex.Peek(); peek.Type == symbol()(`CO`) {
// 			return pkg.UnexpectedTokenError{
// 				Tok: peek,
// 				Msg: []string{
// 					`composites "` + co + `…` + cc + `" must be declared before ` +
// 						`parameters "` + po + `…` + pc + `"`,
// 				},
// 			}
// 		}
// 	}

// 	if !advance() {
// 		return nil
// 	}

// 	if lex.Peek().Type == symbol()(`SO`) {
// 		if err := n.Sta.Parse(lex); err != nil {
// 			return err
// 		}

// 		if !advance() {
// 			return nil
// 		}

// 		//nolint:exhaustive
// 		switch peek := lex.Peek(); peek.Type {
// 		case symbol()(`PO`):
// 			return pkg.UnexpectedTokenError{
// 				Tok: peek,
// 				Msg: []string{
// 					`parameters "` + po + `…` + pc + `" must be declared before ` +
// 						`statements "` + so + `…` + sc + `"`,
// 				},
// 			}

// 		case symbol()(`CO`):
// 			return pkg.UnexpectedTokenError{
// 				Tok: peek,
// 				Msg: []string{
// 					`composites "` + co + `…` + cc + `" must be declared before ` +
// 						`statements "` + so + `…` + sc + `"`,
// 				},
// 			}
// 		}
// 	}

// 	return nil
// }
