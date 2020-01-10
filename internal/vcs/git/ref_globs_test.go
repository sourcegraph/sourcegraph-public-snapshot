package git

import (
	"testing"
)

func TestRefGlobs(t *testing.T) {
	tests := map[string]struct {
		globs   []RefGlob
		match   []string
		noMatch []string
		want    []string
	}{
		"empty": {
			globs:   nil,
			noMatch: []string{"a"},
		},
		"globs": {
			globs:   []RefGlob{{Include: "refs/heads/"}},
			match:   []string{"refs/heads/a", "refs/heads/b/c"},
			noMatch: []string{"refs/tags/t"},
		},
		"excludes": {
			globs: []RefGlob{
				{Include: "refs/heads/"}, {Exclude: "refs/heads/x"},
			},
			match:   []string{"refs/heads/a", "refs/heads/b", "refs/heads/x/c"},
			noMatch: []string{"refs/tags/t", "refs/heads/x"},
		},
		"implicit leading refs/": {
			globs: []RefGlob{{Include: "heads/"}},
			match: []string{"refs/heads/a"},
		},
		"implicit trailing /*": {
			globs:   []RefGlob{{Include: "refs/heads/a"}},
			match:   []string{"refs/heads/a", "refs/heads/a/b"},
			noMatch: []string{"refs/heads/b"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m, err := CompileRefGlobs(test.globs)
			if err != nil {
				t.Fatal(err)
			}
			for _, ref := range test.match {
				if !m.Match(ref) {
					t.Errorf("want match %q", ref)
				}
			}
			for _, ref := range test.noMatch {
				if m.Match(ref) {
					t.Errorf("want no match %q", ref)
				}
			}
		})
	}
}
