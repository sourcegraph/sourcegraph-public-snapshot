package norwegian

import (
	"github.com/kljensen/snowball/romance"
	"github.com/kljensen/snowball/snowballword"
)

// Find the starting point of the region R1.
//
// R1 is the region after the first non-vowel following a vowel,
// or is the null region at the end of the word if there is no
// such non-vowel. R2 is not used in Norwegian
//
// See http://snowball.tartarus.org/texts/r1r2.html
//
func r1(word *snowballword.SnowballWord) (r1start int) {
	// Like the German R1, the length of the Norwegian R1 is adjusted to be at least three.
	r1start = romance.VnvSuffix(word, isLowerVowel, 0)
	if r1start < 3 && len(word.RS) >= 3 {
		r1start = 3
	}
	return
}

// Checks if a rune is a lowercase Norwegian vowel.
//
func isLowerVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u', 'y', 'æ', 'ø', 'å':
		return true
	}
	return false
}

// Return `true` if the input `word` is a Norwegian stop word.
//
func isStopWord(word string) bool {
	switch word {
	case "ut", "få", "hadde", "hva", "tilbake", "vil", "han", "meget", "men", "vi", "en", "før",
		"samme", "stille", "inn", "er", "kan", "makt", "ved", "forsøke", "hvis", "part", "rett",
		"måte", "denne", "mer", "i", "lang", "ny", "hans", "hvilken", "tid", "vite", "her", "opp",
		"var", "navn", "mye", "om", "sant", "tilstand", "der", "ikke", "mest", "punkt", "hvem",
		"skulle", "mange", "over", "vårt", "alle", "arbeid", "lik", "like", "gå", "når", "siden",
		"å", "begge", "bruke", "eller", "og", "til", "da", "et", "hvorfor", "nå", "sist", "slutt",
		"deres", "det", "hennes", "så", "mens", "bra", "din", "fordi", "gjøre", "god", "ha", "start",
		"andre", "må", "med", "under", "meg", "oss", "innen", "på", "verdi", "ville", "kunne", "uten",
		"vår", "slik", "ene", "folk", "min", "riktig", "enhver", "bort", "enn", "nei", "som", "våre", "disse",
		"gjorde", "lage", "si", "du", "fra", "også", "hvordan", "av", "eneste", "for", "hvor", "først", "hver":
		return true
	}
	return false
}
