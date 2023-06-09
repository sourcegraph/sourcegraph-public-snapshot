package keyword

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
)

func removePunctuation(input string) string {
	return strings.TrimFunc(input, unicode.IsPunct)
}

func stemTerm(input string) string {
	// Attempt to stem words, but only use the stem if it's a prefix of the original term. This
	// avoids cases where the stem is noisy and no longer matches the original term.
	stemmed, err := snowball.Stem(input, "english", false)
	if err != nil || !strings.HasPrefix(input, stemmed) {
		return input
	}

	return stemmed
}

func isCommonTerm(input string) bool {
	return commonCodeSearchTerms.Has(input) || stopWords.Has(input)
}

var commonCodeSearchTerms = stringSet{
	"class":       {},
	"classes":     {},
	"code":        {},
	"codebase":    {},
	"cody":        {},
	"define":      {},
	"defines":     {},
	"defined":     {},
	"design":      {},
	"find":        {},
	"finding":     {},
	"function":    {},
	"functions":   {},
	"implement":   {},
	"implements":  {},
	"implemented": {},
	"method":      {},
	"methods":     {},
	"package":     {},
	"product":     {},
}
