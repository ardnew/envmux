package parse

import (
	"fmt"
	"iter"
	"strings"
)

// StringerAlwaysShowsMeta controls whether String methods include delimiters
// around composite, parameter, and statement metadata.
const StringerAlwaysShowsMeta = true

// Namespace associates a composition of environment variable definitions with
// a Namespace identifier.
//
// Variable definitions are expressed entirely with the [expr-lang] grammar.
//
// [expr-lang]: https://github.com/expr-lang/expr
type Namespace struct {
	Ident      string
	Composites []Composite
	Parameters []Parameter
	Statements []Statement
}

// String renders the namespace in a compact manifest-like representation.
func (n Namespace) String() string {
	if n.Ident == "" {
		return ""
	}

	var com, par, sta string

	if len(n.Composites) > 0 {
		coms := make([]string, len(n.Composites))
		for i, c := range n.Composites {
			coms[i] = c.String()
		}

		com = strings.Join(coms, FS)
		if !StringerAlwaysShowsMeta {
			com = fmt.Sprintf("%s%s%s", co, com, cc)
		}
	}

	if len(n.Parameters) > 0 {
		pars := make([]string, len(n.Parameters))
		for i, p := range n.Parameters {
			pars[i] = p.String()
		}

		par = strings.Join(pars, FS)
		if !StringerAlwaysShowsMeta {
			par = fmt.Sprintf("%s%s%s", po, par, pc)
		}
	}

	if len(n.Statements) > 0 {
		stas := make([]string, len(n.Statements))
		for i, s := range n.Statements {
			stas[i] = s.String()
		}

		sta = strings.Join(stas, RS)
		if !StringerAlwaysShowsMeta {
			sta = fmt.Sprintf("%s%s%s", so, sta, sc)
		}
	}

	if StringerAlwaysShowsMeta {
		com = fmt.Sprintf("%s%s%s", co, com, cc)
		par = fmt.Sprintf("%s%s%s", po, par, pc)
		sta = fmt.Sprintf("%s%s%s", so, sta, sc)
	}

	return fmt.Sprintf("%s%s%s%s", n.Ident, com, par, sta)
}

// Arguments returns a [Parameter.Value] sequence of each [Namespace.Parameter].
// Arguments yields the raw parameter values of the namespace in order.
func (n Namespace) Arguments() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, p := range n.Parameters {
			if !yield(p.Value) {
				return
			}
		}
	}
}
