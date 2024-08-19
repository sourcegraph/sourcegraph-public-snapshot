package english

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 1c is the normalization of various "y" endings.
//
func step1c(w *snowballword.SnowballWord) bool {

	rsLen := len(w.RS)

	// Replace suffix y or Y by i if preceded by a non-vowel which is not
	// the first letter of the word (so cry -> cri, by -> by, say -> say)
	//
	// Note: the unicode code points for
	// y, Y, & i are 121, 89, & 105 respectively.
	//
	if len(w.RS) > 2 && (w.RS[rsLen-1] == 121 || w.RS[rsLen-1] == 89) && !isLowerVowel(w.RS[rsLen-2]) {
		w.RS[rsLen-1] = 105
		return true
	}
	return false
}
