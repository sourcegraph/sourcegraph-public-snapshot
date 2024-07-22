package graphqlbackend

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	godiff "github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestPreviewRepositoryComparisonResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, nil)

	const testDiffFiles = 3
	const testOldFile = `First
Second
Line 1
Line 2
Line 3
Line 4
Line 5
Line 6
Line 7
Line 8
Line 9
Line 10
`

	var testDiff = []byte(`diff --git INSTALL.md INSTALL.md
index e5af166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobar Line 8
 Line 9
 Line 10
diff --git JOKES.md JOKES.md
index ea80abf..1b86505 100644
--- JOKES.md
+++ JOKES.md
@@ -4,10 +4,10 @@ Joke #1
 Joke #2
 Joke #3
 Joke #4
-Joke #5
+This is not funny: Joke #5
 Joke #6
-Joke #7
+This one is good: Joke #7
 Joke #8
-Joke #9
+Waffle: Joke #9
 Joke #10
 Joke #11
diff --git README.md README.md
index 9bd8209..d2acfa9 100644
--- README.md
+++ README.md
@@ -1,12 +1,13 @@
 # README

-Line 1
+Foobar Line 1
 Line 2
 Line 3
 Line 4
 Line 5
-Line 6
+Barfoo Line 6
 Line 7
 Line 8
 Line 9
 Line 10
+Another line
`)

	wantBaseRef := "refs/heads/master"
	wantHeadRevision := api.CommitID("b69072d5f687b31b9f6ae3ceafdc24c259c4b9ec")

	repo := &types.Repo{ID: api.RepoID(1), Name: "github.com/sourcegraph/sourcegraph", CreatedAt: time.Now()}

	t.Run("EmptyCommit", func(t *testing.T) {
		gitserverClient := gitserver.NewMockClient()
		gitserverClient.ResolveRevisionFunc.SetDefaultReturn("", &gitdomain.RevisionNotFoundError{})
		_, err := NewPreviewRepositoryComparisonResolver(ctx, db, gitserverClient, NewRepositoryResolver(db, gitserverClient, repo), string(wantHeadRevision), testDiff)
		if err == nil {
			t.Fatal("unexpected empty err")
		}
		if !errors.HasType[*gitdomain.RevisionNotFoundError](err) {
			t.Fatalf("incorrect err returned %T", err)
		}
	})

	gitserverClient := gitserver.NewMockClient()

	byRev := map[api.CommitID]struct{}{
		api.CommitID(wantBaseRef): {},
		wantHeadRevision:          {},
	}

	gitserverClient.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, rev string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		if _, ok := byRev[api.CommitID(rev)]; !ok {
			t.Fatalf("ResolveRev received unexpected rev: %q", rev)
		}
		return api.CommitID(rev), nil
	})

	previewComparisonResolver, err := NewPreviewRepositoryComparisonResolver(ctx, db, gitserverClient, NewRepositoryResolver(db, gitserverClient, repo), string(wantHeadRevision), testDiff)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Pagination", func(t *testing.T) {
		endCursors := []string{"1", "2"}
		totalCount := int32(testDiffFiles)

		tests := []struct {
			first int32
			after string

			wantNodeCount int

			wantTotalCount *int32

			wantHasNextPage bool
			wantEndCursor   *string
		}{
			{
				first:           1,
				after:           "",
				wantNodeCount:   1,
				wantHasNextPage: true,
				wantEndCursor:   &endCursors[0],
				wantTotalCount:  nil,
			},
			{
				first:           1,
				after:           endCursors[0],
				wantNodeCount:   1,
				wantHasNextPage: true,
				wantEndCursor:   &endCursors[1],
				wantTotalCount:  nil,
			},
			{
				first:           1,
				after:           endCursors[1],
				wantNodeCount:   1,
				wantHasNextPage: false,
				wantEndCursor:   nil,
				wantTotalCount:  &totalCount,
			},
			{
				first:           testDiffFiles + 1,
				after:           "",
				wantNodeCount:   testDiffFiles,
				wantHasNextPage: false,
				wantEndCursor:   nil,
				wantTotalCount:  &totalCount,
			},
		}

		for _, tc := range tests {
			args := &FileDiffsConnectionArgs{First: &tc.first}
			if tc.after != "" {
				args.After = &tc.after
			}

			conn, err := previewComparisonResolver.FileDiffs(ctx, args)
			if err != nil {
				t.Fatal(err)
			}

			nodes, err := conn.Nodes(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if len(nodes) != tc.wantNodeCount {
				t.Fatalf("wrong length of nodes. want=%d, have=%d", tc.wantNodeCount, len(nodes))
			}

			pageInfo, err := conn.PageInfo(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if pageInfo.HasNextPage() != tc.wantHasNextPage {
				t.Fatalf("pageInfo HasNextPage wrong. want=%t, have=%t", tc.wantHasNextPage, pageInfo.HasNextPage())
			}

			if diff := cmp.Diff(tc.wantEndCursor, pageInfo.EndCursor()); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}

			totalCount, err := conn.TotalCount(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.wantTotalCount, totalCount); diff != "" {
				t.Fatalf("wrong totalCount: %s", diff)
			}
		}
	})

	t.Run("NewFile resolver", func(t *testing.T) {
		fileDiffConnection, err := previewComparisonResolver.FileDiffs(ctx, &FileDiffsConnectionArgs{})
		if err != nil {
			t.Fatal(err)
		}
		fileDiffs, err := fileDiffConnection.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(fileDiffs), testDiffFiles; have != want {
			t.Fatalf("invalid len(FileDiffs.Nodes). want=%d have=%d", want, len(fileDiffs))
		}
		fileDiff := fileDiffs[0]

		gitserverClient.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, name string) (io.ReadCloser, error) {
			if name != "INSTALL.md" {
				t.Fatalf("ReadFile received call for wrong file: %s", name)
			}
			return io.NopCloser(bytes.NewReader([]byte(testOldFile))), nil
		})

		newFile := fileDiff.NewFile()
		if newFile == nil {
			t.Fatal("NewFile is null")
		}

		wantNewFileContent := `First
Second
Line 1
Line 2
Line 3
This is cool: Line 4
Line 5
Line 6
Another Line 7
Foobar Line 8
Line 9
Line 10
`

		haveContent, err := newFile.Content(ctx, &GitTreeContentPageArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if haveContent != wantNewFileContent {
			t.Fatalf("wrong file content. want=%q have=%q", wantNewFileContent, haveContent)
		}
	})
}

func TestApplyPatch(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		patch    string
		wantFile string
	}{
		{
			name: "replace in middle",
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
18 oh yes
`,
			patch: `diff --git a/test b/test
index 38dea4a..d81676e 100644
--- a/test
+++ b/test
@@ -4,7 +4,7 @@
 4
 5
 6
-7 super awesome
+7 super mega awesome
 8
 9
 10
`,
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
18 oh yes
`,
		},
		{
			name: "delete file",
			file: `1 some
2
3
`,
			patch: `diff --git a/test b/test
deleted file mode 100644
index 2e0cf96..0000000
--- a/test
+++ /dev/null
@@ -1,3 +0,0 @@
-1 some
-2
-3
`,
			wantFile: "",
		},
		{
			name: "New file, additional newline at end",
			file: "",
			patch: `diff --git a/file2.txt b/file2.txt
new file mode 100644
index 0000000..122f5d9
--- /dev/null
+++ b/file2.txt
@@ -0,0 +1 @@
+filecontent
+
`,
			wantFile: `filecontent

`,
		},
		{
			name: "New file",
			file: "",
			patch: `diff --git a/file2.txt b/file2.txt
new file mode 100644
index 0000000..122f5d9
--- /dev/null
+++ b/file2.txt
@@ -0,0 +1 @@
+filecontent
`,
			wantFile: `filecontent
`,
		},
		{
			name: "New file without newline",
			file: "",
			patch: `diff --git a/README.md b/README.md
new file mode 100644
index 0000000..373ae20
--- /dev/null
+++ b/README.md
@@ -0,0 +1 @@
+No newline after this
\ No newline at end of file
`,
			// Note: No newline.
			wantFile: `No newline after this`,
		},
		{
			name: "Add newline to file without newline",
			// Note: No newline.
			file: `No newline after this`,
			patch: `diff --git a/README.md b/README.md
index 373ae20..7e17295 100644
--- a/README.md
+++ b/README.md
@@ -1 +1 @@
-No newline after this
\ No newline at end of file
+No newline after this
`,
			// Note: Has a newline now.
			wantFile: `No newline after this
`,
		},
		{
			name: "Remove newline at end of file",
			file: `No newline after this
`,
			patch: `diff --git a/README.md b/README.md
index 7e17295..373ae20 100644
--- a/README.md
+++ b/README.md
@@ -1 +1 @@
-No newline after this
+No newline after this
\ No newline at end of file
`,
			// Note: Has no newline anymore.
			wantFile: `No newline after this`,
		},
		{
			name: "Add line without newline to file that ended with no newline",
			file: `No newline after this`,
			patch: `diff --git a/README.md b/README.md
index 373ae20..89ad131 100644
--- a/README.md
+++ b/README.md
@@ -1 +1,2 @@
-No newline after this
\ No newline at end of file
+No newline after this
+Also no newline after this
\ No newline at end of file
`,
			// Note: Has no newline at the end.
			wantFile: `No newline after this
Also no newline after this`,
		},
		{
			name: "Add line without newline to file that ended with no newline",
			file: `No newline after this`,
			patch: `diff --git a/README.md b/README.md
index 373ae20..89ad131 100644
--- a/README.md
+++ b/README.md
@@ -1 +1,2 @@
-No newline after this
\ No newline at end of file
+No newline after this
+Also no newline after this
\ No newline at end of file
`,
			// Note: Has no newline at the end.
			wantFile: `No newline after this
Also no newline after this`,
		},
		{
			name: "No newline and last hunk ends before EOF",
			file: `1
3
4
5
6
7
8
9
10`,
			patch: `diff --git a/README.md b/README.md
index 373ae20..89ad131 100644
--- a/README.md
+++ b/README.md
@@ -1,4 +1,5 @@
 1
+2
 3
 4
 5
`,
			// Note: Has no newline at the end.
			wantFile: `1
2
3
4
5
6
7
8
9
10`,
		},
		{
			name: "Multiple hunks and no newline at the end",
			file: `1
3
4
5
6
7
8
9
10
11
12`,
			patch: `diff --git a/README.md b/README.md
index 373ae20..89ad131 100644
--- a/README.md
+++ b/README.md
@@ -1,4 +1,5 @@
 1
+2
 3
 4
 5
@@ -6,6 +7,7 @@
 7
 8
 9
+9.5
 10
 11
 12
\ No newline at end of file
`,
			// Note: Has no newline at the end.
			wantFile: `1
2
3
4
5
6
7
8
9
9.5
10
11
12`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fileDiff, err := godiff.ParseFileDiff([]byte(tc.patch))
			if err != nil {
				t.Fatal(err)
			}
			have, err := applyPatch(tc.file, fileDiff)
			if err != nil {
				t.Fatal(err)
			}
			if have != tc.wantFile {
				t.Fatalf("wrong patched file content %q, want=%q", have, tc.wantFile)
			}
		})
	}
}
