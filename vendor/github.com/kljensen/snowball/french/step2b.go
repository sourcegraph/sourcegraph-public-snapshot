package french

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 2b is the removal of Verb suffixes in RV
// that do not begin with "i".
//
func step2b(word *snowballword.SnowballWord) bool {

	// Search for the longest among the following suffixes in RV.
	//
	suffix, suffixRunes := word.FirstSuffixIn(word.RVstart, len(word.RS),
		"eraIent", "assions", "erions", "assiez", "assent",
		"èrent", "eront", "erons", "eriez", "erait", "erais",
		"asses", "antes", "aIent", "âtes", "âmes", "ions",
		"erez", "eras", "erai", "asse", "ants", "ante", "ées",
		"iez", "era", "ant", "ait", "ais", "és", "ée", "ât",
		"ez", "er", "as", "ai", "é", "a",
	)

	switch suffix {
	case "ions":

		// Delete if in R2
		suffixLen := len(suffixRunes)
		if word.FitsInR2(suffixLen) {
			word.RemoveLastNRunes(suffixLen)
			return true
		}
		return false

	case "é", "ée", "ées", "és", "èrent", "er", "era",
		"erai", "eraIent", "erais", "erait", "eras", "erez",
		"eriez", "erions", "erons", "eront", "ez", "iez":

		// Delete
		word.RemoveLastNRunes(len(suffixRunes))
		return true

	case "âmes", "ât", "âtes", "a", "ai", "aIent",
		"ais", "ait", "ant", "ante", "antes", "ants", "as",
		"asse", "assent", "asses", "assiez", "assions":

		// Delete
		word.RemoveLastNRunes(len(suffixRunes))

		// If preceded by e (unicode code point 101), delete
		//
		idx := len(word.RS) - 1
		if idx >= 0 && word.RS[idx] == 101 && word.FitsInRV(1) {
			word.RemoveLastNRunes(1)
		}
		return true

	}
	return false
}
