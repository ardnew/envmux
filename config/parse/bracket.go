package parse

import (
	"slices"
	"strconv"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

type (
	bracket       struct{ open, close string }
	bracketParser func(next func(*lexer.Token, int) (terminate, error)) error
)

//nolint:gochecknoglobals
var (
	compositeBracket = bracket{open: co, close: cc}
	statementBracket = bracket{open: so, close: sc}
	parameterBracket = bracket{open: po, close: pc}
	aggregateBracket = bracket{open: ao, close: ac}
)

func (b bracket) isClosedBy(token string) bool { return b.close == token }

func isOpeningBracket(token string) (bracket, bool) {
	switch token {
	case compositeBracket.open:
		return compositeBracket, true
	case statementBracket.open:
		return statementBracket, true
	case aggregateBracket.open:
		return aggregateBracket, true
	case parameterBracket.open:
		return parameterBracket, true
	default:
		return bracket{}, false //nolint:exhaustruct
	}
}

func isClosingBracket(token string) bool {
	switch token {
	case compositeBracket.close,
		statementBracket.close,
		parameterBracket.close,
		aggregateBracket.close:
		return true
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
	except ...bracket,
) bracketParser {
	return func(next func(*lexer.Token, int) (terminate, error)) error {
		stack := []bracket{}

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
								stack[len(stack)-1].close,
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
					if stack[len(stack)-1].isClosedBy(tok.Value) {
						stack = stack[:len(stack)-1] // Pop
					} else {
						return &pkg.UnexpectedTokenError{
							Tok: tok,
							Msg: []string{`expected close-bracket ` + strconv.Quote(stack[len(stack)-1].close)},
						}
					}
				} else {
					return &pkg.UnexpectedTokenError{
						Tok: tok,
					}
				}
			}
		}
		// If we have consumed all input without satisfying the stop condition,
		// then we have incomplete input.
		//
		//	if len(stack) > 0 {
		//		return &pkg.UnexpectedTokenError{
		//			Tok: tok,
		// 			Msg: []string{`expected close-bracket ` +
		// strconv.Quote(stack[len(stack)-1].close)},
		//		}
		//	}
		//
		// If we reach here, we have consumed all input with an empty stack.
		// return nil
	}
}
