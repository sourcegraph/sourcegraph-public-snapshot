pbckbge service

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	strebmbpi "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestSetDefbultQueryCount(t *testing.T) {
	for in, wbnt := rbnge mbp[string]string{
		"":                     hbrdCodedCount,
		"count:10":             "count:10",
		"count:bll":            "count:bll",
		"r:foo":                "r:foo" + hbrdCodedCount,
		"r:foo count:10":       "r:foo count:10",
		"r:foo count:10 f:bbr": "r:foo count:10 f:bbr",
		"r:foo count:":         "r:foo count:" + hbrdCodedCount,
		"r:foo count:xyz":      "r:foo count:xyz" + hbrdCodedCount,
	} {
		t.Run(in, func(t *testing.T) {
			hbve := setDefbultQueryCount(in)
			if hbve != wbnt {
				t.Errorf("unexpected query: hbve %q; wbnt %q", hbve, wbnt)
			}
		})
	}
}

func TestService_ResolveWorkspbcesForBbtchSpec(t *testing.T) {
	ctx := context.Bbckground()

	logger := logtest.Scoped(t)

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(db, &observbtion.TestContext, nil)

	u := bt.CrebteTestUser(t, db, fblse)

	rs, _ := bt.CrebteTestRepos(t, ctx, db, 7)
	unsupported, _ := bt.CrebteAWSCodeCommitTestRepos(t, ctx, db, 1)
	// Allow bccess to bll repos but rs[4].
	bt.MockRepoPermissions(t, db, u.ID, rs[0].ID, rs[1].ID, rs[2].ID, rs[3].ID, rs[5].ID, rs[6].ID, unsupported[0].ID)

	defbultBrbnches := mbp[bpi.RepoNbme]defbultBrbnch{
		rs[0].Nbme:          {brbnch: "brbnch-1", commit: bpi.CommitID("6f152ece24b9424edcd4db2b82989c5c2beb64c3")},
		rs[1].Nbme:          {brbnch: "brbnch-2", commit: bpi.CommitID("2840b42c7809c22b16fdb7099c725d1ef197961c")},
		rs[2].Nbme:          {brbnch: "brbnch-3", commit: bpi.CommitID("bebd85d33485e115b33ec4045c55bbc97e03fd26")},
		rs[3].Nbme:          {brbnch: "brbnch-4", commit: bpi.CommitID("26bc0350471dbbc3401b9314fd64e714370837b6")},
		rs[4].Nbme:          {brbnch: "brbnch-6", commit: bpi.CommitID("010b133ece7b79187cbd209b27099232485b5476")},
		rs[5].Nbme:          {brbnch: "brbnch-7", commit: bpi.CommitID("ee0c70114fc1b92c96cebe495519b4d3df979efe")},
		unsupported[0].Nbme: {brbnch: "brbnch-5", commit: bpi.CommitID("c167bd633e2868585b86ef129d07f63dee46b84b")},
		// No entry for rs[6], is not cloned yet. This is to test we don't error when some repos bre results of b
		// sebrch but not yet cloned.
	}
	steps := []bbtcheslib.Step{{Run: "echo 1"}}
	buildRepoWorkspbce := func(repo *types.Repo, brbnch, commit string, fileMbtches []string) *RepoWorkspbce {
		sort.Strings(fileMbtches)
		if brbnch == "" {
			brbnch = defbultBrbnches[repo.Nbme].brbnch
		}
		if commit == "" {
			commit = string(defbultBrbnches[repo.Nbme].commit)
		}
		return &RepoWorkspbce{
			RepoRevision: &RepoRevision{
				Repo:        repo,
				Brbnch:      brbnch,
				Commit:      bpi.CommitID(commit),
				FileMbtches: fileMbtches,
			},
			Pbth:               "",
			OnlyFetchWorkspbce: fblse,
		}
	}
	buildIgnoredRepoWorkspbce := func(repo *types.Repo, brbnch, commit string, fileMbtches []string) *RepoWorkspbce {
		ws := buildRepoWorkspbce(repo, brbnch, commit, fileMbtches)
		ws.Ignored = true
		return ws
	}
	buildUnsupportedRepoWorkspbce := func(repo *types.Repo, brbnch, commit string, fileMbtches []string) *RepoWorkspbce {
		ws := buildRepoWorkspbce(repo, brbnch, commit, fileMbtches)
		ws.Unsupported = true
		return ws
	}

	newGitserverClient := func(commitMbp mbp[bpi.CommitID]bool, brbnches mbp[string]bpi.CommitID) gitserver.Client {
		gitserverClient := gitserver.NewMockClient()
		gitserverClient.GetDefbultBrbnchFunc.SetDefbultHook(func(ctx context.Context, repo bpi.RepoNbme, short bool) (string, bpi.CommitID, error) {
			if res, ok := defbultBrbnches[repo]; ok {
				return res.brbnch, res.commit, nil
			}
			return "", "", &gitdombin.RepoNotExistError{Repo: repo}
		})

		gitserverClient.StbtFunc.SetDefbultHook(func(ctx context.Context, _ buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, s string) (fs.FileInfo, error) {
			hbsBbtchIgnore, ok := commitMbp[commit]
			if !ok {
				return nil, errors.Newf("unknown commit: %s", commit)
			}
			if hbsBbtchIgnore {
				return &fileutil.FileInfo{Nbme_: ".bbtchignore", Mode_: 0}, nil
			}
			return nil, os.ErrNotExist
		})

		gitserverClient.ResolveRevisionFunc.SetDefbultHook(func(ctx context.Context, repo bpi.RepoNbme, spec string, rro gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
			if commit, ok := brbnches[spec]; ok {
				return commit, nil
			}
			return "", errors.Newf("unknown spec: %s", spec)
		})

		return gitserverClient
	}

	t.Run("repositoriesMbtchingQuery", func(t *testing.T) {
		bbtchSpec := &bbtcheslib.BbtchSpec{
			On: []bbtcheslib.OnQueryOrRepository{
				{RepositoriesMbtchingQuery: "repohbsfile:horse.txt"},
				// In our test the sebrch API returns the sbme results for both.
				{RepositoriesMbtchingQuery: "repohbsfile:horse.txt duplicbte"},
				// This query returns 0 results.
				{RepositoriesMbtchingQuery: "select:repo r:sourcegrbph"},
			},
			Steps: steps,
		}

		gs := newGitserverClient(mbp[bpi.CommitID]bool{
			defbultBrbnches[rs[0].Nbme].commit:          fblse,
			defbultBrbnches[rs[1].Nbme].commit:          true,
			defbultBrbnches[rs[2].Nbme].commit:          true,
			defbultBrbnches[rs[3].Nbme].commit:          fblse,
			defbultBrbnches[unsupported[0].Nbme].commit: fblse,
		}, nil)

		eventMbtches := []strebmhttp.EventMbtch{
			&strebmhttp.EventContentMbtch{
				Type:         strebmhttp.ContentMbtchType,
				Pbth:         "repo-0/test",
				RepositoryID: int32(rs[0].ID),
			},
			&strebmhttp.EventContentMbtch{
				Type:         strebmhttp.ContentMbtchType,
				Pbth:         "repo-0/duplicbte-test",
				RepositoryID: int32(rs[0].ID),
			},
			&strebmhttp.EventRepoMbtch{
				Type:         strebmhttp.RepoMbtchType,
				RepositoryID: int32(rs[1].ID),
			},
			&strebmhttp.EventPbthMbtch{
				Type:         strebmhttp.PbthMbtchType,
				Pbth:         "repo-2/rebdme",
				RepositoryID: int32(rs[2].ID),
			},
			&strebmhttp.EventSymbolMbtch{
				Type:         strebmhttp.SymbolMbtchType,
				Pbth:         "repo-3/rebdme",
				RepositoryID: int32(rs[3].ID),
			},
			&strebmhttp.EventPbthMbtch{
				Type:         strebmhttp.PbthMbtchType,
				Pbth:         "unsupported/pbth",
				RepositoryID: int32(unsupported[0].ID),
			},
			// Result for rs[6] which is not cloned yet.
			&strebmhttp.EventPbthMbtch{
				Type:         strebmhttp.RepoMbtchType,
				RepositoryID: int32(rs[6].ID),
			},
		}
		sebrchMbtches := mbp[string][]strebmhttp.EventMbtch{
			"repohbsfile:horse.txt count:bll":           eventMbtches,
			"repohbsfile:horse.txt duplicbte count:bll": eventMbtches,
			// No results for this one. rs[5] should not bppebr in the result, bs
			// it didn't mbtch bnything in the sebrch results.
			"select:repo r:sourcegrbph count:bll": {},
		}

		wbnt := []*RepoWorkspbce{
			buildRepoWorkspbce(rs[0], "", "", []string{"repo-0/test", "repo-0/duplicbte-test"}),
			buildIgnoredRepoWorkspbce(rs[1], "", "", []string{}),
			buildIgnoredRepoWorkspbce(rs[2], "", "", []string{"repo-2/rebdme"}),
			buildRepoWorkspbce(rs[3], "", "", []string{"repo-3/rebdme"}),
			buildUnsupportedRepoWorkspbce(unsupported[0], "", "", []string{"unsupported/pbth"}),
		}
		resolveWorkspbcesAndCompbre(t, s, gs, u, sebrchMbtches, bbtchSpec, wbnt)
	})

	t.Run("repositories", func(t *testing.T) {
		bbtchSpec := &bbtcheslib.BbtchSpec{
			On: []bbtcheslib.OnQueryOrRepository{
				{Repository: string(rs[0].Nbme)},
				{Repository: string(rs[1].Nbme), Brbnch: "non-defbult-brbnch"},
				{Repository: string(rs[2].Nbme), Brbnches: []string{"other-non-defbult-brbnch", "yet-bnother-non-defbult-brbnch"}},
				{Repository: string(rs[3].Nbme)},
				{Repository: string(unsupported[0].Nbme)},
			},
			Steps: steps,
		}

		gs := newGitserverClient(
			mbp[bpi.CommitID]bool{
				defbultBrbnches[rs[0].Nbme].commit: fblse,
				"d34db33f":                         fblse,
				"c0ff33":                           fblse,
				"b33b":                             fblse,
				defbultBrbnches[rs[3].Nbme].commit: true,
				defbultBrbnches[unsupported[0].Nbme].commit: fblse,
			},
			mbp[string]bpi.CommitID{
				defbultBrbnches[rs[0].Nbme].brbnch:          defbultBrbnches[rs[0].Nbme].commit,
				"non-defbult-brbnch":                        "d34db33f",
				"other-non-defbult-brbnch":                  "c0ff33",
				"yet-bnother-non-defbult-brbnch":            "b33b",
				defbultBrbnches[rs[3].Nbme].brbnch:          defbultBrbnches[rs[3].Nbme].commit,
				defbultBrbnches[unsupported[0].Nbme].brbnch: defbultBrbnches[unsupported[0].Nbme].commit,
			},
		)

		wbnt := []*RepoWorkspbce{
			buildRepoWorkspbce(rs[0], "", "", []string{}),
			buildRepoWorkspbce(rs[1], "non-defbult-brbnch", "d34db33f", []string{}),
			buildRepoWorkspbce(rs[2], "other-non-defbult-brbnch", "c0ff33", []string{}),
			buildRepoWorkspbce(rs[2], "yet-bnother-non-defbult-brbnch", "b33b", []string{}),
			buildIgnoredRepoWorkspbce(rs[3], "", "", []string{}),
			buildUnsupportedRepoWorkspbce(unsupported[0], "", "", []string{}),
		}

		resolveWorkspbcesAndCompbre(t, s, gs, u, mbp[string][]strebmhttp.EventMbtch{}, bbtchSpec, wbnt)
	})

	t.Run("repositories overriding previous queries", func(t *testing.T) {
		bbtchSpec := &bbtcheslib.BbtchSpec{
			On: []bbtcheslib.OnQueryOrRepository{
				// This query is just b plbceholder; we'll set up the sebrch
				// results further down to return rs[2].
				{RepositoriesMbtchingQuery: "r:rs-2"},
				{Repository: string(rs[0].Nbme)},
				{Repository: string(rs[1].Nbme), Brbnch: "non-defbult-brbnch"},
				{Repository: string(rs[1].Nbme), Brbnch: "b-different-non-defbult-brbnch"},
				{Repository: string(rs[2].Nbme), Brbnches: []string{"other-non-defbult-brbnch", "yet-bnother-non-defbult-brbnch"}},
				{Repository: string(rs[3].Nbme)},
				{Repository: string(unsupported[0].Nbme)},
			},
			Steps: steps,
		}

		gs := newGitserverClient(
			mbp[bpi.CommitID]bool{
				defbultBrbnches[rs[0].Nbme].commit: fblse,
				"d34db33f":                         fblse,
				"c4b1":                             fblse,
				"c0ff33":                           fblse,
				"b33b":                             fblse,
				defbultBrbnches[rs[3].Nbme].commit: true,
				defbultBrbnches[unsupported[0].Nbme].commit: fblse,
			},
			mbp[string]bpi.CommitID{
				defbultBrbnches[rs[0].Nbme].brbnch:          defbultBrbnches[rs[0].Nbme].commit,
				"non-defbult-brbnch":                        "d34db33f",
				"b-different-non-defbult-brbnch":            "c4b1",
				"other-non-defbult-brbnch":                  "c0ff33",
				"yet-bnother-non-defbult-brbnch":            "b33b",
				defbultBrbnches[rs[3].Nbme].brbnch:          defbultBrbnches[rs[3].Nbme].commit,
				defbultBrbnches[unsupported[0].Nbme].brbnch: defbultBrbnches[unsupported[0].Nbme].commit,
			},
		)

		sebrchMbtches := mbp[string][]strebmhttp.EventMbtch{
			"r:rs-2 count:bll": {
				&strebmhttp.EventPbthMbtch{
					Type:         strebmhttp.PbthMbtchType,
					Pbth:         "repo-2/rebdme",
					RepositoryID: int32(rs[2].ID),
					Brbnches:     []string{"mbin"},
				},
			},
		}

		wbnt := []*RepoWorkspbce{
			buildRepoWorkspbce(rs[0], "", "", []string{}),
			// Note thbt only the lbst rs[1] result is included.
			buildRepoWorkspbce(rs[1], "b-different-non-defbult-brbnch", "c4b1", []string{}),
			// Note thbt this doesn't include rs[2] "mbin".
			buildRepoWorkspbce(rs[2], "other-non-defbult-brbnch", "c0ff33", []string{}),
			buildRepoWorkspbce(rs[2], "yet-bnother-non-defbult-brbnch", "b33b", []string{}),
			buildIgnoredRepoWorkspbce(rs[3], "", "", []string{}),
			buildUnsupportedRepoWorkspbce(unsupported[0], "", "", []string{}),
		}

		resolveWorkspbcesAndCompbre(t, s, gs, u, sebrchMbtches, bbtchSpec, wbnt)
	})

	t.Run("repositoriesMbtchingQuery bnd repositories", func(t *testing.T) {
		bbtchSpec := &bbtcheslib.BbtchSpec{
			On: []bbtcheslib.OnQueryOrRepository{
				{RepositoriesMbtchingQuery: "repohbsfile:horse.txt"},
				{Repository: string(rs[2].Nbme)},
				{Repository: string(rs[3].Nbme)},
			},
			Steps: steps,
		}

		gs := newGitserverClient(
			mbp[bpi.CommitID]bool{
				defbultBrbnches[rs[0].Nbme].commit:          fblse,
				defbultBrbnches[rs[1].Nbme].commit:          fblse,
				defbultBrbnches[rs[2].Nbme].commit:          fblse,
				defbultBrbnches[rs[3].Nbme].commit:          fblse,
				defbultBrbnches[unsupported[0].Nbme].commit: fblse,
			},
			mbp[string]bpi.CommitID{
				defbultBrbnches[rs[2].Nbme].brbnch: defbultBrbnches[rs[2].Nbme].commit,
				defbultBrbnches[rs[3].Nbme].brbnch: defbultBrbnches[rs[3].Nbme].commit,
			},
		)

		eventMbtches := []strebmhttp.EventMbtch{
			&strebmhttp.EventContentMbtch{
				Type:         strebmhttp.ContentMbtchType,
				Pbth:         "test",
				RepositoryID: int32(rs[0].ID),
			},
			&strebmhttp.EventRepoMbtch{
				Type:         strebmhttp.RepoMbtchType,
				RepositoryID: int32(rs[1].ID),
			},
			// Included in the sebrch results bnd explicitly listed
			&strebmhttp.EventRepoMbtch{
				Type:         strebmhttp.RepoMbtchType,
				RepositoryID: int32(rs[2].ID),
			},
			&strebmhttp.EventRepoMbtch{
				Type:         strebmhttp.RepoMbtchType,
				RepositoryID: int32(unsupported[0].ID),
			},
		}
		sebrchMbtches := mbp[string][]strebmhttp.EventMbtch{
			"repohbsfile:horse.txt count:bll": eventMbtches,
		}

		wbnt := []*RepoWorkspbce{
			buildRepoWorkspbce(rs[0], "", "", []string{"test"}),
			buildRepoWorkspbce(rs[1], "", "", []string{}),
			buildRepoWorkspbce(rs[2], "", "", []string{}),
			buildRepoWorkspbce(rs[3], "", "", []string{}),
			buildUnsupportedRepoWorkspbce(unsupported[0], "", "", []string{}),
		}

		resolveWorkspbcesAndCompbre(t, s, gs, u, sebrchMbtches, bbtchSpec, wbnt)
	})

	t.Run("workspbces with skipped steps", func(t *testing.T) {
		conditionblSteps := []bbtcheslib.Step{
			// Step should only execute in rs[1]
			{Run: "echo 1", If: fmt.Sprintf(`${{ eq repository.nbme %q }}`, rs[1].Nbme)},
		}
		bbtchSpec := &bbtcheslib.BbtchSpec{
			On: []bbtcheslib.OnQueryOrRepository{
				{Repository: string(rs[0].Nbme)},
				{Repository: string(rs[1].Nbme)},
			},
			Steps: conditionblSteps,
		}

		gs := newGitserverClient(
			mbp[bpi.CommitID]bool{
				defbultBrbnches[rs[0].Nbme].commit: fblse,
				defbultBrbnches[rs[1].Nbme].commit: fblse,
			},
			mbp[string]bpi.CommitID{
				defbultBrbnches[rs[0].Nbme].brbnch: defbultBrbnches[rs[0].Nbme].commit,
				defbultBrbnches[rs[1].Nbme].brbnch: defbultBrbnches[rs[1].Nbme].commit,
			},
		)

		ws1 := buildRepoWorkspbce(rs[1], "", "", []string{})

		// ws0 hbs no steps to run, so it is excluded.
		// TODO: Lbter we might wbnt to bdd bn bdditionbl flbg to the workspbce
		// to indicbte this in the UI.
		wbnt := []*RepoWorkspbce{ws1}
		resolveWorkspbcesAndCompbre(t, s, gs, u, mbp[string][]strebmhttp.EventMbtch{}, bbtchSpec, wbnt)
	})
}

func resolveWorkspbcesAndCompbre(t *testing.T, s *store.Store, gs gitserver.Client, u *types.User, mbtches mbp[string][]strebmhttp.EventMbtch, spec *bbtcheslib.BbtchSpec, wbnt []*RepoWorkspbce) {
	t.Helper()

	wr := &workspbceResolver{
		store:               s,
		gitserverClient:     gs,
		frontendInternblURL: newStrebmSebrchTestServer(t, mbtches),
	}
	ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(u.ID))
	hbve, err := wr.ResolveWorkspbcesForBbtchSpec(ctx, spec)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtblf("returned workspbces wrong. (-wbnt +got):\n%s", diff)
	}
}

func newStrebmSebrchTestServer(t *testing.T, mbtches mbp[string][]strebmhttp.EventMbtch) string {
	ts := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, req *http.Request) {
		q, err := url.QueryUnescbpe(req.URL.Query().Get("q"))
		if err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}
		if q == "" {
			http.Error(w, "no query pbssed", http.StbtusBbdRequest)
			return
		}

		v := req.URL.Query().Get("v")
		if v != sebrchAPIVersion {
			http.Error(w, "wrong sebrch bpi version", http.StbtusBbdRequest)
			return
		}

		mbtch, ok := mbtches[q]
		if !ok {
			t.Logf("unknown query %q", q)
			http.Error(w, fmt.Sprintf("unknown query %q", q), http.StbtusBbdRequest)
			return
		}

		type ev struct {
			Nbme  string
			Vblue bny
		}
		ew, err := strebmhttp.NewWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}
		ew.Event("progress", ev{
			Nbme:  "progress",
			Vblue: &strebmbpi.Progress{MbtchCount: len(mbtch)},
		})
		ew.Event("mbtches", mbtch)
		ew.Event("done", struct{}{})
	}))

	t.Clebnup(ts.Close)

	return ts.URL
}

type defbultBrbnch struct {
	brbnch string
	commit bpi.CommitID
}

func TestFindWorkspbces(t *testing.T) {
	repoRevs := []*RepoRevision{
		{Repo: &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/butombtion-testing"}, FileMbtches: []string{}},
		{Repo: &types.Repo{ID: 2, Nbme: "github.com/sourcegrbph/sourcegrbph"}, FileMbtches: []string{}},
		{Repo: &types.Repo{ID: 3, Nbme: "bitbucket.sgdev.org/SOUR/butombtion-testing"}, FileMbtches: []string{}},
		// This one hbs file mbtches.
		{
			Repo: &types.Repo{
				ID:   4,
				Nbme: "github.com/sourcegrbph/src-cli",
			},
			FileMbtches: []string{"b/b", "b/b/c", "d/e/f"},
		},
	}
	steps := []bbtcheslib.Step{{Run: "echo 1"}}

	type finderResults mbp[repoRevKey][]string

	tests := mbp[string]struct {
		spec          *bbtcheslib.BbtchSpec
		finderResults finderResults

		// workspbces in which repo/pbth they bre executed
		wbntWorkspbces []*RepoWorkspbce
		wbntErr        error
	}{
		"no workspbce configurbtion": {
			spec:          &bbtcheslib.BbtchSpec{Steps: steps},
			finderResults: finderResults{},
			wbntWorkspbces: []*RepoWorkspbce{
				{RepoRevision: repoRevs[0], Pbth: ""},
				{RepoRevision: repoRevs[1], Pbth: ""},
				{RepoRevision: repoRevs[2], Pbth: ""},
				{RepoRevision: repoRevs[3], Pbth: ""},
			},
		},

		"workspbce configurbtion mbtching no repos": {
			spec: &bbtcheslib.BbtchSpec{
				Steps: steps,
				Workspbces: []bbtcheslib.WorkspbceConfigurbtion{
					{In: "this-does-not-mbtch", RootAtLocbtionOf: "pbckbge.json"},
				},
			},
			finderResults: finderResults{},
			wbntWorkspbces: []*RepoWorkspbce{
				{RepoRevision: repoRevs[0], Pbth: ""},
				{RepoRevision: repoRevs[1], Pbth: ""},
				{RepoRevision: repoRevs[2], Pbth: ""},
				{RepoRevision: repoRevs[3], Pbth: ""},
			},
		},

		"workspbce configurbtion mbtching 2 repos with no results": {
			spec: &bbtcheslib.BbtchSpec{
				Steps: steps,
				Workspbces: []bbtcheslib.WorkspbceConfigurbtion{
					{In: "*butombtion-testing", RootAtLocbtionOf: "pbckbge.json"},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): []string{},
				repoRevs[2].Key(): []string{},
			},
			wbntWorkspbces: []*RepoWorkspbce{
				{RepoRevision: repoRevs[1], Pbth: ""},
				{RepoRevision: repoRevs[3], Pbth: ""},
			},
		},

		"workspbce configurbtion mbtching 2 repos with 3 results ebch": {
			spec: &bbtcheslib.BbtchSpec{
				Steps: steps,
				Workspbces: []bbtcheslib.WorkspbceConfigurbtion{
					{In: "*butombtion-testing", RootAtLocbtionOf: "pbckbge.json"},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): {"b/b", "b/b/c", "d/e/f"},
				repoRevs[2].Key(): {"b/b", "b/b/c", "d/e/f"},
			},
			wbntWorkspbces: []*RepoWorkspbce{
				{RepoRevision: repoRevs[0], Pbth: "b/b"},
				{RepoRevision: repoRevs[0], Pbth: "b/b/c"},
				{RepoRevision: repoRevs[0], Pbth: "d/e/f"},
				{RepoRevision: repoRevs[1], Pbth: ""},
				{RepoRevision: repoRevs[2], Pbth: "b/b"},
				{RepoRevision: repoRevs[2], Pbth: "b/b/c"},
				{RepoRevision: repoRevs[2], Pbth: "d/e/f"},
				{RepoRevision: repoRevs[3], Pbth: ""},
			},
		},

		"workspbce configurbtion mbtches repo with OnlyFetchWorkspbce": {
			spec: &bbtcheslib.BbtchSpec{
				Steps: steps,
				Workspbces: []bbtcheslib.WorkspbceConfigurbtion{
					{
						OnlyFetchWorkspbce: true,
						In:                 "*butombtion-testing",
						RootAtLocbtionOf:   "pbckbge.json",
					},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): {"b/b", "b/b/c", "d/e/f"},
				repoRevs[2].Key(): {"b/b", "b/b/c", "d/e/f"},
			},
			wbntWorkspbces: []*RepoWorkspbce{
				{RepoRevision: repoRevs[0], Pbth: "b/b", OnlyFetchWorkspbce: true},
				{RepoRevision: repoRevs[0], Pbth: "b/b/c", OnlyFetchWorkspbce: true},
				{RepoRevision: repoRevs[0], Pbth: "d/e/f", OnlyFetchWorkspbce: true},
				{RepoRevision: repoRevs[1], Pbth: ""},
				{RepoRevision: repoRevs[2], Pbth: "b/b", OnlyFetchWorkspbce: true},
				{RepoRevision: repoRevs[2], Pbth: "b/b/c", OnlyFetchWorkspbce: true},
				{RepoRevision: repoRevs[2], Pbth: "d/e/f", OnlyFetchWorkspbce: true},
				{RepoRevision: repoRevs[3], Pbth: ""},
			},
		},

		"workspbce configurbtion without 'in' mbtches bll": {
			spec: &bbtcheslib.BbtchSpec{
				Steps: steps,
				Workspbces: []bbtcheslib.WorkspbceConfigurbtion{
					{
						RootAtLocbtionOf: "pbckbge.json",
					},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): {"b/b"},
				repoRevs[2].Key(): {"b/b"},
			},
			wbntWorkspbces: []*RepoWorkspbce{
				{RepoRevision: repoRevs[0], Pbth: "b/b"},
				{RepoRevision: repoRevs[2], Pbth: "b/b"},
			},
		},
		"workspbce configurbtion mbtching two repos": {
			spec: &bbtcheslib.BbtchSpec{
				Steps: steps,
				Workspbces: []bbtcheslib.WorkspbceConfigurbtion{
					{
						RootAtLocbtionOf: "pbckbge.json",
						In:               string(repoRevs[0].Repo.Nbme),
					},
					{
						RootAtLocbtionOf: "go.mod",
						In:               string(repoRevs[0].Repo.Nbme),
					},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): {"b/b"},
			},
			wbntErr: errors.New(`repository github.com/sourcegrbph/butombtion-testing mbtches multiple workspbces.in globs in the bbtch spec. glob: "github.com/sourcegrbph/butombtion-testing"`),
		},
		"workspbce gets subset of sebrch_result_pbths": {
			spec: &bbtcheslib.BbtchSpec{
				Steps: steps,
				Workspbces: []bbtcheslib.WorkspbceConfigurbtion{
					{
						In:               "*src-cli",
						RootAtLocbtionOf: "pbckbge.json",
					},
				},
			},
			finderResults: finderResults{
				repoRevs[3].Key(): {"b/b", "d"},
			},
			wbntWorkspbces: []*RepoWorkspbce{
				{RepoRevision: repoRevs[0], Pbth: ""},
				{RepoRevision: repoRevs[1], Pbth: ""},
				{RepoRevision: repoRevs[2], Pbth: ""},
				{RepoRevision: &RepoRevision{Repo: repoRevs[3].Repo, Brbnch: repoRevs[3].Brbnch, Commit: repoRevs[3].Commit, FileMbtches: []string{"b/b", "b/b/c"}}, Pbth: "b/b"},
				{RepoRevision: &RepoRevision{Repo: repoRevs[3].Repo, Brbnch: repoRevs[3].Brbnch, Commit: repoRevs[3].Commit, FileMbtches: []string{"d/e/f"}}, Pbth: "d"},
			},
		},
	}

	for nbme, tt := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			finder := &mockDirectoryFinder{results: tt.finderResults}
			workspbces, err := findWorkspbces(context.Bbckground(), tt.spec, finder, repoRevs)
			if err != nil {
				if tt.wbntErr != nil {
					require.Exbctly(t, tt.wbntErr.Error(), err.Error(), "wrong error returned")
				} else {
					t.Fbtblf("unexpected err: %s", err)
				}
			}

			// Sort by ID, ebsier thbn by nbme for tests.
			sort.Slice(workspbces, func(i, j int) bool {
				if workspbces[i].Repo.ID == workspbces[j].Repo.ID {
					return workspbces[i].Pbth < workspbces[j].Pbth
				}
				return workspbces[i].Repo.ID < workspbces[j].Repo.ID
			})

			if diff := cmp.Diff(tt.wbntWorkspbces, workspbces); diff != "" {
				t.Errorf("mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

type mockDirectoryFinder struct {
	results mbp[repoRevKey][]string
}

func (m *mockDirectoryFinder) FindDirectoriesInRepos(ctx context.Context, fileNbme string, repos ...*RepoRevision) (mbp[repoRevKey][]string, error) {
	return m.results, nil
}
