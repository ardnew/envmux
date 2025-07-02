package parse

import (
	"fmt"
	"iter"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/config/env/vars"
	"github.com/ardnew/envmux/pkg"
)

// Statement represents a statement node in the AST.
//
// It assigns or amends value to a variable identifier in a namespace.
// The value can be a literal or an evaluated expression.
//
// Expressions are evaluated in the context of the enclosing namespace
// and the implicit parameter (identified with [vars.ParameterKey])
// for each parameter to the namespace.
//
// Each parameter's evaluation is assigned to the variable based on the formal
// syntax used by the parameter.
type Statement struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	Ev string
	Op string
	Ex *Expression
}

func (s *Statement) String() string {
	if s == nil {
		return ""
	}

	str := []string{s.Ev}
	if s.Op != "" && s.Ex != nil {
		str = append(str, s.Op, s.Ex.String())
	}

	return vars.Export(str...)
}

func (s *Statement) Parse(lex *lexer.PeekingLexer) error {
	if err := s.parse(lex); err != nil {
		if s.Ev != "" {
			return fmt.Errorf("%w %q: %w", pkg.ErrInvalidStatement, s.Ev, err)
		}

		return err
	}

	return nil
}

func (s *Statement) parse(lex *lexer.PeekingLexer) error {
	advance := consume(lex, `XX`)

	tid := lex.Next()
	if tid.Type != symbol()(`ID`) {
		return &pkg.UnexpectedTokenError{
			Tok: tid,
			Msg: []string{`expected identifier in assignment statement`},
		}
	}

	s.Ev = tid.Value
	s.Tokens = append(s.Tokens, *tid)

	advance()

	top := lex.Next()
	if top.Type != symbol()(`OP`) {
		return &pkg.UnexpectedTokenError{
			Tok: top,
			Msg: []string{`expected operator '=' in assignment statement`},
		}
	}

	s.Op = top.Value
	s.Tokens = append(s.Tokens, *top)

	advance()

	s.Ex = new(Expression)
	if err := s.Ex.Parse(lex); err != nil {
		return err
	}

	s.Pos = tid.Pos
	s.EndPos = s.Ex.EndPos
	s.Tokens = append(s.Tokens, s.Ex.Tokens...)

	return nil
}

// Statements represents a sequence of Statement nodes in the AST.
type Statements struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	List []*Statement
}

// Seq returns the receiver's sequence of statements.
func (s *Statements) Seq() iter.Seq[string] {
	if s == nil {
		return nil
	}

	return func(yield func(string) bool) {
		for _, v := range s.List {
			if !yield(v.Ev) {
				return
			}
		}
	}
}

func (s *Statements) Parse(lex *lexer.PeekingLexer) error {
	if err := s.parse(lex); err != nil {
		return err
	}

	return nil
}

func (s *Statements) parse(lex *lexer.PeekingLexer) error {
	if tok := lex.Next(); tok.Type != symbol()(`SO`) {
		return &pkg.UnexpectedTokenError{
			Tok: tok,
			Msg: []string{`expected open-bracket '{' to begin statements`},
		}
	}

	advance := consume(lex, `XX`)

	for advance() {
		if tok := lex.Peek(); tok.Type == symbol()(`SC`) {
			_ = lex.Next() // Consume the closing curly bracket.

			return nil
		}

		sta := new(Statement)
		if err := sta.Parse(lex); err != nil {
			return err
		}

		s.List = append(s.List, sta)
	}

	return &pkg.UnexpectedTokenError{
		Tok: lex.Next(),
		Msg: []string{`expected close-bracket '}' to end statements`},
	}
}
