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
	Ident     string
	arguments []Argument
}

// Argument is an actual [Parameter] value passed to the [Namespace] being
// composited.
type Argument Parameter

func (c Composite) String() string {
	if c.Ident == "" {
		return ""
	}

	var par string

	if len(c.arguments) > 0 {
		pars := make([]string, len(c.arguments))
		for i, p := range c.arguments {
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
func (c Composite) Arguments() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, p := range c.arguments {
			if !yield(p.Value) {
				return
			}
		}
	}
}
