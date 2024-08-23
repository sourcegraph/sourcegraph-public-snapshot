// Package wildmatch used to match strings against a simple wildcard pattern.
// Tests a wildcard pattern `p` against an input string `s`. Returns true only when `p` matches the entirety of `s`.
//
// See also the example described on [wikipedia](https://en.wikipedia.org/wiki/Matching_wildcards) for matching wildcards.
//
// No escape characters are defined.
//
// - `?` matches exactly one occurrence of any character.
// - `*` matches arbitrary many (including zero) occurrences of any character.
//
// Examples matching wildcards:
// ``` go
// import "github.com/becheran/wildmatch-go"
// wildmatch.NewWildMatch("cat").IsMatch("cat")
// wildmatch.NewWildMatch("*cat*").IsMatch("dog_cat_dog")
// wildmatch.NewWildMatch("c?t").IsMatch("cat")
// wildmatch.NewWildMatch("c?t").IsMatch("cot")
// ```
// Examples not matching wildcards:
// ``` go
// import "github.com/becheran/wildmatch-go"
// wildmatch.NewWildMatch("dog").IsMatch("cat")
// wildmatch.NewWildMatch("*d").IsMatch("cat")
// wildmatch.NewWildMatch("????").IsMatch("cat")
// wildmatch.NewWildMatch("?").IsMatch("cat")
// ```
package wildmatch

import "strings"

/// WildMatch is a wildcard matcher used to match strings.
type WildMatch struct {
	pattern []state
}

type state struct {
	NextChar    *rune
	HasWildcard bool
}

func (w *WildMatch) String() string {
	var sb strings.Builder
	for _, p := range w.pattern {
		if p.NextChar == nil {
			break
		}
		sb.WriteString(string(*p.NextChar))
	}
	return sb.String()
}

// NewWildMatch creates new pattern matcher.
func NewWildMatch(pattern string) *WildMatch {
	simplified := make([]state, 0)
	prevWasStar := false
	for _, currentChar := range pattern {
		copyCurrentChar := currentChar
		if currentChar == '*' {
			prevWasStar = true
		} else {
			s := state{
				NextChar:    &copyCurrentChar,
				HasWildcard: prevWasStar,
			}
			simplified = append(simplified, s)
			prevWasStar = false
		}
	}

	if len(pattern) > 0 {
		final := state{
			NextChar:    nil,
			HasWildcard: prevWasStar,
		}
		simplified = append(simplified, final)
	}

	return &WildMatch{
		pattern: simplified,
	}
}

// IsMatch indicates whether the matcher finds a match in the input string.
func (w *WildMatch) IsMatch(input string) bool {
	if len(w.pattern) == 0 {
		return false
	}

	patternIdx := 0
	for _, inputChar := range input {
		if patternIdx > len(w.pattern) {
			return false
		}

		p := w.pattern[patternIdx]

		if p.NextChar != nil && (*p.NextChar == '?' || *p.NextChar == inputChar) {
			patternIdx += 1
		} else if p.HasWildcard {
			if p.NextChar == nil {
				return true
			}
		} else {
			// Go back to last state with wildcard
			for {
				pattern := w.pattern[patternIdx]
				if pattern.HasWildcard {
					if pattern.NextChar != nil && (*pattern.NextChar == '?' || *pattern.NextChar == inputChar) {
						patternIdx += 1
					}
					break
				}
				if patternIdx == 0 {
					return false
				}
				patternIdx -= 1
			}
		}
	}
	return w.pattern[patternIdx].NextChar == nil
}
