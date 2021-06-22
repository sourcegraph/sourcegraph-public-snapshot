package git

import (
	"bytes"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/sourcegraph/go-diff/diff"
)

func TestFilterAndHighlightDiff(t *testing.T) {
	const sampleRawDiff = `diff --git f f
index a29bdeb434d874c9b1d8969c40c42161b03fafdc..c0d0fb45c382919737f8d0c20aaf57cf89b74af8 100644
--- f
+++ f
@@ -1,1 +1,2 @@
 line1
+line2
`

	const sampleUnicodeDiff = `diff --git f f
index a29bdeb434d874c9b1d8969c40c42161b03fafdc..c0d0fb45c382919737f8d0c20aaf57cf89b74af8 100644
--- f
+++ f
@@ -1,1 +1,2 @@
 line1
+before › after
`
	tests := map[string]struct {
		rawDiff        string
		query          string
		paths          PathOptions
		want           string
		wantHighlights []Highlight
	}{
		"no matches": {
			rawDiff:        sampleRawDiff,
			query:          "line3",
			want:           "",
			wantHighlights: nil,
		},
		"changed line and context line match": {
			rawDiff: sampleRawDiff,
			query:   "line",
			want:    sampleRawDiff,
			wantHighlights: []Highlight{
				{Line: 6, Character: 1, Length: 4},
				{Line: 7, Character: 1, Length: 4},
			},
		},
		"only context line matches": {
			rawDiff:        sampleRawDiff,
			query:          "line1",
			want:           "",
			wantHighlights: nil,
		},
		"only changed line matches": {
			rawDiff:        sampleRawDiff,
			query:          "line2",
			want:           sampleRawDiff,
			wantHighlights: []Highlight{{Line: 7, Character: 1, Length: 5}},
		},
		"multi-byte character before": {
			rawDiff:        sampleUnicodeDiff,
			query:          "before",
			want:           sampleUnicodeDiff,
			wantHighlights: []Highlight{{Line: 7, Character: 1, Length: 6}},
		},
		// https://github.com/sourcegraph/sourcegraph/issues/22066
		"multi-byte character after": {
			rawDiff:        sampleUnicodeDiff,
			query:          "after",
			want:           sampleUnicodeDiff,
			wantHighlights: []Highlight{{Line: 7, Character: 10, Length: 5}},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			query, err := regexp.Compile(test.query)
			if err != nil {
				t.Fatal(err)
			}
			pathMatcher, err := compilePathMatcher(test.paths)
			if err != nil {
				t.Fatal(err)
			}
			rawDiff, highlights, err := filterAndHighlightDiff([]byte(test.rawDiff), query, true, pathMatcher)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(string(rawDiff), test.want) {
				t.Errorf("got diff %q, want %q", rawDiff, test.want)
			}
			if !reflect.DeepEqual(highlights, test.wantHighlights) {
				t.Errorf("got highlights %v, want %v", highlights, test.wantHighlights)
			}
		})
	}
}

func TestSplitHunkMatches(t *testing.T) {
	tests := []struct {
		hunks             string
		query             string
		matchContextLines int
		maxLinesPerHunk   int
		want              string
	}{
		{
			hunks: `@@ -1,2 +1,3 @@ mysection
 line1
+line2
 line3`,
			query: "doesntmatch",
			want:  ``,
		},

		// matchContextLines == 0
		{
			hunks: `@@ -1,2 +1,3 @@ mysection
 line1
+line2
 line3`,
			query: "line",
			want: `@@ -2,0 +2,1 @@ mysection
+line2`,
		},
		{
			hunks: `@@ -1,2 +1,3 @@ mysection
 line1
+line2
 line3
@@ -10,2 +11,2 @@ mysection2
 line10
+line11
-line12`,
			query: "line",
			want: `@@ -2,0 +2,1 @@ mysection
+line2
@@ -11,1 +12,1 @@ mysection2
+line11
-line12`,
		},
		{
			hunks: `@@ -1,1 +1,2 @@
+line1
+line2
-line3`,
			query: "line",
			want: `@@ -1,1 +1,2 @@
+line1
+line2
-line3`,
		},
		{
			hunks: `@@ -1,3 +1,2 @@
 line1
-line2
 line3`,
			query: "line2",
			want: `@@ -2,1 +2,0 @@
-line2`,
		},
		{
			hunks: `@@ -1,3 +1,3 @@
 line1
-line2
+line3
 line4`,
			query: "line[23]",
			want: `@@ -2,1 +2,1 @@
-line2
+line3`,
		},
		{
			hunks: `@@ -1,3 +1,4 @@ mysection
 line1
+line2
-line3
+line4
 line5`,
			query: "line[24]",
			want: `@@ -2,0 +2,1 @@ mysection
+line2
@@ -3,0 +3,1 @@ mysection
+line4`,
		},

		// matchContextLines >= 1
		{
			hunks: `@@ -1,2 +1,3 @@ mysection
 line1
+line2
 line3`,
			query:             "line2",
			matchContextLines: 1,
			want: `@@ -1,2 +1,3 @@ mysection
 line1
+line2
 line3`,
		},
		{
			hunks: `@@ -1,2 +1,3 @@ mysection
 line1
+line2
 line3`,
			query:             "line2",
			matchContextLines: 100,
			want: `@@ -1,2 +1,3 @@ mysection
 line1
+line2
 line3`,
		},
		{
			hunks: `@@ -1,3 +1,2 @@
 line1
-line2
 line3`,
			query:             "line2",
			matchContextLines: 1,
			want: `@@ -1,3 +1,2 @@
 line1
-line2
 line3`,
		},
		{
			hunks: `@@ -1,5 +1,5 @@
 line1
 line2
-line3
+line4
 line5
 line6`,
			query:             "line[34]",
			matchContextLines: 1,
			want: `@@ -2,3 +2,3 @@
 line2
-line3
+line4
 line5`,
		},
		{
			hunks: `@@ -1,5 +1,6 @@
 line1
 line2
+line3
-line4
+line5
 line6
 line7`,
			query:             "line[35]",
			matchContextLines: 1,
			want: `@@ -2,3 +2,4 @@
 line2
+line3
-line4
+line5
 line6`,
		},
		{
			hunks: `@@ -1,7 +1,8 @@
 line1
 line2
+line3
 line4
-line5
 line6
+line7
 line8
 line9`,
			query:             "line[37]",
			matchContextLines: 1,
			want: `@@ -2,2 +2,3 @@
 line2
+line3
 line4
@@ -5,2 +5,3 @@
 line6
+line7
 line8`,
		},

		// maxLinesPerHunk >= 1
		{
			hunks: `@@ -1,3 +1,3 @@ mysection
 line1
+line2
-line3
 line4`,
			query:           "line[23]",
			maxLinesPerHunk: 1,
			want: `@@ -2,0 +2,1 @@ mysection ... +1
+line2`,
		},

		// matchContextLines >= 1 && maxLinesPerHunk >= 1
		{
			hunks: `@@ -1,4 +1,4 @@ mysection
 line1
+line2
-line3
+line4
-line5
 line6`,
			query:             "line[2345]",
			matchContextLines: 1,
			maxLinesPerHunk:   1,
			want: `@@ -1,2 +1,2 @@ mysection ... +2
 line1
+line2
-line3`,
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			hunks, err := diff.ParseHunks([]byte(test.hunks))
			if err != nil {
				t.Fatal(err)
			}
			query, err := regexp.Compile(test.query)
			if err != nil {
				t.Fatal(err)
			}
			gotHunks := splitHunkMatches(hunks, query, test.matchContextLines, test.maxLinesPerHunk)
			got, err := diff.PrintHunks(gotHunks)
			if err != nil {
				t.Fatal(err)
			}
			got = bytes.TrimSpace(got)
			if string(got) != test.want {
				t.Errorf("hunks\ngot:\n%s\n\nwant:\n%s", got, test.want)
			}
		})
	}
}

func TestTruncateLongLines(t *testing.T) {
	const maxCharsPerLine = 5

	tests := map[string]string{
		"":           "",
		"1":          "1",
		"12345":      "12345",
		"123456":     "12345",
		"一二三四五六七八九十": "一二三四五",
		"一二三四五六七\n一二三四五六七": "一二三四五\n一二三四五",
		"😄😱👽😁😘😤😸":          "😄😱👽😁😘",
		"😄😱👽😁😘😤😸\n😄😱👽😁😘😤😸": "😄😱👽😁😘\n😄😱👽😁😘",
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got := truncateLongLines([]byte(input), maxCharsPerLine)
			if string(got) != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
