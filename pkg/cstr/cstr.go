package cstr

func IsASCIILetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func IsASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// ParseASCII ensures the given byte slice contains nothing or a sequence
// of commonly-used ASCII characters, which may optionally contain a
// terminating null byte ('\0'), and returns that sequence as a string
// without the terminating null byte.
//
// The commonly-used ASCII characters are defined as:
//
//   - ASCII characters in the range 0x20 to 0x7E (inclusive)
//   - 0x09 '\t' (TAB)
//   - 0x0A '\n' (LF)
//   - 0x0D '\r' (CR)
//
// The empty byte slice satisfies these conditions and returns an empty string.
func ParseASCII(b []byte) (string, bool) {
	if len(b) == 0 {
		return "", true
	}

	n := len(b)

	// Checking n > 1 ensures we do not have the single-null-byte slice {'\0'}.
	// If n == 1, the condition in the loop below will then return false.
	if n > 1 && b[n-1] == 0 {
		n-- // Exclude the terminating null byte.
	}

	for _, r := range b[:n] {
		if (r < 0x20 || 0x7F <= r) && r != '\t' && r != '\n' && r != '\r' {
			return "", false
		}
	}

	return string(b[:n]), true
}
