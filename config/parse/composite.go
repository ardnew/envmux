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

// Composite contains a namespace identifier whose evaluated environment can be
// inherited in the definition of a [namespace].
//
// A Composite can optionally specify an in-line [parameter] list that is
// appended to that [namespace]'s own parameter list definition.
type Composite struct {
	Ident string

	Parameters []Parameter
}

func (c Composite) String() string {
	if c.Ident == "" {
		return ""
	}

	var par string

	if len(c.Parameters) > 0 {
		pars := make([]string, len(c.Parameters))
		for i, p := range c.Parameters {
			pars[i] = fmt.Sprintf(`%v`, p.Value)
		}

		par = strings.Join(pars, FS)
	}

	return fmt.Sprintf(`%s%s%s%s`, c.Ident, po, par, pc)
}

// Arguments returns a [Parameter.Value] sequence of all [Composite.Parameter]s.
//
// These parameters are specified in-line in the [Composite] list of a
// [Namespace] definition, which get appended to the composited
// [Namespace.Parameters], but only for that evaluated instance.
func (c Composite) Arguments() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, p := range c.Parameters {
			if !yield(p.Value) {
				return
			}
		}
	}
}

func composites(
	ctx context.Context,
	sg *stream.Group[stream.Token],
) stream.Group[Composite] {
	var getType stream.TypeResolver = tokenType()

	count := 0 // number of composite namespace identifiers parsed by the stage

	var stage stream.Stage[Composite] = func() (Composite, error) {
		var sym, msg string

		// There are two possible starting states for the stage:
		//  1. No composites have been parsed yet, so the first token must be the
		//     composite list opening meta-token.
		//  2. At least one composite has been parsed, so the first token must be
		//     the composite list delimiter or closing meta-token.
		if count == 0 {
			sym = `CO` // composite list open meta-token
			msg = fmt.Sprintf(`expected composite list opening meta-token %q`, co)
		} else {
			sym = `FS` // composite list delimiter meta-token (field separator)
			msg = fmt.Sprintf(`expected composite list delimiter meta-token %q`, FS)

			// Check for the composite list close meta-token
			// after processing each composite,
			// but before the composite list delimiter meta-token.
			if _, err := sg.Accept(getType.Predicate(`CC`)); err == nil {
				return Composite{}, pkg.ErrEOF
			}
		}

		tok, err := sg.Accept(getType.Predicate(sym))

		switch {
		case errors.Is(err, pkg.ErrClosedStream):
			if count > 0 {
				return Composite{}, pkg.ErrUnexpectedEOF
			}

			return Composite{}, pkg.ErrEOF
		case errors.Is(err, pkg.ErrUnacceptableStream):
			return Composite{}, pkg.UnexpectedTokenError{
				Tok: tok.Lexeme(), Msg: []string{msg},
			}
		}

		// Check for the composite list close meta-token immediately
		// after the opening or delimiting meta-token.
		if _, err := sg.Accept(getType.Predicate(`CC`)); err == nil {
			return Composite{}, pkg.ErrEOF
		}

		// After either opening or delimiting the composite list,
		// the next token must be a composite namespace identifier.
		tok, err = sg.AcceptAny(getType.Predicates(`NS`, `QQ`)...)

		switch {
		case errors.Is(err, pkg.ErrClosedStream):
			return Composite{}, pkg.ErrUnexpectedEOF
		case errors.Is(err, pkg.ErrUnacceptableStream):
			return Composite{}, pkg.UnexpectedTokenError{
				Tok: tok.Lexeme(),
				Msg: []string{`expected composite namespace identifier`},
			}
		}

		count++
		co := Composite{Ident: tok.Value} //nolint:exhaustruct

		// (optional) Capture all parameters specified in-line
		// with the composite namespace identifier.
		for p := range parameters(ctx, sg).Channel {
			co.Parameters = append(co.Parameters, p)
		}

		return co, nil
	}

	return pkg.Make(stage.Pipe(ctx))
}
