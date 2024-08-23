package parser

// isNewlineChar returns whether the given character is '\n' or '\r'.
func isNewlineChar(b byte) bool {
	return b == '\n' || b == '\r'
}

// NewlineIndex returns the index of the first occurrence of a newline sequence (\n, \r, or \r\n).
// It also returns the sequence's length. If no sequence is found, index is equal to len(s)
// and length is 0.
//
// The newline is defined in the Event Stream standard's documentation:
// https://html.spec.whatwg.org/multipage/server-sent-events.html#server-sent-events
func NewlineIndex(s string) (index, length int) {
	for l := len(s); index < l; index++ {
		b := s[index]

		if isNewlineChar(b) {
			length++
			if b == '\r' && index < l-1 && s[index+1] == '\n' {
				length++
			}

			break
		}
	}

	return
}

// NextChunk retrieves the next chunk of data from the given string
// along with the data remaining after the returned chunk.
// A chunk is a string of data delimited by a newline.
// If the returned chunk is the last one, len(remaining) will be 0.
//
// The newline is defined in the Event Stream standard's documentation:
// https://html.spec.whatwg.org/multipage/server-sent-events.html#server-sent-events
func NextChunk(s string) (chunk, remaining string, hasNewline bool) {
	index, endlineLen := NewlineIndex(s)
	return s[:index], s[index+endlineLen:], endlineLen != 0
}
