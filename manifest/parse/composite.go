package parse

import (
	"fmt"
	"iter"
	"strings"
)

// Composite contains a namespace identifier whose evaluated environment can be
// inherited in the definition of a [namespace].
//
// A Composite can optionally specify an in-line [parameter] list that is
// appended to that [namespace]'s own parameter list definition.
type Composite struct {
	Ident      string
	Parameters []Parameter
}

// String renders the composite in a compact manifest-like representation.
func (c Composite) String() string {
	if c.Ident == "" {
		return ""
	}

	var par string

	if len(c.Parameters) > 0 {
		pars := make([]string, len(c.Parameters))
		for i, p := range c.Parameters {
			pars[i] = fmt.Sprintf(`%v`, p.Value)
		}

		par = strings.Join(pars, FS)
	}

	return fmt.Sprintf(`%s%s%s%s`, c.Ident, po, par, pc)
}

// Arguments returns a [Parameter.Value] sequence of all [Composite.Parameter]s.
//
// These parameters are specified in-line in the [Composite] list of a
// [Namespace] definition, which get appended to the composited
// [Namespace.Parameters], but only for that evaluated instance.
//
// Arguments yields the ordered, raw parameter values of the composite.
func (c Composite) Arguments() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, p := range c.Parameters {
			if !yield(p.Value) {
				return
			}
		}
	}
}
