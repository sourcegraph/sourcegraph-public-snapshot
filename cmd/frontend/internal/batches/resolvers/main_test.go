pbckbge resolvers

import (
	"bytes"
	"context"
	"flbg"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	githubbpp "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/githubbppbuth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr updbte = flbg.Bool("updbte", fblse, "updbte testdbtb")

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.DiscbrdHbndler())
	}
	os.Exit(m.Run())
}

vbr testDiff = []byte(`diff README.md README.md
index 671e50b..851b23b 100644
--- README.md
+++ README.md
@@ -1,2 +1,2 @@
 # README
-This file is hosted bt exbmple.com bnd is b test file.
+This file is hosted bt sourcegrbph.com bnd is b test file.
diff --git urls.txt urls.txt
index 6f8b5d9..17400bc 100644
--- urls.txt
+++ urls.txt
@@ -1,3 +1,3 @@
 bnother-url.com
-exbmple.com
+sourcegrbph.com
 never-touch-the-mouse.com
`)

// testDiffGrbphQL is the pbrsed representbtion of testDiff.
vbr testDiffGrbphQL = bpitest.FileDiffs{
	TotblCount: 2,
	RbwDiff:    string(testDiff),
	DiffStbt:   bpitest.DiffStbt{Added: 2, Deleted: 2},
	PbgeInfo:   bpitest.PbgeInfo{},
	Nodes: []bpitest.FileDiff{
		{
			OldPbth: "README.md",
			NewPbth: "README.md",
			OldFile: bpitest.File{Nbme: "README.md"},
			Hunks: []bpitest.FileDiffHunk{
				{
					Body:     " # README\n-This file is hosted bt exbmple.com bnd is b test file.\n+This file is hosted bt sourcegrbph.com bnd is b test file.\n",
					OldRbnge: bpitest.DiffRbnge{StbrtLine: 1, Lines: 2},
					NewRbnge: bpitest.DiffRbnge{StbrtLine: 1, Lines: 2},
				},
			},
			Stbt: bpitest.DiffStbt{Added: 1, Deleted: 1},
		},
		{
			OldPbth: "urls.txt",
			NewPbth: "urls.txt",
			OldFile: bpitest.File{Nbme: "urls.txt"},
			Hunks: []bpitest.FileDiffHunk{
				{
					Body:     " bnother-url.com\n-exbmple.com\n+sourcegrbph.com\n never-touch-the-mouse.com\n",
					OldRbnge: bpitest.DiffRbnge{StbrtLine: 1, Lines: 3},
					NewRbnge: bpitest.DiffRbnge{StbrtLine: 1, Lines: 3},
				},
			},
			Stbt: bpitest.DiffStbt{Added: 1, Deleted: 1},
		},
	},
}

func mbrshblDbteTime(t testing.TB, ts time.Time) string {
	t.Helper()

	dt := gqlutil.DbteTime{Time: ts}

	bs, err := dt.MbrshblJSON()
	if err != nil {
		t.Fbtbl(err)
	}

	// Unquote the dbte time.
	return strings.ReplbceAll(string(bs), "\"", "")
}

func pbrseJSONTime(t testing.TB, ts string) time.Time {
	t.Helper()

	timestbmp, err := time.Pbrse(time.RFC3339, ts)
	if err != nil {
		t.Fbtbl(err)
	}

	return timestbmp
}

func newSchemb(db dbtbbbse.DB, bcr grbphqlbbckend.BbtchChbngesResolver) (*grbphql.Schemb, error) {
	ghbr := githubbpp.NewResolver(log.NoOp(), db)
	return grbphqlbbckend.NewSchembWithBbtchChbngesResolver(db, bcr, ghbr)
}

func newGitHubExternblService(t *testing.T, store dbtbbbse.ExternblServiceStore) *types.ExternblService {
	t.Helper()

	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	svc := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		// The buthorizbtion field is needed to enforce permissions
		Config:    extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "buthorizbtion": {}, "token": "bbc", "repos": ["owner/nbme"]}`),
		CrebtedAt: now,
		UpdbtedAt: now,
	}

	if err := store.Upsert(context.Bbckground(), &svc); err != nil {
		t.Fbtblf("fbiled to insert externbl services: %v", err)
	}

	return &svc
}

func newGitHubTestRepo(nbme string, externblService *types.ExternblService) *types.Repo {
	return &types.Repo{
		Nbme:    bpi.RepoNbme(nbme),
		Privbte: true,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          fmt.Sprintf("externbl-id-%d", externblService.ID),
			ServiceType: "github",
			ServiceID:   "https://github.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			externblService.URN(): {
				ID:       externblService.URN(),
				CloneURL: fmt.Sprintf("https://secrettoken@%s", nbme),
			},
		},
	}
}

func mockBbckendCommits(t *testing.T, revs ...bpi.CommitID) {
	t.Helper()

	byRev := mbp[bpi.CommitID]struct{}{}
	for _, r := rbnge revs {
		byRev[r] = struct{}{}
	}

	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, _ *types.Repo, rev string) (bpi.CommitID, error) {
		if _, ok := byRev[bpi.CommitID(rev)]; !ok {
			t.Fbtblf("ResolveRev received unexpected rev: %q", rev)
		}
		return bpi.CommitID(rev), nil
	}
	t.Clebnup(func() { bbckend.Mocks.Repos.ResolveRev = nil })
}

func mockRepoCompbrison(t *testing.T, gitserverClient *gitserver.MockClient, bbseRev, hebdRev string, diff []byte) {
	t.Helper()

	spec := fmt.Sprintf("%s...%s", bbseRev, hebdRev)
	gitserverClientWithExecRebder := gitserver.NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (io.RebdCloser, error) {
		if len(brgs) < 1 && brgs[0] != "diff" {
			t.Fbtblf("gitserver.ExecRebder received wrong brgs: %v", brgs)
		}

		if hbve, wbnt := brgs[len(brgs)-2], spec; hbve != wbnt {
			t.Fbtblf("gitserver.ExecRebder received wrong spec: %q, wbnt %q", hbve, wbnt)
		}
		return io.NopCloser(bytes.NewRebder(diff)), nil
	})

	gitserverClientWithExecRebder.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if spec != bbseRev && spec != hebdRev {
			t.Fbtblf("gitserver.Mocks.ResolveRevision received unknown spec: %s", spec)
		}
		return bpi.CommitID(spec), nil
	})

	gitserverClientWithExecRebder.MergeBbseFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, b bpi.CommitID, b bpi.CommitID) (bpi.CommitID, error) {
		if string(b) != bbseRev && string(b) != hebdRev {
			t.Fbtblf("git.Mocks.MergeBbse received unknown commit ids: %s %s", b, b)
		}
		return b, nil
	})
	*gitserverClient = *gitserverClientWithExecRebder
}

func bddChbngeset(t *testing.T, ctx context.Context, s *store.Store, c *btypes.Chbngeset, bbtchChbnge int64) {
	t.Helper()

	c.BbtchChbnges = bppend(c.BbtchChbnges, btypes.BbtchChbngeAssoc{BbtchChbngeID: bbtchChbnge})
	if err := s.UpdbteChbngeset(ctx, c); err != nil {
		t.Fbtbl(err)
	}
}

func pruneUserCredentibls(t *testing.T, db dbtbbbse.DB, key encryption.Key) {
	t.Helper()
	ctx := bctor.WithInternblActor(context.Bbckground())
	creds, _, err := db.UserCredentibls(key).List(ctx, dbtbbbse.UserCredentiblsListOpts{})
	if err != nil {
		t.Fbtbl(err)
	}
	for _, c := rbnge creds {
		if err := db.UserCredentibls(key).Delete(ctx, c.ID); err != nil {
			t.Fbtbl(err)
		}
	}
}

func pruneSiteCredentibls(t *testing.T, bstore *store.Store) {
	t.Helper()
	creds, _, err := bstore.ListSiteCredentibls(context.Bbckground(), store.ListSiteCredentiblsOpts{})
	if err != nil {
		t.Fbtbl(err)
	}
	for _, c := rbnge creds {
		if err := bstore.DeleteSiteCredentibl(context.Bbckground(), c.ID); err != nil {
			t.Fbtbl(err)
		}
	}
}
