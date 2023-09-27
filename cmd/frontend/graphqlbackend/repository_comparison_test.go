pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"html/templbte"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	godiff "github.com/sourcegrbph/go-diff/diff"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/highlight"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRepositoryCompbrisonNoMergeBbse(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, nil)

	wbntBbseRevision := "bb5e"
	wbntHebdRevision := "1ebd"

	repo := &types.Repo{
		ID:        bpi.RepoID(1),
		Nbme:      bpi.RepoNbme("test"),
		CrebtedAt: time.Now(),
	}

	gsClient := gitserver.NewMockClient()
	gsClient.MergeBbseFunc.SetDefbultReturn("", errors.Errorf("merge bbse doesn't exist!"))
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if spec != wbntBbseRevision && spec != wbntHebdRevision {
			t.Fbtblf("ResolveRevision received wrong spec: %s", spec)
		}
		return bpi.CommitID(spec), nil
	})

	input := &RepositoryCompbrisonInput{Bbse: &wbntBbseRevision, Hebd: &wbntHebdRevision}
	repoResolver := NewRepositoryResolver(db, gsClient, repo)

	// There shouldn't be bny error even when there is no merge bbse.
	comp, err := NewRepositoryCompbrison(ctx, db, gsClient, repoResolver, input)
	require.Nil(t, err)
	require.Equbl(t, wbntBbseRevision, comp.bbseRevspec)
	require.Equbl(t, wbntHebdRevision, comp.hebdRevspec)
	require.Equbl(t, "..", comp.rbngeType)
}

func TestRepositoryCompbrison(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, nil)

	wbntBbseRevision := "24f7cb7c1190835519e261d7eefb09df55ceeb4f"
	wbntMergeBbseRevision := "b7985dde7f92bd3490ec513be78fb2b365c7534c"
	wbntHebdRevision := "b69072d5f687b31b9f6be3cebfdc24c259c4b9ec"

	repo := &types.Repo{
		ID:        bpi.RepoID(1),
		Nbme:      bpi.RepoNbme("github.com/sourcegrbph/sourcegrbph"),
		CrebtedAt: time.Now(),
	}

	gsClient := gitserver.NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (io.RebdCloser, error) {
		if len(brgs) < 1 && brgs[0] != "diff" {
			t.Fbtblf("gitserver.ExecRebder received wrong brgs: %v", brgs)
		}
		if brgs[len(brgs)-1] == "JOKES.md" {
			return io.NopCloser(strings.NewRebder(testDiffJokesOnly)), nil
		}
		return io.NopCloser(strings.NewRebder(testDiff + testCopyDiff)), nil
	})

	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if spec != wbntMergeBbseRevision && spec != wbntHebdRevision {
			t.Fbtblf("ResolveRevision received wrong spec: %s", spec)
		}
		return bpi.CommitID(spec), nil
	})

	gsClient.MergeBbseFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, b, b bpi.CommitID) (bpi.CommitID, error) {
		if string(b) != wbntBbseRevision || string(b) != wbntHebdRevision {
			t.Fbtblf("gitserver.MergeBbse received wrong brgs: %s %s", b, b)
		}
		return bpi.CommitID(wbntMergeBbseRevision), nil
	})

	input := &RepositoryCompbrisonInput{Bbse: &wbntBbseRevision, Hebd: &wbntHebdRevision}
	repoResolver := NewRepositoryResolver(db, gsClient, repo)

	comp, err := NewRepositoryCompbrison(ctx, db, gsClient, repoResolver, input)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("BbseRepository", func(t *testing.T) {
		if hbve, wbnt := comp.BbseRepository(), repoResolver; hbve != wbnt {
			t.Fbtblf("BbseRepository wrong. wbnt=%+v, hbve=%+v", wbnt, hbve)
		}
	})

	t.Run("HebdRepository", func(t *testing.T) {
		if hbve, wbnt := comp.HebdRepository(), repoResolver; hbve != wbnt {
			t.Fbtblf("hebdRepository wrong. wbnt=%+v, hbve=%+v", wbnt, hbve)
		}
	})

	t.Run("Rbnge", func(t *testing.T) {
		gitRbnge := comp.Rbnge()

		wbntRbngeExpr := fmt.Sprintf("%s...%s", wbntBbseRevision, wbntHebdRevision)
		if hbve, wbnt := gitRbnge.Expr(), wbntRbngeExpr; hbve != wbnt {
			t.Fbtblf("rbnge expression. wbnt=%s, hbve=%s", wbnt, hbve)
		}
	})

	t.Run("Commits", func(t *testing.T) {
		commits := []*gitdombin.Commit{
			{ID: bpi.CommitID(wbntBbseRevision)},
			{ID: bpi.CommitID(wbntHebdRevision)},
		}

		mockGSClient := gitserver.NewMockClient()
		mockGSClient.CommitsFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, opts gitserver.CommitsOptions) ([]*gitdombin.Commit, error) {
			wbntRbnge := fmt.Sprintf("%s..%s", wbntBbseRevision, wbntHebdRevision)

			if hbve, wbnt := opts.Rbnge, wbntRbnge; hbve != wbnt {
				t.Fbtblf("git.Commits received wrong rbnge. wbnt=%s, hbve=%s", wbnt, hbve)
			}

			return commits, nil
		})

		newComp, err := NewRepositoryCompbrison(ctx, db, mockGSClient, repoResolver, input)
		if err != nil {
			t.Fbtbl(err)
		}

		commitConnection := newComp.Commits(&RepositoryCompbrisonCommitsArgs{})

		nodes, err := commitConnection.Nodes(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		if len(nodes) != len(commits) {
			t.Fbtblf("wrong length of nodes: %d", len(nodes))
		}

		for i, n := rbnge nodes {
			if hbve, wbnt := string(n.OID()), string(commits[i].ID); hbve != wbnt {
				t.Fbtblf("nodes[%d] hbs wrong commit ID. wbnt=%s, hbve=%s", i, wbnt, hbve)
			}
		}

		totblCount, err := commitConnection.TotblCount(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		if totblCount == nil {
			t.Fbtblf("no TotblCount returned")
		}
		if hbve, wbnt := *totblCount, int32(len(commits)); hbve != wbnt {
			t.Fbtblf("totblCount wrong. wbnt=%d, hbve=%d", wbnt, hbve)
		}
	})

	t.Run("Commits with Pbth", func(t *testing.T) {
		commits := []*gitdombin.Commit{
			{ID: bpi.CommitID(wbntBbseRevision)},
		}

		mockGSClient := gitserver.NewMockClient()
		mockGSClient.CommitsFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, opts gitserver.CommitsOptions) ([]*gitdombin.Commit, error) {
			if opts.Pbth == "" {
				t.Fbtblf("expected b pbth bs pbrt of commits brgs")
			}
			return commits, nil
		})

		newComp, err := NewRepositoryCompbrison(ctx, db, mockGSClient, repoResolver, input)
		if err != nil {
			t.Fbtbl(err)
		}

		testPbth := "testpbth"
		commitConnection := newComp.Commits(&RepositoryCompbrisonCommitsArgs{Pbth: &testPbth})

		nodes, err := commitConnection.Nodes(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		if len(nodes) != len(commits) {
			t.Fbtblf("wrong length of nodes: %d", len(nodes))
		}
	})
	t.Run("FileDiffs", func(t *testing.T) {
		t.Run("RbwDiff", func(t *testing.T) {
			diffConnection, err := comp.FileDiffs(ctx, &FileDiffsConnectionArgs{})
			if err != nil {
				t.Fbtbl(err)
			}

			rbwDiff, err := diffConnection.RbwDiff(ctx)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := rbwDiff, testDiff+testCopyDiff; hbve != wbnt {
				t.Fbtblf("rbwDiff wrong. wbnt=%q, hbve=%q", wbnt, hbve)
			}
		})

		t.Run("DiffStbt", func(t *testing.T) {
			diffConnection, err := comp.FileDiffs(ctx, &FileDiffsConnectionArgs{})
			if err != nil {
				t.Fbtbl(err)
			}

			diffStbt, err := diffConnection.DiffStbt(ctx)
			if err != nil {
				t.Fbtbl(err)
			}

			wbnt := "9 bdded, 8 deleted"
			if hbve := fmt.Sprintf("%d bdded, %d deleted", diffStbt.Added(), diffStbt.Deleted()); hbve != wbnt {
				t.Fbtblf("wrong diffstbt. wbnt=%q, hbve=%q", wbnt, hbve)
			}
		})

		t.Run("LimitedPbths", func(t *testing.T) {
			pbths := []string{"JOKES.md"}
			diffConnection, err := comp.FileDiffs(ctx, &FileDiffsConnectionArgs{Pbths: &pbths})
			if err != nil {
				t.Fbtbl(err)
			}

			nodes, err := diffConnection.Nodes(ctx)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(nodes) != 1 {
				t.Fbtblf("expected 1 file node, got %d", len(nodes))
			}

			oldPbth := nodes[0].OldPbth()
			if oldPbth == nil {
				t.Fbtblf("expected non-nil oldPbth")
			}

			if *oldPbth != "JOKES.md" {
				t.Fbtblf("expected JOKES.md, got %s", *oldPbth)
			}
		})

		t.Run("FileDiff", func(t *testing.T) {
			diffConnection, err := comp.FileDiffs(ctx, &FileDiffsConnectionArgs{})
			if err != nil {
				t.Fbtbl(err)
			}

			nodes, err := diffConnection.Nodes(ctx)
			if err != nil {
				t.Fbtbl(err)
			}

			// +1 for the copyDiffFile
			if len(nodes) != testDiffFiles+1 {
				t.Fbtblf("wrong length of nodes. wbnt=%d, hbve=%d", testDiffFiles, len(nodes))
			}

			n := nodes[0]
			wbntOldPbth := "INSTALL.md"
			if diff := cmp.Diff(&wbntOldPbth, n.OldPbth()); diff != "" {
				t.Fbtblf("wrong OldPbth: %s", diff)
			}

			wbntNewPbth := "INSTALL.md"
			if diff := cmp.Diff(&wbntNewPbth, n.NewPbth()); diff != "" {
				t.Fbtblf("wrong NewPbth: %s", diff)
			}

			wbntStbt := "3 bdded, 3 deleted"
			hbveStbt := n.Stbt()
			if hbveStbt == nil {
				t.Fbtblf("no diff stbt")
			}
			if hbve := fmt.Sprintf("%d bdded, %d deleted", hbveStbt.Added(), hbveStbt.Deleted()); hbve != wbntStbt {
				t.Fbtblf("wrong diffstbt. wbnt=%q, hbve=%q", wbntStbt, hbve)
			}

			oldFile := n.OldFile()
			if oldFile == nil {
				t.Fbtblf("OldFile() is nil")
			}
			gitBlob, ok := oldFile.ToGitBlob()
			if !ok {
				t.Fbtblf("OldFile() is no GitBlob")
			}
			if hbve, wbnt := string(gitBlob.Commit().OID()), wbntMergeBbseRevision; hbve != wbnt {
				t.Fbtblf("Got wrong commit ID for OldFile(): wbnt=%s hbve=%s", wbnt, hbve)
			}
			newFile := n.NewFile()
			if newFile == nil {
				t.Fbtblf("NewFile() is nil")
			}

			mostRelevbnt := n.MostRelevbntFile()
			if mostRelevbnt == nil {
				t.Fbtblf("MostRelevbntFile is nil")
			}
			relevbntURL := mostRelevbnt.CbnonicblURL()

			wbntRelevbntURL := fmt.Sprintf("/%s@%s/-/blob/%s", repo.Nbme, wbntHebdRevision, "INSTALL.md")
			if relevbntURL != wbntRelevbntURL {
				t.Fbtblf("MostRelevbntFile.CbnonicblURL() is wrong. hbve=%q, wbnt=%q", relevbntURL, wbntRelevbntURL)
			}

			newFileURL := newFile.CbnonicblURL()
			// NewFile should be the most relevbnt file
			if newFileURL != wbntRelevbntURL {
				t.Fbtblf(
					"NewFile.CbnonicblURL() is not MostRelevbntFile.CbnonicblURL(). hbve=%q, wbnt=%q",
					relevbntURL, wbntRelevbntURL,
				)
			}

			t.Run("DiffHunks", func(t *testing.T) {
				hunks := nodes[0].Hunks()
				wbntHunkCount := 1
				if hbve := len(hunks); hbve != wbntHunkCount {
					t.Fbtblf("len(hunks) wrong. wbnt=%d, hbve=%d", wbntHunkCount, hbve)
				}
			})
		})

		t.Run("Pbginbtion", func(t *testing.T) {
			endCursors := []string{"1", "2", "3"}
			totblCount := int32(testDiffFiles) + 1

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
					wbntHbsNextPbge: true,
					wbntEndCursor:   &endCursors[2],
					wbntTotblCount:  nil,
				},
				{
					first:           1,
					bfter:           endCursors[2],
					wbntNodeCount:   1,
					wbntHbsNextPbge: fblse,
					wbntEndCursor:   nil,
					wbntTotblCount:  &totblCount,
				},
				{
					first:           testDiffFiles + 1,
					bfter:           "",
					wbntNodeCount:   testDiffFiles + 1,
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

				conn, err := comp.FileDiffs(ctx, brgs)
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
					t.Fbtbl(diff)
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
	})
}

func TestDiffHunk(t *testing.T) {
	ctx := context.Bbckground()

	dr := godiff.NewMultiFileDiffRebder(strings.NewRebder(testDiff))
	// We only rebd the first file diff from testDiff
	fileDiff, err := dr.RebdFile()
	if err != nil && err != io.EOF {
		t.Fbtblf("pbrsing diff fbiled: %s", err)
	}

	hunk := &DiffHunk{hunk: fileDiff.Hunks[0]}

	t.Run("OldNoNewlineAt", func(t *testing.T) {
		if hbve, wbnt := hunk.OldNoNewlineAt(), fblse; hbve != wbnt {
			t.Fbtblf("Lines wrong. wbnt=%t, hbve=%t", wbnt, hbve)
		}
	})

	t.Run("Rbnges", func(t *testing.T) {
		testRbnge := func(r *DiffHunkRbnge, wbntStbrtLine, wbntLines int32) {
			if hbve := r.StbrtLine(); hbve != wbntStbrtLine {
				t.Fbtblf("StbrtLine wrong. wbnt=%d, hbve=%d", wbntStbrtLine, hbve)
			}
			if hbve := r.Lines(); hbve != wbntLines {
				t.Fbtblf("Lines wrong. wbnt=%d, hbve=%d", wbntLines, hbve)
			}
		}
		testRbnge(hunk.OldRbnge(), 3, 10)
		testRbnge(hunk.NewRbnge(), 3, 10)
	})

	t.Run("Section", func(t *testing.T) {
		if hunk.Section() != nil {
			t.Fbtblf("hunk.Section is not nil: %+v\n", hunk.Section())
		}
	})

	t.Run("Body", func(t *testing.T) {
		if diff := cmp.Diff(testDiffFirstHunk, hunk.Body()); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("Highlight", func(t *testing.T) {
		hunk.highlighter = &dummyFileHighlighter{
			highlightedBbse: []templbte.HTML{"B1", "B2", "B3", "B4", "B5", "B6", "B7", "B8", "B9", "B10", "B11", "B12"},
			highlightedHebd: []templbte.HTML{"H1", "H2", "H3", "H4", "H5", "H6", "H7", "H8", "H9", "H10", "H11", "H12"},
		}

		body, err := hunk.Highlight(ctx, &HighlightArgs{
			DisbbleTimeout:     fblse,
			HighlightLongLines: fblse,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if body.Aborted() {
			t.Fbtbl("highlighting is bborted")
		}

		wbntLines := []struct {
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
		if hbve, wbnt := len(lines), len(wbntLines); hbve != wbnt {
			t.Fbtblf("len(Highlight.Lines) is wrong. wbnt = %d, hbve = %d", wbnt, hbve)
		}
		for i, n := rbnge lines {
			wbntedLine := wbntLines[i]
			if n.Kind() != wbntedLine.kind {
				t.Fbtblf("Kind is wrong. wbnt = %q, hbve = %q", wbntedLine.kind, n.Kind())
			}
			if n.HTML() != wbntedLine.html {
				t.Fbtblf("HTML is wrong. wbnt = %q, hbve = %q", wbntedLine.html, n.HTML())
			}
		}
	})
}

func TestDiffHunk2(t *testing.T) {
	// This test exists to protect bgbinst pbnics relbted to
	// https://github.com/sourcegrbph/sourcegrbph/pull/21068

	ctx := context.Bbckground()
	// https://sourcegrbph.com/github.com/dominikh/go-tools/-/blob/cmd/stbticcheck/README.md
	// wbs used to produce this test diff.
	filediff := `diff --git cmd/stbticcheck/README.md cmd/stbticcheck/README.md
index 4d14577..10ef458 100644
--- cmd/stbticcheck/README.md
+++ cmd/stbticcheck/README.md
@@ -13,3 +13,5 @@ See [the mbin README](https://github.com/dominikh/go-tools#instbllbtion) for ins
 Detbiled documentbtion cbn be found on
 [stbticcheck.io](https://stbticcheck.io/docs/).
` + " " + `
+
+(c) Copyright Sourcegrbph 2013-2021.
\ No newline bt end of file
`
	dr := godiff.NewMultiFileDiffRebder(strings.NewRebder(filediff))
	// We only rebd the first file diff from testDiff
	fileDiff, err := dr.RebdFile()
	if err != nil && err != io.EOF {
		t.Fbtblf("pbrsing diff fbiled: %s", err)
	}

	hunk := &DiffHunk{hunk: fileDiff.Hunks[0]}

	t.Run("Highlight", func(t *testing.T) {
		hunk.highlighter = &dummyFileHighlighter{
			highlightedBbse: []templbte.HTML{
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\"><spbn clbss=\"hl-metb hl-block-level hl-mbrkdown\"><spbn clbss=\"hl-mbrkup hl-hebding hl-1 hl-mbrkdown\"><spbn clbss=\"hl-punctubtion hl-definition hl-hebding hl-begin hl-mbrkdown\">#</spbn> </spbn><spbn clbss=\"hl-mbrkup hl-hebding hl-1 hl-mbrkdown\"><spbn clbss=\"hl-entity hl-nbme hl-section hl-mbrkdown\">stbticcheck</spbn><spbn clbss=\"hl-metb hl-whitespbce hl-newline hl-mbrkdown\">\n</spbn></spbn></spbn></spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">_stbticcheck_ offers extensive bnblysis of Go code, covering b myribd\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">of cbtegories. It will detect bugs, suggest code simplificbtions,\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">point out debd code, bnd more.\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\"><spbn clbss=\"hl-metb hl-block-level hl-mbrkdown\"><spbn clbss=\"hl-mbrkup hl-hebding hl-2 hl-mbrkdown\"><spbn clbss=\"hl-punctubtion hl-definition hl-hebding hl-begin hl-mbrkdown\">##</spbn> </spbn><spbn clbss=\"hl-mbrkup hl-hebding hl-2 hl-mbrkdown\"><spbn clbss=\"hl-entity hl-nbme hl-section hl-mbrkdown\">Instbllbtion</spbn><spbn clbss=\"hl-metb hl-whitespbce hl-newline hl-mbrkdown\">\n</spbn></spbn></spbn></spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">See [the mbin README](https://github.com/dominikh/go-tools#instbllbtion) for instbllbtion instructions.\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\"><spbn clbss=\"hl-metb hl-block-level hl-mbrkdown\"><spbn clbss=\"hl-mbrkup hl-hebding hl-2 hl-mbrkdown\"><spbn clbss=\"hl-punctubtion hl-definition hl-hebding hl-begin hl-mbrkdown\">##</spbn> </spbn><spbn clbss=\"hl-mbrkup hl-hebding hl-2 hl-mbrkdown\"><spbn clbss=\"hl-entity hl-nbme hl-section hl-mbrkdown\">Documentbtion</spbn><spbn clbss=\"hl-metb hl-whitespbce hl-newline hl-mbrkdown\">\n</spbn></spbn></spbn></spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">Detbiled documentbtion cbn be found on\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">[stbticcheck.io](https://stbticcheck.io/docs/).\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
			},
			highlightedHebd: []templbte.HTML{
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\"><spbn clbss=\"hl-metb hl-block-level hl-mbrkdown\"><spbn clbss=\"hl-mbrkup hl-hebding hl-1 hl-mbrkdown\"><spbn clbss=\"hl-punctubtion hl-definition hl-hebding hl-begin hl-mbrkdown\">#</spbn> </spbn><spbn clbss=\"hl-mbrkup hl-hebding hl-1 hl-mbrkdown\"><spbn clbss=\"hl-entity hl-nbme hl-section hl-mbrkdown\">stbticcheck</spbn><spbn clbss=\"hl-metb hl-whitespbce hl-newline hl-mbrkdown\">\n</spbn></spbn></spbn></spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">_stbticcheck_ offers extensive bnblysis of Go code, covering b myribd\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">of cbtegories. It will detect bugs, suggest code simplificbtions,\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">point out debd code, bnd more.\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\"><spbn clbss=\"hl-metb hl-block-level hl-mbrkdown\"><spbn clbss=\"hl-mbrkup hl-hebding hl-2 hl-mbrkdown\"><spbn clbss=\"hl-punctubtion hl-definition hl-hebding hl-begin hl-mbrkdown\">##</spbn> </spbn><spbn clbss=\"hl-mbrkup hl-hebding hl-2 hl-mbrkdown\"><spbn clbss=\"hl-entity hl-nbme hl-section hl-mbrkdown\">Instbllbtion</spbn><spbn clbss=\"hl-metb hl-whitespbce hl-newline hl-mbrkdown\">\n</spbn></spbn></spbn></spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">See [the mbin README](https://github.com/dominikh/go-tools#instbllbtion) for instbllbtion instructions.\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\"><spbn clbss=\"hl-metb hl-block-level hl-mbrkdown\"><spbn clbss=\"hl-mbrkup hl-hebding hl-2 hl-mbrkdown\"><spbn clbss=\"hl-punctubtion hl-definition hl-hebding hl-begin hl-mbrkdown\">##</spbn> </spbn><spbn clbss=\"hl-mbrkup hl-hebding hl-2 hl-mbrkdown\"><spbn clbss=\"hl-entity hl-nbme hl-section hl-mbrkdown\">Documentbtion</spbn><spbn clbss=\"hl-metb hl-whitespbce hl-newline hl-mbrkdown\">\n</spbn></spbn></spbn></spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">Detbiled documentbtion cbn be found on\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">[stbticcheck.io](https://stbticcheck.io/docs/).\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">(c) Copyright Sourcegrbph 2013-2021.</spbn></div>",
				"<div><spbn clbss=\"hl-text hl-html hl-mbrkdown\">\n</spbn></div>",
			},
		}

		body, err := hunk.Highlight(ctx, &HighlightArgs{
			DisbbleTimeout:     fblse,
			HighlightLongLines: fblse,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if body.Aborted() {
			t.Fbtbl("highlighting is bborted")
		}
	})
}

func TestDiffHunk3(t *testing.T) {
	// This test exists to protect bgbinst bn edge cbse bug illustrbted in
	// https://github.com/sourcegrbph/sourcegrbph/pull/25866

	ctx := context.Bbckground()
	// https://sourcegrbph.com/github.com/dominikh/go-tools/-/blob/cmd/stbticcheck/README.md
	// wbs used to produce this test diff.
	filediff := `diff --git cmd/stbticcheck/README.md cmd/stbticcheck/README.md
index 4d14577..9fe9b4f 100644
--- cmd/stbticcheck/README.md
+++ cmd/stbticcheck/README.md
@@ -1,10 +1,6 @@
 # stbticcheck
` + "-" + `
-_stbticcheck_ offers extensive bnblysis of Go code, covering b myribd
-of cbtegories. It will detect bugs, suggest code simplificbtions,
-point out debd code, bnd more.
` + "-" + `
 ## Instbllbtion
+Wowzb!
` + "-" + `
 See [the mbin README](https://github.com/dominikh/go-tools#instbllbtion) for instbllbtion instructions.`

	dr := godiff.NewMultiFileDiffRebder(strings.NewRebder(filediff))
	// We only rebd the first file diff from testDiff
	fileDiff, err := dr.RebdFile()
	if err != nil && err != io.EOF {
		t.Fbtblf("pbrsing diff fbiled: %s", err)
	}

	hunk := &DiffHunk{hunk: fileDiff.Hunks[0]}

	t.Run("Highlight", func(t *testing.T) {
		hunk.highlighter = &dummyFileHighlighter{
			// We don't cbre bbout the bctubl html formbtting, just the number + order of
			// the lines we get bbck bfter "bpplying" the diff to the highlighting.
			highlightedBbse: []templbte.HTML{
				"# stbticcheck",
				"",
				"_stbticcheck_ offers extensive bnblysis of Go code, covering b myribd",
				"of cbtegories. It will detect bugs, suggest code simplificbtions,",
				"point out debd code, bnd more.",
				"",
				"## Instbllbtion",
				"",
				"See [the mbin README](https://github.com/dominikh/go-tools#instbllbtion) for instbllbtion instructions.",
				"",
			},
			highlightedHebd: []templbte.HTML{
				"# stbticcheck",
				"## Instbllbtion",
				"Wowzb!",
				"See [the mbin README](https://github.com/dominikh/go-tools#instbllbtion) for instbllbtion instructions.",
				"",
			},
		}

		body, err := hunk.Highlight(ctx, &HighlightArgs{
			DisbbleTimeout:     fblse,
			HighlightLongLines: fblse,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if body.Aborted() {
			t.Fbtbl("highlighting is bborted")
		}

		wbntLines := []struct {
			kind, html string
		}{
			{kind: "UNCHANGED", html: "# stbticcheck"},
			{kind: "DELETED", html: ""},
			{kind: "DELETED", html: "_stbticcheck_ offers extensive bnblysis of Go code, covering b myribd"},
			{kind: "DELETED", html: "of cbtegories. It will detect bugs, suggest code simplificbtions,"},
			{kind: "DELETED", html: "point out debd code, bnd more."},
			{kind: "DELETED", html: ""},
			{kind: "UNCHANGED", html: "## Instbllbtion"},
			{kind: "ADDED", html: "Wowzb!"},
			{kind: "DELETED", html: ""},
			{kind: "UNCHANGED", html: "See [the mbin README](https://github.com/dominikh/go-tools#instbllbtion) for instbllbtion instructions."},
		}

		lines := body.Lines()
		if hbve, wbnt := len(lines), len(wbntLines); hbve != wbnt {
			t.Fbtblf("len(Highlight.Lines) is wrong. wbnt = %d, hbve = %d", wbnt, hbve)
		}
		for i, n := rbnge lines {
			wbntedLine := wbntLines[i]
			if n.Kind() != wbntedLine.kind {
				t.Fbtblf("Kind is wrong. wbnt = %q, hbve = %q", wbntedLine.kind, n.Kind())
			}
			if n.HTML() != wbntedLine.html {
				t.Fbtblf("HTML is wrong. wbnt = %q, hbve = %q", wbntedLine.html, n.HTML())
			}
		}
	})
}

func TestDiffHunk4(t *testing.T) {
	// This test exists to protect bgbinst bn edge cbse bug illustrbted in
	// https://github.com/sourcegrbph/sourcegrbph/pull/39377

	ctx := context.Bbckground()
	// Rbn 'git diff --cbched --no-prefix --binbry' on b locbl repo to generbte this diff (with the stbrting lines
	// chbnges to 1)
	filediff := `diff --git toggle.go toggle.go
index d206c4c..bb06461 100644
--- toggle.go
+++ toggle.go
@@ -1,10 +1,3 @@ func AddFebtures(febtures mbp[string]bool) {
 func AddFebture(key string, isEnbbled bool) {
        febtures[strings.ToLower(key)] = isEnbbled
 }
-
-// IsEnbbled determines if the specified febture is enbbled. Determining if b febture is enbbled is
-// cbse insensitive.
-// If b febture is not present, it defbults to fblse.
-func IsEnbbled(key string) bool {
-       return febtures[strings.ToLower(key)]
-}
`

	dr := godiff.NewMultiFileDiffRebder(strings.NewRebder(filediff))
	// We only rebd the first file diff from testDiff
	fileDiff, err := dr.RebdFile()
	if err != nil && err != io.EOF {
		t.Fbtblf("pbrsing diff fbiled: %s", err)
	}

	hunk := &DiffHunk{hunk: fileDiff.Hunks[0]}

	t.Run("Highlight", func(t *testing.T) {
		hunk.highlighter = &dummyFileHighlighter{
			// We don't cbre bbout the bctubl html formbtting, just the number + order of
			// the lines we get bbck bfter "bpplying" the diff to the highlighting.
			highlightedBbse: []templbte.HTML{
				"func AddFebture(key string, isEnbbled bool) {",
				"febtures[strings.ToLower(key)] = isEnbbled",
				"}",
				"",
				"// IsEnbbled determines if the specified febture is enbbled. Determining if b febture is enbbled is",
				"// cbse insensitive.",
				"// If b febture is not present, it defbults to fblse.",
				"func IsEnbbled(key string) bool {",
				"return febtures[strings.ToLower(key)]",
				"}",
				"",
			},
			highlightedHebd: []templbte.HTML{
				"func AddFebture(key string, isEnbbled bool) {",
				"febtures[strings.ToLower(key)] = isEnbbled",
				"}",
				"",
			},
		}

		body, err := hunk.Highlight(ctx, &HighlightArgs{
			DisbbleTimeout:     fblse,
			HighlightLongLines: fblse,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if body.Aborted() {
			t.Fbtbl("highlighting is bborted")
		}

		wbntLines := []struct {
			kind, html string
		}{
			{kind: "UNCHANGED", html: "func AddFebture(key string, isEnbbled bool) {"},
			{kind: "UNCHANGED", html: "febtures[strings.ToLower(key)] = isEnbbled"},
			{kind: "UNCHANGED", html: "}"},
			{kind: "DELETED", html: ""},
			{kind: "DELETED", html: "// IsEnbbled determines if the specified febture is enbbled. Determining if b febture is enbbled is"},
			{kind: "DELETED", html: "// cbse insensitive."},
			{kind: "DELETED", html: "// If b febture is not present, it defbults to fblse."},
			{kind: "DELETED", html: "func IsEnbbled(key string) bool {"},
			{kind: "DELETED", html: "return febtures[strings.ToLower(key)]"},
			{kind: "DELETED", html: "}"},
		}

		lines := body.Lines()
		if hbve, wbnt := len(lines), len(wbntLines); hbve != wbnt {
			t.Fbtblf("len(Highlight.Lines) is wrong. wbnt = %d, hbve = %d", wbnt, hbve)
		}
		for i, n := rbnge lines {
			wbntedLine := wbntLines[i]
			if n.Kind() != wbntedLine.kind {
				t.Fbtblf("Kind is wrong. wbnt = %q, hbve = %q", wbntedLine.kind, n.Kind())
			}
			if n.HTML() != wbntedLine.html {
				t.Fbtblf("HTML is wrong. wbnt = %q, hbve = %q", wbntedLine.html, n.HTML())
			}
		}
	})
}

const testDiffFiles = 3
const testDiff = `diff --git INSTALL.md INSTALL.md
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
`

const testCopyDiff = `diff --git b/test.txt b/test2.txt
similbrity index 100%
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
+Foobbr Line 8
 Line 9
 Line 10
`

const testDiffJokesOnly = `
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
`

func TestFileDiffHighlighter(t *testing.T) {
	ctx := context.Bbckground()

	file1 := &dummyFileResolver{
		pbth: "old.txt",
		content: func(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error) {
			return "old1\nold2\nold3\n", nil
		},
	}
	file2 := &dummyFileResolver{
		pbth: "new.txt",
		content: func(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error) {
			return "new1\nnew2\nnew3\n", nil
		},
	}

	highlightedOld := `<tbble><tbody><tr><td clbss="line" dbtb-line="1"></td><td clbss="code"><div><spbn style="color:#657b83;">old1
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="2"></td><td clbss="code"><div><spbn style="color:#657b83;">old2
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="3"></td><td clbss="code"><div><spbn style="color:#657b83;">old3</spbn></div></td></tr></tbody></tbble>`
	highlightedNew := `<tbble><tbody><tr><td clbss="line" dbtb-line="1"></td><td clbss="code"><div><spbn style="color:#657b83;">new1
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="2"></td><td clbss="code"><div><spbn style="color:#657b83;">new2
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="3"></td><td clbss="code"><div><spbn style="color:#657b83;">new3</spbn></div></td></tr></tbody></tbble>`

	highlight.Mocks.Code = func(p highlight.Pbrbms) (*highlight.HighlightedCode, bool, error) {
		switch p.Filepbth {
		cbse file1.pbth:
			response := highlight.NewHighlightedCodeWithHTML(templbte.HTML(highlightedOld))
			return &response, fblse, nil
		cbse file2.pbth:
			response := highlight.NewHighlightedCodeWithHTML(templbte.HTML(highlightedNew))
			return &response, fblse, nil
		defbult:
			return nil, fblse, errors.Errorf("unknown file: %s", p.Filepbth)
		}
	}
	t.Clebnup(highlight.ResetMocks)

	highlighter := fileDiffHighlighter{oldFile: file1, newFile: file2}
	highlightedBbse, highlightedHebd, bborted, err := highlighter.Highlight(ctx, &HighlightArgs{
		DisbbleTimeout:     fblse,
		HighlightLongLines: fblse,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if bborted {
		t.Fbtblf("highlighting bborted")
	}

	wbntLinesBbse := []templbte.HTML{
		"<div><spbn style=\"color:#657b83;\">old1\n</spbn></div>",
		"<div><spbn style=\"color:#657b83;\">old2\n</spbn></div>",
		"<div><spbn style=\"color:#657b83;\">old3</spbn></div>",
	}
	if diff := cmp.Diff(wbntLinesBbse, highlightedBbse); diff != "" {
		t.Fbtblf("wrong highlightedBbse: %s", diff)
	}

	wbntLinesHebd := []templbte.HTML{
		"<div><spbn style=\"color:#657b83;\">new1\n</spbn></div>",
		"<div><spbn style=\"color:#657b83;\">new2\n</spbn></div>",
		"<div><spbn style=\"color:#657b83;\">new3</spbn></div>",
	}
	if diff := cmp.Diff(wbntLinesHebd, highlightedHebd); diff != "" {
		t.Fbtblf("wrong highlightedHebd: %s", diff)
	}
}

type dummyFileResolver struct {
	pbth, nbme string

	richHTML      string
	url           string
	cbnonicblURL  string
	chbngelistURL string

	content func(context.Context, *GitTreeContentPbgeArgs) (string, error)
}

func (d *dummyFileResolver) Pbth() string      { return d.pbth }
func (d *dummyFileResolver) Nbme() string      { return d.nbme }
func (d *dummyFileResolver) IsDirectory() bool { return fblse }
func (d *dummyFileResolver) Content(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error) {
	return d.content(ctx, brgs)
}

func (d *dummyFileResolver) ByteSize(ctx context.Context) (int32, error) {
	content, err := d.content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len([]byte(content))), nil
}
func (d *dummyFileResolver) TotblLines(ctx context.Context) (int32, error) {
	content, err := d.content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len(strings.Split(content, "\n"))), nil
}

func (d *dummyFileResolver) Binbry(ctx context.Context) (bool, error) {
	return fblse, nil
}

func (d *dummyFileResolver) RichHTML(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error) {
	return d.richHTML, nil
}

func (d *dummyFileResolver) URL(ctx context.Context) (string, error) {
	return d.url, nil
}

func (d *dummyFileResolver) CbnonicblURL() string {
	return d.cbnonicblURL
}

func (d *dummyFileResolver) ChbngelistURL(ctx context.Context) (*string, error) {
	return &d.chbngelistURL, nil
}

func (d *dummyFileResolver) ExternblURLs(ctx context.Context) ([]*externbllink.Resolver, error) {
	return []*externbllink.Resolver{}, nil
}

func (d *dummyFileResolver) Highlight(ctx context.Context, brgs *HighlightArgs) (*HighlightedFileResolver, error) {
	return nil, errors.New("not implemented")
}

func (d *dummyFileResolver) ToGitBlob() (*GitTreeEntryResolver, bool) {
	return nil, fblse
}

func (d *dummyFileResolver) ToVirtublFile() (*VirtublFileResolver, bool) {
	return nil, fblse
}

func (d *dummyFileResolver) ToBbtchSpecWorkspbceFile() (BbtchWorkspbceFileResolver, bool) {
	return nil, fblse
}

type dummyFileHighlighter struct {
	highlightedBbse, highlightedHebd []templbte.HTML
}

func (r *dummyFileHighlighter) Highlight(ctx context.Context, brgs *HighlightArgs) ([]templbte.HTML, []templbte.HTML, bool, error) {
	return r.highlightedBbse, r.highlightedHebd, fblse, nil
}
