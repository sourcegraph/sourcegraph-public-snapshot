package syntaxhighlight

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

// Matches multiline comments /* ... */
var MultiLineCommentMatcher = func(source []byte) []int {
	state := 0
	for i, c := range source {
		switch state {
		case 0:
			if c == '/' {
				state = 1
			} else {
				return nil
			}
			break
		case 1:
			if c == '*' {
				state = 2
			} else {
				return nil
			}
			break
		case 2:
			if c == '*' {
				state = 3
			}
			break
		case 3:
			switch c {
			case '/':
				return []int{0, i + 1}
			case '*':
				break
			default:
				state = 2
				break
			}
			break
		}
	}
	return nil
}

// Matches single-line comment that starts with specific prefix
var SingleLineCommentMatcher = func(prefix string) Matcher {
	return func(source []byte) []int {
		l := len(source)
		lPrefix := len(prefix)
		if lPrefix > l {
			return nil
		}
		if !bytes.HasPrefix(source, []byte(prefix)) {
			return nil
		}
		end := bytes.IndexAny(source, "\r\n")
		if end < 0 {
			return []int{0, l}
		}
		return []int{0, end}
	}
}

// Matches characters java-style
var JavaCharMatcher = func(source []byte) []int {
	state := 0
	ucounter := 0
	buf := make([]byte, 0, 3)
	for i, c := range source {
		switch state {
		case 0:
			if c == '\'' {
				state = 1
			} else {
				return nil
			}
			break
		case 1:
			if c == '\\' {
				state = 2 // '\
			} else {
				buf = append(buf, c)
				r, _ := utf8.DecodeLastRune(buf)
				if r != utf8.RuneError {
					state = 3 // '...
				} else if len(buf) == 3 {
					return nil
				}
			}
			break
		case 2:
			// '\
			if c == 'u' {
				ucounter = 0
				state = 4 // '\u....
			} else {
				state = 3 // '\C
			}
			break
		case 3:
			if c == '\'' {
				return []int{0, i + 1}
			}
			return nil
		case 4:
			if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' {
				ucounter++
				if ucounter == 4 {
					state = 3
				}
			} else {
				return nil
			}
			break
		}
	}
	return nil
}

// Matches hex numbers in form 0[xX][A-Fa-f0-9]
var HexNumberMatcher = func(source []byte) []int {
	state := 0
	for i, c := range source {
		switch state {
		case 0:
			if c == '0' {
				state = 1
			} else {
				return nil
			}
			break
		case 1:
			if c == 'x' || c == 'X' {
				state = 2
			} else {
				return nil
			}
			break
		case 2:
			if c == 'l' || c == 'L' {
				return []int{0, i + 1}
			} else if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' {
				// continue
			} else {
				return nil
			}
			break
		}
	}
	return nil
}

// Checks if given byte indicates number literal modifier
var isNumberModifier = func(c byte, modifiers string) bool {
	r := rune(c)
	return strings.ContainsRune(modifiers, r)
}

// Returns index where potential decimal part literal ends in given slice,
// so number = source[0:decimalPart]
// decimal part could be: \d+.\d+, \.\d+, \d+
var decimalPart = func(source []byte) int {
	state := 0
	for i, c := range source {
		switch state {
		case 0:
			if c == '.' {
				state = 1
			} else if c >= '0' && c <= '9' {
				// continue
			} else {
				return i
			}
			break
		case 1:
			if c >= '0' && c <= '9' {
				// continue
			} else {
				return i
			}
			break
		}
	}
	return len(source)
}

// Returns index where exponent part literal ends in given slice,
// exponent part could be: [+-]\d+
var exponentPart = func(source []byte, modifiers string) int {
	state := 0
	for i, c := range source {
		switch state {
		case 0:
			if c == '+' || c == '-' || (c >= '0' && c <= '9') {
				state = 1
			} else {
				return 0
			}
			break
		case 1:
			if c >= '0' && c <= '9' {
				// continue
			} else if isNumberModifier(c, modifiers) {
				return i + 1
			} else {
				return i
			}
			break
		}
	}
	return len(source)
}

// Matches numbers, modifiers include allows number modifiers (e.g. 10l, 1D)
var NumberMatcher = func(modifiers string) Matcher {
	return func(source []byte) []int {
		pos := decimalPart(source)
		if pos == 0 || (pos == 1 && source[0] == '.') {
			return nil
		}
		l := len(source)
		if pos == l {
			return []int{0, l}
		}
		c := source[pos]

		if isNumberModifier(c, modifiers) {
			return []int{0, pos + 1}
		}
		if c == 'E' || c == 'e' && pos < l-1 {
			ePos := exponentPart(source[pos+1:], modifiers)
			if ePos > 0 {
				return []int{0, pos + 1 + ePos}
			}
			return nil
		}
		return []int{0, pos}
	}
}

// Matches string literals QUOTE...QUOTE,
func StringMatcher(quote byte) Matcher {
	return func(source []byte) []int {
		state := 0
		for i, c := range source {
			switch state {
			case 0:
				if c == quote {
					state = 1
				} else {
					return nil
				}
				break
			case 1:
				if c == '\\' {
					state = 2
				} else if c == quote {
					return []int{0, i + 1}
				}
				break
			case 2:
				state = 1
				break
			}
		}
		return nil
	}
}
