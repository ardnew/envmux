package parse

import (
	"fmt"
	"iter"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/config/env/vars"
	"github.com/ardnew/envmux/pkg"
)

// Statement is a container for the semantic information of a Statement node.
type Statement struct {
	ID string
	Op string
	Ex *expression
}

// statement represents a statement node in the AST.
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
type statement struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	Statement
}

func (s *statement) String() string {
	if s == nil {
		return ""
	}

	return vars.Export(s.ID, s.Ex)
}

func (s *statement) Parse(lex *lexer.PeekingLexer) error {
	if err := s.parse(lex); err != nil {
		if s.ID != "" {
			return fmt.Errorf("%w %q: %w", pkg.ErrInvalidStatement, s.ID, err)
		}

		return err
	}

	return nil
}

func (s *statement) parse(lex *lexer.PeekingLexer) error {
	advance := consume(lex, `XX`)

	tid := lex.Next()
	if tid.Type != symbol()(`ID`) {
		return &pkg.UnexpectedTokenError{
			Tok: tid,
			Msg: []string{`expected identifier in assignment statement`},
		}
	}

	s.ID = tid.Value
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

	s.Ex = new(expression)
	if err := s.Ex.Parse(lex); err != nil {
		return err
	}

	s.Pos = tid.Pos
	s.EndPos = s.Ex.EndPos
	s.Tokens = append(s.Tokens, s.Ex.Tokens...)

	return nil
}

// statements represents a sequence of Statement nodes in the AST.
type statements struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	list []*statement
}

// Len returns the number of statements in the sequence.
func (s *statements) Len() int {
	if s == nil {
		return 0
	}

	return len(s.list)
}

// Seq returns the receiver's sequence of statements.
func (s *statements) Seq() iter.Seq[Statement] {
	if s == nil {
		return nil
	}

	return func(yield func(Statement) bool) {
		for _, v := range s.list {
			if !yield(v.Statement) {
				return
			}
		}
	}
}

func (s *statements) Parse(lex *lexer.PeekingLexer) error {
	if err := s.parse(lex); err != nil {
		return err
	}

	return nil
}

func (s *statements) parse(lex *lexer.PeekingLexer) error {
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

		sta := new(statement)
		if err := sta.Parse(lex); err != nil {
			return err
		}

		s.list = append(s.list, sta)
	}

	return &pkg.UnexpectedTokenError{
		Tok: lex.Next(),
		Msg: []string{`expected close-bracket '}' to end statements`},
	}
}
