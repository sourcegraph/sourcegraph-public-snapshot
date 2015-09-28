package tmpl

import "testing"

func TestUnixPath(t *testing.T) {
	var paths = map[string]string{
		"a\\b\\c\\d/e\\": "a/b/c/d/e",
		"a/b/c/":         "a/b/c",
		"a/b":            "a/b",
	}
	for p, correct := range paths {
		got := unixPath(p)
		if got != correct {
			t.Fatalf("got %q expected %q", got, correct)
		}
	}
}

func TestStripExt(t *testing.T) {
	var files = map[string]string{
		"hello.txt":      "hello",
		"no_ext":         "no_ext",
		"long.extension": "long",
	}
	for f, correct := range files {
		got := stripExt(f)
		if got != correct {
			t.Fatalf("got %q expected %q", got, correct)
		}
	}
}

func TestComments(t *testing.T) {
	var tests = map[string][]string{
		"we like to\nkeep width\nbelow 10\n\nbut sometimes we go over\n\t   \ncrazy, right?\n": []string{
			"we like to keep width below 10",
			"but sometimes we go over",
			"crazy, right?",
		},
		"one line":   []string{"one line"},
		"two\nlines": []string{"two lines"},
		"\nbegin":    []string{"begin"},
	}
	for input, want := range tests {
		got := comments(input)
		if len(got) != len(want) {
			t.Fatalf("got %#q want %#q", got, want)
		}
		for i, w := range want {
			if w != got[i] {
				t.Fatalf("%d. got %q expected %q", i, got[i], w)
			}
		}
	}
}
