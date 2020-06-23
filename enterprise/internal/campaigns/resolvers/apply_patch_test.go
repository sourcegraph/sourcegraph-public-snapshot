package resolvers

import (
	"testing"

	"github.com/sourcegraph/go-diff/diff"
)

func TestApplyPatch(t *testing.T) {
	tests := []struct {
		file          string
		patch         string
		origStartLine int32
		wantFile      string
	}{
		{
			file: `1 some
2
3
4
5
6
7 super awesome
8
9
10
11
12
13
14 file
15
16
17
18 oh yes`,
			patch: ` 4
 5
 6
-7 super awesome
+7 super mega awesome
 8
 9
 10
`,
			origStartLine: 4,
			wantFile: `1 some
2
3
4
5
6
7 super mega awesome
8
9
10
11
12
13
14 file
15
16
17
18 oh yes`,
		},
	}

	for _, tc := range tests {
		have := applyPatch(tc.file, &diff.FileDiff{Hunks: []*diff.Hunk{{OrigStartLine: tc.origStartLine, Body: []byte(tc.patch)}}})
		if have != tc.wantFile {
			t.Fatalf("wrong patched file content %q, want=%q", have, tc.wantFile)
		}
	}
}
