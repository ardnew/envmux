package parse

import (
	"fmt"
	"iter"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

// Composite represents a composition node in the AST.
type Composite struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node.

	ID string
}

func (c *Composite) String() string {
	if c == nil {
		return ""
	}

	return c.ID
}

func (c *Composite) Parse(lex *lexer.PeekingLexer) error {
	if err := c.parse(lex); err != nil {
		if c.ID != "" {
			return fmt.Errorf("%w %q: %w", pkg.ErrInvalidComposite, c.ID, err)
		}

		return err
	}

	return nil
}

func (c *Composite) parse(lex *lexer.PeekingLexer) error {
	switch tok := lex.Next(); tok.Type { //nolint:exhaustive
	case symbol()(`NS`):
		*c = makeComposite(tok)

		return nil

	default:
		return &pkg.UnexpectedTokenError{
			Tok: tok, Msg: []string{
				`expected namespace identifier as composite`,
			},
		}
	}
}

func makeComposite(tok *lexer.Token) Composite {
	if tok == nil || tok.EOF() {
		return Composite{} //nolint:exhaustruct
	}

	endPos := tok.Pos
	endPos.Advance(tok.Value)

	return Composite{
		ID:     tok.Value,
		Pos:    tok.Pos,
		EndPos: endPos,
		Tokens: []lexer.Token{*tok},
	}
}

// Composites represents a sequence of Composite nodes in the AST.
type Composites struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	List []*Composite
}

// Seq returns the receiver's unique sequence of composite identifiers.
// Only the first occurrence of duplicate identifiers is yielded.
func (c *Composites) Seq() iter.Seq[string] {
	if c == nil {
		return nil
	}

	unique := make(pkg.Unique[string])

	return func(yield func(string) bool) {
		for _, v := range c.List {
			if unique.Set(v.ID) && !yield(v.ID) {
				return
			}
		}
	}
}

func (c *Composites) Parse(lex *lexer.PeekingLexer) error {
	if err := c.parse(lex); err != nil {
		return err
	}

	return nil
}

func (c *Composites) parse(lex *lexer.PeekingLexer) error {
	if tok := lex.Next(); tok.Type != symbol()(`CO`) {
		return &pkg.UnexpectedTokenError{
			Tok: tok,
			Msg: []string{
				`expected open-angle-bracket '<' to begin composites`,
			},
		}
	}

	advance := consume(lex, `XX`, `FS`)

	for advance() {
		if tok := lex.Peek(); tok.Type == symbol()(`CC`) {
			_ = lex.Next() // Consume the closing angle bracket.

			return nil
		}

		com := new(Composite)
		if err := com.Parse(lex); err != nil {
			return err
		}

		c.List = append(c.List, com)
	}

	return &pkg.UnexpectedTokenError{
		Tok: lex.Peek(),
		Msg: []string{
			`expected close-angle-bracket '>' to end composites`,
		},
	}
}
