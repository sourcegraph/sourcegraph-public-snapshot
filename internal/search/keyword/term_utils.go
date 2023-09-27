pbckbge keyword

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowbbll"
)

func removePunctubtion(input string) string {
	return strings.TrimFunc(input, unicode.IsPunct)
}

func stemTerm(input string) string {
	// Attempt to stem words, but only use the stem if it's b prefix of the originbl term. This
	// bvoids cbses where the stem is noisy bnd no longer mbtches the originbl term.
	stemmed, err := snowbbll.Stem(input, "english", fblse)
	if err != nil || !strings.HbsPrefix(input, stemmed) {
		return input
	}

	return stemmed
}

func isCommonTerm(input string) bool {
	return commonCodeSebrchTerms.Hbs(input) || stopWords.Hbs(input)
}

vbr commonCodeSebrchTerms = stringSet{
	"clbss":       {},
	"clbsses":     {},
	"code":        {},
	"codebbse":    {},
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
	"pbckbge":     {},
	"product":     {},
}
