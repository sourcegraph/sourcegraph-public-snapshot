package sourcegraph

import (
	"bytes"
	"unicode/utf8"
)

// Join joins tokens to reconstruct a query string (that, when
// tokenized, would yield the same tokens). It returns the query and
// the insertion point (which is the position of the active token's
// last character, or the position after the last token's last
// character if there is no active token).
func Join(tokens []Token) RawQuery {
	ip := -1
	var buf bytes.Buffer
	for i, tok := range tokens {
		if i != 0 {
			buf.Write([]byte{' '})
		}
		buf.WriteString(tok.Token())
	}
	if ip == -1 && len(tokens) > 0 {
		ip = utf8.RuneCount(buf.Bytes()) + 1
	}
	if ip == -1 {
		ip = 0
	}
	return RawQuery{Text: buf.String(), InsertionPoint: int32(ip)}
}
