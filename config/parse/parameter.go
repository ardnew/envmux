package parse

import (
	"fmt"
	"iter"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

// Parameter represents a parameter node in the AST.
type Parameter struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	ID string
}

func (p *Parameter) String() string {
	if p == nil {
		return ""
	}

	return p.ID
}

func (p *Parameter) Parse(lex *lexer.PeekingLexer) error {
	if err := p.parse(lex); err != nil {
		if p.ID != "" {
			return fmt.Errorf("%w %q: %w", pkg.ErrInvalidParameter, p.ID, err)
		}

		return err
	}

	return nil
}

func (p *Parameter) parse(lex *lexer.PeekingLexer) error {
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

func makeParameter(tok *lexer.Token) Parameter {
	if tok == nil || tok.EOF() {
		return Parameter{} //nolint:exhaustruct
	}

	endPos := tok.Pos
	endPos.Advance(tok.Value)

	return Parameter{
		ID:     tok.Value,
		Pos:    tok.Pos,
		EndPos: endPos,
		Tokens: []lexer.Token{*tok},
	}
}

// Parameters represents a sequence of Parameter nodes in the AST.
type Parameters struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	List []*Parameter
}

// Seq returns the receiver's sequence of parameters.
func (p *Parameters) Seq() iter.Seq[string] {
	if p == nil {
		return nil
	}

	return func(yield func(string) bool) {
		for _, v := range p.List {
			if !yield(v.ID) {
				return
			}
		}
	}
}

func (p *Parameters) Parse(lex *lexer.PeekingLexer) error {
	if err := p.parse(lex); err != nil {
		return err
	}

	return nil
}

func (p *Parameters) parse(lex *lexer.PeekingLexer) error {
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

		par := new(Parameter)
		if err := par.Parse(lex); err != nil {
			return err
		}

		p.List = append(p.List, par)
	}

	return &pkg.UnexpectedTokenError{
		Tok: lex.Peek(),
		Msg: []string{`expected close-parenthesis ')' to end parameters`},
	}
}
