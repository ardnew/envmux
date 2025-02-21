package pkg

import "strings"

func FormatEnvVar(pre ...string) string {
	var sb strings.Builder
	uscore := false
	for _, s := range pre {
		sb.WriteString(
			strings.Map(
				func(u rune) rune {
					// Verify u is a letter or digit
					if (u >= 'A' && u <= 'Z') || (u >= '0' && u <= '9') {
						uscore = false
						return u
					}
					// Replace all others with at most 1 underscore
					if !uscore {
						uscore = true
						return '_'
					}
					return -1 // discard consecutive underscores
				},
				strings.ToUpper(s),
			),
		)
	}
	return sb.String()
}
