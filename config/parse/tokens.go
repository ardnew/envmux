package parse

import (
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

type token struct{ *lexer.Token }

func withToken(tok *lexer.Token) pkg.Option[token] {
	return func(t token) token {
		t.Token = tok
		return t
	}
}

func withTokenType(typ lexer.TokenType) pkg.Option[token] {
	return func(t token) token {
		t.Type = typ
		return t
	}
}

func withEOF() pkg.Option[token] { return withTokenType(lexer.EOF) }

type capture[T any] func() (T, error)

func withLexer(lex *lexer.PeekingLexer) pkg.Option[capture[token]] {
	return func(c capture[token]) capture[token] {
		return func() (token, error) {
			t := pkg.Make(withToken(lex.Next()))
			if t.Token == nil || t.Type == lexer.EOF {
				return pkg.Make(withEOF()), pkg.ErrEOF
			}

			return t, nil
		}
	}
}

type cap[T any] struct {
	capture capture[T]
	undo    []T
}

func (s *cap[T]) Accept(p func(T) error) (T, error) {
	if len(s.undo) > 0 {
		t := s.undo[len(s.undo)-1]
		s.undo = s.undo[:len(s.undo)-1]
		return t, p(t)
	}

	t, err := s.capture()
	if err != nil {
		return t, err
	}

	if err := p(t); err != nil {
		s.undo = append(s.undo, t)
		return t, err
	}

	return t, nil
}
