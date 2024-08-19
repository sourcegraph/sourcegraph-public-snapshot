package russian

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 2 is the removal of the "и" suffix.
//
func step2(word *snowballword.SnowballWord) bool {
	suffix, _ := word.RemoveFirstSuffixIn(word.RVstart, "и")
	if suffix != "" {
		return true
	}
	return false
}
