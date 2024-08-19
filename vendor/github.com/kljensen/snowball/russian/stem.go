package russian

import (
	"github.com/kljensen/snowball/snowballword"
	"strings"
)

// Stem an Russian word.  This is the only exported
// function in this package.
//
func Stem(word string, stemStopwWords bool) string {

	word = strings.ToLower(strings.TrimSpace(word))
	w := snowballword.New(word)

	// Return small words and stop words
	if len(w.RS) <= 2 || (stemStopwWords == false && isStopWord(word)) {
		return word
	}

	preprocess(w)
	step1(w)
	step2(w)
	step3(w)
	step4(w)
	return w.String()

}
