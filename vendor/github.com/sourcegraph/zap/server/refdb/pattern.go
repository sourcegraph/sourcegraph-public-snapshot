package refdb

// matchPattern matches the pattern described in (RefDB).List.
func matchPattern(pattern string, strings []string) (matches []string) {
	for _, s := range strings {
		if MatchPattern(pattern, s) {
			matches = append(matches, s)
		}
	}
	return matches
}

// MatchPattern reports whether name matches pattern. The "*"
// character is a wildcard that matches any number of characters.
func MatchPattern(pattern, name string) bool {
	return matchRunePattern1([]rune(pattern), []rune(name))
}

func matchRunePattern1(pattern, source []rune) (ok bool) {
	if len(pattern) == 0 {
		return len(source) == 0
	}
	if len(source) == 0 {
		for _, c := range pattern {
			if c != '*' {
				return false
			}
		}
		return true
	}
	if pattern[0] == '*' {
		return matchRunePattern1(pattern[1:], source[1:]) || matchRunePattern1(pattern, source[1:]) || matchRunePattern1(pattern[1:], source)
	}
	if pattern[0] == source[0] {
		return matchRunePattern1(pattern[1:], source[1:])
	}
	return false
}
