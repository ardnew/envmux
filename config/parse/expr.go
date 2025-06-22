package parse

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
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

//nolint:cyclop
func (e *Expr) parse(lex *lexer.PeekingLexer) (string, error) {
	var result strings.Builder

	hasEOL := func(s string) bool {
		for _, suffix := range []string{RS, `\n`} {
			if strings.HasSuffix(s, suffix) {
				return true
			}
		}

		return false
	}

	stack := []balance{}

	for tok := lex.Next(); !tok.EOF(); tok = lex.Next() {
		result.WriteString(tok.Value)

		// Check if we are opening a new balance
		// or closing the one on top of the stack.
		if b, ok := nextBalance(tok.Value); ok {
			stack = append(stack, b)

			continue
		} else if len(stack) > 0 {
			if b := stack[len(stack)-1]; b.close == tok.Value {
				stack = stack[:len(stack)-1]

				continue
			}
		}

		// If we are closing a balance that is not on top of the stack,
		// then we have an unbalanced expression.
		if unbalanced(tok.Value) {
			if len(stack) > 0 {
				return "", fmt.Errorf(
					"%w: unexpected '%s' (expected '%s')",
					pkg.ErrInvalidExpression,
					tok.Value,
					stack[len(stack)-1].close,
				)
			}

			return "", fmt.Errorf(
				"%w: unexpected '%s'",
				pkg.ErrInvalidExpression,
				tok.Value,
			)
		}

		if len(stack) == 0 && hasEOL(tok.Value) {
			return result.String(), nil
		}
	}

	// If we have consumed all input but still have unclosed balances,
	// then we have an unbalanced expression.
	if len(stack) > 0 {
		return "", fmt.Errorf(
			"%w: expected '%s'",
			pkg.ErrInvalidExpression,
			stack[len(stack)-1].close,
		)
	}

	return result.String(), nil
}

type balance struct{ open, close string }

//nolint:gochecknoglobals
var (
	curly  = balance{open: `{`, close: `}`}
	square = balance{open: `[`, close: `]`}
	paren  = balance{open: `(`, close: `)`}
)

func nextBalance(token string) (balance, bool) {
	switch token {
	case curly.open:
		return curly, true
	case square.open:
		return square, true
	case paren.open:
		return paren, true
	default:
		return balance{}, false //nolint:exhaustruct
	}
}

func unbalanced(token string) bool {
	switch token {
	case curly.close, square.close, paren.close:
		return true
	}

	return false
}
