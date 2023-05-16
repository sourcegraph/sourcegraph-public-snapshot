package keyword

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
)

func removePunctuation(input string) string {
	lowerCase := strings.ToLower(input)
	runes := []rune(lowerCase)

	// Remove leading and trailing punctuation
	start := 0
	for ; start < len(runes) && unicode.IsPunct(runes[start]); start++ {
	}
	end := len(runes)
	for ; end > start && unicode.IsPunct(runes[end-1]); end-- {
	}

	return string(runes[start:end])
}

func stemTerm(input string) string {
	// Attempt to stem words, but only use the stem if it's a prefix of the original term.
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
