package russian

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 3 is the removal of the derivational suffix.
//
func step3(word *snowballword.SnowballWord) bool {

	// Search for a DERIVATIONAL ending in R2 (i.e. the entire
	// ending must lie in R2), and if one is found, remove it.

	suffix, _ := word.RemoveFirstSuffixIn(word.R2start, "ост", "ость")
	if suffix != "" {
		return true
	}
	return false
}
