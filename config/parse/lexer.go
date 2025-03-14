// Code generated by Participle. DO NOT EDIT.
//
//go:generate bash internal/lexer.go.bash
package parse

import (
	"fmt"
	"io"
	"regexp/syntax"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var _ syntax.Op
var _ fmt.State

const _ = utf8.RuneError

var ConfigBackRefCache sync.Map
var ConfigLexer lexer.Definition = lexerConfigDefinitionImpl{}

type lexerConfigDefinitionImpl struct{}

func (lexerConfigDefinitionImpl) Symbols() map[string]lexer.TokenType {
	return map[string]lexer.TokenType{
		"Bareword":   -9,
		"Comment":    -3,
		"EOF":        -1,
		"Ident":      -8,
		"Number":     -6,
		"Punct":      -5,
		"String":     -7,
		"Template":   -2,
		"Whitespace": -4,
	}
}

func (lexerConfigDefinitionImpl) LexString(filename string, s string) (lexer.Lexer, error) {
	return &lexerConfigImpl{
		s: s,
		pos: lexer.Position{
			Filename: filename,
			Line:     1,
			Column:   1,
		},
		states: []lexerConfigState{{name: "Root"}},
	}, nil
}

func (d lexerConfigDefinitionImpl) LexBytes(filename string, b []byte) (lexer.Lexer, error) {
	return d.LexString(filename, string(b))
}

func (d lexerConfigDefinitionImpl) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	s := &strings.Builder{}
	_, err := io.Copy(s, r)
	if err != nil {
		return nil, err
	}
	return d.LexString(filename, s.String())
}

type lexerConfigState struct {
	name   string
	groups []string
}

type lexerConfigImpl struct {
	s      string
	p      int
	pos    lexer.Position
	states []lexerConfigState
}

func (l *lexerConfigImpl) Next() (lexer.Token, error) {
	if l.p == len(l.s) {
		return lexer.EOFToken(l.pos), nil
	}
	var (
		state  = l.states[len(l.states)-1]
		groups []int
		sym    lexer.TokenType
	)
	switch state.name {
	case "Root":
		if match := matchConfigTemplate(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -2
			groups = match[:]
		} else if match := matchConfigComment(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -3
			groups = match[:]
		} else if match := matchConfigWhitespace(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -4
			groups = match[:]
		} else if match := matchConfigPunct(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -5
			groups = match[:]
		} else if match := matchConfigNumber(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -6
			groups = match[:]
		} else if match := matchConfigString(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -7
			groups = match[:]
		} else if match := matchConfigIdent(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -8
			groups = match[:]
		} else if match := matchConfigBareword(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -9
			groups = match[:]
		}
	}
	if groups == nil {
		sample := []rune(l.s[l.p:])
		if len(sample) > 16 {
			sample = append(sample[:16], []rune("...")...)
		}
		return lexer.Token{}, participle.Errorf(l.pos, "invalid input text %q", string(sample))
	}
	pos := l.pos
	span := l.s[groups[0]:groups[1]]
	l.p = groups[1]
	l.pos.Advance(span)
	return lexer.Token{
		Type:  sym,
		Value: span,
		Pos:   pos,
	}, nil
}

func (l *lexerConfigImpl) sgroups(match []int) []string {
	sgroups := make([]string, len(match)/2)
	for i := 0; i < len(match)-1; i += 2 {
		sgroups[i/2] = l.s[l.p+match[i] : l.p+match[i+1]]
	}
	return sgroups
}

// <<([^>]|>[^>])*>>
func matchConfigTemplate(s string, p int, backrefs []string) (groups [4]int) {
	// << (Literal)
	l0 := func(s string, p int) int {
		if p+2 <= len(s) && s[p:p+2] == "<<" {
			return p + 2
		}
		return -1
	}
	// [^>] (CharClass)
	l1 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn >= '\x00' && rn <= '=':
			return p + 1
		case rn >= '?' && rn <= '\U0010ffff':
			return p + n
		}
		return -1
	}
	// > (Literal)
	l2 := func(s string, p int) int {
		if p < len(s) && s[p] == '>' {
			return p + 1
		}
		return -1
	}
	// >[^>] (Concat)
	l3 := func(s string, p int) int {
		if p = l2(s, p); p == -1 {
			return -1
		}
		if p = l1(s, p); p == -1 {
			return -1
		}
		return p
	}
	// [^>]|>[^>] (Alternate)
	l4 := func(s string, p int) int {
		if np := l1(s, p); np != -1 {
			return np
		}
		if np := l3(s, p); np != -1 {
			return np
		}
		return -1
	}
	// ([^>]|>[^>]) (Capture)
	l5 := func(s string, p int) int {
		np := l4(s, p)
		if np != -1 {
			groups[2] = p
			groups[3] = np
		}
		return np
	}
	// ([^>]|>[^>])* (Star)
	l6 := func(s string, p int) int {
		for len(s) > p {
			if np := l5(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// >> (Literal)
	l7 := func(s string, p int) int {
		if p+2 <= len(s) && s[p:p+2] == ">>" {
			return p + 2
		}
		return -1
	}
	// <<([^>]|>[^>])*>> (Concat)
	l8 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		if p = l6(s, p); p == -1 {
			return -1
		}
		if p = l7(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l8(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// //[^\n\r]*
func matchConfigComment(s string, p int, backrefs []string) (groups [2]int) {
	// // (Literal)
	l0 := func(s string, p int) int {
		if p+2 <= len(s) && s[p:p+2] == "//" {
			return p + 2
		}
		return -1
	}
	// [^\n\r] (CharClass)
	l1 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn >= '\x00' && rn <= '\t':
			return p + 1
		case rn >= '\v' && rn <= '\f':
			return p + 1
		case rn >= '\x0e' && rn <= '\U0010ffff':
			return p + n
		}
		return -1
	}
	// [^\n\r]* (Star)
	l2 := func(s string, p int) int {
		for len(s) > p {
			if np := l1(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// //[^\n\r]* (Concat)
	l3 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		if p = l2(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l3(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// [\t\n\f\r ]+
func matchConfigWhitespace(s string, p int, backrefs []string) (groups [2]int) {
	// [\t\n\f\r ] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch {
		case rn >= '\t' && rn <= '\n':
			return p + 1
		case rn >= '\f' && rn <= '\r':
			return p + 1
		case rn == ' ':
			return p + 1
		}
		return -1
	}
	// [\t\n\f\r ]+ (Plus)
	l1 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		for len(s) > p {
			if np := l0(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	np := l1(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// [,:=\[\]\{\}]
func matchConfigPunct(s string, p int, backrefs []string) (groups [2]int) {
	// [,:=\[\]\{\}] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch rn {
		case ',', ':', '=', '[', ']', '{', '}':
			return p + 1
		}
		return -1
	}
	np := l0(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// [\+\-]?([0-9]*\.)?[0-9]+\b
func matchConfigNumber(s string, p int, backrefs []string) (groups [4]int) {
	// [\+\-] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		if rn == '+' || rn == '-' {
			return p + 1
		}
		return -1
	}
	// [\+\-]? (Quest)
	l1 := func(s string, p int) int {
		if np := l0(s, p); np != -1 {
			return np
		}
		return p
	}
	// [0-9] (CharClass)
	l2 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch {
		case rn >= '0' && rn <= '9':
			return p + 1
		}
		return -1
	}
	// [0-9]* (Star)
	l3 := func(s string, p int) int {
		for len(s) > p {
			if np := l2(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// \. (Literal)
	l4 := func(s string, p int) int {
		if p < len(s) && s[p] == '.' {
			return p + 1
		}
		return -1
	}
	// [0-9]*\. (Concat)
	l5 := func(s string, p int) int {
		if p = l3(s, p); p == -1 {
			return -1
		}
		if p = l4(s, p); p == -1 {
			return -1
		}
		return p
	}
	// ([0-9]*\.) (Capture)
	l6 := func(s string, p int) int {
		np := l5(s, p)
		if np != -1 {
			groups[2] = p
			groups[3] = np
		}
		return np
	}
	// ([0-9]*\.)? (Quest)
	l7 := func(s string, p int) int {
		if np := l6(s, p); np != -1 {
			return np
		}
		return p
	}
	// [0-9]+ (Plus)
	l8 := func(s string, p int) int {
		if p = l2(s, p); p == -1 {
			return -1
		}
		for len(s) > p {
			if np := l2(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// \b (WordBoundary)
	l9 := func(s string, p int) int {
		var l, u rune = -1, -1
		if p == 0 {
			if p < len(s) {
				if s[0] < utf8.RuneSelf {
					u, _ = rune(s[0]), 1
				} else {
					u, _ = utf8.DecodeRuneInString(s[0:])
				}
			}
		} else if p == len(s) {
			l, _ = utf8.DecodeLastRuneInString(s)
		} else {
			l, _ = utf8.DecodeLastRuneInString(s[0:p])
			if s[p] < utf8.RuneSelf {
				u, _ = rune(s[p]), 1
			} else {
				u, _ = utf8.DecodeRuneInString(s[p:])
			}
		}
		op := syntax.EmptyOpContext(l, u)
		if op&syntax.EmptyWordBoundary != 0 {
			return p
		}
		return -1
	}
	// [\+\-]?([0-9]*\.)?[0-9]+\b (Concat)
	l10 := func(s string, p int) int {
		if p = l1(s, p); p == -1 {
			return -1
		}
		if p = l7(s, p); p == -1 {
			return -1
		}
		if p = l8(s, p); p == -1 {
			return -1
		}
		if p = l9(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l10(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// (?-s:"(\\.|[^"])*")
func matchConfigString(s string, p int, backrefs []string) (groups [4]int) {
	// " (Literal)
	l0 := func(s string, p int) int {
		if p < len(s) && s[p] == '"' {
			return p + 1
		}
		return -1
	}
	// \\ (Literal)
	l1 := func(s string, p int) int {
		if p < len(s) && s[p] == '\\' {
			return p + 1
		}
		return -1
	}
	// (?-s:.) (AnyCharNotNL)
	l2 := func(s string, p int) int {
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		if len(s) <= p+n || rn == '\n' {
			return -1
		}
		return p + n
	}
	// (?-s:\\.) (Concat)
	l3 := func(s string, p int) int {
		if p = l1(s, p); p == -1 {
			return -1
		}
		if p = l2(s, p); p == -1 {
			return -1
		}
		return p
	}
	// [^"] (CharClass)
	l4 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn >= '\x00' && rn <= '!':
			return p + 1
		case rn >= '#' && rn <= '\U0010ffff':
			return p + n
		}
		return -1
	}
	// (?-s:\\.|[^"]) (Alternate)
	l5 := func(s string, p int) int {
		if np := l3(s, p); np != -1 {
			return np
		}
		if np := l4(s, p); np != -1 {
			return np
		}
		return -1
	}
	// (?-s:(\\.|[^"])) (Capture)
	l6 := func(s string, p int) int {
		np := l5(s, p)
		if np != -1 {
			groups[2] = p
			groups[3] = np
		}
		return np
	}
	// (?-s:(\\.|[^"])*) (Star)
	l7 := func(s string, p int) int {
		for len(s) > p {
			if np := l6(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// (?-s:"(\\.|[^"])*") (Concat)
	l8 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		if p = l7(s, p); p == -1 {
			return -1
		}
		if p = l0(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l8(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// [A-Z_a-zſK][\-\.0-9A-Z_a-zſK]*
func matchConfigIdent(s string, p int, backrefs []string) (groups [2]int) {
	// [A-Z_a-zſK] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn >= 'A' && rn <= 'Z':
			return p + 1
		case rn == '_':
			return p + 1
		case rn >= 'a' && rn <= 'z':
			return p + 1
		case rn == 'ſ':
			return p + n
		case rn == 'K':
			return p + n
		}
		return -1
	}
	// [\-\.0-9A-Z_a-zſK] (CharClass)
	l1 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn >= '-' && rn <= '.':
			return p + 1
		case rn >= '0' && rn <= '9':
			return p + 1
		case rn >= 'A' && rn <= 'Z':
			return p + 1
		case rn == '_':
			return p + 1
		case rn >= 'a' && rn <= 'z':
			return p + 1
		case rn == 'ſ':
			return p + n
		case rn == 'K':
			return p + n
		}
		return -1
	}
	// [\-\.0-9A-Z_a-zſK]* (Star)
	l2 := func(s string, p int) int {
		for len(s) > p {
			if np := l1(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// [A-Z_a-zſK][\-\.0-9A-Z_a-zſK]* (Concat)
	l3 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		if p = l2(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l3(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// [/A-Z_a-zſK][^\t\n\f\r ,:=\[\]\{\}]*
func matchConfigBareword(s string, p int, backrefs []string) (groups [2]int) {
	// [/A-Z_a-zſK] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn == '/':
			return p + 1
		case rn >= 'A' && rn <= 'Z':
			return p + 1
		case rn == '_':
			return p + 1
		case rn >= 'a' && rn <= 'z':
			return p + 1
		case rn == 'ſ':
			return p + n
		case rn == 'K':
			return p + n
		}
		return -1
	}
	// [^\t\n\f\r ,:=\[\]\{\}] (CharClass)
	l1 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn >= '\x00' && rn <= '\b':
			return p + 1
		case rn == '\v':
			return p + 1
		case rn >= '\x0e' && rn <= '\x1f':
			return p + 1
		case rn >= '!' && rn <= '+':
			return p + 1
		case rn >= '-' && rn <= '9':
			return p + 1
		case rn >= ';' && rn <= '<':
			return p + 1
		case rn >= '>' && rn <= 'Z':
			return p + 1
		case rn == '\\':
			return p + 1
		case rn >= '^' && rn <= 'z':
			return p + 1
		case rn == '|':
			return p + 1
		case rn >= '~' && rn <= '\U0010ffff':
			return p + n
		}
		return -1
	}
	// [^\t\n\f\r ,:=\[\]\{\}]* (Star)
	l2 := func(s string, p int) int {
		for len(s) > p {
			if np := l1(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// [/A-Z_a-zſK][^\t\n\f\r ,:=\[\]\{\}]* (Concat)
	l3 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		if p = l2(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l3(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}
