package query

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var fieldRx = lazyregexp.New(`^-?[a-zA-Z]+:`)

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

var fieldWithQuotedTokenValue = lazyregexp.New(`(\b-?[a-zA-Z]+:("([^"\\]|[\\].)*"|'([^'\\]|[\\].)*'))`)
var tokenRx = lazyregexp.New(`("([^"\\]|[\\].)*"|\s+|\S+)`)

// tokenize returns a slice of the double-quoted strings, contiguous chunks
// of non-whitespace, and contiguous chunks of whitespace in the input.
func tokenize(input string) []string {
	// Find all tokens with quoted values, and then remove them from the original input
	matchedTokens := fieldWithQuotedTokenValue.FindAllString(input, -1)
	modifiedInput := fieldWithQuotedTokenValue.ReplaceAllString(input, "")

	// Find all remaining tokens
	return append(matchedTokens, tokenRx.FindAllString(modifiedInput, -1)...)
}
