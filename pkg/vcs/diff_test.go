package vcs_test

import (
	"bytes"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestRepository_Diff(t *testing.T) {
	t.Parallel()

	// TODO(sqs): test ExcludeReachableFromBoth

	gitCommands := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testhead",
	}
	tests := map[string]struct {
		repo       vcs.Repository
		base, head string // can be any revspec; is resolved during the test
		opt        *vcs.DiffOptions

		// wantDiff is the expected diff. In the Raw field,
		// %(headCommitID) is replaced with the actual head commit ID
		// (it seems to change in hg).
		wantDiff *vcs.Diff
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			base: "testbase", head: "testhead",
			wantDiff: &vcs.Diff{
				Raw: "diff --git f f\nindex a29bdeb434d874c9b1d8969c40c42161b03fafdc..c0d0fb45c382919737f8d0c20aaf57cf89b74af8 100644\n--- f\n+++ f\n@@ -1 +1,2 @@\n line1\n+line2\n",
			},
		},
	}

	// TODO(sqs): implement diff for hg native

	for label, test := range tests {
		baseCommitID, err := test.repo.ResolveRevision(ctx, test.base)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.base, err)
			continue
		}

		headCommitID, err := test.repo.ResolveRevision(ctx, test.head)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on head: %s", label, test.head, err)
			continue
		}

		diff, err := test.repo.Diff(ctx, baseCommitID, headCommitID, test.opt)
		if err != nil {
			t.Errorf("%s: Diff(%s, %s, %v): %s", label, baseCommitID, headCommitID, test.opt, err)
			continue
		}

		// Substitute for easier test expectation definition. See the
		// wantDiff field doc for more info.
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(baseCommitID)", string(baseCommitID), -1)
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(headCommitID)", string(headCommitID), -1)
		if runtime.GOOS == "windows" {
			test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "/dev/null", `\dev\null`, -1)
		}

		if !reflect.DeepEqual(diff, test.wantDiff) {
			t.Errorf("%s: diff != wantDiff\n\ndiff ==========\n%s\n\nwantDiff ==========\n%s", label, asJSON(diff), asJSON(test.wantDiff))
		}

		if _, err := test.repo.Diff(ctx, nonexistentCommitID, headCommitID, test.opt); err != vcs.ErrRevisionNotFound {
			t.Errorf("%s: Diff with bad base commit ID: want ErrRevisionNotFound, got %v", label, err)
			continue
		}

		if _, err := test.repo.Diff(ctx, baseCommitID, nonexistentCommitID, test.opt); err != vcs.ErrRevisionNotFound {
			t.Errorf("%s: Diff with bad head commit ID: want ErrRevisionNotFound, got %v", label, err)
			continue
		}
	}
}

func TestRepository_Diff_rename(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"git mv f g",
		"git add g",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testhead",
	}
	opt := &vcs.DiffOptions{DetectRenames: true}
	tests := map[string]struct {
		repo       vcs.Repository
		base, head string // can be any revspec; is resolved during the test
		opt        *vcs.DiffOptions

		// wantDiff is the expected diff. In the Raw field,
		// %(headCommitID) is replaced with the actual head commit ID
		// (it seems to change in hg).
		wantDiff *vcs.Diff
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			base: "testbase", head: "testhead",
			wantDiff: &vcs.Diff{
				Raw: "diff --git f g\nsimilarity index 100%\nrename from f\nrename to g\n",
			},
			opt: opt,
		},
	}

	// TODO(sqs): implement diff for hg native

	for label, test := range tests {
		baseCommitID, err := test.repo.ResolveRevision(ctx, test.base)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.base, err)
			continue
		}

		headCommitID, err := test.repo.ResolveRevision(ctx, test.head)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on head: %s", label, test.head, err)
			continue
		}

		diff, err := test.repo.Diff(ctx, baseCommitID, headCommitID, test.opt)
		if err != nil {
			t.Errorf("%s: Diff(%s, %s, %v): %s", label, baseCommitID, headCommitID, test.opt, err)
			continue
		}

		// Substitute for easier test expectation definition. See the
		// wantDiff field doc for more info.
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(baseCommitID)", string(baseCommitID), -1)
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(headCommitID)", string(headCommitID), -1)
		if runtime.GOOS == "windows" {
			test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "/dev/null", `\dev\null`, -1)
		}

		if !reflect.DeepEqual(diff, test.wantDiff) {
			t.Errorf("%s: diff != wantDiff\n\ndiff ==========\n%s\n\nwantDiff ==========\n%s", label, asJSON(diff), asJSON(test.wantDiff))
		}

		if _, err := test.repo.Diff(ctx, nonexistentCommitID, headCommitID, test.opt); err != vcs.ErrRevisionNotFound {
			t.Errorf("%s: Diff with bad base commit ID: want ErrRevisionNotFound, got %v", label, err)
			continue
		}

		if _, err := test.repo.Diff(ctx, baseCommitID, nonexistentCommitID, test.opt); err != vcs.ErrRevisionNotFound {
			t.Errorf("%s: Diff with bad head commit ID: want ErrRevisionNotFound, got %v", label, err)
			continue
		}
	}
}

func TestFilterAndHighlightDiff(t *testing.T) {
	const sampleRawDiff = `diff --git f f
index a29bdeb434d874c9b1d8969c40c42161b03fafdc..c0d0fb45c382919737f8d0c20aaf57cf89b74af8 100644
--- f
+++ f
@@ -1,1 +1,2 @@
 line1
+line2
`
	tests := map[string]struct {
		rawDiff        string
		query          string
		paths          vcs.PathOptions
		want           string
		wantHighlights []vcs.Highlight
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
			wantHighlights: []vcs.Highlight{
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
			wantHighlights: []vcs.Highlight{{Line: 7, Character: 1, Length: 5}},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			query, err := regexp.Compile(test.query)
			if err != nil {
				t.Fatal(err)
			}
			pathMatcher, err := vcs.CompilePathMatcher(test.paths)
			if err != nil {
				t.Fatal(err)
			}
			rawDiff, highlights, err := vcs.FilterAndHighlightDiff([]byte(test.rawDiff), query, true, pathMatcher)
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
			gotHunks := vcs.SplitHunkMatches(hunks, query, test.matchContextLines, test.maxLinesPerHunk)
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
