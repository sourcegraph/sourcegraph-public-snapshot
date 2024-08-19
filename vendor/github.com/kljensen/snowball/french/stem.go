package french

import (
	"github.com/kljensen/snowball/snowballword"
	"strings"
)

// Stem an French word.  This is the only exported
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
	var (
		changeInStep1  bool
		changeInStep2a bool
		changeInStep2b bool
	)

	changeInStep1 = step1(w)
	if changeInStep1 == false {
		changeInStep2a = step2a(w)
		if changeInStep2a == false {
			changeInStep2b = step2b(w)
		}
	}

	// If the last step was successful, do step 3.  Note that,
	// since we only do 2a if 1 is unsuccessful, the following
	// "if" condition tests to see if the previous step was
	// successful.
	//
	if changeInStep1 || changeInStep2a || changeInStep2b {
		step3(w)
	} else {
		step4(w)
	}

	step5(w)
	step6(w)
	postprocess(w)
	return w.String()

}
