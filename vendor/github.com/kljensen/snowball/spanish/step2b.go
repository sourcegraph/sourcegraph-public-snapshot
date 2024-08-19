package spanish

import (
	"github.com/kljensen/snowball/snowballword"
)

// Step 2b is the removal of verb suffixes beginning y,
// Search for the longest among the following suffixes
// in RV, and if found, delete if preceded by u.
//
func step2b(word *snowballword.SnowballWord) bool {
	suffix, suffixRunes := word.FirstSuffixIn(word.RVstart, len(word.RS),
		"iésemos", "iéramos", "iríamos", "eríamos", "aríamos", "ásemos",
		"áramos", "ábamos", "isteis", "iríais", "iremos", "ieseis",
		"ierais", "eríais", "eremos", "asteis", "aríais", "aremos",
		"íamos", "irías", "irían", "iréis", "ieses", "iesen", "ieron",
		"ieras", "ieran", "iendo", "erías", "erían", "eréis", "aseis",
		"arías", "arían", "aréis", "arais", "abais", "íais", "iste",
		"iría", "irás", "irán", "imos", "iese", "iera", "idos", "idas",
		"ería", "erás", "erán", "aste", "ases", "asen", "aría", "arás",
		"arán", "aron", "aras", "aran", "ando", "amos", "ados", "adas",
		"abas", "aban", "ías", "ían", "éis", "áis", "iré", "irá", "ido",
		"ida", "eré", "erá", "emos", "ase", "aré", "ará", "ara", "ado",
		"ada", "aba", "ís", "ía", "ió", "ir", "id", "es", "er", "en",
		"ed", "as", "ar", "an", "ad",
	)
	switch suffix {
	case "":
		return false

	case "en", "es", "éis", "emos":

		// Delete, and if preceded by gu delete the u (the gu need not be in RV)
		word.RemoveLastNRunes(len(suffixRunes))
		guSuffix, _ := word.FirstSuffix("gu")
		if guSuffix != "" {
			word.RemoveLastNRunes(1)
		}

	default:

		// Delete
		word.RemoveLastNRunes(len(suffixRunes))
	}
	return true
}
