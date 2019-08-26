package query

import (
	"fmt"
	"regexp"
	"strings"
)

var fieldRx = regexp.MustCompile(`^-?[a-zA-Z]+:`)

// HandlePatternType returns a modified version of the input query where it has
// been either quoted because it has patternType:literal, not quoted because it
// has patternType:regex, or been possibly quoted in the default case according
// to the defaultToRegexp parameter.
func HandlePatternType(input string, defaultToRegexp bool) string {
	tokens := tokenize(input)
	isRegex := defaultToRegexp
	var tokens2 []string
	for _, t := range tokens {
		switch t {
		case "patternType:regex":
			isRegex = true
		case "patternType:regexp":
			isRegex = true
		case "patternType:literal":
			isRegex = false
		default:
			tokens2 = append(tokens2, t)
		}
	}
	if isRegex {
		// Rebuild the input from the remaining tokens.
		input = strings.TrimSpace(strings.Join(tokens2, ""))
	} else {
		// Sort the tokens into fields and non-fields.
		var fields, nonFields []string
		for _, t := range tokens2 {
			if fieldRx.MatchString(t) {
				fields = append(fields, t)
			} else {
				nonFields = append(nonFields, t)
			}
		}

		// Rebuild the input as fields followed by non-fields quoted together.
		var pieces []string
		if len(fields) > 0 {
			pieces = append(pieces, strings.Join(fields, " "))
		}
		if len(nonFields) > 0 {
			// Count up the number of non-whitespace tokens in the nonFields slice.
			q := strings.Join(nonFields, "")
			q = strings.TrimSpace(q)
			q = strings.ReplaceAll(q, `\`, `\\`)
			q = strings.ReplaceAll(q, `"`, `\"`)
			q = fmt.Sprintf(`"%s"`, q)
			if q != `""` {
				pieces = append(pieces, q)
			}
		}
		input = strings.Join(pieces, " ")
	}
	return input
}

var tokenRx = regexp.MustCompile(`("([^"\\]|[\\].)*"|\s+|\S+)`)

// tokenize returns a slice of the double-quoted strings, contiguous chunks
// of non-whitespace, and contiguous chunks of whitespace in the input.
func tokenize(input string) []string {
	return tokenRx.FindAllString(input, -1)
}
