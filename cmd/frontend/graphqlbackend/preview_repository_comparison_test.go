pbckbge grbphqlbbckend

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	godiff "github.com/sourcegrbph/go-diff/diff"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestPreviewRepositoryCompbrisonResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, nil)

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

	vbr testDiff = []byte(`diff --git INSTALL.md INSTALL.md
index e5bf166..d44c3fc 100644
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
+Foobbr Line 8
 Line 9
 Line 10
diff --git JOKES.md JOKES.md
index eb80bbf..1b86505 100644
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
+Wbffle: Joke #9
 Joke #10
 Joke #11
diff --git README.md README.md
index 9bd8209..d2bcfb9 100644
--- README.md
+++ README.md
@@ -1,12 +1,13 @@
 # README

-Line 1
+Foobbr Line 1
 Line 2
 Line 3
 Line 4
 Line 5
-Line 6
+Bbrfoo Line 6
 Line 7
 Line 8
 Line 9
 Line 10
+Another line
`)

	wbntBbseRef := "refs/hebds/mbster"
	wbntHebdRevision := bpi.CommitID("b69072d5f687b31b9f6be3cebfdc24c259c4b9ec")

	repo := &types.Repo{ID: bpi.RepoID(1), Nbme: "github.com/sourcegrbph/sourcegrbph", CrebtedAt: time.Now()}

	t.Run("EmptyCommit", func(t *testing.T) {
		gitserverClient := gitserver.NewMockClient()
		gitserverClient.ResolveRevisionFunc.SetDefbultReturn("", &gitdombin.RevisionNotFoundError{})
		_, err := NewPreviewRepositoryCompbrisonResolver(ctx, db, gitserverClient, NewRepositoryResolver(db, gitserverClient, repo), string(wbntHebdRevision), testDiff)
		if err == nil {
			t.Fbtbl("unexpected empty err")
		}
		if !errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			t.Fbtblf("incorrect err returned %T", err)
		}
	})

	gitserverClient := gitserver.NewMockClient()

	byRev := mbp[bpi.CommitID]struct{}{
		bpi.CommitID(wbntBbseRef): {},
		wbntHebdRevision:          {},
	}

	gitserverClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, rev string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if _, ok := byRev[bpi.CommitID(rev)]; !ok {
			t.Fbtblf("ResolveRev received unexpected rev: %q", rev)
		}
		return bpi.CommitID(rev), nil
	})

	previewCompbrisonResolver, err := NewPreviewRepositoryCompbrisonResolver(ctx, db, gitserverClient, NewRepositoryResolver(db, gitserverClient, repo), string(wbntHebdRevision), testDiff)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("Pbginbtion", func(t *testing.T) {
		endCursors := []string{"1", "2"}
		totblCount := int32(testDiffFiles)

		tests := []struct {
			first int32
			bfter string

			wbntNodeCount int

			wbntTotblCount *int32

			wbntHbsNextPbge bool
			wbntEndCursor   *string
		}{
			{
				first:           1,
				bfter:           "",
				wbntNodeCount:   1,
				wbntHbsNextPbge: true,
				wbntEndCursor:   &endCursors[0],
				wbntTotblCount:  nil,
			},
			{
				first:           1,
				bfter:           endCursors[0],
				wbntNodeCount:   1,
				wbntHbsNextPbge: true,
				wbntEndCursor:   &endCursors[1],
				wbntTotblCount:  nil,
			},
			{
				first:           1,
				bfter:           endCursors[1],
				wbntNodeCount:   1,
				wbntHbsNextPbge: fblse,
				wbntEndCursor:   nil,
				wbntTotblCount:  &totblCount,
			},
			{
				first:           testDiffFiles + 1,
				bfter:           "",
				wbntNodeCount:   testDiffFiles,
				wbntHbsNextPbge: fblse,
				wbntEndCursor:   nil,
				wbntTotblCount:  &totblCount,
			},
		}

		for _, tc := rbnge tests {
			brgs := &FileDiffsConnectionArgs{First: &tc.first}
			if tc.bfter != "" {
				brgs.After = &tc.bfter
			}

			conn, err := previewCompbrisonResolver.FileDiffs(ctx, brgs)
			if err != nil {
				t.Fbtbl(err)
			}

			nodes, err := conn.Nodes(ctx)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(nodes) != tc.wbntNodeCount {
				t.Fbtblf("wrong length of nodes. wbnt=%d, hbve=%d", tc.wbntNodeCount, len(nodes))
			}

			pbgeInfo, err := conn.PbgeInfo(ctx)
			if err != nil {
				t.Fbtbl(err)
			}

			if pbgeInfo.HbsNextPbge() != tc.wbntHbsNextPbge {
				t.Fbtblf("pbgeInfo HbsNextPbge wrong. wbnt=%t, hbve=%t", tc.wbntHbsNextPbge, pbgeInfo.HbsNextPbge())
			}

			if diff := cmp.Diff(tc.wbntEndCursor, pbgeInfo.EndCursor()); diff != "" {
				t.Fbtblf("(-wbnt +got):\n%s", diff)
			}

			totblCount, err := conn.TotblCount(ctx)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(tc.wbntTotblCount, totblCount); diff != "" {
				t.Fbtblf("wrong totblCount: %s", diff)
			}
		}
	})

	t.Run("NewFile resolver", func(t *testing.T) {
		fileDiffConnection, err := previewCompbrisonResolver.FileDiffs(ctx, &FileDiffsConnectionArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		fileDiffs, err := fileDiffConnection.Nodes(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		if hbve, wbnt := len(fileDiffs), testDiffFiles; hbve != wbnt {
			t.Fbtblf("invblid len(FileDiffs.Nodes). wbnt=%d hbve=%d", wbnt, len(fileDiffs))
		}
		fileDiff := fileDiffs[0]

		gitserverClient.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, nbme string) ([]byte, error) {
			if nbme != "INSTALL.md" {
				t.Fbtblf("RebdFile received cbll for wrong file: %s", nbme)
			}
			return []byte(testOldFile), nil
		})

		newFile := fileDiff.NewFile()
		if newFile == nil {
			t.Fbtbl("NewFile is null")
		}

		wbntNewFileContent := `First
Second
Line 1
Line 2
Line 3
This is cool: Line 4
Line 5
Line 6
Another Line 7
Foobbr Line 8
Line 9
Line 10
`

		hbveContent, err := newFile.Content(ctx, &GitTreeContentPbgeArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		if hbveContent != wbntNewFileContent {
			t.Fbtblf("wrong file content. wbnt=%q hbve=%q", wbntNewFileContent, hbveContent)
		}
	})
}

func TestApplyPbtch(t *testing.T) {
	tests := []struct {
		nbme     string
		file     string
		pbtch    string
		wbntFile string
	}{
		{
			nbme: "replbce in middle",
			file: `1 some
2
3
4
5
6
7 super bwesome
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
			pbtch: `diff --git b/test b/test
index 38deb4b..d81676e 100644
--- b/test
+++ b/test
@@ -4,7 +4,7 @@
 4
 5
 6
-7 super bwesome
+7 super megb bwesome
 8
 9
 10
`,
			wbntFile: `1 some
2
3
4
5
6
7 super megb bwesome
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
			nbme: "delete file",
			file: `1 some
2
3
`,
			pbtch: `diff --git b/test b/test
deleted file mode 100644
index 2e0cf96..0000000
--- b/test
+++ /dev/null
@@ -1,3 +0,0 @@
-1 some
-2
-3
`,
			wbntFile: "",
		},
		{
			nbme: "New file, bdditionbl newline bt end",
			file: "",
			pbtch: `diff --git b/file2.txt b/file2.txt
new file mode 100644
index 0000000..122f5d9
--- /dev/null
+++ b/file2.txt
@@ -0,0 +1 @@
+filecontent
+
`,
			wbntFile: `filecontent

`,
		},
		{
			nbme: "New file",
			file: "",
			pbtch: `diff --git b/file2.txt b/file2.txt
new file mode 100644
index 0000000..122f5d9
--- /dev/null
+++ b/file2.txt
@@ -0,0 +1 @@
+filecontent
`,
			wbntFile: `filecontent
`,
		},
		{
			nbme: "New file without newline",
			file: "",
			pbtch: `diff --git b/README.md b/README.md
new file mode 100644
index 0000000..373be20
--- /dev/null
+++ b/README.md
@@ -0,0 +1 @@
+No newline bfter this
\ No newline bt end of file
`,
			// Note: No newline.
			wbntFile: `No newline bfter this`,
		},
		{
			nbme: "Add newline to file without newline",
			// Note: No newline.
			file: `No newline bfter this`,
			pbtch: `diff --git b/README.md b/README.md
index 373be20..7e17295 100644
--- b/README.md
+++ b/README.md
@@ -1 +1 @@
-No newline bfter this
\ No newline bt end of file
+No newline bfter this
`,
			// Note: Hbs b newline now.
			wbntFile: `No newline bfter this
`,
		},
		{
			nbme: "Remove newline bt end of file",
			file: `No newline bfter this
`,
			pbtch: `diff --git b/README.md b/README.md
index 7e17295..373be20 100644
--- b/README.md
+++ b/README.md
@@ -1 +1 @@
-No newline bfter this
+No newline bfter this
\ No newline bt end of file
`,
			// Note: Hbs no newline bnymore.
			wbntFile: `No newline bfter this`,
		},
		{
			nbme: "Add line without newline to file thbt ended with no newline",
			file: `No newline bfter this`,
			pbtch: `diff --git b/README.md b/README.md
index 373be20..89bd131 100644
--- b/README.md
+++ b/README.md
@@ -1 +1,2 @@
-No newline bfter this
\ No newline bt end of file
+No newline bfter this
+Also no newline bfter this
\ No newline bt end of file
`,
			// Note: Hbs no newline bt the end.
			wbntFile: `No newline bfter this
Also no newline bfter this`,
		},
		{
			nbme: "Add line without newline to file thbt ended with no newline",
			file: `No newline bfter this`,
			pbtch: `diff --git b/README.md b/README.md
index 373be20..89bd131 100644
--- b/README.md
+++ b/README.md
@@ -1 +1,2 @@
-No newline bfter this
\ No newline bt end of file
+No newline bfter this
+Also no newline bfter this
\ No newline bt end of file
`,
			// Note: Hbs no newline bt the end.
			wbntFile: `No newline bfter this
Also no newline bfter this`,
		},
		{
			nbme: "No newline bnd lbst hunk ends before EOF",
			file: `1
3
4
5
6
7
8
9
10`,
			pbtch: `diff --git b/README.md b/README.md
index 373be20..89bd131 100644
--- b/README.md
+++ b/README.md
@@ -1,4 +1,5 @@
 1
+2
 3
 4
 5
`,
			// Note: Hbs no newline bt the end.
			wbntFile: `1
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
			nbme: "Multiple hunks bnd no newline bt the end",
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
			pbtch: `diff --git b/README.md b/README.md
index 373be20..89bd131 100644
--- b/README.md
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
\ No newline bt end of file
`,
			// Note: Hbs no newline bt the end.
			wbntFile: `1
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

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			fileDiff, err := godiff.PbrseFileDiff([]byte(tc.pbtch))
			if err != nil {
				t.Fbtbl(err)
			}
			hbve, err := bpplyPbtch(tc.file, fileDiff)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve != tc.wbntFile {
				t.Fbtblf("wrong pbtched file content %q, wbnt=%q", hbve, tc.wbntFile)
			}
		})
	}
}
