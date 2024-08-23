package norwegian

import (
	"github.com/kljensen/snowball/snowballword"
)

// Get the r1 of the word
//
func preprocess(word *snowballword.SnowballWord) {
	// Find the region R1. R2 is not used
	word.R1start = r1(word)
}
