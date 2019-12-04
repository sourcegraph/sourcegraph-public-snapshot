package search

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

// Tests that structural search correctly infers the Go matcher from the .go
// file extension.
func TestInferredMatcher(t *testing.T) {
	// If we are not on CI skip the test.
	if os.Getenv("CI") == "" {
		t.Skip("Not on CI, skipping comby-dependent test")
	}

	input := map[string]string{
		"main.go": `
/* This foo(ignore string) {} is in a Go comment should not match */
func foo(real string) {}
`,
	}

	pattern := "foo(:[args])"
	want := "foo(real string)"

	includePatterns := []string{"main.go"}

	zipData, err := testutil.CreateZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf, cleanup, err := testutil.TempZipFileOnDisk(zipData)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	p := &protocol.PatternInfo{
		Pattern:         pattern,
		IncludePatterns: includePatterns,
	}
	m, _, err := structuralSearch(context.Background(), zf, p.Pattern, p.IncludePatterns, "foo")
	if err != nil {
		t.Fatal(err)
	}
	got := m[0].LineMatches[0].Preview
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("got file matches %v, want %v", got, want)
	}
}

// Tests that includePatterns works. includePatterns serve a similar role in
// structural search compared to regex search, but is interpreted _differently_.
// includePatterns cannot be a regex expression (as in traditional search), but
// instead (currently) expects a list of patterns that represent a set of file
// paths to search.
func TestIncludePatterns(t *testing.T) {
	// If we are not on CI skip the test.
	if os.Getenv("CI") == "" {
		t.Skip("Not on CI, skipping comby-dependent test")
	}

	input := map[string]string{
		"/a/b/c":         "",
		"/a/b/c/foo.go":  "",
		"c/foo.go":       "",
		"bar.go":         "",
		"/x/y/z/bar.go":  "",
		"/a/b/c/nope.go": "",
		"nope.go":        "",
	}

	want := []string{
		"/a/b/c/foo.go",
		"/x/y/z/bar.go",
		"bar.go",
	}

	includePatterns := []string{"a/b/c/foo.go", "bar.go"}

	zipData, err := testutil.CreateZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf, cleanup, err := testutil.TempZipFileOnDisk(zipData)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	p := &protocol.PatternInfo{
		Pattern:         "",
		IncludePatterns: includePatterns,
	}
	fileMatches, _, err := structuralSearch(context.Background(), zf, p.Pattern, p.IncludePatterns, "foo")
	if err != nil {
		t.Fatal(err)
	}

	got := make([]string, len(fileMatches))
	for i, fm := range fileMatches {
		got[i] = fm.Path
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got file matches %v, want %v", got, want)
	}
}

func TestHighlightMultipleLines(t *testing.T) {
	cases := []struct {
		Name  string
		Match *comby.Match
		Want  []protocol.LineMatch
	}{
		{
			Name: "Single line",
			Match: &comby.Match{
				Range: comby.Range{
					Start: comby.Location{
						Line:   1,
						Column: 1,
					},
					End: comby.Location{
						Line:   1,
						Column: 2,
					},
				},
				Matched: "this is a single line match",
			},
			Want: []protocol.LineMatch{
				{
					LineNumber: 0,
					OffsetAndLengths: [][2]int{
						{
							0,
							1,
						},
					},
					Preview: "this is a single line match",
				},
			},
		},
		{
			Name: "Three lines",
			Match: &comby.Match{
				Range: comby.Range{
					Start: comby.Location{
						Line:   1,
						Column: 1,
					},
					End: comby.Location{
						Line:   3,
						Column: 5,
					},
				},
				Matched: "this is a match across\nthree\nlines",
			},
			Want: []protocol.LineMatch{
				{
					LineNumber: 0,
					OffsetAndLengths: [][2]int{
						{
							0,
							22,
						},
					},
					Preview: "this is a match across",
				},
				{
					LineNumber: 1,
					OffsetAndLengths: [][2]int{
						{
							0,
							5,
						},
					},
					Preview: "three",
				},
				{
					LineNumber: 2,
					OffsetAndLengths: [][2]int{
						{
							0,
							4, // don't include trailing newline
						},
					},
					Preview: "lines",
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			got := highlightMultipleLines(tt.Match)
			if !reflect.DeepEqual(got, tt.Want) {
				jsonGot, _ := json.Marshal(got)
				jsonWant, _ := json.Marshal(tt.Want)
				t.Errorf("got: %s, want: %s", jsonGot, jsonWant)
			}
		})
	}
}
