package search

import (
	"testing"
)

func TestFileFilter(t *testing.T) {
	cases := []struct {
		include     []string
		exclude     string
		file        string
		shouldMatch bool
	}{
		{[]string{}, "", "test.go", true},
		{[]string{`.*\.go`}, "", "test.go", true},
		{[]string{`.*\.go`}, "", "test.rb", false},
		{[]string{`.*\.go`, `test.*`}, "", "test.go", true},
		{[]string{`.*\.go`, `test.*`}, "", "foo.go", false},
		{[]string{}, `.*`, "foo.go", false},
		{[]string{}, `foo\.*`, "foo.go", false},
		{[]string{`.*`}, `foo\.*`, "foo.go", false},
		{[]string{`.*`}, `bar\.*`, "foo.go", true},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			fm, err := newFileFilter(tc.include, tc.exclude)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			matches := fm.Match(tc.file)
			if matches != tc.shouldMatch {
				t.Fatalf("Expect matches to be %v, but got %v", tc.shouldMatch, matches)
			}
		})
	}
}
