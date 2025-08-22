package ev

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ardnew/envmux/pkg/cstr"
	"github.com/ardnew/envmux/pkg/errs"
)

// EnvVarOption configures environment variable identifier formatting.
type EnvVarOption struct {
	Case    cases.Caser // Letter case transformation.
	Break   []byte      // String to insert between runs.
	Unicode bool        // Accept Unicode code points as valid glyphs.
}

// DefaultEnvVarOption is the default formatting for variable identifiers.
//
// It replaces all runs of invalid glyphs with a single underscore and converts
// all valid glyphs to uppercase.
//
//nolint:gochecknoglobals
var DefaultEnvVarOption = EnvVarOption{
	Case:    cases.Upper(language.Und), // Use all-uppercase glyphs.
	Break:   []byte{'_'},               // Separate valid runs with an underscore.
	Unicode: false,                     // Do not accept Unicode (ASCII-only).
}

// FormatEnvVar formats a run of strings as an environment variable identifier
// using [DefaultEnvVarOption].
func FormatEnvVar(run ...string) string {
	return DefaultEnvVarOption.FormatEnvVar(run...)
}

// FormatEnvVar formats a run of strings as an environment variable identifier.
func (o EnvVarOption) FormatEnvVar(run ...string) string {
	o = o.asValid()

	var sb strings.Builder

	brk := false
	for i, s := range run {
		if i > 0 && !brk {
			brk = true

			sb.Write(o.Break)
		}

		brk = o.formatEnvVarWord(&sb, s, i == 0, brk)
	}

	return sb.String()
}

// formatEnvVarWord formats a single word as an environment variable identifier
// and appends it to the current identifier constructed so far in sb.
func (o EnvVarOption) formatEnvVarWord(
	sb *strings.Builder,
	s string,
	isFirstRun, brk bool,
) bool {
	r := []rune(strings.TrimSpace(s))
	t := []rune(o.Case.String(string(r)))

	for j := range r {
		isLetter := cstr.IsASCIILetter(r[j])
		isDigit := cstr.IsASCIIDigit(r[j])

		switch {
		case isFirstRun && j == 0 && isDigit:
			brk = false

			sb.Write(o.Break)
			sb.WriteRune(t[j])
		case isLetter || isDigit:
			brk = false

			sb.WriteRune(t[j])
		default:
			if !brk {
				brk = true

				sb.Write(o.Break)
			}
		}
	}

	return brk
}

func (o EnvVarOption) isValid() bool {
	return o.Case != (cases.Caser{}) && o.Break != nil
}

func (o EnvVarOption) asValid() EnvVarOption {
	if o.isValid() {
		return o
	}

	if !DefaultEnvVarOption.isValid() {
		panic(errs.ErrInvalidEnvVar)
	}

	if o.Case == (cases.Caser{}) {
		o.Case = DefaultEnvVarOption.Case
	}

	if o.Break == nil {
		o.Break = DefaultEnvVarOption.Break
	}

	return o
}
