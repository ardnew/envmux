package parse

import (
	"fmt"
)

// Parameter represents a value that can be referenced
// using the implicit variable named by
// [github.com/ardnew/envmux/spec/env/vars.ParameterKey]
// in each [statement.expression] of a [namespace].
type Parameter struct {
	Value any
}

func (p Parameter) String() string {
	if p.Value == nil {
		return ""
	}

	return fmt.Sprintf("%v", p.Value)
}
