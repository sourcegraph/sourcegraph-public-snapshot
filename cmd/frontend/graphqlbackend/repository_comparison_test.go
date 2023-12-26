package graphqlbackend

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/go-enry/go-enry/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	godiff "github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestRepositoryComparisonNoMergeBase(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, nil)

	wantBaseRevision := "ba5e"
	wantHeadRevision := "1ead"

	repo := &types.Repo{
		ID:        api.RepoID(1),
		Name:      api.RepoName("test"),
		CreatedAt: time.Now(),
	}

	gsClient := gitserver.NewMockClient()
	gsClient.MergeBaseFunc.SetDefaultReturn("", errors.Errorf("merge base doesn't exist!"))
	gsClient.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		if spec != wantBaseRevision && spec != wantHeadRevision {
			t.Fatalf("ResolveRevision received wrong spec: %s", spec)
		}
		return api.CommitID(spec), nil
	})

	input := &RepositoryComparisonInput{Base: &wantBaseRevision, Head: &wantHeadRevision}
	repoResolver := NewRepositoryResolver(db, gsClient, repo)

	// There shouldn't be any error even when there is no merge base.
	comp, err := NewRepositoryComparison(ctx, db, gsClient, repoResolver, input)
	require.Nil(t, err)
	require.Equal(t, wantBaseRevision, comp.baseRevspec)
	require.Equal(t, wantHeadRevision, comp.headRevspec)
	require.Equal(t, "..", comp.rangeType)
}

func TestRepositoryComparison(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, nil)

	wantBaseRevision := "24f7ca7c1190835519e261d7eefa09df55ceea4f"
	wantMergeBaseRevision := "a7985dde7f92ad3490ec513be78fa2b365c7534c"
	wantHeadRevision := "b69072d5f687b31b9f6ae3ceafdc24c259c4b9ec"

	repo := &types.Repo{
		ID:        api.RepoID(1),
		Name:      api.RepoName("github.com/sourcegraph/sourcegraph"),
		CreatedAt: time.Now(),
	}

	gsClient := gitserver.NewMockClientWithExecReader(nil, func(_ context.Context, _ api.RepoName, args []string) (io.ReadCloser, error) {
		if len(args) < 1 && args[0] != "diff" {
			t.Fatalf("gitserver.ExecReader received wrong args: %v", args)
		}
		if args[len(args)-1] == "JOKES.md" {
			return io.NopCloser(strings.NewReader(testDiffJokesOnly)), nil
		}
		return io.NopCloser(strings.NewReader(testDiff + testCopyDiff)), nil
	})

	gsClient.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		if spec != wantMergeBaseRevision && spec != wantHeadRevision {
			t.Fatalf("ResolveRevision received wrong spec: %s", spec)
		}
		return api.CommitID(spec), nil
	})

	gsClient.MergeBaseFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, a, b api.CommitID) (api.CommitID, error) {
		if string(a) != wantBaseRevision || string(b) != wantHeadRevision {
			t.Fatalf("gitserver.MergeBase received wrong args: %s %s", a, b)
		}
		return api.CommitID(wantMergeBaseRevision), nil
	})

	input := &RepositoryComparisonInput{Base: &wantBaseRevision, Head: &wantHeadRevision}
	repoResolver := NewRepositoryResolver(db, gsClient, repo)

	comp, err := NewRepositoryComparison(ctx, db, gsClient, repoResolver, input)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("BaseRepository", func(t *testing.T) {
		if have, want := comp.BaseRepository(), repoResolver; have != want {
			t.Fatalf("BaseRepository wrong. want=%+v, have=%+v", want, have)
		}
	})

	t.Run("HeadRepository", func(t *testing.T) {
		if have, want := comp.HeadRepository(), repoResolver; have != want {
			t.Fatalf("headRepository wrong. want=%+v, have=%+v", want, have)
		}
	})

	t.Run("Range", func(t *testing.T) {
		gitRange := comp.Range()

		wantRangeExpr := fmt.Sprintf("%s...%s", wantBaseRevision, wantHeadRevision)
		if have, want := gitRange.Expr(), wantRangeExpr; have != want {
			t.Fatalf("range expression. want=%s, have=%s", want, have)
		}
	})

	t.Run("Commits", func(t *testing.T) {
		commits := []*gitdomain.Commit{
			{ID: api.CommitID(wantBaseRevision)},
			{ID: api.CommitID(wantHeadRevision)},
		}

		mockGSClient := gitserver.NewMockClient()
		mockGSClient.CommitsFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, opts gitserver.CommitsOptions) ([]*gitdomain.Commit, error) {
			wantRange := fmt.Sprintf("%s..%s", wantBaseRevision, wantHeadRevision)

			if have, want := opts.Range, wantRange; have != want {
				t.Fatalf("git.Commits received wrong range. want=%s, have=%s", want, have)
			}

			return commits, nil
		})

		newComp, err := NewRepositoryComparison(ctx, db, mockGSClient, repoResolver, input)
		if err != nil {
			t.Fatal(err)
		}

		commitConnection := newComp.Commits(&RepositoryComparisonCommitsArgs{})

		nodes, err := commitConnection.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if len(nodes) != len(commits) {
			t.Fatalf("wrong length of nodes: %d", len(nodes))
		}

		for i, n := range nodes {
			if have, want := string(n.OID()), string(commits[i].ID); have != want {
				t.Fatalf("nodes[%d] has wrong commit ID. want=%s, have=%s", i, want, have)
			}
		}

		totalCount, err := commitConnection.TotalCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if totalCount == nil {
			t.Fatalf("no TotalCount returned")
		}
		if have, want := *totalCount, int32(len(commits)); have != want {
			t.Fatalf("totalCount wrong. want=%d, have=%d", want, have)
		}
	})

	t.Run("Commits with Path", func(t *testing.T) {
		commits := []*gitdomain.Commit{
			{ID: api.CommitID(wantBaseRevision)},
		}

		mockGSClient := gitserver.NewMockClient()
		mockGSClient.CommitsFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, opts gitserver.CommitsOptions) ([]*gitdomain.Commit, error) {
			if opts.Path == "" {
				t.Fatalf("expected a path as part of commits args")
			}
			return commits, nil
		})

		newComp, err := NewRepositoryComparison(ctx, db, mockGSClient, repoResolver, input)
		if err != nil {
			t.Fatal(err)
		}

		testPath := "testpath"
		commitConnection := newComp.Commits(&RepositoryComparisonCommitsArgs{Path: &testPath})

		nodes, err := commitConnection.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if len(nodes) != len(commits) {
			t.Fatalf("wrong length of nodes: %d", len(nodes))
		}
	})
	t.Run("FileDiffs", func(t *testing.T) {
		t.Run("RawDiff", func(t *testing.T) {
			diffConnection, err := comp.FileDiffs(ctx, &FileDiffsConnectionArgs{})
			if err != nil {
				t.Fatal(err)
			}

			rawDiff, err := diffConnection.RawDiff(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if have, want := rawDiff, testDiff+testCopyDiff; have != want {
				t.Fatalf("rawDiff wrong. want=%q, have=%q", want, have)
			}
		})

		t.Run("DiffStat", func(t *testing.T) {
			diffConnection, err := comp.FileDiffs(ctx, &FileDiffsConnectionArgs{})
			if err != nil {
				t.Fatal(err)
			}

			diffStat, err := diffConnection.DiffStat(ctx)
			if err != nil {
				t.Fatal(err)
			}

			want := "9 added, 8 deleted"
			if have := fmt.Sprintf("%d added, %d deleted", diffStat.Added(), diffStat.Deleted()); have != want {
				t.Fatalf("wrong diffstat. want=%q, have=%q", want, have)
			}
		})

		t.Run("LimitedPaths", func(t *testing.T) {
			paths := []string{"JOKES.md"}
			diffConnection, err := comp.FileDiffs(ctx, &FileDiffsConnectionArgs{Paths: &paths})
			if err != nil {
				t.Fatal(err)
			}

			nodes, err := diffConnection.Nodes(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if len(nodes) != 1 {
				t.Fatalf("expected 1 file node, got %d", len(nodes))
			}

			oldPath := nodes[0].OldPath()
			if oldPath == nil {
				t.Fatalf("expected non-nil oldPath")
			}

			if *oldPath != "JOKES.md" {
				t.Fatalf("expected JOKES.md, got %s", *oldPath)
			}
		})

		t.Run("FileDiff", func(t *testing.T) {
			diffConnection, err := comp.FileDiffs(ctx, &FileDiffsConnectionArgs{})
			if err != nil {
				t.Fatal(err)
			}

			nodes, err := diffConnection.Nodes(ctx)
			if err != nil {
				t.Fatal(err)
			}

			// +1 for the copyDiffFile
			if len(nodes) != testDiffFiles+1 {
				t.Fatalf("wrong length of nodes. want=%d, have=%d", testDiffFiles, len(nodes))
			}

			n := nodes[0]
			wantOldPath := "INSTALL.md"
			if diff := cmp.Diff(&wantOldPath, n.OldPath()); diff != "" {
				t.Fatalf("wrong OldPath: %s", diff)
			}

			wantNewPath := "INSTALL.md"
			if diff := cmp.Diff(&wantNewPath, n.NewPath()); diff != "" {
				t.Fatalf("wrong NewPath: %s", diff)
			}

			wantStat := "3 added, 3 deleted"
			haveStat := n.Stat()
			if haveStat == nil {
				t.Fatalf("no diff stat")
			}
			if have := fmt.Sprintf("%d added, %d deleted", haveStat.Added(), haveStat.Deleted()); have != wantStat {
				t.Fatalf("wrong diffstat. want=%q, have=%q", wantStat, have)
			}

			oldFile := n.OldFile()
			if oldFile == nil {
				t.Fatalf("OldFile() is nil")
			}
			gitBlob, ok := oldFile.ToGitBlob()
			if !ok {
				t.Fatalf("OldFile() is no GitBlob")
			}
			if have, want := string(gitBlob.Commit().OID()), wantMergeBaseRevision; have != want {
				t.Fatalf("Got wrong commit ID for OldFile(): want=%s have=%s", want, have)
			}
			newFile := n.NewFile()
			if newFile == nil {
				t.Fatalf("NewFile() is nil")
			}

			mostRelevant := n.MostRelevantFile()
			if mostRelevant == nil {
				t.Fatalf("MostRelevantFile is nil")
			}
			relevantURL := mostRelevant.CanonicalURL()

			wantRelevantURL := fmt.Sprintf("/%s@%s/-/blob/%s", repo.Name, wantHeadRevision, "INSTALL.md")
			if relevantURL != wantRelevantURL {
				t.Fatalf("MostRelevantFile.CanonicalURL() is wrong. have=%q, want=%q", relevantURL, wantRelevantURL)
			}

			newFileURL := newFile.CanonicalURL()
			// NewFile should be the most relevant file
			if newFileURL != wantRelevantURL {
				t.Fatalf(
					"NewFile.CanonicalURL() is not MostRelevantFile.CanonicalURL(). have=%q, want=%q",
					relevantURL, wantRelevantURL,
				)
			}

			t.Run("DiffHunks", func(t *testing.T) {
				hunks := nodes[0].Hunks()
				wantHunkCount := 1
				if have := len(hunks); have != wantHunkCount {
					t.Fatalf("len(hunks) wrong. want=%d, have=%d", wantHunkCount, have)
				}
			})
		})

		t.Run("Pagination", func(t *testing.T) {
			endCursors := []string{"1", "2", "3"}
			totalCount := int32(testDiffFiles) + 1

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
					wantHasNextPage: true,
					wantEndCursor:   &endCursors[2],
					wantTotalCount:  nil,
				},
				{
					first:           1,
					after:           endCursors[2],
					wantNodeCount:   1,
					wantHasNextPage: false,
					wantEndCursor:   nil,
					wantTotalCount:  &totalCount,
				},
				{
					first:           testDiffFiles + 1,
					after:           "",
					wantNodeCount:   testDiffFiles + 1,
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

				conn, err := comp.FileDiffs(ctx, args)
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
					t.Fatal(diff)
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
	})
}

func TestDiffHunk(t *testing.T) {
	ctx := context.Background()

	dr := godiff.NewMultiFileDiffReader(strings.NewReader(testDiff))
	// We only read the first file diff from testDiff
	fileDiff, err := dr.ReadFile()
	if err != nil && err != io.EOF {
		t.Fatalf("parsing diff failed: %s", err)
	}

	hunk := &DiffHunk{hunk: fileDiff.Hunks[0]}

	t.Run("OldNoNewlineAt", func(t *testing.T) {
		if have, want := hunk.OldNoNewlineAt(), false; have != want {
			t.Fatalf("Lines wrong. want=%t, have=%t", want, have)
		}
	})

	t.Run("Ranges", func(t *testing.T) {
		testRange := func(r *DiffHunkRange, wantStartLine, wantLines int32) {
			if have := r.StartLine(); have != wantStartLine {
				t.Fatalf("StartLine wrong. want=%d, have=%d", wantStartLine, have)
			}
			if have := r.Lines(); have != wantLines {
				t.Fatalf("Lines wrong. want=%d, have=%d", wantLines, have)
			}
		}
		testRange(hunk.OldRange(), 3, 10)
		testRange(hunk.NewRange(), 3, 10)
	})

	t.Run("Section", func(t *testing.T) {
		if hunk.Section() != nil {
			t.Fatalf("hunk.Section is not nil: %+v\n", hunk.Section())
		}
	})

	t.Run("Body", func(t *testing.T) {
		if diff := cmp.Diff(testDiffFirstHunk, hunk.Body()); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Highlight", func(t *testing.T) {
		hunk.highlighter = &dummyFileHighlighter{
			highlightedBase: []template.HTML{"B1", "B2", "B3", "B4", "B5", "B6", "B7", "B8", "B9", "B10", "B11", "B12"},
			highlightedHead: []template.HTML{"H1", "H2", "H3", "H4", "H5", "H6", "H7", "H8", "H9", "H10", "H11", "H12"},
		}

		body, err := hunk.Highlight(ctx, &HighlightArgs{
			DisableTimeout:     false,
			HighlightLongLines: false,
		})
		if err != nil {
			t.Fatal(err)
		}
		if body.Aborted() {
			t.Fatal("highlighting is aborted")
		}

		wantLines := []struct {
			kind, html string
		}{
			{kind: "UNCHANGED", html: "B3"},
			{kind: "UNCHANGED", html: "B4"},
			{kind: "UNCHANGED", html: "B5"},
			{kind: "DELETED", html: "B6"},
			{kind: "ADDED", html: "H6"},
			{kind: "UNCHANGED", html: "B7"},
			{kind: "UNCHANGED", html: "B8"},
			{kind: "DELETED", html: "B9"},
			{kind: "DELETED", html: "B10"},
			{kind: "ADDED", html: "H9"},
			{kind: "ADDED", html: "H10"},
			{kind: "UNCHANGED", html: "B11"},
			{kind: "UNCHANGED", html: "B12"},
		}

		lines := body.Lines()
		if have, want := len(lines), len(wantLines); have != want {
			t.Fatalf("len(Highlight.Lines) is wrong. want = %d, have = %d", want, have)
		}
		for i, n := range lines {
			wantedLine := wantLines[i]
			if n.Kind() != wantedLine.kind {
				t.Fatalf("Kind is wrong. want = %q, have = %q", wantedLine.kind, n.Kind())
			}
			if n.HTML() != wantedLine.html {
				t.Fatalf("HTML is wrong. want = %q, have = %q", wantedLine.html, n.HTML())
			}
		}
	})
}

func TestDiffHunk2(t *testing.T) {
	// This test exists to protect against panics related to
	// https://github.com/sourcegraph/sourcegraph/pull/21068

	ctx := context.Background()
	// https://sourcegraph.com/github.com/dominikh/go-tools/-/blob/cmd/staticcheck/README.md
	// was used to produce this test diff.
	filediff := `diff --git cmd/staticcheck/README.md cmd/staticcheck/README.md
index 4d14577..10ef458 100644
--- cmd/staticcheck/README.md
+++ cmd/staticcheck/README.md
@@ -13,3 +13,5 @@ See [the main README](https://github.com/dominikh/go-tools#installation) for ins
 Detailed documentation can be found on
 [staticcheck.io](https://staticcheck.io/docs/).
` + " " + `
+
+(c) Copyright Sourcegraph 2013-2021.
\ No newline at end of file
`
	dr := godiff.NewMultiFileDiffReader(strings.NewReader(filediff))
	// We only read the first file diff from testDiff
	fileDiff, err := dr.ReadFile()
	if err != nil && err != io.EOF {
		t.Fatalf("parsing diff failed: %s", err)
	}

	hunk := &DiffHunk{hunk: fileDiff.Hunks[0]}

	t.Run("Highlight", func(t *testing.T) {
		hunk.highlighter = &dummyFileHighlighter{
			highlightedBase: []template.HTML{
				"<div><span class=\"hl-text hl-html hl-markdown\"><span class=\"hl-meta hl-block-level hl-markdown\"><span class=\"hl-markup hl-heading hl-1 hl-markdown\"><span class=\"hl-punctuation hl-definition hl-heading hl-begin hl-markdown\">#</span> </span><span class=\"hl-markup hl-heading hl-1 hl-markdown\"><span class=\"hl-entity hl-name hl-section hl-markdown\">staticcheck</span><span class=\"hl-meta hl-whitespace hl-newline hl-markdown\">\n</span></span></span></span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">_staticcheck_ offers extensive analysis of Go code, covering a myriad\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">of categories. It will detect bugs, suggest code simplifications,\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">point out dead code, and more.\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\"><span class=\"hl-meta hl-block-level hl-markdown\"><span class=\"hl-markup hl-heading hl-2 hl-markdown\"><span class=\"hl-punctuation hl-definition hl-heading hl-begin hl-markdown\">##</span> </span><span class=\"hl-markup hl-heading hl-2 hl-markdown\"><span class=\"hl-entity hl-name hl-section hl-markdown\">Installation</span><span class=\"hl-meta hl-whitespace hl-newline hl-markdown\">\n</span></span></span></span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">See [the main README](https://github.com/dominikh/go-tools#installation) for installation instructions.\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\"><span class=\"hl-meta hl-block-level hl-markdown\"><span class=\"hl-markup hl-heading hl-2 hl-markdown\"><span class=\"hl-punctuation hl-definition hl-heading hl-begin hl-markdown\">##</span> </span><span class=\"hl-markup hl-heading hl-2 hl-markdown\"><span class=\"hl-entity hl-name hl-section hl-markdown\">Documentation</span><span class=\"hl-meta hl-whitespace hl-newline hl-markdown\">\n</span></span></span></span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">Detailed documentation can be found on\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">[staticcheck.io](https://staticcheck.io/docs/).\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
			},
			highlightedHead: []template.HTML{
				"<div><span class=\"hl-text hl-html hl-markdown\"><span class=\"hl-meta hl-block-level hl-markdown\"><span class=\"hl-markup hl-heading hl-1 hl-markdown\"><span class=\"hl-punctuation hl-definition hl-heading hl-begin hl-markdown\">#</span> </span><span class=\"hl-markup hl-heading hl-1 hl-markdown\"><span class=\"hl-entity hl-name hl-section hl-markdown\">staticcheck</span><span class=\"hl-meta hl-whitespace hl-newline hl-markdown\">\n</span></span></span></span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">_staticcheck_ offers extensive analysis of Go code, covering a myriad\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">of categories. It will detect bugs, suggest code simplifications,\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">point out dead code, and more.\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\"><span class=\"hl-meta hl-block-level hl-markdown\"><span class=\"hl-markup hl-heading hl-2 hl-markdown\"><span class=\"hl-punctuation hl-definition hl-heading hl-begin hl-markdown\">##</span> </span><span class=\"hl-markup hl-heading hl-2 hl-markdown\"><span class=\"hl-entity hl-name hl-section hl-markdown\">Installation</span><span class=\"hl-meta hl-whitespace hl-newline hl-markdown\">\n</span></span></span></span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">See [the main README](https://github.com/dominikh/go-tools#installation) for installation instructions.\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\"><span class=\"hl-meta hl-block-level hl-markdown\"><span class=\"hl-markup hl-heading hl-2 hl-markdown\"><span class=\"hl-punctuation hl-definition hl-heading hl-begin hl-markdown\">##</span> </span><span class=\"hl-markup hl-heading hl-2 hl-markdown\"><span class=\"hl-entity hl-name hl-section hl-markdown\">Documentation</span><span class=\"hl-meta hl-whitespace hl-newline hl-markdown\">\n</span></span></span></span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">Detailed documentation can be found on\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">[staticcheck.io](https://staticcheck.io/docs/).\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">(c) Copyright Sourcegraph 2013-2021.</span></div>",
				"<div><span class=\"hl-text hl-html hl-markdown\">\n</span></div>",
			},
		}

		body, err := hunk.Highlight(ctx, &HighlightArgs{
			DisableTimeout:     false,
			HighlightLongLines: false,
		})
		if err != nil {
			t.Fatal(err)
		}
		if body.Aborted() {
			t.Fatal("highlighting is aborted")
		}
	})
}

func TestDiffHunk3(t *testing.T) {
	// This test exists to protect against an edge case bug illustrated in
	// https://github.com/sourcegraph/sourcegraph/pull/25866

	ctx := context.Background()
	// https://sourcegraph.com/github.com/dominikh/go-tools/-/blob/cmd/staticcheck/README.md
	// was used to produce this test diff.
	filediff := `diff --git cmd/staticcheck/README.md cmd/staticcheck/README.md
index 4d14577..9fe9a4f 100644
--- cmd/staticcheck/README.md
+++ cmd/staticcheck/README.md
@@ -1,10 +1,6 @@
 # staticcheck
` + "-" + `
-_staticcheck_ offers extensive analysis of Go code, covering a myriad
-of categories. It will detect bugs, suggest code simplifications,
-point out dead code, and more.
` + "-" + `
 ## Installation
+Wowza!
` + "-" + `
 See [the main README](https://github.com/dominikh/go-tools#installation) for installation instructions.`

	dr := godiff.NewMultiFileDiffReader(strings.NewReader(filediff))
	// We only read the first file diff from testDiff
	fileDiff, err := dr.ReadFile()
	if err != nil && err != io.EOF {
		t.Fatalf("parsing diff failed: %s", err)
	}

	hunk := &DiffHunk{hunk: fileDiff.Hunks[0]}

	t.Run("Highlight", func(t *testing.T) {
		hunk.highlighter = &dummyFileHighlighter{
			// We don't care about the actual html formatting, just the number + order of
			// the lines we get back after "applying" the diff to the highlighting.
			highlightedBase: []template.HTML{
				"# staticcheck",
				"",
				"_staticcheck_ offers extensive analysis of Go code, covering a myriad",
				"of categories. It will detect bugs, suggest code simplifications,",
				"point out dead code, and more.",
				"",
				"## Installation",
				"",
				"See [the main README](https://github.com/dominikh/go-tools#installation) for installation instructions.",
				"",
			},
			highlightedHead: []template.HTML{
				"# staticcheck",
				"## Installation",
				"Wowza!",
				"See [the main README](https://github.com/dominikh/go-tools#installation) for installation instructions.",
				"",
			},
		}

		body, err := hunk.Highlight(ctx, &HighlightArgs{
			DisableTimeout:     false,
			HighlightLongLines: false,
		})
		if err != nil {
			t.Fatal(err)
		}
		if body.Aborted() {
			t.Fatal("highlighting is aborted")
		}

		wantLines := []struct {
			kind, html string
		}{
			{kind: "UNCHANGED", html: "# staticcheck"},
			{kind: "DELETED", html: ""},
			{kind: "DELETED", html: "_staticcheck_ offers extensive analysis of Go code, covering a myriad"},
			{kind: "DELETED", html: "of categories. It will detect bugs, suggest code simplifications,"},
			{kind: "DELETED", html: "point out dead code, and more."},
			{kind: "DELETED", html: ""},
			{kind: "UNCHANGED", html: "## Installation"},
			{kind: "ADDED", html: "Wowza!"},
			{kind: "DELETED", html: ""},
			{kind: "UNCHANGED", html: "See [the main README](https://github.com/dominikh/go-tools#installation) for installation instructions."},
		}

		lines := body.Lines()
		if have, want := len(lines), len(wantLines); have != want {
			t.Fatalf("len(Highlight.Lines) is wrong. want = %d, have = %d", want, have)
		}
		for i, n := range lines {
			wantedLine := wantLines[i]
			if n.Kind() != wantedLine.kind {
				t.Fatalf("Kind is wrong. want = %q, have = %q", wantedLine.kind, n.Kind())
			}
			if n.HTML() != wantedLine.html {
				t.Fatalf("HTML is wrong. want = %q, have = %q", wantedLine.html, n.HTML())
			}
		}
	})
}

func TestDiffHunk4(t *testing.T) {
	// This test exists to protect against an edge case bug illustrated in
	// https://github.com/sourcegraph/sourcegraph/pull/39377

	ctx := context.Background()
	// Ran 'git diff --cached --no-prefix --binary' on a local repo to generate this diff (with the starting lines
	// changes to 1)
	filediff := `diff --git toggle.go toggle.go
index d206c4c..bb06461 100644
--- toggle.go
+++ toggle.go
@@ -1,10 +1,3 @@ func AddFeatures(features map[string]bool) {
 func AddFeature(key string, isEnabled bool) {
        features[strings.ToLower(key)] = isEnabled
 }
-
-// IsEnabled determines if the specified feature is enabled. Determining if a feature is enabled is
-// case insensitive.
-// If a feature is not present, it defaults to false.
-func IsEnabled(key string) bool {
-       return features[strings.ToLower(key)]
-}
`

	dr := godiff.NewMultiFileDiffReader(strings.NewReader(filediff))
	// We only read the first file diff from testDiff
	fileDiff, err := dr.ReadFile()
	if err != nil && err != io.EOF {
		t.Fatalf("parsing diff failed: %s", err)
	}

	hunk := &DiffHunk{hunk: fileDiff.Hunks[0]}

	t.Run("Highlight", func(t *testing.T) {
		hunk.highlighter = &dummyFileHighlighter{
			// We don't care about the actual html formatting, just the number + order of
			// the lines we get back after "applying" the diff to the highlighting.
			highlightedBase: []template.HTML{
				"func AddFeature(key string, isEnabled bool) {",
				"features[strings.ToLower(key)] = isEnabled",
				"}",
				"",
				"// IsEnabled determines if the specified feature is enabled. Determining if a feature is enabled is",
				"// case insensitive.",
				"// If a feature is not present, it defaults to false.",
				"func IsEnabled(key string) bool {",
				"return features[strings.ToLower(key)]",
				"}",
				"",
			},
			highlightedHead: []template.HTML{
				"func AddFeature(key string, isEnabled bool) {",
				"features[strings.ToLower(key)] = isEnabled",
				"}",
				"",
			},
		}

		body, err := hunk.Highlight(ctx, &HighlightArgs{
			DisableTimeout:     false,
			HighlightLongLines: false,
		})
		if err != nil {
			t.Fatal(err)
		}
		if body.Aborted() {
			t.Fatal("highlighting is aborted")
		}

		wantLines := []struct {
			kind, html string
		}{
			{kind: "UNCHANGED", html: "func AddFeature(key string, isEnabled bool) {"},
			{kind: "UNCHANGED", html: "features[strings.ToLower(key)] = isEnabled"},
			{kind: "UNCHANGED", html: "}"},
			{kind: "DELETED", html: ""},
			{kind: "DELETED", html: "// IsEnabled determines if the specified feature is enabled. Determining if a feature is enabled is"},
			{kind: "DELETED", html: "// case insensitive."},
			{kind: "DELETED", html: "// If a feature is not present, it defaults to false."},
			{kind: "DELETED", html: "func IsEnabled(key string) bool {"},
			{kind: "DELETED", html: "return features[strings.ToLower(key)]"},
			{kind: "DELETED", html: "}"},
		}

		lines := body.Lines()
		if have, want := len(lines), len(wantLines); have != want {
			t.Fatalf("len(Highlight.Lines) is wrong. want = %d, have = %d", want, have)
		}
		for i, n := range lines {
			wantedLine := wantLines[i]
			if n.Kind() != wantedLine.kind {
				t.Fatalf("Kind is wrong. want = %q, have = %q", wantedLine.kind, n.Kind())
			}
			if n.HTML() != wantedLine.html {
				t.Fatalf("HTML is wrong. want = %q, have = %q", wantedLine.html, n.HTML())
			}
		}
	})
}

const testDiffFiles = 3
const testDiff = `diff --git INSTALL.md INSTALL.md
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
`

const testCopyDiff = `diff --git a/test.txt b/test2.txt
similarity index 100%
copy from test.txt
copy to test2.txt
`

const testDiffFirstHunk = ` Line 1
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
`

const testDiffJokesOnly = `
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
`

func TestFileDiffHighlighter(t *testing.T) {
	ctx := context.Background()

	file1 := &dummyFileResolver{
		path: "old.txt",
		content: func(ctx context.Context, args *GitTreeContentPageArgs) (string, error) {
			return "old1\nold2\nold3\n", nil
		},
	}
	file2 := &dummyFileResolver{
		path: "new.txt",
		content: func(ctx context.Context, args *GitTreeContentPageArgs) (string, error) {
			return "new1\nnew2\nnew3\n", nil
		},
	}

	highlightedOld := `<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#657b83;">old1
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#657b83;">old2
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#657b83;">old3</span></div></td></tr></tbody></table>`
	highlightedNew := `<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#657b83;">new1
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#657b83;">new2
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#657b83;">new3</span></div></td></tr></tbody></table>`

	highlight.Mocks.Code = func(p highlight.Params) (*highlight.HighlightedCode, bool, error) {
		switch p.Filepath {
		case file1.path:
			response := highlight.NewHighlightedCodeWithHTML(template.HTML(highlightedOld))
			return &response, false, nil
		case file2.path:
			response := highlight.NewHighlightedCodeWithHTML(template.HTML(highlightedNew))
			return &response, false, nil
		default:
			return nil, false, errors.Errorf("unknown file: %s", p.Filepath)
		}
	}
	t.Cleanup(highlight.ResetMocks)

	highlighter := fileDiffHighlighter{oldFile: file1, newFile: file2}
	highlightedBase, highlightedHead, aborted, err := highlighter.Highlight(ctx, &HighlightArgs{
		DisableTimeout:     false,
		HighlightLongLines: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if aborted {
		t.Fatalf("highlighting aborted")
	}

	wantLinesBase := []template.HTML{
		"<div><span style=\"color:#657b83;\">old1\n</span></div>",
		"<div><span style=\"color:#657b83;\">old2\n</span></div>",
		"<div><span style=\"color:#657b83;\">old3</span></div>",
	}
	if diff := cmp.Diff(wantLinesBase, highlightedBase); diff != "" {
		t.Fatalf("wrong highlightedBase: %s", diff)
	}

	wantLinesHead := []template.HTML{
		"<div><span style=\"color:#657b83;\">new1\n</span></div>",
		"<div><span style=\"color:#657b83;\">new2\n</span></div>",
		"<div><span style=\"color:#657b83;\">new3</span></div>",
	}
	if diff := cmp.Diff(wantLinesHead, highlightedHead); diff != "" {
		t.Fatalf("wrong highlightedHead: %s", diff)
	}
}

type dummyFileResolver struct {
	path, name string

	richHTML      string
	url           string
	canonicalURL  string
	changelistURL string

	content func(context.Context, *GitTreeContentPageArgs) (string, error)
}

func (d *dummyFileResolver) Path() string      { return d.path }
func (d *dummyFileResolver) Name() string      { return d.name }
func (d *dummyFileResolver) IsDirectory() bool { return false }
func (d *dummyFileResolver) Content(ctx context.Context, args *GitTreeContentPageArgs) (string, error) {
	return d.content(ctx, args)
}

func (d *dummyFileResolver) ByteSize(ctx context.Context) (int32, error) {
	content, err := d.content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len([]byte(content))), nil
}
func (d *dummyFileResolver) TotalLines(ctx context.Context) (int32, error) {
	content, err := d.content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len(strings.Split(content, "\n"))), nil
}

func (d *dummyFileResolver) Binary(ctx context.Context) (bool, error) {
	return false, nil
}

func (d *dummyFileResolver) RichHTML(ctx context.Context, args *GitTreeContentPageArgs) (string, error) {
	return d.richHTML, nil
}

func (d *dummyFileResolver) URL(ctx context.Context) (string, error) {
	return d.url, nil
}

func (d *dummyFileResolver) CanonicalURL() string {
	return d.canonicalURL
}

func (d *dummyFileResolver) ChangelistURL(ctx context.Context) (*string, error) {
	return &d.changelistURL, nil
}

func (d *dummyFileResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return []*externallink.Resolver{}, nil
}

func (d *dummyFileResolver) Highlight(ctx context.Context, args *HighlightArgs) (*HighlightedFileResolver, error) {
	return nil, errors.New("not implemented")
}

func (d *dummyFileResolver) Languages(ctx context.Context) ([]string, error) {
	filename := d.Name()
	languages := enry.GetLanguages(filename, nil)
	if len(languages) <= 1 {
		return languages, nil
	}
	content, err := d.Content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return nil, err
	}
	return enry.GetLanguages(filename, []byte(content)), nil
}

func (d *dummyFileResolver) ToGitBlob() (*GitTreeEntryResolver, bool) {
	return nil, false
}

func (d *dummyFileResolver) ToVirtualFile() (*VirtualFileResolver, bool) {
	return nil, false
}

func (d *dummyFileResolver) ToBatchSpecWorkspaceFile() (BatchWorkspaceFileResolver, bool) {
	return nil, false
}

type dummyFileHighlighter struct {
	highlightedBase, highlightedHead []template.HTML
}

func (r *dummyFileHighlighter) Highlight(ctx context.Context, args *HighlightArgs) ([]template.HTML, []template.HTML, bool, error) {
	return r.highlightedBase, r.highlightedHead, false, nil
}
