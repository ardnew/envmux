package parse

import (
	"context"
	"errors"
	"fmt"

	"github.com/ardnew/envmux/config/parse/stream"
	"github.com/ardnew/envmux/pkg"
)

// Parameter represents a value that can be referenced
// using the implicit variable named by
// [github.com/ardnew/envmux/config/env/vars.ParameterKey]
// in each [statement.expression] of a [namespace].
type Parameter struct {
	Value any
}

func (p Parameter) String() string {
	if p.Value == nil {
		return ""
	}

	return fmt.Sprintf("%v", p.Value)
}

func parameters(
	ctx context.Context,
	sg *stream.Group[stream.Token],
) stream.Group[Parameter] {
	var getType stream.TypeResolver = tokenType()

	count := 0 // number of parameters parsed by the stage

	var stage stream.Stage[Parameter] = func() (Parameter, error) {
		var sym, msg string

		// There are two possible starting states for the stage:
		//  1. No parameters have been parsed yet, so the first token must
		//     be the parameter list opening meta-token.
		//  2. At least one parameter has been parsed, so the first token
		//     must be the parameter list delimiter or closing meta-token.
		if count == 0 {
			sym = `PO` // parameter list open meta-token
			msg = fmt.Sprintf(`expected parameter list opening meta-token %q`, po)
		} else {
			sym = `FS` // parameter list delimiter meta-token (field separator)
			msg = fmt.Sprintf(`expected parameter list delimiter meta-token %q`, FS)

			// Check for the parameter close meta-token
			// after processing each parameter,
			// and before the parameter list delimiter meta-token.
			if _, err := sg.Accept(getType.Predicate(`PC`)); err == nil {
				return Parameter{}, pkg.ErrEOF
			}
		}

		tok, err := sg.Accept(getType.Predicate(sym))

		switch {
		case errors.Is(err, pkg.ErrClosedStream):
			if count > 0 {
				return Parameter{}, pkg.ErrUnexpectedEOF
			}

			return Parameter{}, pkg.ErrEOF
		case errors.Is(err, pkg.ErrUnacceptableStream):
			return Parameter{}, pkg.UnexpectedTokenError{
				Tok: tok.Lexeme(), Msg: []string{msg},
			}
		}

		// Check for the parameter list close meta-token immediately
		// after the opening or delimiting meta-token.
		if _, err := sg.Accept(getType.Predicate(`PC`)); err == nil {
			return Parameter{}, pkg.ErrEOF
		}

		// After either opening or delimiting the parameter list,
		// the next token must be a parameter value.
		tok, err = sg.AcceptAny(getType.Predicates(`ID`, `QQ`, `NU`)...)

		switch {
		case errors.Is(err, pkg.ErrClosedStream):
			return Parameter{}, pkg.ErrUnexpectedEOF
		case errors.Is(err, pkg.ErrUnacceptableStream):
			return Parameter{}, pkg.UnexpectedTokenError{
				Tok: tok.Lexeme(),
				Msg: []string{`expected parameter value`},
			}
		}

		count++
		pv := Parameter{Value: tok.Value}

		return pv, nil
	}

	return pkg.Make(stage.Pipe(ctx))
}
