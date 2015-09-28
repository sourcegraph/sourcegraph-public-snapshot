package base

import (
	"regexp"
	. "testing"
)

func TestRandomIdString(t *T) {

	for i := 1; i <= 100; i++ {
		if s := RandomIdString(i); len(s) != i {
			t.Error("Incorrect length")
		}
	}

	containsAVowel, _ := regexp.Compile(`[aeiou]`)

	history := make(map[string]bool)
	collisions := 0
	count := 0
	for i := 0; i < 10000; i++ {

		s := RandomIdString(10)
		count++

		// The chance of a collision here here should be very, very low.
		val, found := history[s]
		if val || found {
			t.Errorf("Collision in generated string! %v", s)
			collisions++
		}
		history[s] = true

		if containsAVowel.MatchString(s) {
			t.Errorf("RandomIdString contains a vowel: %v", s)
		}
	}
	if collisions > 0 {
		percent := 100.0 * float64(collisions) / float64(count)
		t.Errorf("%v collisions in %v.  %v%%", collisions, count, percent)
	}

}
