package french

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 6 Un-accent
//
func step6(word *snowballword.SnowballWord) bool {

	// If the words ends é or è (unicode code points 233 and 232)
	// followed by at least one non-vowel, remove the accent from the e.

	// Note, this step is oddly articulated on Porter's Snowball website:
	// http://snowball.tartarus.org/algorithms/french/stemmer.html
	// More clearly stated, we should replace é or è with e in the
	// case where the suffix of the word is é or è followed by
	// one-or-more non-vowels.

	numNonVowels := 0
	for i := len(word.RS) - 1; i >= 0; i-- {
		r := word.RS[i]

		if isLowerVowel(r) == false {
			numNonVowels += 1
		} else {

			// `r` is a vowel

			if (r == 233 || r == 232) && numNonVowels > 0 {

				// Replace with "e", or unicode code point 101
				word.RS[i] = 101
				return true

			}
			return false
		}

	}
	return false
}
