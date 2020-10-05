package fs

import (
	"strconv"
	"strings"
)

// Unicode return unicode
func Unicode(s string) string {
	r, _ := strconv.ParseInt(strings.TrimPrefix(s, "\\U"), 16, 32)
	return string(r)
}
