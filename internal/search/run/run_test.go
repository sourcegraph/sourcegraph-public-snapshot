package run

import "testing"

func TestOrderedFuzzyRegexp(t *testing.T) {
	got := orderedFuzzyRegexp([]string{})
	if want := ""; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = orderedFuzzyRegexp([]string{"a"})
	if want := "a"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = orderedFuzzyRegexp([]string{"a", "b|c"})
	if want := "(a).*?(b|c)"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
