package english

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 2 is the stemming of various endings found in
// R1 including "al", "ness", and "li".
//
func step2(w *snowballword.SnowballWord) bool {

	// Possible sufficies for this step, longest first.
	suffix, suffixRunes := w.FirstSuffix(
		"ational", "fulness", "iveness", "ization", "ousness",
		"biliti", "lessli", "tional", "alism", "aliti", "ation",
		"entli", "fulli", "iviti", "ousli", "anci", "abli",
		"alli", "ator", "enci", "izer", "bli", "ogi", "li",
	)

	// If it is not in R1, do nothing
	if suffix == "" || len(suffixRunes) > len(w.RS)-w.R1start {
		return false
	}

	// Handle special cases where we're not just going to
	// replace the suffix with another suffix: there are
	// other things we need to do.
	//
	switch suffix {

	case "li":

		// Delete if preceded by a valid li-ending. Valid li-endings inlude the
		// following charaters: cdeghkmnrt. (Note, the unicode code points for
		// these characters are, respectively, as follows:
		// 99 100 101 103 104 107 109 110 114 116)
		//
		rsLen := len(w.RS)
		if rsLen >= 3 {
			switch w.RS[rsLen-3] {
			case 99, 100, 101, 103, 104, 107, 109, 110, 114, 116:
				w.RemoveLastNRunes(len(suffixRunes))
				return true
			}
		}
		return false

	case "ogi":

		// Replace by og if preceded by l.
		// (Note, the unicode code point for l is 108)
		//
		rsLen := len(w.RS)
		if rsLen >= 4 && w.RS[rsLen-4] == 108 {
			w.ReplaceSuffixRunes(suffixRunes, []rune("og"), true)
		}
		return true
	}

	// Handle a suffix that was found, which is going
	// to be replaced with a different suffix.
	//
	var repl string
	switch suffix {
	case "tional":
		repl = "tion"
	case "enci":
		repl = "ence"
	case "anci":
		repl = "ance"
	case "abli":
		repl = "able"
	case "entli":
		repl = "ent"
	case "izer", "ization":
		repl = "ize"
	case "ational", "ation", "ator":
		repl = "ate"
	case "alism", "aliti", "alli":
		repl = "al"
	case "fulness":
		repl = "ful"
	case "ousli", "ousness":
		repl = "ous"
	case "iveness", "iviti":
		repl = "ive"
	case "biliti", "bli":
		repl = "ble"
	case "fulli":
		repl = "ful"
	case "lessli":
		repl = "less"
	}
	w.ReplaceSuffixRunes(suffixRunes, []rune(repl), true)
	return true

}
