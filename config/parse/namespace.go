package parse

import (
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/ardnew/envmux/pkg"
)

// Namespace associates a composition of environment variable definitions with
// a namespace identifier.
//
// Variable definitions are expressed entirely with the [expr-lang] grammar.
//
// [expr-lang]: https://github.com/expr-lang/expr
type Namespace struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	ID string

	Com Composites
	Par Parameters
	Sta Statements
}

// Parse parses an expression using the provided lexer
// and stores the unevaluated source code in the Expr's Src field.
// Returns an error if parsing fails.
func (n *Namespace) Parse(lex *lexer.PeekingLexer) error {
	err := n.parse(lex)
	if err != nil {
		return &pkg.NamespaceError{ID: n.ID, Err: err}
	}

	return nil
}

func (n *Namespace) parse(lex *lexer.PeekingLexer) error {
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
	}

	return nil
}

// Namespaces represents a sequence of Namespace nodes in the AST.
type Namespaces struct {
	Pos    lexer.Position // Pos records the start position of the node.
	EndPos lexer.Position // EndPos records the end position of the node.
	Tokens []lexer.Token  // Tokens records the tokens consumed by the node

	List []*Namespace
}

func (n *Namespaces) Parse(lex *lexer.PeekingLexer) error {
	if err := n.parse(lex); err != nil {
		return err
	}

	return nil
}

func (n *Namespaces) parse(lex *lexer.PeekingLexer) error {
	advance := consume(lex, `XX`, `RS`)

	for advance() {
		ns := new(Namespace)
		if err := ns.Parse(lex); err != nil {
			return err
		}

		n.List = append(n.List, ns)
	}

	return nil
}
