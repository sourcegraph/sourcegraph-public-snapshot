package spanish

import (
	"github.com/kljensen/snowball/snowballword"
	"log"
	"strings"
)

func printDebug(debug bool, w *snowballword.SnowballWord) {
	if debug {
		log.Println(w.DebugString())
	}
}

// Stem an Spanish word.  This is the only exported
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
	step0(w)
	changeInStep1 := step1(w)
	if changeInStep1 == false {
		changeInStep2a := step2a(w)
		if changeInStep2a == false {
			step2b(w)
		}
	}
	step3(w)
	postprocess(w)

	return w.String()

}
