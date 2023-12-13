package search

import (
	"archive/tar"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

func TestMatcherLookupByLanguage(t *testing.T) {
	maybeSkipComby(t)

	input := map[string]string{
		"file_without_extension": `
/* This foo(plain string) {} is in a Go comment should not match in Go, but should match in plaintext */
func foo(go string) {}
`,
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

	zipData, err := createZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf := tempZipFileOnDisk(t, zipData)

	t.Run("group", func(t *testing.T) {
		for _, tt := range cases {
			tt := tt
			t.Run(tt.Name, func(t *testing.T) {
				t.Parallel()

				p := &protocol.PatternInfo{
					Pattern:         "foo(:[args])",
					IncludePatterns: []string{"file_without_extension"},
					Languages:       tt.Languages,
				}

				ctx, cancel, sender := newLimitedStreamCollector(context.Background(), 100000000)
				defer cancel()
				err := structuralSearch(ctx, logtest.Scoped(t), comby.ZipPath(zf), subset(p.IncludePatterns), "", p.Pattern, p.CombyRule, p.Languages, "repo_foo", 0, sender)
				if err != nil {
					t.Fatal(err)
				}
				var got []string
				for _, fileMatches := range sender.collected {
					for _, m := range fileMatches.ChunkMatches {
						got = append(got, m.MatchedContent()...)
					}
				}

				if !reflect.DeepEqual(got, tt.Want) {
					t.Fatalf("got file matches %q, want %q", got, tt.Want)
				}
			})
		}
	})
}

func TestMatcherLookupByExtension(t *testing.T) {
	maybeSkipComby(t)

	t.Parallel()

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

	zipData, err := createZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf := tempZipFileOnDisk(t, zipData)

	test := func(language, filename string) string {
		var languages []string
		if language != "" {
			languages = []string{language}
		}

		extensionHint := filepath.Ext(filename)
		ctx, cancel, sender := newLimitedStreamCollector(context.Background(), 1000000000)
		defer cancel()
		err := structuralSearch(ctx, logtest.Scoped(t), comby.ZipPath(zf), all, extensionHint, "foo(:[args])", "", languages, "repo_foo", 0, sender)
		if err != nil {
			return "ERROR: " + err.Error()
		}
		var got []string
		for _, fileMatches := range sender.collected {
			for _, m := range fileMatches.ChunkMatches {
				got = append(got, m.MatchedContent()...)
			}
		}
		sort.Strings(got)
		return strings.Join(got, " ")
	}

	cases := []struct {
		name     string
		want     string
		language string
		filename string
	}{{
		name:     "No language and no file extension => .generic matcher",
		want:     "foo(go.empty) foo(go.go) foo(go.txt) foo(plain.empty) foo(plain.go) foo(plain.txt)",
		language: "",
		filename: "file_without_extension",
	}, {
		name:     "No language and .go file extension => .go matcher",
		want:     "foo(go.empty) foo(go.go) foo(go.txt)",
		language: "",
		filename: "a/b/c/file.go",
	}, {
		name:     "Language Go and no file extension => .go matcher",
		want:     "foo(go.empty) foo(go.go) foo(go.txt)",
		language: "go",
		filename: "",
	}, {
		name:     "Language .go and .txt file extension => .go matcher",
		want:     "foo(go.empty) foo(go.go) foo(go.txt)",
		language: "go",
		filename: "file.txt",
	}}
	t.Run("group", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got := test(tc.language, tc.filename)
				if d := cmp.Diff(tc.want, got); d != "" {
					t.Errorf("mismatch (-want +got):\n%s", d)
				}
			})
		}
	})
}

// Tests that structural search correctly infers the Go matcher from the .go
// file extension.
func TestInferredMatcher(t *testing.T) {
	maybeSkipComby(t)

	input := map[string]string{
		"main.go": `
/* This foo(ignore string) {} is in a Go comment should not match */
func foo(real string) {}
`,
	}

	pattern := "foo(:[args])"
	want := "foo(real string)"

	zipData, err := createZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zPath := tempZipFileOnDisk(t, zipData)

	zFile, _ := mockZipFile(zipData)
	if err != nil {
		t.Fatal(err)
	}

	p := &protocol.PatternInfo{
		Pattern: pattern,
		Limit:   30,
	}
	ctx, cancel, sender := newLimitedStreamCollector(context.Background(), 1000000000)
	defer cancel()
	err = filteredStructuralSearch(ctx, logtest.Scoped(t), zPath, zFile, p, "foo", sender, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := sender.collected[0].ChunkMatches[0].MatchedContent()[0]
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
		language        []string
		includePatterns []string
		want            string
	}{
		{
			name:            "Empty values",
			language:        nil,
			includePatterns: []string{},
			want:            ".generic",
		},
		{
			name:            "Include patterns no extension",
			language:        nil,
			includePatterns: []string{"foo", "bar.go"},
			want:            ".generic",
		},
		{
			name:            "Include patterns first extension",
			language:        nil,
			includePatterns: []string{"foo.c", "bar.go"},
			want:            ".c",
		},
		{
			name:            "Non-empty language",
			language:        []string{"xml"},
			includePatterns: []string{"foo.c", "bar.go"},
			want:            ".xml",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var extensionHint string
			if len(tt.includePatterns) > 0 {
				filename := tt.includePatterns[0]
				extensionHint = filepath.Ext(filename)
			}
			got := toMatcher(tt.language, extensionHint)
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
	maybeSkipComby(t)

	input := map[string]string{
		"a/b/c":         "",
		"a/b/c/foo.go":  "",
		"c/foo.go":      "",
		"bar.go":        "",
		"x/y/z/bar.go":  "",
		"a/b/c/nope.go": "",
		"nope.go":       "",
	}

	want := []string{
		"a/b/c/foo.go",
		"bar.go",
		"x/y/z/bar.go",
	}

	includePatterns := []string{"a/b/c/foo.go", "bar.go"}

	zipData, err := createZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf := tempZipFileOnDisk(t, zipData)

	p := &protocol.PatternInfo{
		Pattern:         "",
		IncludePatterns: includePatterns,
	}
	ctx, cancel, sender := newLimitedStreamCollector(context.Background(), 1000000000)
	defer cancel()
	err = structuralSearch(ctx, logtest.Scoped(t), comby.ZipPath(zf), subset(p.IncludePatterns), "", p.Pattern, p.CombyRule, p.Languages, "foo", 0, sender)
	if err != nil {
		t.Fatal(err)
	}
	fileMatches := sender.collected

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
	maybeSkipComby(t)

	input := map[string]string{
		"file.go": "func foo(success) {} func bar(fail) {}",
	}

	zipData, err := createZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf := tempZipFileOnDisk(t, zipData)

	p := &protocol.PatternInfo{
		Pattern:         "func :[[fn]](:[args])",
		IncludePatterns: []string{".go"},
		CombyRule:       `where :[args] == "success"`,
	}

	ctx, cancel, sender := newLimitedStreamCollector(context.Background(), 1000000000)
	defer cancel()
	err = structuralSearch(ctx, logtest.Scoped(t), comby.ZipPath(zf), subset(p.IncludePatterns), "", p.Pattern, p.CombyRule, p.Languages, "repo", 0, sender)
	if err != nil {
		t.Fatal(err)
	}
	got := sender.collected

	want := []protocol.FileMatch{{
		Path:     "file.go",
		LimitHit: false,
		ChunkMatches: []protocol.ChunkMatch{{
			Content:      "func foo(success) {} func bar(fail) {}",
			ContentStart: protocol.Location{Offset: 0, Line: 0, Column: 0},
			Ranges: []protocol.Range{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 17, Line: 0, Column: 17},
			}},
		}},
	}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got file matches %v, want %v", got, want)
	}
}

func TestStructuralLimits(t *testing.T) {
	maybeSkipComby(t)

	input := map[string]string{
		"test1.go": `
func foo() {
    fmt.Println("foo")
}

func bar() {
    fmt.Println("bar")
}
`,
		"test2.go": `
func foo() {
    fmt.Println("foo")
}

func bar() {
    fmt.Println("bar")
}
`,
	}

	zipData, err := createZip(input)
	require.NoError(t, err)

	zf := tempZipFileOnDisk(t, zipData)

	count := func(matches []protocol.FileMatch) int {
		c := 0
		for _, match := range matches {
			c += match.MatchCount()
		}
		return c
	}

	test := func(limit, wantCount int, p *protocol.PatternInfo) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel, sender := newLimitedStreamCollector(context.Background(), limit)
			defer cancel()
			err := structuralSearch(ctx, logtest.Scoped(t), comby.ZipPath(zf), subset(p.IncludePatterns), "", p.Pattern, p.CombyRule, p.Languages, "repo_foo", 0, sender)
			require.NoError(t, err)

			require.Equal(t, wantCount, count(sender.collected))
		}
	}

	t.Run("unlimited", test(10000, 4, &protocol.PatternInfo{Pattern: "{:[body]}"}))
	t.Run("exact limit", func(t *testing.T) { t.Skip("disabled because flaky") }) // test(4, 4, &protocol.PatternInfo{Pattern: "{:[body]}"}))
	t.Run("limited", func(t *testing.T) { t.Skip("disabled because flaky") })     // test(2, 2, &protocol.PatternInfo{Pattern: "{:[body]}"}))
	t.Run("many", test(12, 8, &protocol.PatternInfo{Pattern: "(:[_])"}))
}

func TestMatchCountForMultilineMatches(t *testing.T) {
	maybeSkipComby(t)

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

	zipData, err := createZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf := tempZipFileOnDisk(t, zipData)

	t.Run("Strutural search match count", func(t *testing.T) {
		ctx, cancel, sender := newLimitedStreamCollector(context.Background(), 1000000000)
		defer cancel()
		err := structuralSearch(ctx, logtest.Scoped(t), comby.ZipPath(zf), subset(p.IncludePatterns), "", p.Pattern, p.CombyRule, p.Languages, "repo_foo", 0, sender)
		if err != nil {
			t.Fatal(err)
		}
		matches := sender.collected
		var gotMatchCount int
		for _, fileMatches := range matches {
			gotMatchCount += fileMatches.MatchCount()
		}
		if gotMatchCount != wantMatchCount {
			t.Fatalf("got match count %d, want %d", gotMatchCount, wantMatchCount)
		}
	})
}

func TestMultilineMatches(t *testing.T) {
	maybeSkipComby(t)

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

	p := &protocol.PatternInfo{Pattern: "{:[body]}"}

	zipData, err := createZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf := tempZipFileOnDisk(t, zipData)

	t.Run("Strutural search match count", func(t *testing.T) {
		ctx, cancel, sender := newLimitedStreamCollector(context.Background(), 1000000000)
		defer cancel()
		err := structuralSearch(ctx, logtest.Scoped(t), comby.ZipPath(zf), subset(p.IncludePatterns), "", p.Pattern, p.CombyRule, p.Languages, "repo_foo", 0, sender)
		if err != nil {
			t.Fatal(err)
		}
		matches := sender.collected
		expected := []protocol.FileMatch{{
			Path: "main.go",
			ChunkMatches: []protocol.ChunkMatch{{
				Content:      "func foo() {\n    fmt.Println(\"foo\")\n}",
				ContentStart: protocol.Location{Offset: 1, Line: 1},
				Ranges: []protocol.Range{{
					Start: protocol.Location{Offset: 12, Line: 1, Column: 11},
					End:   protocol.Location{Offset: 38, Line: 3, Column: 1},
				}},
			}, {
				Content:      "func bar() {\n    fmt.Println(\"bar\")\n}",
				ContentStart: protocol.Location{Offset: 40, Line: 5},
				Ranges: []protocol.Range{{
					Start: protocol.Location{Offset: 51, Line: 5, Column: 11},
					End:   protocol.Location{Offset: 77, Line: 7, Column: 1},
				}},
			}},
		}}
		require.Equal(t, expected, matches)
	})
}

func TestBuildQuery(t *testing.T) {
	pattern := ":[x~*]"
	want := "error parsing regexp: missing argument to repetition operator: `*`"
	t.Run("build query", func(t *testing.T) {
		_, err := buildQuery(&search.TextPatternInfo{Pattern: pattern}, nil, nil, false)
		if diff := cmp.Diff(err.Error(), want); diff != "" {
			t.Error(diff)
		}
	})
}

func Test_chunkRanges(t *testing.T) {
	cases := []struct {
		ranges         []protocol.Range
		mergeThreshold int32
		output         []rangeChunk
	}{{
		// Single range
		ranges: []protocol.Range{{
			Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Location{Offset: 20, Line: 1, Column: 10},
		}},
		mergeThreshold: 0,
		output: []rangeChunk{{
			cover: protocol.Range{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 20, Line: 1, Column: 10},
			},
			ranges: []protocol.Range{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 20, Line: 1, Column: 10},
			}},
		}},
	}, {
		// Overlapping ranges
		ranges: []protocol.Range{{
			Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Location{Offset: 20, Line: 1, Column: 10},
		}, {
			Start: protocol.Location{Offset: 5, Line: 0, Column: 5},
			End:   protocol.Location{Offset: 25, Line: 1, Column: 15},
		}},
		mergeThreshold: 0,
		output: []rangeChunk{{
			cover: protocol.Range{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 25, Line: 1, Column: 15},
			},
			ranges: []protocol.Range{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 20, Line: 1, Column: 10},
			}, {
				Start: protocol.Location{Offset: 5, Line: 0, Column: 5},
				End:   protocol.Location{Offset: 25, Line: 1, Column: 15},
			}},
		}},
	}, {
		// Non-overlapping ranges, but share a line
		ranges: []protocol.Range{{
			Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Location{Offset: 20, Line: 1, Column: 10},
		}, {
			Start: protocol.Location{Offset: 25, Line: 1, Column: 15},
			End:   protocol.Location{Offset: 35, Line: 2, Column: 5},
		}},
		mergeThreshold: 0,
		output: []rangeChunk{{
			cover: protocol.Range{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 35, Line: 2, Column: 5},
			},
			ranges: []protocol.Range{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 20, Line: 1, Column: 10},
			}, {
				Start: protocol.Location{Offset: 25, Line: 1, Column: 15},
				End:   protocol.Location{Offset: 35, Line: 2, Column: 5},
			}},
		}},
	}, {
		// Ranges on adjacent lines, but not merged because of low merge threshold
		ranges: []protocol.Range{{
			Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Location{Offset: 10, Line: 0, Column: 10},
		}, {
			Start: protocol.Location{Offset: 11, Line: 1, Column: 0},
			End:   protocol.Location{Offset: 20, Line: 1, Column: 9},
		}},
		mergeThreshold: 0,
		output: []rangeChunk{{
			cover: protocol.Range{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 10, Line: 0, Column: 10},
			},
			ranges: []protocol.Range{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 10, Line: 0, Column: 10},
			}},
		}, {
			cover: protocol.Range{
				Start: protocol.Location{Offset: 11, Line: 1, Column: 0},
				End:   protocol.Location{Offset: 20, Line: 1, Column: 9},
			},
			ranges: []protocol.Range{{
				Start: protocol.Location{Offset: 11, Line: 1, Column: 0},
				End:   protocol.Location{Offset: 20, Line: 1, Column: 9},
			}},
		}},
	}, {
		// Ranges on adjacent lines, merged because of high merge threshold
		ranges: []protocol.Range{{
			Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Location{Offset: 10, Line: 0, Column: 10},
		}, {
			Start: protocol.Location{Offset: 11, Line: 1, Column: 0},
			End:   protocol.Location{Offset: 20, Line: 1, Column: 9},
		}},
		mergeThreshold: 1,
		output: []rangeChunk{{
			cover: protocol.Range{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 20, Line: 1, Column: 9},
			},
			ranges: []protocol.Range{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 10, Line: 0, Column: 10},
			}, {
				Start: protocol.Location{Offset: 11, Line: 1, Column: 0},
				End:   protocol.Location{Offset: 20, Line: 1, Column: 9},
			}},
		}},
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := chunkRanges(tc.ranges, tc.mergeThreshold)
			require.Equal(t, tc.output, got)
		})
	}
}

func TestTarInput(t *testing.T) {
	maybeSkipComby(t)

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

	p := &protocol.PatternInfo{Pattern: "{:[body]}"}

	tarInputEventC := make(chan comby.TarInputEvent, 1)
	hdr := tar.Header{
		Name: "main.go",
		Mode: 0600,
		Size: int64(len(input["main.go"])),
	}
	tarInputEventC <- comby.TarInputEvent{
		Header:  hdr,
		Content: []byte(input["main.go"]),
	}
	close(tarInputEventC)

	t.Run("Structural search tar input to comby", func(t *testing.T) {
		ctx, cancel, sender := newLimitedStreamCollector(context.Background(), 1000000000)
		defer cancel()
		err := structuralSearch(ctx, logtest.Scoped(t), comby.Tar{TarInputEventC: tarInputEventC}, all, "", p.Pattern, p.CombyRule, p.Languages, "repo_foo", 0, sender)
		if err != nil {
			t.Fatal(err)
		}
		matches := sender.collected
		expected := []protocol.FileMatch{{
			Path: "main.go",
			ChunkMatches: []protocol.ChunkMatch{{
				Content:      "func foo() {\n    fmt.Println(\"foo\")\n}",
				ContentStart: protocol.Location{Offset: 1, Line: 1},
				Ranges: []protocol.Range{{
					Start: protocol.Location{Offset: 12, Line: 1, Column: 11},
					End:   protocol.Location{Offset: 38, Line: 3, Column: 1},
				}},
			}, {
				Content:      "func bar() {\n    fmt.Println(\"bar\")\n}",
				ContentStart: protocol.Location{Offset: 40, Line: 5},
				Ranges: []protocol.Range{{
					Start: protocol.Location{Offset: 51, Line: 5, Column: 11},
					End:   protocol.Location{Offset: 77, Line: 7, Column: 1},
				}},
			}},
		}}
		require.Equal(t, expected, matches)
	})
}

func maybeSkipComby(t *testing.T) {
	t.Helper()
	if os.Getenv("CI") != "" {
		return
	}
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		t.Skip("Skipping due to limitations in comby and M1")
	}
	if _, err := exec.LookPath("comby"); err != nil {
		t.Skipf("skipping comby test when not on CI: %v", err)
	}
}

func Test_addContext(t *testing.T) {
	l := func(offset, line, column int32) protocol.Location {
		return protocol.Location{Offset: offset, Line: line, Column: column}
	}

	r := func(start, end protocol.Location) protocol.Range {
		return protocol.Range{Start: start, End: end}
	}

	testCases := []struct {
		file         string
		contextLines int32
		inputRange   protocol.Range
		expected     string
	}{
		{
			"",
			0,
			r(l(0, 0, 0), l(0, 0, 0)),
			"",
		},
		{
			"",
			1,
			r(l(0, 0, 0), l(0, 0, 0)),
			"",
		},
		{
			"\n",
			0,
			r(l(0, 0, 0), l(0, 0, 0)),
			"",
		},
		{
			"\n",
			1,
			r(l(0, 0, 0), l(0, 0, 0)),
			"",
		},
		{
			"\n\n\n",
			0,
			r(l(1, 1, 0), l(1, 1, 0)),
			"",
		},
		{
			"\n\n\n\n",
			1,
			r(l(1, 1, 0), l(1, 1, 0)),
			"\n\n",
		},
		{
			"\n\n\n\n",
			2,
			r(l(1, 1, 0), l(1, 1, 0)),
			"\n\n\n",
		},
		{
			"abc\ndef\nghi\n",
			0,
			r(l(1, 0, 1), l(1, 0, 1)),
			"abc",
		},
		{
			"abc\ndef\nghi\n",
			1,
			r(l(1, 0, 1), l(1, 0, 1)),
			"abc\ndef",
		},
		{
			"abc\ndef\nghi\n",
			2,
			r(l(1, 0, 1), l(1, 0, 1)),
			"abc\ndef\nghi",
		},
		{
			"abc\ndef\nghi",
			0,
			r(l(1, 0, 1), l(1, 0, 1)),
			"abc",
		},
		{
			"abc\ndef\nghi",
			1,
			r(l(1, 0, 1), l(1, 0, 1)),
			"abc\ndef",
		},
		{
			"abc\ndef\nghi",
			2,
			r(l(1, 0, 1), l(1, 0, 1)),
			"abc\ndef\nghi",
		},
		{
			"abc\ndef\nghi",
			2,
			r(l(5, 1, 1), l(6, 1, 2)),
			"abc\ndef\nghi",
		},
		{
			"abc",
			0,
			r(l(1, 0, 1), l(2, 0, 2)),
			"abc",
		},
		{
			"abc",
			1,
			r(l(1, 0, 1), l(2, 0, 2)),
			"abc",
		},
		{
			"abc\r\ndef\r\nghi\r\n",
			1,
			r(l(1, 0, 1), l(2, 0, 2)),
			"abc\r\ndef",
		},
		{
			"abc\r\ndef\r\nghi",
			3,
			r(l(1, 0, 1), l(2, 0, 2)),
			"abc\r\ndef\r\nghi",
		},
		{
			"\r\n",
			0,
			r(l(0, 0, 0), l(0, 0, 0)),
			"",
		},
		{
			"\r\n",
			1,
			r(l(0, 0, 0), l(0, 0, 0)),
			"",
		},
		{
			"abc\nd\xE2\x9D\x89f\nghi",
			0,
			r(l(4, 1, 0), l(5, 1, 1)),
			"d\xE2\x9D\x89f",
		},
		{
			"abc\nd\xE2\x9D\x89f\nghi",
			1,
			r(l(4, 1, 0), l(5, 1, 1)),
			"abc\nd\xE2\x9D\x89f\nghi",
		},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			buf := []byte(testCase.file)
			extendedRange := extendRangeToLines(testCase.inputRange, buf)
			contextedRange := addContextLines(extendedRange, buf, testCase.contextLines)
			require.Equal(t, testCase.expected, string(buf[contextedRange.Start.Offset:contextedRange.End.Offset]))
		})
	}
}
