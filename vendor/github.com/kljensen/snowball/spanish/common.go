package spanish

import (
	"github.com/kljensen/snowball/romance"
	"github.com/kljensen/snowball/snowballword"
)

// Change the vowels "áéíóú" into "aeiou".
//
func removeAccuteAccents(word *snowballword.SnowballWord) (didReplacement bool) {
	for i := 0; i < len(word.RS); i++ {
		switch word.RS[i] {
		case 225:
			// á -> a
			word.RS[i] = 97
			didReplacement = true
		case 233:
			// é -> e
			word.RS[i] = 101
			didReplacement = true
		case 237:
			// í -> i
			word.RS[i] = 105
			didReplacement = true
		case 243:
			// ó -> o
			word.RS[i] = 111
			didReplacement = true
		case 250:
			// ú -> u
			word.RS[i] = 117
			didReplacement = true
		}
	}
	return
}

// Find the starting point of the regions R1, R2, & RV
//
func findRegions(word *snowballword.SnowballWord) (r1start, r2start, rvstart int) {

	r1start = romance.VnvSuffix(word, isLowerVowel, 0)
	r2start = romance.VnvSuffix(word, isLowerVowel, r1start)
	rvstart = len(word.RS)

	if len(word.RS) >= 3 {
		switch {

		case !isLowerVowel(word.RS[1]):

			// If the second letter is a consonant, RV is the region after the
			// next following vowel.
			for i := 2; i < len(word.RS); i++ {
				if isLowerVowel(word.RS[i]) {
					rvstart = i + 1
					break
				}
			}

		case isLowerVowel(word.RS[0]) && isLowerVowel(word.RS[1]):

			// Or if the first two letters are vowels, RV
			// is the region after the next consonant.
			for i := 2; i < len(word.RS); i++ {
				if !isLowerVowel(word.RS[i]) {
					rvstart = i + 1
					break
				}
			}
		default:

			// Otherwise (consonant-vowel case) RV is the region after the
			// third letter. But RV is the end of the word if these
			// positions cannot be found.
			rvstart = 3
		}
	}

	return
}

// Checks if a rune is a lowercase Spanish vowel.
//
func isLowerVowel(r rune) bool {

	// The spanish vowels are "aeiouáéíóúü", which
	// are referenced by their unicode code points
	// in the switch statement below.
	switch r {
	case 97, 101, 105, 111, 117, 225, 233, 237, 243, 250, 252:
		return true
	}
	return false
}

// Return `true` if the input `word` is a Spanish stop word.
//
func isStopWord(word string) bool {
	switch word {
	case "de", "la", "que", "el", "en", "y", "a", "los", "del", "se", "las",
		"por", "un", "para", "con", "no", "una", "su", "al", "lo", "como",
		"más", "pero", "sus", "le", "ya", "o", "este", "sí", "porque", "esta",
		"entre", "cuando", "muy", "sin", "sobre", "también", "me", "hasta",
		"hay", "donde", "quien", "desde", "todo", "nos", "durante", "todos",
		"uno", "les", "ni", "contra", "otros", "ese", "eso", "ante", "ellos",
		"e", "esto", "mí", "antes", "algunos", "qué", "unos", "yo", "otro",
		"otras", "otra", "él", "tanto", "esa", "estos", "mucho", "quienes",
		"nada", "muchos", "cual", "poco", "ella", "estar", "estas", "algunas",
		"algo", "nosotros", "mi", "mis", "tú", "te", "ti", "tu", "tus", "ellas",
		"nosotras", "vosostros", "vosostras", "os", "mío", "mía", "míos", "mías",
		"tuyo", "tuya", "tuyos", "tuyas", "suyo", "suya", "suyos", "suyas",
		"nuestro", "nuestra", "nuestros", "nuestras", "vuestro", "vuestra",
		"vuestros", "vuestras", "esos", "esas", "estoy", "estás", "está", "estamos",
		"estáis", "están", "esté", "estés", "estemos", "estéis", "estén", "estaré",
		"estarás", "estará", "estaremos", "estaréis", "estarán", "estaría",
		"estarías", "estaríamos", "estaríais", "estarían", "estaba", "estabas",
		"estábamos", "estabais", "estaban", "estuve", "estuviste", "estuvo",
		"estuvimos", "estuvisteis", "estuvieron", "estuviera", "estuvieras",
		"estuviéramos", "estuvierais", "estuvieran", "estuviese", "estuvieses",
		"estuviésemos", "estuvieseis", "estuviesen", "estando", "estado",
		"estada", "estados", "estadas", "estad", "he", "has", "ha", "hemos",
		"habéis", "han", "haya", "hayas", "hayamos", "hayáis", "hayan",
		"habré", "habrás", "habrá", "habremos", "habréis", "habrán", "habría",
		"habrías", "habríamos", "habríais", "habrían", "había", "habías",
		"habíamos", "habíais", "habían", "hube", "hubiste", "hubo", "hubimos",
		"hubisteis", "hubieron", "hubiera", "hubieras", "hubiéramos", "hubierais",
		"hubieran", "hubiese", "hubieses", "hubiésemos", "hubieseis", "hubiesen",
		"habiendo", "habido", "habida", "habidos", "habidas", "soy", "eres",
		"es", "somos", "sois", "son", "sea", "seas", "seamos", "seáis", "sean",
		"seré", "serás", "será", "seremos", "seréis", "serán", "sería", "serías",
		"seríamos", "seríais", "serían", "era", "eras", "éramos", "erais",
		"eran", "fui", "fuiste", "fue", "fuimos", "fuisteis", "fueron", "fuera",
		"fueras", "fuéramos", "fuerais", "fueran", "fuese", "fueses", "fuésemos",
		"fueseis", "fuesen", "sintiendo", "sentido", "sentida", "sentidos",
		"sentidas", "siente", "sentid", "tengo", "tienes", "tiene", "tenemos",
		"tenéis", "tienen", "tenga", "tengas", "tengamos", "tengáis", "tengan",
		"tendré", "tendrás", "tendrá", "tendremos", "tendréis", "tendrán",
		"tendría", "tendrías", "tendríamos", "tendríais", "tendrían", "tenía",
		"tenías", "teníamos", "teníais", "tenían", "tuve", "tuviste", "tuvo",
		"tuvimos", "tuvisteis", "tuvieron", "tuviera", "tuvieras", "tuviéramos",
		"tuvierais", "tuvieran", "tuviese", "tuvieses", "tuviésemos", "tuvieseis",
		"tuviesen", "teniendo", "tenido", "tenida", "tenidos", "tenidas", "tened":
		return true
	}
	return false
}
