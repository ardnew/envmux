package parse

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/ardnew/envmux/pkg"
	"github.com/pkg/errors"
)

// Expr represents an expression node in the AST.
type Expr struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node.

	Src string
}

// Parse parses an expression using the provided lexer
// and stores the unevaluated source code in the Expr's Src field.
// Returns an error if parsing fails.
func (e *Expr) Parse(lex *lexer.PeekingLexer) (err error) {
	e.Src, err = e.parse(lex)
	return err
}

func (e *Expr) parse(lex *lexer.PeekingLexer) (string, error) {
	var braceDepth int
	var result strings.Builder
	for {
		tok := lex.Next()
		if tok.EOF() {
			break
		}
		if tok.Value == "{" {
			braceDepth++
			continue
		} else if tok.Value == "}" {
			braceDepth--
			if braceDepth < 0 {
				return "", errors.Wrap(pkg.ErrInvalidExpression, "unexpected '}'")
			}
			continue
		} else if tok.Value == ";" && braceDepth == 0 {
			break
		}
		result.WriteString(tok.Value)
	}
	if braceDepth != 0 {
		return "", errors.Wrap(pkg.ErrInvalidExpression, "expected '}'")
	}
	return result.String(), nil
}
