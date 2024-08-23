package russian

import (
	"github.com/kljensen/snowball/romance"
	"github.com/kljensen/snowball/snowballword"
)

// Checks if a rune is a lowercase Russian vowel.
//
func isLowerVowel(r rune) bool {

	// The Russian vowels are "аеиоуыэюя", which
	// are referenced by their unicode code points
	// in the switch statement below.
	switch r {
	case 1072, 1077, 1080, 1086, 1091, 1099, 1101, 1102, 1103:
		return true
	}
	return false
}

// Return `true` if the input `word` is a French stop word.
//
func isStopWord(word string) bool {
	switch word {
	case "и", "в", "во", "не", "что", "он", "на", "я", "с",
		"со", "как", "а", "то", "все", "она", "так", "его",
		"но", "да", "ты", "к", "у", "же", "вы", "за", "бы",
		"по", "только", "ее", "мне", "было", "вот", "от",
		"меня", "еще", "нет", "о", "из", "ему", "теперь",
		"когда", "даже", "ну", "вдруг", "ли", "если", "уже",
		"или", "ни", "быть", "был", "него", "до", "вас",
		"нибудь", "опять", "уж", "вам", "ведь", "там", "потом",
		"себя", "ничего", "ей", "может", "они", "тут", "где",
		"есть", "надо", "ней", "для", "мы", "тебя", "их",
		"чем", "была", "сам", "чтоб", "без", "будто", "чего",
		"раз", "тоже", "себе", "под", "будет", "ж", "тогда",
		"кто", "этот", "того", "потому", "этого", "какой",
		"совсем", "ним", "здесь", "этом", "один", "почти",
		"мой", "тем", "чтобы", "нее", "сейчас", "были", "куда",
		"зачем", "всех", "никогда", "можно", "при", "наконец",
		"два", "об", "другой", "хоть", "после", "над", "больше",
		"тот", "через", "эти", "нас", "про", "всего", "них",
		"какая", "много", "разве", "три", "эту", "моя",
		"впрочем", "хорошо", "свою", "этой", "перед", "иногда",
		"лучше", "чуть", "том", "нельзя", "такой", "им", "более",
		"всегда", "конечно", "всю", "между":
		return true
	}
	return false
}

// Find the starting point of the regions R1, R2, & RV
//
func findRegions(word *snowballword.SnowballWord) (r1start, r2start, rvstart int) {

	// R1 & R2 are defined in the standard manner.
	r1start = romance.VnvSuffix(word, isLowerVowel, 0)
	r2start = romance.VnvSuffix(word, isLowerVowel, r1start)

	// Set RV, by default, as empty.
	rvstart = len(word.RS)

	// RV is the region after the first vowel, or the end of
	// the word if it contains no vowel.
	//
	for i := 0; i < len(word.RS); i++ {
		if isLowerVowel(word.RS[i]) {
			rvstart = i + 1
			break
		}
	}

	return
}
