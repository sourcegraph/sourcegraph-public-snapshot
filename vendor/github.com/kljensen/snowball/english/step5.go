package english

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 5 is the stemming of "e" and "l" sufficies
// found in R2.
//
func step5(w *snowballword.SnowballWord) bool {

	// Last rune index = `lri`
	lri := len(w.RS) - 1

	// If R1 is emtpy, R2 is also empty, and we
	// need not do anything in step 5.
	//
	if w.R1start > lri {
		return false
	}

	if w.RS[lri] == 101 {

		// The word ends with "e", which is unicode code point 101.

		// Delete "e" suffix if in R2, or in R1 and not preceded
		// by a short syllable.
		if w.R2start <= lri || !endsShortSyllable(w, lri) {
			w.ReplaceSuffix("e", "", true)
			return true
		}
		return false

	} else if w.R2start <= lri && w.RS[lri] == 108 && lri-1 >= 0 && w.RS[lri-1] == 108 {

		// The word ends in double "l", and the final "l" is
		// in R2. (Note, the unicode code point for "l" is 108.)

		// Delete the second "l".
		w.ReplaceSuffix("l", "", true)
		return true

	}
	return false
}
