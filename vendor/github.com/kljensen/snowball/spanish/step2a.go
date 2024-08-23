package spanish

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 2a is the removal of verb suffixes beginning y,
// Search for the longest among the following suffixes
// in RV, and if found, delete if preceded by u.
//
func step2a(word *snowballword.SnowballWord) bool {
	suffix, suffixRunes := word.FirstSuffixIn(word.RVstart, len(word.RS), "ya", "ye", "yan", "yen", "yeron", "yendo", "yo", "yó", "yas", "yes", "yais", "yamos")
	if suffix != "" {
		idx := len(word.RS) - len(suffixRunes) - 1
		if idx >= 0 && word.RS[idx] == 117 {
			word.RemoveLastNRunes(len(suffixRunes))
			return true
		}
	}
	return false
}
