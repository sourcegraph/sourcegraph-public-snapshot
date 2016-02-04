// Package trim contains helpers for trimming strings.
package trim

// LastNewline trims the last newline character from s, if any.
func LastNewline(s string) string {
	if len(s) < 1 || s[len(s)-1] != '\n' {
		return s
	}
	return s[:len(s)-1]
}

// FirstSpace trims the first space character from s, if any.
func FirstSpace(s string) string {
	if len(s) < 1 || s[0] != ' ' {
		return s
	}
	return s[1:]
}
