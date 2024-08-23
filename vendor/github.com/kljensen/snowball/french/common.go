package french

import (
	"github.com/kljensen/snowball/romance"
	"github.com/kljensen/snowball/snowballword"
)

// Return `true` if the input `word` is a French stop word.
//
func isStopWord(word string) bool {
	switch word {
	case "au", "aux", "avec", "ce", "ces", "dans", "de", "des", "du",
		"elle", "en", "et", "eux", "il", "je", "la", "le", "leur",
		"lui", "ma", "mais", "me", "même", "mes", "moi", "mon", "ne",
		"nos", "notre", "nous", "on", "ou", "par", "pas", "pour", "qu",
		"que", "qui", "sa", "se", "ses", "son", "sur", "ta", "te",
		"tes", "toi", "ton", "tu", "un", "une", "vos", "votre", "vous",
		"c", "d", "j", "l", "à", "m", "n", "s", "t", "y", "été",
		"étée", "étées", "étés", "étant", "étante", "étants", "étantes",
		"suis", "es", "est", "sommes", "êtes", "sont", "serai",
		"seras", "sera", "serons", "serez", "seront", "serais",
		"serait", "serions", "seriez", "seraient", "étais", "était",
		"étions", "étiez", "étaient", "fus", "fut", "fûmes", "fûtes",
		"furent", "sois", "soit", "soyons", "soyez", "soient", "fusse",
		"fusses", "fût", "fussions", "fussiez", "fussent", "ayant",
		"ayante", "ayantes", "ayants", "eu", "eue", "eues", "eus",
		"ai", "as", "avons", "avez", "ont", "aurai", "auras", "aura",
		"aurons", "aurez", "auront", "aurais", "aurait", "aurions",
		"auriez", "auraient", "avais", "avait", "avions", "aviez",
		"avaient", "eut", "eûmes", "eûtes", "eurent", "aie", "aies",
		"ait", "ayons", "ayez", "aient", "eusse", "eusses", "eût",
		"eussions", "eussiez", "eussent":
		return true
	}
	return false
}

// Checks if a rune is a lowercase French vowel.
//
func isLowerVowel(r rune) bool {

	// The French vowels are "aeiouyâàëéêèïîôûù", which
	// are referenced by their unicode code points
	// in the switch statement below.
	switch r {
	case 97, 101, 105, 111, 117, 121, 226, 224, 235, 233, 234, 232, 239, 238, 244, 251, 249:
		return true
	}
	return false
}

// Capitalize Y, I, and U runes that are acting as consanants.
// Put into upper case "u" or "i" preceded and followed by a
// vowel, and "y" preceded or followed by a vowel. "u" after q is
// also put into upper case.
//
func capitalizeYUI(word *snowballword.SnowballWord) {

	// Keep track of vowels that we see
	vowelPreviously := false

	// Peak ahead to see if the next rune is a vowel
	vowelNext := func(j int) bool {
		return (j+1 < len(word.RS) && isLowerVowel(word.RS[j+1]))
	}

	// Look at all runes
	for i := 0; i < len(word.RS); i++ {

		// Nothing to do for non-vowels
		if isLowerVowel(word.RS[i]) == false {
			vowelPreviously = false
			continue
		}

		vowelHere := true

		switch word.RS[i] {
		case 121: // y

			// Is this "y" preceded OR followed by a vowel?
			if vowelPreviously || vowelNext(i) {
				word.RS[i] = 89 // Y
				vowelHere = false
			}

		case 117: // u

			// Is this "u" is flanked by vowels OR preceded by a "q"?
			if (vowelPreviously && vowelNext(i)) || (i >= 1 && word.RS[i-1] == 113) {
				word.RS[i] = 85 // U
				vowelHere = false
			}

		case 105: // i

			// Is this "i" is flanked by vowels?
			if vowelPreviously && vowelNext(i) {
				word.RS[i] = 73 // I
				vowelHere = false
			}
		}
		vowelPreviously = vowelHere
	}
}

// Find the starting point of the regions R1, R2, & RV
//
func findRegions(word *snowballword.SnowballWord) (r1start, r2start, rvstart int) {

	// R1 & R2 are defined in the standard manner.
	r1start = romance.VnvSuffix(word, isLowerVowel, 0)
	r2start = romance.VnvSuffix(word, isLowerVowel, r1start)

	// Set RV, by default, as empty.
	rvstart = len(word.RS)

	// Handle the three special cases: "par", "col", & "tap"
	//
	prefix, prefixRunes := word.FirstPrefix("par", "col", "tap")
	if prefix != "" {
		rvstart = len(prefixRunes)
		return
	}

	// If the word begins with two vowels, RV is the region after the third letter
	if len(word.RS) >= 3 && isLowerVowel(word.RS[0]) && isLowerVowel(word.RS[1]) {
		rvstart = 3
		return
	}

	// Otherwise the region after the first vowel not at the beginning of the word.
	for i := 1; i < len(word.RS); i++ {
		if isLowerVowel(word.RS[i]) {
			rvstart = i + 1
			return
		}
	}

	return
}
