/*
	This file contains test runners that are common to
	the romance languages.
*/
package romance

import (
	"fmt"
	"github.com/kljensen/snowball/snowballword"
	"testing"
)

type stepFunc func(*snowballword.SnowballWord) bool
type StepTestCase struct {
	WordIn     string
	R1start    int
	R2start    int
	RVstart    int
	Changed    bool
	WordOut    string
	R1startOut int
	R2startOut int
	RVstartOut int
}

func RunStepTest(t *testing.T, f stepFunc, tcs []StepTestCase) {
	for _, testCase := range tcs {
		w := snowballword.New(testCase.WordIn)
		w.R1start = testCase.R1start
		w.R2start = testCase.R2start
		w.RVstart = testCase.RVstart
		retval := f(w)
		if retval != testCase.Changed || w.String() != testCase.WordOut || w.R1start != testCase.R1startOut || w.R2start != testCase.R2startOut || w.RVstart != testCase.RVstartOut {
			t.Errorf("Expected %v -> \"{%v, %v, %v, %v, %v}\", but got \"{%v, %v, %v, %v, %v}\"", testCase.WordIn, testCase.WordOut, testCase.R1startOut, testCase.R2startOut, testCase.RVstartOut, testCase.Changed, w.String(), w.R1start, w.R2start, w.RVstart, retval)
		}
		if w.String() != testCase.WordOut {
			fmt.Printf("{\"%v\", %v, %v, %v, true, \"%v\", %v, %v, %v},\n", testCase.WordIn, testCase.R1start, testCase.R2start, testCase.RVstart, testCase.WordOut, w.R1start, w.R2start, w.RVstart)
		}
	}
}

// Test case for functions that take a word and return a bool.
type WordBoolTestCase struct {
	Word   string
	Result bool
}

// Test runner for functions that take a word and return a bool.
//
func RunWordBoolTest(t *testing.T, f func(string) bool, tcs []WordBoolTestCase) {
	for _, testCase := range tcs {
		result := f(testCase.Word)
		if result != testCase.Result {
			t.Errorf("Expected %v -> %v, but got %v", testCase.Word, testCase.Result, result)
		}
	}
}

// Test runner for functions that should be fed each rune of
// a string and that return a bool for each rune.  Usually used
// to test functions that return true if a rune is a vowel, etc.
//
func RunRunewiseBoolTest(t *testing.T, f func(rune) bool, tcs []WordBoolTestCase) {
	for _, testCase := range tcs {
		for _, r := range testCase.Word {
			result := f(r)
			if result != testCase.Result {
				t.Errorf("Expected %v -> %v, but got %v", r, testCase.Result, result)
			}
		}
	}
}

type FindRegionsTestCase struct {
	Word    string
	R1start int
	R2start int
	RVstart int
}

// Test isLowerVowel for things we know should be true
// or false.
//
func RunFindRegionsTest(t *testing.T, f func(*snowballword.SnowballWord) (int, int, int), tcs []FindRegionsTestCase) {
	for _, testCase := range tcs {
		w := snowballword.New(testCase.Word)
		r1start, r2start, rvstart := f(w)
		if r1start != testCase.R1start || r2start != testCase.R2start || rvstart != testCase.RVstart {
			t.Errorf("Expect \"%v\" -> %v, %v, %v, but got %v, %v, %v",
				testCase.Word, testCase.R1start, testCase.R2start, testCase.RVstart,
				r1start, r2start, rvstart,
			)
		}

	}
}
