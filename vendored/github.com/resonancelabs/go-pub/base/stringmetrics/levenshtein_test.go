package stringmetrics

import (
	"fmt"
	"testing"
)

var tests = []struct {
	a, b string
	dist int
}{
	{"levenshtein", "levenshtein", 0},
	{"levenshtein", "levenAshtein", 1},
	{"levenshtein", "levenAshteinB", 2},
	{"levenshtein", "evenshtein", 1},
	{"levenshtein", "evenshtei", 2},
	{"levenshtein", "evenAshtei", 3},
	{"wokr", "work", 2},
	{"", "testing", 7},
	{"testing", "", 7},
	{"penny", "pickle", 5},
}

func TestLevenshtein(t *testing.T) {
	for _, test := range tests {
		dist := Levenshtein(test.a, test.b)
		if dist != test.dist {
			t.Errorf(fmt.Sprintf("Expected edit distance (%s, %s) => %v, got %v",
				test.a, test.b, test.dist, dist))
		}
	}
}

func BenchmarkLevenshtein(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			Levenshtein(test.a, test.b)
		}
	}
}
