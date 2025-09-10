package parse

import (
	"fmt"
)

// Statement associates an expression with a variable identifier and operator.
// Expressions are evaluated in the context of the enclosing namespace.
//
// Expressions use an entirely different grammar than what is recognized by this
// module. The grammar is defined by [github.com/expr-lang/expr].
// Our grammar was designed to accommodate the embedded expression grammar.
type Statement struct {
	Text       string
	Ident      string
	Operator   string
	Expression *Expression
}

func (s Statement) String() string {
	if s.Ident == "" || s.Operator == "" || s.Expression == nil {
		return ""
	}

	return fmt.Sprintf("%s%s%s", s.Ident, s.Operator, s.Expression.String())
}
