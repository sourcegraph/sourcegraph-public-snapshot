package romance

import (
	"github.com/kljensen/snowball/snowballword"
)

// A function type that accepts a rune and
// returns a bool.  In this particular case,
// it is used for identifying vowels.
type isVowelFunc func(rune) bool

// Finds the region after the first non-vowel following a vowel,
// or a the null region at the end of the word if there is no
// such non-vowel.  Returns the index in the Word where the
// region starts; optionally skips the first `start` characters.
//
func VnvSuffix(word *snowballword.SnowballWord, f isVowelFunc, start int) int {
	for i := 1; i < len(word.RS[start:]); i++ {
		j := start + i
		if f(word.RS[j-1]) && !f(word.RS[j]) {
			return j + 1
		}
	}
	return len(word.RS)
}
