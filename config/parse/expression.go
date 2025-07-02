package parse

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

// Expression represents an expression node in the AST.
type Expression struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node.

	Src string
}

func (e *Expression) String() string {
	if e == nil {
		return ""
	}

	return e.Src
}

// Parse parses an expression using the provided lexer
// and stores the unevaluated source code in the Expr's Src field.
// Returns an error if parsing fails.
func (e *Expression) Parse(lex *lexer.PeekingLexer) (err error) {
	if e.Src, err = e.parse(lex); err != nil {
		if e.Src != "" {
			return fmt.Errorf("%w %q: %w", pkg.ErrInvalidExpression, e.Src, err)
		}

		return err
	}

	return nil
}

func (e *Expression) parse(lex *lexer.PeekingLexer) (string, error) {
	var sb strings.Builder

	err := makeBracketParser(lex, bracketComposite)(
		func(token *lexer.Token, depth int) (terminate, error) {
			var result terminate

			if token.EOF() {
				result = atEOF
			} else {
				switch token.Value {
				case NL:
					result = atNL
				case RS:
					result = atRS
				case sc:
					if depth == 0 {
						token.Value = RS
						result = atSC
					}
				}
			}

			if len(e.Tokens) == 0 {
				e.Pos = token.Pos
				e.EndPos = token.Pos
			}

			e.EndPos.Advance(token.Value)
			e.Tokens = append(e.Tokens, *token)

			if _, err := sb.WriteString(token.Value); err != nil {
				return atError, err
			}

			return result, nil
		},
	)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}
