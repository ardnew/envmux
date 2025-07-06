package parse

import (
	"fmt"
	"iter"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

// Parameter is a container for the semantic information of a Parameter node.
type Parameter struct {
	Value any
}

// parameter represents a parameter node in the AST.
type parameter struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	Parameter
}

func (p *parameter) String() string {
	if p == nil {
		return ""
	}

	return fmt.Sprint(p.Value)
}

func (p *parameter) Parse(lex *lexer.PeekingLexer) error {
	if err := p.parse(lex); err != nil {
		if p.Value != "" {
			return fmt.Errorf("%w %q: %w", pkg.ErrInvalidParameter, p.Value, err)
		}

		return err
	}

	return nil
}

func (p *parameter) parse(lex *lexer.PeekingLexer) error {
	switch tok := lex.Next(); tok.Type { //nolint:exhaustive
	case symbol()(`QQ`), symbol()(`NU`), symbol()(`ID`):
		*p = makeParameter(tok)

		return nil

	default:
		return &pkg.UnexpectedTokenError{
			Tok: tok,
			Msg: []string{`expected string, number, or identifier as parameter`},
		}
	}
}

func makeParameter(tok *lexer.Token) parameter {
	if tok == nil || tok.EOF() {
		return parameter{} //nolint:exhaustruct
	}

	endPos := tok.Pos
	endPos.Advance(tok.Value)

	return parameter{
		Pos:    tok.Pos,
		EndPos: endPos,
		Tokens: []lexer.Token{*tok},
		Parameter: Parameter{
			Value: tok.Value,
		},
	}
}

// parameters represents a sequence of Parameter nodes in the AST.
type parameters struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	list []*parameter
}

// Len returns the number of parameters in the sequence.
func (p *parameters) Len() int {
	if p == nil {
		return 0
	}

	return len(p.list)
}

// Seq returns the receiver's sequence of parameters.
func (p *parameters) Seq() iter.Seq[Parameter] {
	if p == nil {
		return nil
	}

	return func(yield func(Parameter) bool) {
		for _, v := range p.list {
			if !yield(v.Parameter) {
				return
			}
		}
	}
}

// Values returns the receiver's sequence of parameter values.
func (p *parameters) Values() iter.Seq[any] {
	if p == nil {
		return nil
	}

	return func(yield func(any) bool) {
		for _, v := range p.list {
			if !yield(v.Value) {
				return
			}
		}
	}
}

func (p *parameters) Parse(lex *lexer.PeekingLexer) error {
	if err := p.parse(lex); err != nil {
		return err
	}

	return nil
}

func (p *parameters) parse(lex *lexer.PeekingLexer) error {
	if tok := lex.Next(); tok.Type != symbol()(`PO`) {
		return &pkg.UnexpectedTokenError{
			Tok: tok,
			Msg: []string{`expected open-parenthesis '(' to begin parameters`},
		}
	}

	advance := consume(lex, `XX`, `FS`)

	for advance() {
		if tok := lex.Peek(); tok.Type == symbol()(`PC`) {
			_ = lex.Next() // Consume the closing parenthesis.

			return nil
		}

		par := new(parameter)
		if err := par.Parse(lex); err != nil {
			return err
		}

		p.list = append(p.list, par)
	}

	return &pkg.UnexpectedTokenError{
		Tok: lex.Peek(),
		Msg: []string{`expected close-parenthesis ')' to end parameters`},
	}
}
