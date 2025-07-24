package parse

import (
	"context"
	"errors"
	"strings"

	"github.com/ardnew/envmux/config/parse/stream"
	"github.com/ardnew/envmux/pkg"
)

// Expression contains the text of an Expression recognized by the
// [github.com/expr-lang/expr] grammar.
type Expression struct {
	Src string
}

func (e *Expression) String() string {
	if e == nil {
		return ""
	}

	return e.Src
}

func expressions(
	ctx context.Context,
	ns, id, op string,
	sg *stream.Group[stream.Token],
) stream.Group[Expression] {
	var getType stream.TypeResolver = tokenType()

	var bracketDepth int

	// Include a predicate that ensures brackets are balanced.
	balancedBracket := bracketBalancer(&bracketDepth, bracketAngles)

	accept := func(tok stream.Token) bool {
		return balancedBracket(tok) &&
			true // TBD: other predicates
	}

	var stage stream.Stage[Expression] = func() (Expression, error) {
		var sb strings.Builder
		for {
			tok, err := sg.Accept(accept)

			if bracketDepth > 0 {
				switch {
				case errors.Is(err, pkg.ErrClosedStream):
					return Expression{Src: sb.String()}, pkg.ErrUnexpectedEOF
				case errors.Is(err, pkg.ErrUnacceptableStream):
					return Expression{Src: sb.String()}, pkg.ExpressionError{
						Namespace: ns,
						Statement: id,
						Err: pkg.UnexpectedTokenError{
							Tok: tok.Lexeme(),
							Msg: []string{`unbalanced brackets in expression`},
						},
					}
				}
			} else {
				switch {
				case errors.Is(err, pkg.ErrClosedStream):
					return Expression{Src: sb.String()}, pkg.ErrEOF
				case errors.Is(err, pkg.ErrUnacceptableStream):
					return Expression{Src: sb.String()}, pkg.ExpressionError{
						Namespace: ns,
						Statement: id,
						Err: pkg.UnexpectedTokenError{
							Tok: tok.Lexeme(),
							Msg: []string{`expected expression`},
						},
					}
				}

				sb.WriteString(tok.Value)

				if tok, err = sg.Accept(getType.Predicate(`EX`)); err == nil {
					sb.WriteString(tok.Value)
				}

				if _, err := sg.AcceptAny(getType.Predicates(`RS`, `SC`)...); err == nil {
					if tok.Type != getType(`RS`) {
						sb.WriteString(RS) // terminate the expression
					}

					return Expression{Src: sb.String()}, nil
				}
			}
		}
	}

	return pkg.Make(stage.Pipe(ctx))
}
