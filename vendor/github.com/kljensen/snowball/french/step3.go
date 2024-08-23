package french

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 3 is the cleaning up of "Y" and "ç" suffixes.
//
func step3(word *snowballword.SnowballWord) bool {

	// Replace final Y with i or final ç with c
	if idx := len(word.RS) - 1; idx >= 0 {

		switch word.RS[idx] {

		case 89:
			// Replace Y (89) with "i" (105)
			word.RS[idx] = 105
			return true

		case 231:
			// Replace ç (231) with "c" (99)
			word.RS[idx] = 99
			return true
		}
	}
	return false
}
