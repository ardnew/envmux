package stream

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

type (
	// Token represents a token in the stream.
	//
	// Offset is the 0-based, absolute byte index of the token in a stream.
	Token struct {
		Value  string
		Type   lexer.TokenType
		Offset uint64
		Pos    Pos
	}

	// Pos represents a position in the stream as a line and rune-column offset.
	// Both offsets begin at 1.
	Pos struct {
		Line uint64
		Rune uint64
	}

	// TypeResolver is a function that returns the type of a token
	// given the symbolic name of the rule that matches it.
	//
	// Lexer rules and their names are defined in [parse.LexerGenerator].
	TypeResolver func(string) lexer.TokenType
)

// Predicate creates a function that compares the resolved type of symbol to
// its argument's [Token.Type].
func (t TypeResolver) Predicate(symbol string) func(Token) bool {
	typ := t(symbol)

	return func(tok Token) bool { return tok.Type == typ }
}

// Predicates returns a slice of [TypeResolver.Predicate]ed [Token]s.
func (t TypeResolver) Predicates(symbols ...string) []func(Token) bool {
	preds := make([]func(Token) bool, len(symbols))

	for i, symbol := range symbols {
		preds[i] = t.Predicate(symbol)
	}

	return preds
}

//nolint:gochecknoglobals
var eof = Token{Type: lexer.EOF} //nolint:exhaustruct

func resolveEOF(string) lexer.TokenType { return eof.Type }

func (t Token) IsType(typ lexer.TokenType) bool { return t.Type == typ }

// Lexeme returns a lexer.Token with the value and type of the token.
//
// This method fully constructs an external [lexer.Token]
// compatible with the [github.com/alecthomas/participle/v2/lexer] API
// converted from the local [Token] type of the receiver.
func (t Token) Lexeme() *lexer.Token {
	//nolint:exhaustruct
	return &lexer.Token{
		Type:  t.Type,
		Value: t.Value,
		//nolint:gosec
		Pos: lexer.Position{
			Offset: int(t.Offset),
			Line:   int(t.Pos.Line),
			Column: int(t.Pos.Rune),
		},
	}
}

// next returns a new [Token] with the position updated based on the
// receiver [Token], and the value and type from the provided [lexer.Token].
func withLexer(lex *lexer.PeekingLexer) pkg.Option[Token] {
	var tok *lexer.Token

	return func(t Token) Token {
		// t is the previous token, tok is the current token.
		if tok = lex.Next(); tok == nil {
			return t
		}

		// Offset always increments by exactly the size of the token value (bytes).
		t.Offset += uint64(len(t.Value))

		// Initialize the position with the previous token's position, and
		// locate the last newline in the previous value to determine what line
		// the current token starts on.
		p, s := t.Pos, strings.LastIndexByte(t.Value, '\n')

		// Determine if we are continuing on the previous token's starting line, or
		// continuing on the previous token's ending line.
		//
		//nolint:gosec
		if s >= 0 {
			// Continue on the previous token's ending line.
			p.Line += uint64(strings.Count(t.Value[:s], "\n")) + 1
			p.Rune = uint64( // Reset column offset.
				utf8.RuneCountInString(t.Value[s:]),
			)
		} else {
			// Continue on the previous token's starting line.
			// Do not modify the line offset, and increment rune offset without reset.
			p.Rune += uint64(utf8.RuneCountInString(t.Value))
		}

		p.Line = max(p.Line, 1)
		p.Rune = max(p.Rune, 1)

		return Token{
			Value:  tok.Value,
			Type:   tok.Type,
			Offset: t.Offset,
			Pos:    p,
		}
	}
}

// Tokens returns a [Group] of [Token]s from the provided [lexer.PeekingLexer],
// excluding whitespace and comments.
// The position of the token is updated before it is emitted to the next stage.
//
// The context is canceled by any task in [Group.Group] returning an error,
// or by calling [Group.Cancel] with an error cause.
//
// The context is canceled in the error handling of [Stream.Pipe].
//
// When EOF is reached by the lexer, it is emitted to the next stage and
// [pkg.ErrEOF] is returned, canceling the context.
//
// Tokens is intended to be used as the first stage in a pipeline.
// Subsequent stages can perform additional filtering.
//
// [pipeline]: https://go.dev/blog/pipelines
func Tokens(ctx context.Context, lex *lexer.PeekingLexer) Group[Token] {
	resolveType := fromContext(ctx)

	// Explicit declaration required for use as [Stage.Pipe] receiver below.
	var stage Stage[Token] = func() (Token, error) {
	next:
		cur := pkg.Make(withLexer(lex))

		fmt.Printf("token: %#v\n", cur)

		switch cur.Type {
		case lexer.EOF:
			return cur, pkg.ErrEOF

		case resolveType(`XX`):
			goto next // goto is so underrated
		}

		return cur, nil
	}

	return pkg.Make(stage.Pipe(ctx))
}

type contextKey struct{}

var typeResolver contextKey //nolint:gochecknoglobals

func MakeContext(ctx context.Context, resolve TypeResolver) context.Context {
	if resolve == nil {
		resolve = resolveEOF
	}

	return context.WithValue(ctx, typeResolver, resolve)
}

func fromContext(ctx context.Context) TypeResolver {
	res, ok := ctx.Value(typeResolver).(TypeResolver)
	if !ok {
		return resolveEOF
	}

	return res
}
