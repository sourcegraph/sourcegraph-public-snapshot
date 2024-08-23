package english

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 4:
// Search for the longest among the following suffixes,
// and, if found and in R2, perform the action indicated.

// al, ance, ence, er, ic, able, ible, ant, ement, ment,
// ent, ism, ate, iti, ous, ive, ize
// delete
//
// ion
// delete if preceded by s or t
func step4(w *snowballword.SnowballWord) bool {

	// Find all endings in R1
	suffix, suffixRunes := w.FirstSuffix(
		"ement", "ance", "ence", "able", "ible", "ment",
		"ent", "ant", "ism", "ate", "iti", "ous", "ive",
		"ize", "ion", "al", "er", "ic",
	)

	// If it does not fit in R2, do nothing.
	if len(suffixRunes) > len(w.RS)-w.R2start {
		return false
	}

	// Handle special cases
	switch suffix {
	case "":
		return false

	case "ion":
		// Replace by og if preceded by l
		// l = 108
		rsLen := len(w.RS)
		if rsLen >= 4 {
			switch w.RS[rsLen-4] {
			case 115, 116:
				w.RemoveLastNRunes(len(suffixRunes))
				return true
			}

		}
		return false
	}

	// Handle basic replacements
	w.RemoveLastNRunes(len(suffixRunes))
	return true

}
