package parse

import (
	"slices"
	"strconv"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

type (
	bracketType   byte
	bracket       struct{ open, close string }
	bracketParser func(next func(*lexer.Token, int) (terminate, error)) error
)

const (
	bracketUndefined bracketType = iota
	bracketComposite
	bracketStatement
	bracketParameter
	bracketAggregate
)

func (b bracketType) get() bracket {
	if b <= bracketUndefined || int(b) >= len(brackets) {
		return brackets[bracketUndefined]
	}

	return brackets[b]
}
func (b bracketType) open() string  { return b.get().open }
func (b bracketType) close() string { return b.get().close }

//nolint:gochecknoglobals
var brackets = []bracket{
	{open: ZZ, close: ZZ}, // Undefined
	{open: co, close: cc}, // Composite
	{open: so, close: sc}, // Statement
	{open: po, close: pc}, // Parameter
	{open: ao, close: ac}, // Aggregate
}

func isOpeningBracket(token string) (bracketType, bool) {
	for i := range brackets[bracketUndefined+1:] {
		b := bracketType(i)
		if b.open() == token {
			return b, true
		}
	}

	return bracketUndefined, false
}

func isClosingBracket(token string) bool {
	for i := range brackets[bracketUndefined+1:] {
		b := bracketType(i)
		if b.close() == token {
			return true
		}
	}

	return false
}

// makeBracketParser returns a bracketParser that verifies all brackets in the
// input token stream are properly balanced.
//
// It accepts a PeekingLexer and an optional list of bracket types to ignore.
//
// It returns a closure over the lexer that accepts a callback function for
// processing each token in the input stream. The callback function returns a
// boolean indicating whether parsing should continue.
//
// The bracket balancing is performed after each token is processed by the
// callback. This forms a two-phase parsing pipeline with bracket validation
// being the second phase.
func makeBracketParser(
	lex *lexer.PeekingLexer,
	except ...bracketType,
) bracketParser {
	return func(next func(*lexer.Token, int) (terminate, error)) error {
		stack := []bracketType{}

		var tok *lexer.Token

		// Inspect the next token without consuming it.
		// This allows us to recover from certain unexpected tokens
		// without losing the context of the current parsing state.
		for {
			tok = lex.Peek()

			var (
				term, err = next(tok, len(stack))
				brkt, lhs = isOpeningBracket(tok.Value)
				rhs       = isClosingBracket(tok.Value)
			)

			switch term {
			case unterminated:
				_ = lex.Next() // Consume the token.

			case atError:
				return err

			case atEOF, atNL, atRS:
				_ = lex.Next() // Consume the token.

				fallthrough

			case atSC:
				if len(stack) > 0 {
					return &pkg.UnexpectedTokenError{
						Tok: tok,
						Msg: []string{
							`expected close-bracket ` + strconv.Quote(
								stack[len(stack)-1].get().close,
							),
						},
					}
				}

				return nil // No unclosed brackets, we are done.
			}

			switch {
			case slices.Contains(except, brkt):
				// If the bracket is in the except list, we skip it.
				// This allows us to ignore brackets that are not relevant
				// in the current context (e.g., comparison operation "x < y").
				continue

			case lhs:
				// Opening brackets are always accepted.
				stack = append(stack, brkt) // Push

			case rhs:
				if len(stack) > 0 {
					if stack[len(stack)-1].get().close == tok.Value {
						stack = stack[:len(stack)-1] // Pop
					} else {
						return &pkg.UnexpectedTokenError{
							Tok: tok,
							Msg: []string{
								`expected close-bracket ` + strconv.Quote(
									stack[len(stack)-1].get().close,
								),
							},
						}
					}
				} else {
					//nolint:exhaustruct
					return &pkg.UnexpectedTokenError{
						Tok: tok,
					}
				}
			}
		}
	}
}
