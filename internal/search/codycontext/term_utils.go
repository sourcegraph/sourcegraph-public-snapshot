package codycontext

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
)

func removePunctuation(input string) string {
	return strings.TrimFunc(input, unicode.IsPunct)
}

var separators = map[rune]struct{}{
	':': {}, // C++, Rust
	'.': {}, // Java, Go, Python
}

// tokenize splits on runes that usually separate namespaces (class, package,
// ...) from methods or functions.
func tokenize(input string) []string {
	return strings.FieldsFunc(input, func(r rune) bool {
		_, ok := separators[r]
		return ok
	})
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
