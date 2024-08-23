// Package wildcard implements simple pattern matching for strings with wildcard characters.
// Asterisk ("*"), which is the only supported wildcard, matches any character including another asterisk and an empty string.
package wildcard

import "strings"

// Match reports whether string s contains any match of the pattern, which may use 0+ wildcards.
func Match(pattern, s string) bool {
	if pattern == "*" || pattern == s {
		return true
	} else if (pattern == "") != (s == "") {
		return false
	}
	return match(pattern, s)
}

// match expects that both pattern and s are non-empty strings.
func match(pattern, s string) bool {
	pp, sp := newParser(pattern), newParser(s)

	pv, sv := pp.read(), sp.read()
	for {
		pNext, pHasNext := pp.peek()
		sNext, sHasNext := sp.peek()

		switch pv {
		default:
			// Characters not equal, or the pattern has been ended prematurely.
			if pv != sv {
				return false
			}

			// Both pattern and s have reached their ends.
			if !pHasNext && !sHasNext {
				return true
			}

			// One of the strings has been exhausted before the other one.
			if pHasNext != sHasNext {
				// Allow for one or more trailing wildcards, fail otherwise.
				if pNext == '*' {
					pv = pp.read()
					continue
				}
				return false
			}
			pv, sv = pp.read(), sp.read()
		case '*':
			switch {
			case !pHasNext: // * at the end matches anything
				return true
			case pNext == '*': // consume redundant wildcard
				pv = pp.read()
				continue
			case !sHasNext: // s is shorter than pattern
				return false
			}

			// consume wildcard if pNext matches sv / sNext.
			if pNext == sNext || pNext == sv {
				pv = pp.read()
			}

			// consume current character if the wildcard has nothing to match.
			// E.g.: pattern "g*lang" and s "golang".
			if pNext == sNext || pNext != sv {
				sv = sp.read()
			}
		}
	}
}

// newParser creates a parser for s.
func newParser(s string) *parser {
	return &parser{r: []rune(s), len: len(s)}
}

// parser consumes the string 1 rune at a time and allows peeking 1 step ahead.
type parser struct {
	i   int
	r   []rune
	len int
}

// peek returns the next character and the check for its existence.
func (p *parser) peek() (rune, bool) {
	if p.i >= p.len {
		return 0, false
	}
	return p.r[p.i], true
}

// read returns the character at the current position and advances the index.
func (p *parser) read() rune {
	r := p.r[p.i]
	p.i++
	return r
}

// Count counts the number of instances of wildcard (*) in s.
func Count(s string) int {
	return strings.Count(s, "*")
}
