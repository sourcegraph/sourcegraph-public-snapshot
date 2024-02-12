package search

import (
	"testing"
	"testing/quick"
)

func TestTruncateLines(t *testing.T) {
	type testCase struct {
		name       string
		content    string
		maxLineLen int
		want       string
	}

	testCases := []testCase{{
		name:       "empty",
		content:    "",
		maxLineLen: 10,
	}, {
		name:       "no truncation",
		content:    "line 1\nline 2",
		maxLineLen: 10,
	}, {
		name:       "no suffix",
		content:    "line 1\nline 2",
		maxLineLen: 4,
		want:       "line\nline",
	}, {
		name:       "trunc 1st",
		content:    "line 1 is too long to show\nline 2\nline 3",
		maxLineLen: 15,
		want:       "lin...truncated\nline 2\nline 3",
	}, {
		name:       "trunc 2st",
		content:    "line 1\nline 2 is too long to show\nline 3",
		maxLineLen: 15,
		want:       "line 1\nlin...truncated\nline 3",
	}, {
		name:       "trunc 3rd",
		content:    "line 1\nline 2\nline 3 is too long to show",
		maxLineLen: 15,
		want:       "line 1\nline 2\nlin...truncated",
	}, {
		name:       "trunc 1st and third",
		content:    "line 1 is too long to show\nline 2\nline 3 is too long to show",
		maxLineLen: 15,
		want:       "lin...truncated\nline 2\nlin...truncated",
	}}

	// For each test case we ensure that if we have a trailing nl it is still
	// treated well.
	for _, tc := range testCases {
		tcnl := testCase{
			name:       tc.name + " nl",
			content:    tc.content + "\n",
			maxLineLen: tc.maxLineLen,
		}
		if tc.want != "" {
			tcnl.want = tc.want + "\n"
		}
		testCases = append(testCases, tcnl)
	}

	for _, tc := range testCases {
		want := tc.want
		wantTruncated := true
		if want == "" {
			want = tc.content
			wantTruncated = false
		}

		got, truncated := truncateLines(tc.content, tc.maxLineLen)
		if got != want || truncated != wantTruncated {
			t.Errorf("%s: got %q %v, want %q %v", tc.name, got, truncated, want, wantTruncated)
		}
	}
}

func TestTruncateLines_disabled(t *testing.T) {
	fn := func(content string) bool {
		for _, maxLineLen := range []int{0, -1, -10} {
			got, gotTrunc := truncateLines(content, maxLineLen)
			if got != content {
				t.Logf("got content %q want %q", got, content)
				return false
			}
			if gotTrunc {
				t.Logf("truncated for %q", content)
				return false
			}
		}
		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Fatal(err)
	}
}
