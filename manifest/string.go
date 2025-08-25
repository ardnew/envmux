package manifest

import (
	"strconv"
)

// Unquote returns the unquoted value of a string or byte slice.
// If the value is not a string or byte slice, it is returned unchanged.
func unquote(value any) any {
	var s string

	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	case interface{ String() string }:
		s = v.String()
	}

	if us, err := strconv.Unquote(s); err == nil {
		return us
	}

	return value
}
