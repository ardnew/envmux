package parse

import (
	"iter"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

// Namespace is a container for the semantic information of a Namespace node.
type Namespace struct {
	ID string

	Com composites
	Par parameters
	Sta statements
}

// namespace associates a composition of environment variable definitions with
// a namespace identifier.
//
// Variable definitions are expressed entirely with the [expr-lang] grammar.
//
// [expr-lang]: https://github.com/expr-lang/expr
type namespace struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	Namespace
}

func (n *namespace) Composites() iter.Seq[Composite] {
	if n == nil {
		return nil
	}

	return n.Com.Seq()
}

func (n *namespace) Parameters() iter.Seq[Parameter] {
	if n == nil {
		return nil
	}

	return nil
}

// Parse parses an expression using the provided lexer
// and stores the unevaluated source code in the Expr's Src field.
// Returns an error if parsing fails.
func (n *namespace) Parse(lex *lexer.PeekingLexer) error {
	err := n.parse(lex)
	if err != nil {
		return &pkg.NamespaceError{ID: n.ID, Err: err}
	}

	return nil
}

func (n *namespace) parse(lex *lexer.PeekingLexer) error {
	advance := consume(lex, `XX`)

	tok := lex.Next()
	if tok.Type != symbol()(`NS`) {
		return &pkg.UnexpectedTokenError{
			Tok: tok,
			Msg: []string{`expected namespace identifier`},
		}
	}

	n.ID = tok.Value

	if !advance() {
		return nil
	}

	if lex.Peek().Type == symbol()(`CO`) {
		if err := n.Com.Parse(lex); err != nil {
			return err
		}

		if !advance() {
			return nil
		}
	}

	if !advance() {
		return nil
	}

	if lex.Peek().Type == symbol()(`PO`) {
		if err := n.Par.Parse(lex); err != nil {
			return err
		}

		if !advance() {
			return nil
		}

		if peek := lex.Peek(); peek.Type == symbol()(`CO`) {
			return &pkg.UnexpectedTokenError{
				Tok: peek,
				Msg: []string{
					`composites "` + co + `…` + cc + `" must be declared before ` +
						`parameters "` + po + `…` + pc + `"`,
				},
			}
		}
	}

	if !advance() {
		return nil
	}

	if lex.Peek().Type == symbol()(`SO`) {
		if err := n.Sta.Parse(lex); err != nil {
			return err
		}

		if !advance() {
			return nil
		}

		//nolint:exhaustive
		switch peek := lex.Peek(); peek.Type {
		case symbol()(`PO`):
			return &pkg.UnexpectedTokenError{
				Tok: peek,
				Msg: []string{
					`parameters "` + po + `…` + pc + `" must be declared before ` +
						`statements "` + so + `…` + sc + `"`,
				},
			}

		case symbol()(`CO`):
			return &pkg.UnexpectedTokenError{
				Tok: peek,
				Msg: []string{
					`composites "` + co + `…` + cc + `" must be declared before ` +
						`statements "` + so + `…` + sc + `"`,
				},
			}
		}
	}

	return nil
}

// namespaces represents a sequence of Namespace nodes in the AST.
type namespaces struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	list []*namespace
}

// Len returns the number of namespaces in the sequence.
func (n *namespaces) Len() int {
	if n == nil {
		return 0
	}

	return len(n.list)
}

// Seq returns the receiver's sequence of namespaces.
func (n *namespaces) Seq() iter.Seq[Namespace] {
	if n == nil {
		return nil
	}

	return func(yield func(Namespace) bool) {
		for _, v := range n.list {
			if !yield(v.Namespace) {
				return
			}
		}
	}
}

func (n *namespaces) Parse(lex *lexer.PeekingLexer) error {
	if err := n.parse(lex); err != nil {
		return err
	}

	return nil
}

func (n *namespaces) parse(lex *lexer.PeekingLexer) error {
	advance := consume(lex, `XX`, `RS`)

	for advance() {
		ns := new(namespace)
		if err := ns.Parse(lex); err != nil {
			return err
		}

		n.list = append(n.list, ns)
	}

	return nil
}
