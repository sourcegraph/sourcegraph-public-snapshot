package swedish

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 3:
// Search for the longest among the following suffixes,
// and, if found and in R1, perform the action indicated.

// Delete:
// lig, els & ig
// Replace:
// fullt: full, löst: lös

func step3(w *snowballword.SnowballWord) bool {
	// Possible sufficies for this step, longest first.
	suffix, suffixRunes := w.FirstSuffixIn(w.R1start, len(w.RS),
		"fullt", "löst", "lig", "els", "ig",
	)

	// If it is not in R1, do nothing
	if suffix == "" || len(suffixRunes) > len(w.RS)-w.R1start {
		return false
	}

	// Handle a suffix that was found, which is going
	// to be replaced with a different suffix.
	//
	var repl string
	switch suffix {
	case "fullt":
		repl = "full"
	case "löst":
		repl = "lös"
	case "lig", "ig", "els":
		repl = ""
	}
	w.ReplaceSuffixRunes(suffixRunes, []rune(repl), true)
	return true

}
