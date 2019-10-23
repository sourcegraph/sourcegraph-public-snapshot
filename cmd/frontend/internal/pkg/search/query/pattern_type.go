package query

import (
	"fmt"
	"regexp"
	"strings"
)

var fieldRx = regexp.MustCompile(`^-?[a-zA-Z]+:`)

// ConvertToLiteral quotes the input query for literal search.
func ConvertToLiteral(input string) string {
	tokens := tokenize(input)
	// Sort the tokens into fields and non-fields.
	var fields, nonFields []string
	for _, t := range tokens {
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
	return input
}

var tokenRx = regexp.MustCompile(`("([^"\\]|[\\].)*"|\s+|\S+)`)

// tokenize returns a slice of the double-quoted strings, contiguous chunks
// of non-whitespace, and contiguous chunks of whitespace in the input.
func tokenize(input string) []string {
	return tokenRx.FindAllString(input, -1)
}
