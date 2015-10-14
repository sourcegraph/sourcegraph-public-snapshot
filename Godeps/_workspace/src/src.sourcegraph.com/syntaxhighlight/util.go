package syntaxhighlight

import "unicode/utf8"

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

// Matches single-line comments // ....
var SingleLineCommentMatcher = func(source []byte) []int {
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
			if c == '/' {
				state = 2
			} else {
				return nil
			}
			break
		case 2:
			if c == '\n' || c == '\r' {
				return []int{0, i}
			}
			break
		}
	}
	return []int{0, len(source)}
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
			} else {
				return nil
			}
			break
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
