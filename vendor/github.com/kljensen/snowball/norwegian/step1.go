package norwegian

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 1 is the stemming of various endings found in
// R1 including "hetene", "endes", and "ande".
//
func step1(w *snowballword.SnowballWord) bool {

	// Possible sufficies for this step, longest first.
	suffixes := []string{
		"hetenes", "hetene", "hetens", "endes", "heter", "heten", "ende",
		"ande", "edes", "enes", "ene", "ane", "ets", "ers", "ede", "ast",
		"ens", "het", "as", "es", "en", "ar", "er", "et", "e", "a", "s",
	}

	// Using FirstSuffixIn since there are overlapping suffixes, where some might not be in the R1,
	suffix, suffixRunes := w.FirstSuffixIn(w.R1start, len(w.RS), suffixes...)

	if suffix == "s" {
		// Delete if preceded by a valid s-ending. Valid s-endings inlude the
		// following charaters: bcdfghjlmnoprtvyz or k not preceded by a vowel
		rsLen := len(w.RS)

		if rsLen >= 2 {
			switch w.RS[rsLen-2] {
			case 'b', 'c', 'd', 'f', 'g', 'h', 'j',
				'l', 'm', 'n', 'o', 'p', 'r', 't', 'v', 'y', 'z':

				w.RemoveLastNRunes(len(suffixRunes))
				return true
			case 'k':
				if !isLowerVowel(w.RS[rsLen-3]) {
					w.RemoveLastNRunes(len(suffixRunes))
					return true
				}
			}
		}

		return false
	}

	// Remove the suffix
	w.RemoveLastNRunes(len(suffixRunes))

	// replace "erte" and "ert" with "er"
	suffix, suffixRunes = w.FirstSuffix(
		"erte", "ert",
	)

	if suffix == "" || len(suffixRunes) > len(w.RS)-w.R1start {
		return false
	}

	w.ReplaceSuffixRunes(suffixRunes, []rune("er"), true)

	return true
}
