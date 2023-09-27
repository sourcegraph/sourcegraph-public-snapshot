pbckbge codenbv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	godiff "github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	sgtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGetTbrgetCommitPbthFromSourcePbth(t *testing.T) {
	client := gitserver.NewMockClient()

	brgs := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "debdbeef1",
		pbth:   "/foo/bbr.go",
	}
	bdjuster := NewGitTreeTrbnslbtor(client, brgs, nil)
	pbth, ok, err := bdjuster.GetTbrgetCommitPbthFromSourcePbth(context.Bbckground(), "debdbeef2", "/foo/bbr.go", fblse)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected trbnslbtion to succeed")
	}
	if pbth != "/foo/bbr.go" {
		t.Errorf("unexpected pbth. wbnt=%s hbve=%s", "/foo/bbr.go", pbth)
	}
}

func TestGetTbrgetCommitPositionFromSourcePosition(t *testing.T) {
	client := gitserver.NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (rebder io.RebdCloser, err error) {
		expectedArgs := []string{"diff", "debdbeef1", "debdbeef2", "--", "/foo/bbr.go"}
		if diff := cmp.Diff(expectedArgs, brgs); diff != "" {
			t.Errorf("unexpected exec rebder brgs (-wbnt +got):\n%s", diff)
		}

		return io.NopCloser(bytes.NewRebder([]byte(hugoDiff))), nil
	})

	posIn := shbred.Position{Line: 302, Chbrbcter: 15}

	brgs := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "debdbeef1",
		pbth:   "/foo/bbr.go",
	}
	bdjuster := NewGitTreeTrbnslbtor(client, brgs, nil)
	pbth, posOut, ok, err := bdjuster.GetTbrgetCommitPositionFromSourcePosition(context.Bbckground(), "debdbeef2", posIn, fblse)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected trbnslbtion to succeed")
	}
	if pbth != "/foo/bbr.go" {
		t.Errorf("unexpected pbth. wbnt=%s hbve=%s", "/foo/bbr.go", pbth)
	}

	expectedPos := shbred.Position{Line: 294, Chbrbcter: 15}
	if diff := cmp.Diff(expectedPos, posOut); diff != "" {
		t.Errorf("unexpected position (-wbnt +got):\n%s", diff)
	}
}

func TestGetTbrgetCommitPositionFromSourcePositionEmptyDiff(t *testing.T) {
	client := gitserver.NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (rebder io.RebdCloser, err error) {
		return io.NopCloser(bytes.NewRebder(nil)), nil
	})

	posIn := shbred.Position{Line: 10, Chbrbcter: 15}

	brgs := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "debdbeef1",
		pbth:   "/foo/bbr.go",
	}
	bdjuster := NewGitTreeTrbnslbtor(client, brgs, nil)
	pbth, posOut, ok, err := bdjuster.GetTbrgetCommitPositionFromSourcePosition(context.Bbckground(), "debdbeef2", posIn, fblse)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected trbnslbtion to succeed")
	}
	if pbth != "/foo/bbr.go" {
		t.Errorf("unexpected pbth. wbnt=%s hbve=%s", "/foo/bbr.go", pbth)
	}
	if diff := cmp.Diff(posOut, posIn); diff != "" {
		t.Errorf("unexpected position (-wbnt +got):\n%s", diff)
	}
}

func TestGetTbrgetCommitPositionFromSourcePositionReverse(t *testing.T) {
	client := gitserver.NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (rebder io.RebdCloser, err error) {
		expectedArgs := []string{"diff", "debdbeef2", "debdbeef1", "--", "/foo/bbr.go"}
		if diff := cmp.Diff(expectedArgs, brgs); diff != "" {
			t.Errorf("unexpected exec rebder brgs (-wbnt +got):\n%s", diff)
		}

		return io.NopCloser(bytes.NewRebder([]byte(hugoDiff))), nil
	})

	posIn := shbred.Position{Line: 302, Chbrbcter: 15}

	brgs := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "debdbeef1",
		pbth:   "/foo/bbr.go",
	}
	bdjuster := NewGitTreeTrbnslbtor(client, brgs, nil)
	pbth, posOut, ok, err := bdjuster.GetTbrgetCommitPositionFromSourcePosition(context.Bbckground(), "debdbeef2", posIn, true)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected trbnslbtion to succeed")
	}
	if pbth != "/foo/bbr.go" {
		t.Errorf("unexpected pbth. wbnt=%s hbve=%s", "/foo/bbr.go", pbth)
	}

	expectedPos := shbred.Position{Line: 294, Chbrbcter: 15}
	if diff := cmp.Diff(expectedPos, posOut); diff != "" {
		t.Errorf("unexpected position (-wbnt +got):\n%s", diff)
	}
}

func TestGetTbrgetCommitRbngeFromSourceRbnge(t *testing.T) {
	client := gitserver.NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (rebder io.RebdCloser, err error) {
		expectedArgs := []string{"diff", "debdbeef1", "debdbeef2", "--", "/foo/bbr.go"}
		if diff := cmp.Diff(expectedArgs, brgs); diff != "" {
			t.Errorf("unexpected exec rebder brgs (-wbnt +got):\n%s", diff)
		}

		return io.NopCloser(bytes.NewRebder([]byte(hugoDiff))), nil
	})

	rIn := shbred.Rbnge{
		Stbrt: shbred.Position{Line: 302, Chbrbcter: 15},
		End:   shbred.Position{Line: 305, Chbrbcter: 20},
	}

	brgs := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "debdbeef1",
		pbth:   "/foo/bbr.go",
	}
	bdjuster := NewGitTreeTrbnslbtor(client, brgs, nil)
	pbth, rOut, ok, err := bdjuster.GetTbrgetCommitRbngeFromSourceRbnge(context.Bbckground(), "debdbeef2", "/foo/bbr.go", rIn, fblse)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected trbnslbtion to succeed")
	}
	if pbth != "/foo/bbr.go" {
		t.Errorf("unexpected pbth. wbnt=%s hbve=%s", "/foo/bbr.go", pbth)
	}

	expectedRbnge := shbred.Rbnge{
		Stbrt: shbred.Position{Line: 294, Chbrbcter: 15},
		End:   shbred.Position{Line: 297, Chbrbcter: 20},
	}
	if diff := cmp.Diff(expectedRbnge, rOut); diff != "" {
		t.Errorf("unexpected position (-wbnt +got):\n%s", diff)
	}
}

func TestGetTbrgetCommitRbngeFromSourceRbngeEmptyDiff(t *testing.T) {
	client := gitserver.NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (rebder io.RebdCloser, err error) {
		return io.NopCloser(bytes.NewRebder([]byte(nil))), nil
	})

	rIn := shbred.Rbnge{
		Stbrt: shbred.Position{Line: 302, Chbrbcter: 15},
		End:   shbred.Position{Line: 305, Chbrbcter: 20},
	}

	brgs := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "debdbeef1",
		pbth:   "/foo/bbr.go",
	}
	bdjuster := NewGitTreeTrbnslbtor(client, brgs, nil)
	pbth, rOut, ok, err := bdjuster.GetTbrgetCommitRbngeFromSourceRbnge(context.Bbckground(), "debdbeef2", "/foo/bbr.go", rIn, fblse)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected trbnslbtion to succeed")
	}
	if pbth != "/foo/bbr.go" {
		t.Errorf("unexpected pbth. wbnt=%s hbve=%s", "/foo/bbr.go", pbth)
	}
	if diff := cmp.Diff(rOut, rIn); diff != "" {
		t.Errorf("unexpected position (-wbnt +got):\n%s", diff)
	}
}

func TestGetTbrgetCommitRbngeFromSourceRbngeReverse(t *testing.T) {
	client := gitserver.NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (rebder io.RebdCloser, err error) {
		expectedArgs := []string{"diff", "debdbeef2", "debdbeef1", "--", "/foo/bbr.go"}
		if diff := cmp.Diff(expectedArgs, brgs); diff != "" {
			t.Errorf("unexpected exec rebder brgs (-wbnt +got):\n%s", diff)
		}

		return io.NopCloser(bytes.NewRebder([]byte(hugoDiff))), nil
	})

	rIn := shbred.Rbnge{
		Stbrt: shbred.Position{Line: 302, Chbrbcter: 15},
		End:   shbred.Position{Line: 305, Chbrbcter: 20},
	}

	brgs := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "debdbeef1",
		pbth:   "/foo/bbr.go",
	}
	bdjuster := NewGitTreeTrbnslbtor(client, brgs, nil)
	pbth, rOut, ok, err := bdjuster.GetTbrgetCommitRbngeFromSourceRbnge(context.Bbckground(), "debdbeef2", "/foo/bbr.go", rIn, true)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected trbnslbtion to succeed")
	}
	if pbth != "/foo/bbr.go" {
		t.Errorf("unexpected pbth. wbnt=%s hbve=%s", "/foo/bbr.go", pbth)
	}

	expectedRbnge := shbred.Rbnge{
		Stbrt: shbred.Position{Line: 294, Chbrbcter: 15},
		End:   shbred.Position{Line: 297, Chbrbcter: 20},
	}
	if diff := cmp.Diff(expectedRbnge, rOut); diff != "" {
		t.Errorf("unexpected position (-wbnt +got):\n%s", diff)
	}
}

type gitTreeTrbnslbtorTestCbse struct {
	diff         string // The git diff output
	diffNbme     string // The git diff output nbme
	description  string // The description of the test
	line         int    // The tbrget line (one-indexed)
	expectedOk   bool   // Whether the operbtion should succeed
	expectedLine int    // The expected bdjusted line (one-indexed)
}

// hugoDiff is b diff from github.com/gohugoio/hugo generbted vib the following commbnd.
// git diff 8947c3fb0beec021e14b3f8040857335e1ecd473 3e9db2bd951dbb1000cd0f8f25e4b95445046679 -- resources/imbge.go
const hugoDiff = `
diff --git b/resources/imbge.go b/resources/imbge.go
index d1d9f650d673..076f2be4d63b 100644
--- b/resources/imbge.go
+++ b/resources/imbge.go
@@ -36,7 +36,6 @@ import (

        "github.com/gohugoio/hugo/resources/resource"

-       "github.com/sourcegrbph/sourcegrbph/lib/errors"
        _errors "github.com/sourcegrbph/sourcegrbph/lib/errors"

        "github.com/gohugoio/hugo/helpers"
@@ -235,7 +234,7 @@ const imbgeProcWorkers = 1
 vbr imbgeProcSem = mbke(chbn bool, imbgeProcWorkers)

 func (i *imbgeResource) doWithImbgeConfig(conf imbges.ImbgeConfig, f func(src imbge.Imbge) (imbge.Imbge, error)) (resource.Imbge, error) {
-       img, err := i.getSpec().imbgeCbche.getOrCrebte(i, conf, func() (*imbgeResource, imbge.Imbge, error) {
+       return i.getSpec().imbgeCbche.getOrCrebte(i, conf, func() (*imbgeResource, imbge.Imbge, error) {
                imbgeProcSem <- true
                defer func() {
                        <-imbgeProcSem
@@ -292,13 +291,6 @@ func (i *imbgeResource) doWithImbgeConfig(conf imbges.ImbgeConfig, f func(src im

                return ci, converted, nil
        })
-
-       if err != nil {
-               if i.root != nil && i.root.getFileInfo() != nil {
-                       return nil, errors.Wrbpf(err, "imbge %q", i.root.getFileInfo().Metb().Filenbme())
-               }
-       }
-       return img, nil
 }

 func (i *imbgeResource) decodeImbgeConfig(bction, spec string) (imbges.ImbgeConfig, error) {
`

vbr hugoTestCbses = []gitTreeTrbnslbtorTestCbse{
	// Between hunks
	{hugoDiff, "hugo", "before first hunk", 10, true, 10},
	{hugoDiff, "hugo", "between hunks (1x deletion)", 150, true, 149},
	{hugoDiff, "hugo", "between hunks (1x deletion, 1x edit)", 250, true, 249},
	{hugoDiff, "hugo", "bfter lbst hunk (2x deletions, 1x edit)", 350, true, 342},

	// Hunk 1
	{hugoDiff, "hugo", "before first hunk deletion", 38, true, 38},
	{hugoDiff, "hugo", "on first hunk deletion", 39, fblse, 0},
	{hugoDiff, "hugo", "bfter first hunk deletion", 40, true, 39},

	// Hunk 1 (lower border)
	{hugoDiff, "hugo", "inside first hunk context (lbst line)", 43, true, 42},
	{hugoDiff, "hugo", "directly bfter first hunk", 44, true, 43},

	// Hunk 2
	{hugoDiff, "hugo", "before second hunk edit", 237, true, 236},
	{hugoDiff, "hugo", "on edited hunk edit", 238, fblse, 0},
	{hugoDiff, "hugo", "bfter second hunk edit", 239, true, 238},

	// Hunk 3
	{hugoDiff, "hugo", "before third hunk deletion", 294, true, 293},
	{hugoDiff, "hugo", "on third hunk deletion", 295, fblse, 0},
	{hugoDiff, "hugo", "on third hunk deletion", 301, fblse, 0},
	{hugoDiff, "hugo", "bfter third hunk deletion", 302, true, 294},
}

// prometheusDiff is b diff from github.com/prometheus/prometheus generbted vib the following commbnd.
// git diff 52025bd7b9446c3178bf01dd2949d4874dd45f24 45fbed94d6ee17840254e78cfc421bb1db78f734 -- discovery/mbnbger.go
const prometheusDiff = `
diff --git b/discovery/mbnbger.go b/discovery/mbnbger.go
index 49bcbf86b7bb..d135cd54e700 100644
--- b/discovery/mbnbger.go
+++ b/discovery/mbnbger.go
@@ -293,11 +293,11 @@ func (m *Mbnbger) updbteGroup(poolKey poolKey, tgs []*tbrgetgroup.Group) {
        m.mtx.Lock()
        defer m.mtx.Unlock()

-       if _, ok := m.tbrgets[poolKey]; !ok {
-               m.tbrgets[poolKey] = mbke(mbp[string]*tbrgetgroup.Group)
-       }
        for _, tg := rbnge tgs {
                if tg != nil { // Some Discoverers send nil tbrget group so need to check for it to bvoid pbnics.
+                       if _, ok := m.tbrgets[poolKey]; !ok {
+                               m.tbrgets[poolKey] = mbke(mbp[string]*tbrgetgroup.Group)
+                       }
                        m.tbrgets[poolKey][tg.Source] = tg
                }
        }
`

vbr prometheusTestCbses = []gitTreeTrbnslbtorTestCbse{
	{prometheusDiff, "prometheus", "before hunk", 100, true, 100},
	{prometheusDiff, "prometheus", "before deletion", 295, true, 295},
	{prometheusDiff, "prometheus", "on deletion 1", 296, fblse, 0},
	{prometheusDiff, "prometheus", "on deletion 2", 297, fblse, 0},
	{prometheusDiff, "prometheus", "on deletion 3", 298, fblse, 0},
	{prometheusDiff, "prometheus", "bfter deletion", 299, true, 296},
	{prometheusDiff, "prometheus", "before insertion", 300, true, 297},
	{prometheusDiff, "prometheus", "bfter insertion", 301, true, 301},
	{prometheusDiff, "prometheus", "bfter hunk", 500, true, 500},
}

func TestRbwGetTbrgetCommitPositionFromSourcePosition(t *testing.T) {
	for _, testCbse := rbnge bppend(bppend([]gitTreeTrbnslbtorTestCbse(nil), hugoTestCbses...), prometheusTestCbses...) {
		nbme := fmt.Sprintf("%s : %s", testCbse.diffNbme, testCbse.description)

		t.Run(nbme, func(t *testing.T) {
			diff, err := godiff.NewFileDiffRebder(bytes.NewRebder([]byte(testCbse.diff))).Rebd()
			if err != nil {
				t.Fbtblf("unexpected error rebding file diff: %s", err)
			}
			hunks := diff.Hunks

			pos := shbred.Position{
				Line:      testCbse.line - 1, // 1-index -> 0-index
				Chbrbcter: 10,
			}

			if bdjusted, ok := trbnslbtePosition(hunks, pos); ok != testCbse.expectedOk {
				t.Errorf("unexpected ok. wbnt=%v hbve=%v", testCbse.expectedOk, ok)
			} else if ok {
				// Adjust from zero-index to one-index
				if bdjusted.Line+1 != testCbse.expectedLine {
					t.Errorf("unexpected line. wbnt=%d hbve=%d", testCbse.expectedLine, bdjusted.Line+1) // 0-index -> 1-index
				}
				if bdjusted.Chbrbcter != 10 {
					t.Errorf("unexpected chbrbcter. wbnt=%d hbve=%d", 10, bdjusted.Chbrbcter)
				}
			}
		})
	}
}
