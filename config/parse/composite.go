package parse

import (
	"fmt"
	"iter"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

// Composite is a container for the semantic information of a Composite node.
type Composite struct {
	ID     string
	Params parameters
}

// composite represents a composition node in the AST.
type composite struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node.

	Composite
}

func (c *composite) String() string {
	if c == nil {
		return ""
	}

	return c.ID
}

func (c *composite) Parse(lex *lexer.PeekingLexer) error {
	if err := c.parse(lex); err != nil {
		if c.ID != "" {
			return fmt.Errorf("%w %q: %w", pkg.ErrInvalidComposite, c.ID, err)
		}

		return err
	}

	return nil
}

func (c *composite) parse(lex *lexer.PeekingLexer) error {
	switch tok := lex.Next(); tok.Type { //nolint:exhaustive
	case symbol()(`NS`):
		*c = makeComposite(tok)

		ccp := lex.MakeCheckpoint()

		// A composite can also have parameters which are evaluated in the context
		// of the composed namespace.
		if err := c.Params.Parse(lex); err != nil {
			// If parsing parameters fails, reset the lexer to the checkpoint
			// and finish parsing the composite.
			lex.LoadCheckpoint(ccp)
		}

		return nil

	default:
		return &pkg.UnexpectedTokenError{
			Tok: tok, Msg: []string{
				`expected namespace identifier as composite`,
			},
		}
	}
}

func makeComposite(tok *lexer.Token) composite {
	if tok == nil || tok.EOF() {
		return composite{} //nolint:exhaustruct
	}

	endPos := tok.Pos
	endPos.Advance(tok.Value)

	return composite{
		Pos:    tok.Pos,
		EndPos: endPos,
		Tokens: []lexer.Token{*tok},
		Composite: Composite{
			ID:     tok.Value,
			Params: parameters{}, //nolint:exhaustruct
		},
	}
}

// composites represents a sequence of Composite nodes in the AST.
type composites struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	list []*composite
}

// Len returns the number of composites in the sequence.
func (c *composites) Len() int {
	if c == nil {
		return 0
	}

	return len(c.list)
}

// Seq returns the receiver's unique sequence of composite identifiers paired
// with any parameters declared with it in the composition.
//
// Only the first occurrence of duplicate identifiers is yielded.
func (c *composites) Seq() iter.Seq[Composite] {
	if c == nil {
		return nil
	}

	unique := make(pkg.Unique[string])

	return func(yield func(Composite) bool) {
		for _, v := range c.list {
			if unique.Set(v.ID) && !yield(v.Composite) {
				return
			}
		}
	}
}

func (c *composites) Parse(lex *lexer.PeekingLexer) error {
	if err := c.parse(lex); err != nil {
		return err
	}

	return nil
}

func (c *composites) parse(lex *lexer.PeekingLexer) error {
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

		com := new(composite)
		if err := com.Parse(lex); err != nil {
			return err
		}

		c.list = append(c.list, com)
	}

	return &pkg.UnexpectedTokenError{
		Tok: lex.Peek(),
		Msg: []string{
			`expected close-angle-bracket '>' to end composites`,
		},
	}
}
