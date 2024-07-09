package byteutils

import (
	"bytes"
	"strings"
	"testing"
	"testing/quick"
)

func naiveGetLines(contents string, lineStart, lineEnd int) string {
	lines := strings.SplitAfter(contents, "\n")
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	clampedStart := min(max(0, lineStart), len(lines))
	clampedEnd := min(max(clampedStart, lineEnd), len(lines))
	return strings.Join(lines[clampedStart:clampedEnd], "")
}

var testCases = []struct {
	contents           string
	startLine, endLine int
}{
	{"no trailing newline", 0, 1},
	{"trailing newline\n", 0, 1},
	{"trailing newline\nfollowed by no trailing newline", 0, 2},
	{"", 0, 0},
	{"\n", 0, 1},
	{"\n\n\n", 0, 3},

	// Out of bounds
	{"\n\n\n", -1, 4},
	{"\n\n\n", -1, -1},
	{"\n\n\n", 4, 4},
}

func TestNewlineIndex(t *testing.T) {
	lineIndexGetLines := func(contents string, startLine, endLine int) string {
		index := NewLineIndex(contents)
		start, end := index.LinesRange(startLine, endLine)
		return contents[start:end]
	}

	t.Run("cases", func(t *testing.T) {
		for _, tc := range testCases {
			got := lineIndexGetLines(tc.contents, tc.startLine, tc.endLine)
			want := naiveGetLines(tc.contents, tc.startLine, tc.endLine)
			if want != got {
				t.Log(tc)
				t.Fatalf("got: %q, want: %q", got, want)
			}
		}
	})

	t.Run("quick", func(t *testing.T) {
		quick.CheckEqual(lineIndexGetLines, naiveGetLines, nil)
	})

	t.Run("line count", func(t *testing.T) {
		cases := []struct {
			content   string
			lineCount int
		}{
			{"", 0},
			{"test", 1},
			{"test\n", 1},
			{"test\ntest", 2},
			{"test\ntest\n", 2},
			{"\n", 1},
			{"\n\n", 2},
		}

		for _, tc := range cases {
			index := NewLineIndex(tc.content)
			if index.LineCount() != tc.lineCount {
				t.Fatalf("got %q, want %q", index.LineCount(), tc.lineCount)
			}
		}
	})

	t.Run("string allocs", func(t *testing.T) {
		contents := strings.Repeat("testline\n", 1000)
		allocs := testing.AllocsPerRun(10, func() {
			_ = NewLineIndex(contents)
		})
		if allocs != 1 {
			t.Fatalf("expected one alloc got %f", allocs)
		}
	})

	t.Run("byte allocs", func(t *testing.T) {
		contents := bytes.Repeat([]byte("testline\n"), 1000)
		allocs := testing.AllocsPerRun(10, func() {
			_ = NewLineIndex(contents)
		})
		if allocs != 1 {
			t.Fatalf("expected one alloc, got %f", allocs)
		}
	})
}

func FuzzNewlineIndex(f *testing.F) {
	for _, tc := range testCases {
		f.Add(tc.contents, tc.startLine, tc.endLine)
	}
	f.Fuzz(func(t *testing.T, contents string, startLine, endLine int) {
		index := NewLineIndex(contents)
		start, end := index.LinesRange(startLine, endLine)
		got := contents[start:end]
		want := naiveGetLines(contents, startLine, endLine)
		if want != got {
			t.Fatalf("got: %q, want: %q", got, want)
		}
	})
}

func BenchmarkLineIndex(b *testing.B) {
	b.Run("construct string", func(b *testing.B) {
		contents := strings.Repeat("testline\n", 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewLineIndex(contents)
		}
	})

	b.Run("construct bytes", func(b *testing.B) {
		contents := bytes.Repeat([]byte("testline\n"), 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewLineIndex(contents)
		}
	})
}
