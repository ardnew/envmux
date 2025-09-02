// Package shell defines a generic variable declaration syntax compatible with
// most environments.
//
// Users can configure how identifiers are formatted in two ways,
// depending on how much adjustment is needed.
//
//  1. For simple adjustments, modify [DefaultIdentFormat] fields and keep the
//     default formatting function [MakeIdent].
//  2. Otherwise, assign a custom formatting function to [MakeIdent] for full
//     control of the format.
//
// The envmux packages call [MakeIdent] to format variable identifiers,
// assigning to either one of these exported variables is sufficient.
package shell

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ardnew/envmux/pkg"
)

// IdentFormat defines formatting options for variable identifiers.
type IdentFormat struct {
	Case    cases.Caser // Letter case transformation.
	Break   []byte      // String to insert between runs.
	Unicode bool        // Accept Unicode code points as valid glyphs.
}

// DefaultIdentFormat is the default formatting for variable identifiers.
//
// It replaces all runs of invalid glyphs with a single underscore and converts
// all valid glyphs to uppercase.
//
//nolint:gochecknoglobals
var DefaultIdentFormat = IdentFormat{
	Case:    cases.Upper(language.Und), // Use all-uppercase glyphs.
	Break:   []byte{'_'},               // Separate valid words with underscore.
	Unicode: false,                     // Do not accept Unicode (ASCII-only).
}

// MakeIdent is the default formatting function for variable identifiers.
// It uses [DefaultIdentFormat] to format words as a single variable identifier.
var MakeIdent = func(words ...string) string {
	return DefaultIdentFormat.makeIdent(words...)
}

// MakeIdent formats the given words as a single variable identifier
// according to the receiver's formatting options.
func (f IdentFormat) makeIdent(words ...string) string {
	f = f.asValid()

	var sb strings.Builder

	brk := false
	for i, s := range words {
		if i > 0 && !brk {
			brk = true

			sb.Write(f.Break)
		}

		brk = f.appendIdent(&sb, s, i == 0, brk)
	}

	return sb.String()
}

// appendIdent formats a single word as an environment variable identifier
// and appends it to the current identifier constructed so far in sb.
func (f IdentFormat) appendIdent(
	sb *strings.Builder,
	s string,
	isFirstWord, brk bool,
) bool {
	r := []rune(strings.TrimSpace(s))
	t := []rune(f.Case.String(string(r)))

	for j := range r {
		isLetter := isLetter(r[j])
		isDigit := isDigit(r[j])

		switch {
		case isFirstWord && j == 0 && isDigit:
			brk = false

			sb.Write(f.Break)
			sb.WriteRune(t[j])
		case isLetter || isDigit:
			brk = false

			sb.WriteRune(t[j])
		default:
			if !brk {
				brk = true

				sb.Write(f.Break)
			}
		}
	}

	return brk
}

func (f IdentFormat) isValid() bool {
	return f.Case != (cases.Caser{}) && f.Break != nil
}

func (f IdentFormat) asValid() IdentFormat {
	if f.isValid() {
		return f
	}

	if !DefaultIdentFormat.isValid() {
		panic(pkg.ErrInvalidIdentifier)
	}

	if f.Case == (cases.Caser{}) {
		f.Case = DefaultIdentFormat.Case
	}

	if f.Break == nil {
		f.Break = DefaultIdentFormat.Break
	}

	return f
}

func isLetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
