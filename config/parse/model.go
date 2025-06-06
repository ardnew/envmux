package parse

import (
	"io"
	"iter"
	"sync"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

// LexerGenerator is used internally to generate lexer.go, which provides
// the concrete implementation of [ConfigLexer].
//
// It is generated by running "go generate" in the same directory as this file.
// You must regenerate lexer.go if you change the rules in this file.
//
//go:generate bash internal/lexer.bash
var LexerGenerator = sync.OnceValue(func() *lexer.StatefulDefinition {
	return lexer.MustSimple([]lexer.SimpleRule{
		// XS matches whitespace and comments elided from the parse tree.
		// But the lexer must still emit them for input reproduction.
		{Name: `XS`, Pattern: `(?:/\*(?:[^*]|\*[^/])*\*/|(?://|#)[^\r\n]*\r?\n|[ \r\n\t])`},
		{Name: `String`, Pattern: `"(?:\\.|[^"])*"`},
		{Name: `Ident`, Pattern: `[./a-zA-Z_][-./:\w]*`},
		{Name: `Number`, Pattern: `(?:0x[0-9a-fA-F]+|0b[01]+|[+-]?(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?)`},
		{Name: `RS`, Pattern: `;`},
		{Name: `FS`, Pattern: `,`},
		{Name: `Punct`, Pattern: `[-[!@#$%^&*()+_={}\|:"'<>.?/]|]`},
	})
})

type (
	Compose struct {
		Name string `XS* @( Ident ) XS*`
	}
	Subject struct {
		Name string `XS* @( String | Ident ) XS*`
	}
	Mapping struct {
		Name string `XS* @( Ident ) XS*`
		Prec string `@( "?" )? XS*`
		Op   string `@( ( ":" | "^" | "+" )? "=" ) XS*`
		Expr *Expr  `@@ XS*`
	}
	Specification struct {
		Coms []*Compose `XS* ( "<" XS* ( @@ XS* ( FS  XS* @@ XS* )* )? ">" XS* )?`
		Subs []*Subject `XS* ( "(" XS* ( @@ XS* ( FS  XS* @@ XS* )* )? ")" XS* )?`
		Maps []*Mapping `XS* ( "{" XS* ( @@ XS* ( RS* XS* @@ XS* )* )? "}" XS* )?`
	}
	Namespace struct {
		Name string         `XS* @( Ident ) XS*`
		Spec *Specification `@@! XS*`
	}
	Namespaces struct {
		List []*Namespace `@@*`
	}
)

var (
	Options      = []participle.Option{participle.Lexer(ConfigLexer)}
	ParseOptions = []participle.ParseOption{participle.AllowTrailing(true)}
)

var build = sync.OnceValue(
	func() *participle.Parser[Namespaces] {
		return participle.MustBuild[Namespaces](Options...)
	},
)

func Grammar() string { return build().String() }

func Build(r io.Reader) (*Namespaces, error) {
	return build().Parse(pkg.Name, r, ParseOptions...)
}

func (s *Specification) Compositions() iter.Seq[string] {
	if s == nil {
		return nil
	}
	unique := make(pkg.Unique[string])
	return func(yield func(string) bool) {
		for _, v := range s.Coms {
			if unique.Add(v.Name) && !yield(v.Name) {
				return
			}
		}
	}
}

func (s *Specification) Subjects(appends ...string) iter.Seq[string] {
	if s == nil {
		return nil
	}
	unique := make(pkg.Unique[string])
	return func(yield func(string) bool) {
		for _, v := range s.Subs {
			if unique.Add(v.Name) && !yield(v.Name) {
				return
			}
		}
		for _, append := range appends {
			if unique.Add(append) && !yield(append) {
				return
			}
		}
	}
}
