package french

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 1 is the removal of standard suffixes
//
func step1(word *snowballword.SnowballWord) bool {
	suffix, suffixRunes := word.FirstSuffix(
		"issements", "issement", "atrices", "utions", "usions", "logies",
		"emment", "ements", "atrice", "ations", "ateurs", "amment", "ution",
		"usion", "ments", "logie", "istes", "ismes", "iqUes", "euses",
		"ences", "ement", "ation", "ateur", "ances", "ables", "ment",
		"ités", "iste", "isme", "iqUe", "euse", "ence", "eaux", "ance",
		"able", "ives", "ité", "eux", "aux", "ive", "ifs", "if",
	)

	if suffix == "" {
		return false
	}

	isInR1 := (word.R1start <= len(word.RS)-len(suffixRunes))
	isInR2 := (word.R2start <= len(word.RS)-len(suffixRunes))
	isInRV := (word.RVstart <= len(word.RS)-len(suffixRunes))

	// Handle simple replacements & deletions in R2 first
	if isInR2 {

		// Handle simple replacements in R2
		repl := ""
		switch suffix {
		case "logie", "logies":
			repl = "log"
		case "usion", "ution", "usions", "utions":
			repl = "u"
		case "ence", "ences":
			repl = "ent"
		}
		if repl != "" {
			word.ReplaceSuffixRunes(suffixRunes, []rune(repl), true)
			return true
		}

		// Handle simple deletions in R2
		switch suffix {
		case "ance", "iqUe", "isme", "able", "iste", "eux", "ances", "iqUes", "ismes", "ables", "istes":
			word.RemoveLastNRunes(len(suffixRunes))
			return true
		}
	}

	// Handle simple replacements in RV
	if isInRV {

		// NOTE: these are "special" suffixes in that
		// we must still do steps 2a and 2b of the
		// French stemmer even when these suffixes are
		// found in step1.  Therefore, we are returning
		// `false` here.

		repl := ""
		switch suffix {
		case "amment":
			repl = "ant"
		case "emment":
			repl = "ent"
		}
		if repl != "" {
			word.ReplaceSuffixRunes(suffixRunes, []rune(repl), true)
			return false
		}

		// Delete if preceded by a vowel that is also in RV
		if suffix == "ment" || suffix == "ments" {
			idx := len(word.RS) - len(suffixRunes) - 1
			if idx >= word.RVstart && isLowerVowel(word.RS[idx]) {
				word.RemoveLastNRunes(len(suffixRunes))
				return false
			}
			return false
		}
	}

	// Handle all the other "special" cases.  All of these
	// return true immediately after changing the word.
	//
	switch suffix {
	case "eaux":

		// Replace with eau
		word.ReplaceSuffixRunes(suffixRunes, []rune("eau"), true)
		return true

	case "aux":

		// Replace with al if in R1
		if isInR1 {
			word.ReplaceSuffixRunes(suffixRunes, []rune("al"), true)
			return true
		}

	case "euse", "euses":

		// Delete if in R2, else replace by eux if in R1
		if isInR2 {
			word.RemoveLastNRunes(len(suffixRunes))
			return true
		} else if isInR1 {
			word.ReplaceSuffixRunes(suffixRunes, []rune("eux"), true)
			return true
		}

	case "issement", "issements":

		// Delete if in R1 and preceded by a non-vowel
		if isInR1 {
			idx := len(word.RS) - len(suffixRunes) - 1
			if idx >= 0 && isLowerVowel(word.RS[idx]) == false {
				word.RemoveLastNRunes(len(suffixRunes))
				return true
			}
		}
		return false

	case "atrice", "ateur", "ation", "atrices", "ateurs", "ations":

		// Delete if in R2
		if isInR2 {
			word.RemoveLastNRunes(len(suffixRunes))

			// If preceded by "ic", delete if in R2, else replace by "iqU".
			newSuffix, newSuffixRunes := word.FirstSuffix("ic")
			if newSuffix != "" {
				if word.FitsInR2(len(newSuffixRunes)) {
					word.RemoveLastNRunes(len(newSuffixRunes))
				} else {
					word.ReplaceSuffixRunes(newSuffixRunes, []rune("iqU"), true)
				}
			}
			return true
		}

	case "ement", "ements":

		if isInRV {

			// Delete if in RV
			word.RemoveLastNRunes(len(suffixRunes))

			// If preceded by "iv", delete if in R2
			// (and if further preceded by "at", delete if in R2)
			newSuffix, newSuffixRunes := word.RemoveFirstSuffixIfIn(word.R2start, "iv")
			if newSuffix != "" {
				word.RemoveFirstSuffixIfIn(word.R2start, "at")
				return true
			}

			// If preceded by "eus", delete if in R2, else replace by "eux" if in R1
			newSuffix, newSuffixRunes = word.FirstSuffix("eus")
			if newSuffix != "" {
				newSuffixLen := len(newSuffixRunes)
				if word.FitsInR2(newSuffixLen) {
					word.RemoveLastNRunes(newSuffixLen)
				} else if word.FitsInR1(newSuffixLen) {
					word.ReplaceSuffixRunes(newSuffixRunes, []rune("eux"), true)
				}
				return true
			}

			// If preceded by abl or iqU, delete if in R2, otherwise,
			newSuffix, newSuffixRunes = word.FirstSuffix("abl", "iqU")
			if newSuffix != "" {
				newSuffixLen := len(newSuffixRunes)
				if word.FitsInR2(newSuffixLen) {
					word.RemoveLastNRunes(newSuffixLen)
				}
				return true
			}

			// If preceded by ièr or Ièr, replace by i if in RV
			newSuffix, newSuffixRunes = word.FirstSuffix("ièr", "Ièr")
			if newSuffix != "" {
				if word.FitsInRV(len(newSuffixRunes)) {
					word.ReplaceSuffixRunes(newSuffixRunes, []rune("i"), true)
				}
				return true
			}

			return true
		}

	case "ité", "ités":

		if isInR2 {

			// Delete if in R2
			word.RemoveLastNRunes(len(suffixRunes))

			// If preceded by "abil", delete if in R2, else replace by "abl"
			newSuffix, newSuffixRunes := word.FirstSuffix("abil")
			if newSuffix != "" {
				newSuffixLen := len(newSuffixRunes)
				if word.FitsInR2(newSuffixLen) {
					word.RemoveLastNRunes(newSuffixLen)
				} else {
					word.ReplaceSuffixRunes(newSuffixRunes, []rune("abl"), true)
				}
				return true
			}

			// If preceded by "ic", delete if in R2, else replace by "iqU"
			newSuffix, newSuffixRunes = word.FirstSuffix("ic")
			if newSuffix != "" {
				newSuffixLen := len(newSuffixRunes)
				if word.FitsInR2(newSuffixLen) {
					word.RemoveLastNRunes(newSuffixLen)
				} else {
					word.ReplaceSuffixRunes(newSuffixRunes, []rune("iqU"), true)
				}
				return true
			}

			// If preceded by "iv", delete if in R2
			newSuffix, newSuffixRunes = word.RemoveFirstSuffixIfIn(word.R2start, "iv")
			return true
		}
	case "if", "ive", "ifs", "ives":

		if isInR2 {

			// Delete if in R2
			word.RemoveLastNRunes(len(suffixRunes))

			// If preceded by at, delete if in R2
			newSuffix, newSuffixRunes := word.RemoveFirstSuffixIfIn(word.R2start, "at")
			if newSuffix != "" {

				// And if further preceded by ic, delete if in R2, else replace by iqU
				newSuffix, newSuffixRunes = word.FirstSuffix("ic")
				if newSuffix != "" {
					newSuffixLen := len(newSuffixRunes)
					if word.FitsInR2(newSuffixLen) {
						word.RemoveLastNRunes(newSuffixLen)
					} else {
						word.ReplaceSuffixRunes(newSuffixRunes, []rune("iqU"), true)
					}
				}
			}
			return true

		}
	}
	return false
}
