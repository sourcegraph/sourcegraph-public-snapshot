package search

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestMatcherLookupByLanguage(t *testing.T) {
	// If we are not on CI skip the test.
	if os.Getenv("CI") == "" {
		t.Skip("Not on CI, skipping comby-dependent test")
	}

	input := map[string]string{
		"file_without_extension": `
/* This foo(plain string) {} is in a Go comment should not match in Go, but should match in plaintext */
func foo(go string) {}
`,
	}

	p := &protocol.PatternInfo{
		Pattern:         "foo(:[args])",
		IncludePatterns: []string{"file_without_extension"},
	}

	cases := []struct {
		Name      string
		Languages []string
		Want      []string
	}{
		{
			Name:      "Language test for no language",
			Languages: []string{},
			Want:      []string{"foo(plain string)", "foo(go string)"},
		},
		{
			Name:      "Language test for Go",
			Languages: []string{"go"},
			Want:      []string{"foo(go string)"},
		},
		{
			Name:      "Language test for plaintext",
			Languages: []string{"text"},
			Want:      []string{"foo(plain string)", "foo(go string)"},
		},
	}

	zipData, err := testutil.CreateZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf, cleanup, err := testutil.TempZipFileOnDisk(zipData)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			p.Languages = tt.Languages
			matches, _, err := structuralSearch(context.Background(), zf, p.Pattern, p.CombyRule, "", p.Languages, p.IncludePatterns, "repo_foo")
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, fileMatches := range matches {
				for _, m := range fileMatches.LineMatches {
					got = append(got, m.Preview)
				}
			}

			if !reflect.DeepEqual(got, tt.Want) {
				t.Fatalf("got file matches %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestMatcherLookupByExtension(t *testing.T) {
	// If we are not on CI skip the test.
	if os.Getenv("CI") == "" {
		t.Skip("Not on CI, skipping comby-dependent test")
	}

	input := map[string]string{
		"file_without_extension": `
/* This foo(plain.empty) {} is in a Go comment should not match in Go, but should match in plaintext */
func foo(go.empty) {}
`,
		"file.go": `
/* This foo(plain.go) {} is in a Go comment should not match in Go, but should match in plaintext */
func foo(go.go) {}
`,
		"file.txt": `
/* This foo(plain.txt) {} is in a Go comment should not match in Go, but should match in plaintext */
func foo(go.txt) {}
`,
	}

	zipData, err := testutil.CreateZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf, cleanup, err := testutil.TempZipFileOnDisk(zipData)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	test := func(language, filename string) string {
		var languages []string
		if language != "" {
			languages = []string{language}
		}

		extension := filepath.Ext(filename)
		matches, _, err := structuralSearch(context.Background(), zf, "foo(:[args])", "", extension, languages, []string{}, "repo_foo")
		if err != nil {
			return "ERROR"
		}
		var got []string
		for _, fileMatches := range matches {
			for _, m := range fileMatches.LineMatches {
				got = append(got, m.Preview)
			}
		}
		sort.Strings(got)
		return strings.Join(got, " ")
	}

	autogold.Want("No language and no file extension => .generic matcher", "foo(go.empty) foo(go.go) foo(go.txt) foo(plain.empty) foo(plain.go) foo(plain.txt)").Equal(t, test("", "file_without_extension"))
	autogold.Want("No language and .go file extension => .go matcher", "foo(go.empty) foo(go.go) foo(go.txt)").Equal(t, test("", "a/b/c/file.go"))
	autogold.Want("Language Go and no file extension => .go matcher", "foo(go.empty) foo(go.go) foo(go.txt)").Equal(t, test("go", ""))
	autogold.Want("Language .go and .txt file extension => .go matcher", "foo(go.empty) foo(go.go) foo(go.txt)").Equal(t, test("go", "file.txt"))
}

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
	m, _, err := structuralSearch(context.Background(), zf, p.Pattern, p.CombyRule, "", p.Languages, p.IncludePatterns, "foo")
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

func TestRecordMetrics(t *testing.T) {
	cases := []struct {
		name            string
		matcher         string
		includePatterns *[]string
		want            string
	}{
		{
			name:            "Empty values",
			matcher:         "",
			includePatterns: &[]string{},
			want:            "inferred:.generic",
		},
		{
			name:            "Include patterns no extension",
			matcher:         "",
			includePatterns: &[]string{"foo", "bar.go"},
			want:            "inferred:.generic",
		},
		{
			name:            "Include patterns first extension",
			matcher:         "",
			includePatterns: &[]string{"foo.c", "bar.go"},
			want:            "inferred:.c",
		},
		{
			name:            "Non-empty matcher",
			matcher:         ".xml",
			includePatterns: &[]string{"foo.c", "bar.go"},
			want:            ".xml",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := languageMetric(tt.matcher, tt.includePatterns)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
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
	fileMatches, _, err := structuralSearch(context.Background(), zf, p.Pattern, p.CombyRule, "", p.Languages, p.IncludePatterns, "foo")
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

func TestRule(t *testing.T) {
	// If we are not on CI skip the test.
	if os.Getenv("CI") == "" {
		t.Skip("Not on CI, skipping comby-dependent test")
	}

	input := map[string]string{
		"file.go": "func foo(success) {} func bar(fail) {}",
	}

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
		Pattern:         "func :[[fn]](:[args])",
		IncludePatterns: []string{".go"},
		CombyRule:       `where :[args] == "success"`,
	}

	got, _, err := structuralSearch(context.Background(), zf, p.Pattern, p.CombyRule, "", p.Languages, p.IncludePatterns, "repo")
	if err != nil {
		t.Fatal(err)
	}

	want := []protocol.FileMatch{
		{
			Path:     "file.go",
			LimitHit: false,
			LineMatches: []protocol.LineMatch{
				{
					LineNumber:       0,
					OffsetAndLengths: [][2]int{{0, 17}},
					Preview:          "func foo(success)",
				},
			},
			MatchCount: 1,
		},
	}

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

func TestMatchCountForMultilineMatches(t *testing.T) {
	// If we are not on CI skip the test.
	if os.Getenv("CI") == "" {
		t.Skip("Not on CI, skipping comby-dependent test")
	}

	input := map[string]string{
		"main.go": `
func foo() {
    fmt.Println("foo")
}

func bar() {
    fmt.Println("bar")
}
`,
	}

	wantMatchCount := 2

	p := &protocol.PatternInfo{Pattern: "{:[body]}"}

	zipData, err := testutil.CreateZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf, cleanup, err := testutil.TempZipFileOnDisk(zipData)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	t.Run("Strutural search match count", func(t *testing.T) {
		matches, _, err := structuralSearch(context.Background(), zf, p.Pattern, p.CombyRule, "", p.Languages, p.IncludePatterns, "repo_foo")
		if err != nil {
			t.Fatal(err)
		}
		var gotMatchCount int
		for _, fileMatches := range matches {
			gotMatchCount += fileMatches.MatchCount
		}
		if gotMatchCount != wantMatchCount {
			t.Fatalf("got match count %d, want %d", gotMatchCount, wantMatchCount)
		}
	})
}
