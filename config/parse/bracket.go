package parse

import (
	"slices"
	"sync"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/config/parse/stream"
)

const nhs = string(utf8.RuneError) // !LHS && !RHS == NHS ("no hand side")

type bracketType byte

const (
	bracketUndefined bracketType = iota
	bracketAngles
	bracketBraces
	bracketParens
	bracketSquare
)

func (b bracketType) get() bracketPair {
	if !b.isDefined() {
		return brackets[bracketUndefined]
	}

	return brackets[b]
}

func (b bracketType) lhs() string { return b.get().l }
func (b bracketType) rhs() string { return b.get().r }

func (b bracketType) isDefined() bool {
	return b > bracketUndefined && int(b) < len(brackets)
}

type bracketPair struct{ l, r string }

//nolint:gochecknoglobals
var brackets = []bracketPair{
	{l: ZZ, r: ZZ}, // Undefined
	{l: co, r: cc}, // Angles
	{l: so, r: sc}, // Braces
	{l: po, r: pc}, // Parens
	{l: ao, r: ac}, // Square
}

type bracketInfo struct {
	typ   bracketType
	isLHS bool
	mate  string
}

// Private singleton cache for bracketMatch mappings.
//
//nolint:gochecknoglobals
var (
	bracketBond  sync.Once
	bracketMatch map[string]bracketInfo
)

// mateBracket caches the mappings between opening and closing brackets
// in both directions (open -> close and close -> open).
// This enables constant-time lookups for matching brackets by string value.
func matchBracket(token string) (bracketInfo, bool) {
	bracketBond.Do(func() {
		bracketMatch = make(map[string]bracketInfo)
		for i, b := range brackets {
			bracketMatch[b.l] = bracketInfo{
				typ:   bracketType(i),
				isLHS: true,
				mate:  b.r,
			}
			bracketMatch[b.r] = bracketInfo{
				typ:   bracketType(i),
				isLHS: false,
				mate:  b.l,
			}
		}
	})

	mate, ok := bracketMatch[token]

	return mate, ok && mate.typ.isDefined()
}

type bracketStack []bracketType

func (s *bracketStack) push(b bracketType) {
	if b.isDefined() {
		*s = append(*s, b)
	}
}

func (s *bracketStack) pop() (bracketType, bool) {
	if len(*s) == 0 {
		return bracketUndefined, false
	}

	n := len(*s) - 1
	v := (*s)[n]
	*s = (*s)[:n]

	return v, true
}

// bracketBalancer returns a predicate that verifies RHS brackets only occur
// following a matching LHS bracket. It supports nested bracketing to any depth.
//
// Depth is set to len(stack) each time the predicate is called.
//
// See [brackets] for the list of handled brackets.
func bracketBalancer(
	depth *int,
	except ...bracketType,
) func(stream.Token) bool {
	var stack bracketStack

	return func(tok stream.Token) bool {
		if depth != nil {
			defer func() { *depth = len(stack) }()
		}

		if tok.Type == lexer.EOF {
			if len(stack) == 0 {
				return true // EOF is valid if no unclosed brackets.
			}

			return false // EOF is invalid if there are unclosed brackets.
		}

		b, isBracket := matchBracket(tok.Value)

		switch {
		case !isBracket, slices.Contains(except, b.typ): // invalid bracket
			return true // always accept non-bracket tokens

		case b.isLHS: // non-exempt LHS bracket
			stack = append(stack, b.typ) // push unbalanced pair

			return true // always accept LHS brackets

		//  ╔═════════════════════════════════════════════════════════════╗
		// ╶╫╴ NOTE: everything below is guaranteed to be a RHS bracket! ╶╫╴
		//  ╚═════════════════════════════════════════════════════════════╝

		case len(stack) == 0: // empty stack
			return false // reject unmatched brackets

		case stack[len(stack)-1] == b.typ: // match top of stack
			stack = stack[:len(stack)-1] // pop balanced pair

			return true // accept matched brackets
		}

		panic("bracketBalancer: unhandled edge case") // should never happen
	}
}
