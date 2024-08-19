package norwegian

import (
	"github.com/kljensen/snowball/snowballword"
	"strings"
)

// Stem a Norwegian word. This is the only exported
// function in this package.
//
func Stem(word string, stemStopwWords bool) string {

	word = strings.ToLower(strings.TrimSpace(word))

	// Return small words and stop words
	if len(word) <= 2 || (stemStopwWords == false && isStopWord(word)) {
		return word
	}

	w := snowballword.New(word)

	// Stem the word.  Note, each of these
	// steps will alter `w` in place.
	//
	preprocess(w)
	step1(w)
	step2(w)
	step3(w)

	return w.String()

}
