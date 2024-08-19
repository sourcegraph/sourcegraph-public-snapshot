package norwegian

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 3:
// Search for the longest among the following suffixes,
// and, if found and in R1, delete.

func step3(w *snowballword.SnowballWord) bool {
	// Possible sufficies for this step, longest first.
	suffix, suffixRunes := w.FirstSuffixIn(w.R1start, len(w.RS),
		"hetslov", "eleg", "elig", "elov", "slov",
		"leg", "eig", "lig", "els", "lov", "ig",
	)

	// If it is not in R1, do nothing
	if suffix == "" || len(suffixRunes) > len(w.RS)-w.R1start {
		return false
	}

	w.ReplaceSuffixRunes(suffixRunes, []rune(""), true)
	return true

}
