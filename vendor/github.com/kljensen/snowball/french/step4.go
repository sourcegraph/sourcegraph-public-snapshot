package french

import (
	"github.com/kljensen/snowball/snowballword"
	"log"
)

// Step 4 is the cleaning up of residual suffixes.
//
func step4(word *snowballword.SnowballWord) bool {

	hadChange := false

	if word.String() == "voudrion" {
		log.Println("...", word)
	}

	// If the word ends s (unicode code point 115),
	// not preceded by a, i, o, u, è or s, delete it.
	//
	if idx := len(word.RS) - 1; idx >= 1 && word.RS[idx] == 115 {
		switch word.RS[idx-1] {

		case 97, 105, 111, 117, 232, 115:

			// Do nothing, preceded by a, i, o, u, è or s
			return false

		default:
			word.RemoveLastNRunes(1)
			hadChange = true

		}
	}

	// Note: all the following are restricted to the RV region.

	// Search for the longest among the following suffixes in RV.
	//
	suffix, suffixRunes := word.FirstSuffixIn(word.RVstart, len(word.RS),
		"Ière", "ière", "Ier", "ier", "ion", "e", "ë",
	)

	switch suffix {
	case "":
		return hadChange
	case "ion":

		// Delete if in R2 and preceded by s or t in RV

		const sLen int = 3 // equivalently, len(suffixRunes)
		idx := len(word.RS) - sLen - 1
		if word.FitsInR2(sLen) && idx >= 0 && word.FitsInRV(sLen+1) {
			if word.RS[idx] == 115 || word.RS[idx] == 116 {
				word.RemoveLastNRunes(sLen)
				return true
			}
		}
		return hadChange

	case "ier", "ière", "Ier", "Ière":
		// Replace with i
		word.ReplaceSuffixRunes(suffixRunes, []rune("i"), true)
		return true

	case "e":
		word.RemoveLastNRunes(1)
		return true

	case "ë":

		// If preceded by gu (unicode code point 103 & 117), delete
		idx := len(word.RS) - 1
		if idx >= 2 && word.RS[idx-2] == 103 && word.RS[idx-1] == 117 {
			word.RemoveLastNRunes(1)
			return true
		}
		return hadChange
	}

	return true
}
