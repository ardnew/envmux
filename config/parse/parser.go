package parse

import (
	"strconv"
	"sync"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	LexerDefinition = sync.OnceValue(func() *lexer.StatefulDefinition {
		return lexer.MustSimple([]lexer.SimpleRule{
			{Name: "Template", Pattern: `<<([^>]|>[^>])*>>`},
			{Name: "Comment", Pattern: `//[^\r\n]*`},
			{Name: "Whitespace", Pattern: `\s+`},
			{Name: "Punct", Pattern: `[,{}=:\[\]]`},
			{Name: "Number", Pattern: `[-+]?(\d*\.)?\d+\b`},
			{Name: "String", Pattern: `"(\\.|[^"])*"`},
			{Name: "Ident", Pattern: `(?i)[a-z_][a-z0-9_\-\.]*`},
			{Name: "Bareword", Pattern: `(?i)[a-z_/][^\s,{}=:\[\]]*`},
		})
	})
	LexerToken = sync.OnceValue(func() func(string) lexer.TokenType {
		sym := lexer.DefaultDefinition.Symbols()
		return func(s string) lexer.TokenType {
			t, ok := sym[s]
			if !ok {
				return 0
			}
			return t
		}
	})
)

// Terminal productions.
type (
	Boolean  bool
	Template string
)

func (b *Boolean) GoString() string { return strconv.FormatBool(bool(*b)) }
func (b *Boolean) Parse(lex *lexer.PeekingLexer) error {
	val, err := strconv.ParseBool(lex.Peek().Value)
	if err != nil {
		return participle.NextMatch
	}
	lex.Next()
	*b = Boolean(val)
	return nil
}

// Non-terminal productions.
type (
	Value struct {
		Pos lexer.Position

		Template string   `  @Template`
		Boolean  *Boolean `| @("true"|"false")`
		Number   string   `| @Number`
		String   string   `| @(String|Ident|Bareword)`
	}

	Composition struct {
		Pos lexer.Position

		Name []string `":" ( @Ident ","? )*`
	}

	Mapping struct {
		Pos lexer.Position

		Key   string `@Ident`
		Value *Value `"=" @@`
	}

	Dictionary struct {
		Pos lexer.Position

		Map []*Mapping `"{" ( @@ ","? )* "}"`
	}

	Package struct {
		Pos lexer.Position

		Path []string `"[" ( @(String|Ident|Bareword) ","? )* "]"`
	}

	Definition struct {
		Pos    lexer.Position
		Tokens []lexer.Token

		Comp *Composition `@@?`
		Dict *Dictionary  `@@?`
		Pack *Package     `@@?`
	}

	Namespace struct {
		Pos lexer.Position

		Default    string      `( @"default"`
		Name       string      `  | @Ident )`
		Definition *Definition `@@!`
	}

	Source struct {
		Pos lexer.Position

		Block []*Namespace `( @@ ";"* )*`
	}
)

type Config struct {
	*participle.Parser[Source]
}

func Make(def lexer.Definition) Config {
	// maps.Keys(ConfigLexer.Symbols())
	return Config{
		Parser: participle.MustBuild[Source](
			participle.Lexer(def),
			participle.Unquote("String"),
			participle.Elide("Whitespace", "Comment"),
			participle.UseLookahead(participle.MaxLookahead),
		),
	}
}
