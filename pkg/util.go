package pkg

import (
	"errors"
	"fmt"
	"iter"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// EnvVarOption configures environment variable identifier formatting.
type EnvVarOption struct {
	Case  cases.Caser // Letter case transformation.
	Break []byte      // String to insert between runs.
}

// The default format uses all-uppercase glyphs,
// and it uses an underscore to replace invalid glyphs and separate runs.
var DefaultEnvVarOption = EnvVarOption{
	Case:  cases.Upper(language.Und),
	Break: []byte{'_'},
}

var ErrDefaultEnvVarOption = fmt.Errorf(
	"%w: EnvVarOption: invalid default value", errors.ErrUnsupported,
)

func (o EnvVarOption) isValid() bool {
	return o.Case != (cases.Caser{}) && o.Break != nil
}

func (o EnvVarOption) asValid() EnvVarOption {
	if o.isValid() {
		return o
	}
	if !DefaultEnvVarOption.isValid() {
		panic(ErrDefaultEnvVarOption)
	}
	if o.Case == (cases.Caser{}) {
		o.Case = DefaultEnvVarOption.Case
	}
	if o.Break == nil {
		o.Break = DefaultEnvVarOption.Break
	}
	return o
}

// FormatEnvVar formats a string as an environment variable identifier
// using [DefaultEnvVarOption].
func FormatEnvVar(run ...string) string {
	return DefaultEnvVarOption.FormatEnvVar(run...)
}

// FormatEnvVar formats a string as an environment variable identifier.
func (o EnvVarOption) FormatEnvVar(run ...string) string {
	o = o.asValid()
	var sb strings.Builder
	brk := false
	for i, s := range run {
		if i > 0 && !brk {
			brk = true
			sb.Write(o.Break)
		}
		r := []rune(strings.TrimSpace(s))
		t := []rune(o.Case.String(string(r)))
		for j := range r {
			isAlpha := (r[j] >= 'A' && r[j] <= 'Z') || (r[j] >= 'a' && r[j] <= 'z')
			isDigit := (r[j] >= '0' && r[j] <= '9')
			switch {
			case i+j == 0 && isDigit:
				brk = false
				sb.Write(o.Break)
				sb.WriteRune(t[j])
			case isAlpha || isDigit:
				brk = false
				sb.WriteRune(t[j])
			default:
				if !brk {
					brk = true
					sb.Write(o.Break)
				}
			}
		}
	}
	return sb.String()
}

// Map returns a new sequence that yields elements of s transformed by f.
func Map[T any](s iter.Seq[T], f func(T) T) iter.Seq[T] {
	if s == nil {
		return nil
	}
	if f == nil {
		f = func(x T) T { return x }
	}
	return func(yield func(T) bool) {
		for item := range s {
			if !yield(f(item)) {
				return
			}
		}
	}
}
