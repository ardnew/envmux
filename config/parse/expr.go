package parse

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

type Expr struct {
	Pos, EndPos lexer.Position
	Src         string
}

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
				return "", fmt.Errorf("%w: unexpected '}'", pkg.ErrInvalidExpression)
			}
			continue
		} else if tok.Value == ";" && braceDepth == 0 {
			break
		}
		result.WriteString(tok.Value)
	}
	if braceDepth != 0 {
		return "", fmt.Errorf("%w: expected '}'", pkg.ErrInvalidExpression)
	}
	return result.String(), nil
}

// func (e *Expr) String() string {
// 	var res any
// 	var err error
// 	if e.Program != nil {
// 		res, err = expr.Run(e.Program, envContext())
// 	} else {
// 		res, err = expr.Eval(e.Src, envContext())
// 	}
// 	if err != nil {
// 		return fmt.Errorf("%w: %w", pkg.ErrInvalidExpression, err).Error()
// 	}
// 	return fmt.Sprintf("%v", res)
// }
