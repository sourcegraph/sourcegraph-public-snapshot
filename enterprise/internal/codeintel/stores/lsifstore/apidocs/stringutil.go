package apidocs

import (
	"unicode"
	"unicode/utf8"
)

// Truncate truncates a string to limitBytes, taking into account multi-byte runes. If the string
// is truncated, an ellipsis "…" is added to the end.
func Truncate(s string, limitBytes int) string {
	runes := []rune(s)
	bytes := 0
	for i, r := range runes {
		if bytes+len([]byte(string(r))) > limitBytes {
			return string(runes[:i]) + "…"
		}
		bytes += len(string(r))
	}
	return s
}

// Reverse a UTF-8 string accounting for Unicode and combining characters. This is not a part of
// the Go standard library or any of the golang.org/x packages. Note that reversing a slice of
// runes is not enough (would not handle combining characters.)
//
// See http://rosettacode.org/wiki/Reverse_a_string#Go
func Reverse(s string) string {
	if s == "" {
		return ""
	}
	p := []rune(s)
	r := make([]rune, len(p))
	start := len(r)
	for i := 0; i < len(p); {
		// quietly skip invalid UTF-8
		if p[i] == utf8.RuneError {
			i++
			continue
		}
		j := i + 1
		for j < len(p) && (unicode.Is(unicode.Mn, p[j]) ||
			unicode.Is(unicode.Me, p[j]) || unicode.Is(unicode.Mc, p[j])) {
			j++
		}
		for k := j - 1; k >= i; k-- {
			start--
			r[start] = p[k]
		}
		i = j
	}
	return (string(r[start:]))
}
