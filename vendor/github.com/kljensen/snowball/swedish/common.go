package swedish

import (
	"github.com/kljensen/snowball/romance"
	"github.com/kljensen/snowball/snowballword"
)

// Find the starting point of the region R1.
//
// R1 is the region after the first non-vowel following a vowel,
// or is the null region at the end of the word if there is no
// such non-vowel. R2 is not used in Swedish
//
// See http://snowball.tartarus.org/texts/r1r2.html
//
func r1(word *snowballword.SnowballWord) (r1start int) {
	// Like the German R1, the length of the Swedish R1 is adjusted to be at least three.
	r1start = romance.VnvSuffix(word, isLowerVowel, 0)
	if r1start < 3 && len(word.RS) >= 3 {
		r1start = 3
	}
	return
}

// Checks if a rune is a lowercase Swedish vowel.
//
func isLowerVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u', 'y', 'å', 'ä', 'ö':
		return true
	}
	return false
}

// Return `true` if the input `word` is a Swedish stop word.
//
func isStopWord(word string) bool {
	switch word {
	case "och", "det", "att", "i", "en", "jag", "hon", "som", "han",
		"på", "den", "med", "var", "sig", "för", "så", "till", "är", "men",
		"ett", "om", "hade", "de", "av", "icke", "mig", "du", "henne", "då",
		"sin", "nu", "har", "inte", "hans", "honom", "skulle", "hennes",
		"där", "min", "man", "ej", "vid", "kunde", "något", "från", "ut",
		"när", "efter", "upp", "vi", "dem", "vara", "vad", "över", "än",
		"dig", "kan", "sina", "här", "ha", "mot", "alla", "under", "någon",
		"eller", "allt", "mycket", "sedan", "ju", "denna", "själv", "detta",
		"åt", "utan", "varit", "hur", "ingen", "mitt", "ni", "bli", "blev",
		"oss", "din", "dessa", "några", "deras", "blir", "mina", "samma",
		"vilken", "er", "sådan", "vår", "blivit", "dess", "inom", "mellan",
		"sådant", "varför", "varje", "vilka", "ditt", "vem", "vilket",
		"sitta", "sådana", "vart", "dina", "vars", "vårt", "våra",
		"ert", "era", "vilkas":
		return true
	}
	return false
}
