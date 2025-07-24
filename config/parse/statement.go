package parse

import (
	"context"
	"errors"
	"fmt"

	"github.com/ardnew/envmux/config/parse/stream"
	"github.com/ardnew/envmux/pkg"
)

// Statement associates an expression with a variable identifier and operator.
// Expressions are evaluated in the context of the enclosing namespace.
//
// Expressions use an entirely different grammar than what is recognized by this
// module. The grammar is defined by [github.com/expr-lang/expr].
// Our grammar was designed to accommodate the embedded expression grammar.
type Statement struct {
	Ident      string
	Operator   string
	Expression *Expression
}

func (s Statement) String() string {
	if s.Ident == "" || s.Operator == "" || s.Expression == nil {
		return ""
	}

	return fmt.Sprintf("%s%s%s", s.Ident, s.Operator, s.Expression.String())
}

func statements(
	ctx context.Context,
	ns string,
	sg *stream.Group[stream.Token],
) stream.Group[Statement] {
	var getType stream.TypeResolver = tokenType()

	count := 0 // number of statements parsed by the stage

	var stage stream.Stage[Statement] = func() (Statement, error) {
		var sym, msg string

		// There are two possible starting states for the stage:
		//  1. No statements have been parsed yet, so the first token must be the
		//     statement list opening meta-token.
		//  2. At least one statement has been parsed, so the first token must be
		//     the statement list delimiter or closing meta-token.
		if count == 0 {
			sym = `SO` // statement list open meta-token
			msg = fmt.Sprintf(`expected statement list opening meta-token %q`, co)
		} else {
			sym = `RS` // statement list delimiter meta-token (record separator)
			msg = fmt.Sprintf(`expected statement list delimiter meta-token %q`, RS)

			// Check for the statement list close meta-token
			// after processing each statement,
			// but before the statement list delimiter meta-token.
			if _, err := sg.Accept(getType.Predicate(`SC`)); err == nil {
				return Statement{}, pkg.ErrEOF
			}
		}

		tok, err := sg.Accept(getType.Predicate(sym))

		switch {
		case errors.Is(err, pkg.ErrClosedStream):
			if count > 0 {
				return Statement{}, pkg.ErrUnexpectedEOF
			}

			return Statement{}, pkg.ErrEOF
		case errors.Is(err, pkg.ErrUnacceptableStream):
			return Statement{}, pkg.UnexpectedTokenError{
				Tok: tok.Lexeme(), Msg: []string{msg},
			}
		}

		// Check for the statement list close meta-token immediately
		// after the opening or delimiting meta-token.
		if _, err := sg.Accept(getType.Predicate(`SC`)); err == nil {
			return Statement{}, pkg.ErrEOF
		}

		// The next tokens must be an identifier, operator, and expression.
		toks, err := sg.AcceptEach(getType.Predicates(`ID`, `OP`)...)

		switch {
		case errors.Is(err, pkg.ErrClosedStream):
			return Statement{}, pkg.ErrUnexpectedEOF
		case errors.Is(err, pkg.ErrUnacceptableStream):
			return Statement{}, pkg.UnexpectedTokenError{
				Tok: toks[len(toks)-1].Lexeme(),
				Msg: []string{`expected statement namespace identifier`},
			}
		}

		id, op := toks[0].Value, toks[1].Value

		if ex, ok := <-expressions(ctx, ns, id, op, sg).Channel; ok {
			count++

			return Statement{Ident: id, Operator: op, Expression: &ex}, nil
		}

		return Statement{}, pkg.ErrUnexpectedEOF
	}

	return pkg.Make(stage.Pipe(ctx))
}

// // statement represents a statement node in the AST.
// //
// // It assigns or amends value to a variable identifier in a namespace.
// // The value can be a literal or an evaluated expression.
// //
// // Expressions are evaluated in the context of the enclosing namespace
// // and the implicit parameter (identified with [vars.ParameterKey])
// // for each parameter to the namespace.
// //
// // Each parameter's evaluation is assigned to the variable based on the
// formal
// // syntax used by the parameter.
// type statement struct {
// 	Pos    lexer.Position // Pos records the start position of the node.
// 	EndPos lexer.Position // EndPos records the end position of the node.
// 	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

// 	Statement
// }

// func (s *statement) String() string {
// 	if s == nil {
// 		return ""
// 	}

// 	return vars.Export(s.ID, s.Ex)
// }

// func (s *statement) Parse(ts Stream) {
// 	if err := s.parse(ctx, ts); err != nil {
// 		if s.ID != "" {
// 			return fmt.Errorf("%w %q: %w", pkg.ErrInvalidStatement, s.ID, err)
// 		}

// 		return err
// 	}

// 	return nil
// }

// func (s *statement) parse(ts Stream) {
// 	advance := consume(lex, `XX`)

// 	tid := lex.Next()
// 	if tid.Type != symbol()(`ID`) {
// 		return pkg.UnexpectedTokenError{
// 			Tok: tid,
// 			Msg: []string{`expected identifier in assignment statement`},
// 		}
// 	}

// 	s.ID = tid.Value
// 	s.Tokens = append(s.Tokens, *tid)

// 	advance()

// 	top := lex.Next()
// 	if top.Type != symbol()(`OP`) {
// 		return pkg.UnexpectedTokenError{
// 			Tok: top,
// 			Msg: []string{`expected operator '=' in assignment statement`},
// 		}
// 	}

// 	s.Op = top.Value
// 	s.Tokens = append(s.Tokens, *top)

// 	advance()

// 	s.Ex = new(expression)
// 	if err := s.Ex.Parse(ctx, ts); err != nil {
// 		return err
// 	}

// 	s.Pos = tid.Pos
// 	s.EndPos = s.Ex.EndPos
// 	s.Tokens = append(s.Tokens, s.Ex.Tokens...)

// 	return nil
// }

// // statements represents a sequence of Statement nodes in the AST.
// type statements struct {
// 	list []*statement
// }

// // Len returns the number of statements in the sequence.
// func (s *statements) Len() int {
// 	if s == nil {
// 		return 0
// 	}

// 	return len(s.list)
// }

// // Seq returns the receiver's sequence of statements.
// func (s *statements) Seq() iter.Seq[Statement] {
// 	if s == nil {
// 		return nil
// 	}

// 	return func(yield func(Statement) bool) {
// 		for _, v := range s.list {
// 			if !yield(v.Statement) {
// 				return
// 			}
// 		}
// 	}
// }

// func (s *statements) Parse(ts Stream) {
// 	if err := s.parse(ctx, ts); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (s *statements) parse(ts Stream) {
// 	if tok := lex.Next(); tok.Type != symbol()(`SO`) {
// 		return pkg.UnexpectedTokenError{
// 			Tok: tok,
// 			Msg: []string{`expected open-bracket '{' to begin statements`},
// 		}
// 	}

// 	advance := consume(lex, `XX`)

// 	for advance() {
// 		if tok := lex.Peek(); tok.Type == symbol()(`SC`) {
// 			_ = lex.Next() // Consume the closing curly bracket.

// 			return nil
// 		}

// 		sta := new(statement)
// 		if err := sta.Parse(ctx, ts); err != nil {
// 			return err
// 		}

// 		s.list = append(s.list, sta)
// 	}

// 	return pkg.UnexpectedTokenError{
// 		Tok: lex.Next(),
// 		Msg: []string{`expected close-bracket '}' to end statements`},
// 	}
// }
