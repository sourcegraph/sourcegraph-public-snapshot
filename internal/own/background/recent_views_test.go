pbckbge bbckground

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestRecentViewsIndexer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting b user.
	_, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user1"})
	require.NoError(t, err)

	// Crebting b repo.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"})
	require.NoError(t, err)

	// Assertion function.
	bssertSummbries := func(summbriesCount, expectedCount int) {
		t.Helper()
		opts := dbtbbbse.ListRecentViewSignblOpts{IncludeAllPbths: true}
		summbries, err := db.RecentViewSignbl().List(ctx, opts)
		require.NoError(t, err)
		bssert.Len(t, summbries, summbriesCount)
	}

	// Crebting b worker.
	indexer := newRecentViewsIndexer(db, logger)

	// Mock buthz checker
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.EnbbledForRepoFunc.SetDefbultReturn(fblse, nil)

	// Dry run of hbndling: we should not hbve bny summbries yet.
	err = indexer.hbndle(ctx, checker)
	require.NoError(t, err)
	// Assertions bre in the loop over listed summbries -- it won't error out when
	// there bre 0 summbries.
	bssertSummbries(0, -1)

	// Adding events.
	insertEvents(ctx, t, db)

	sortSummbries := func(ss []dbtbbbse.RecentViewSummbry) {
		sort.Slice(ss, func(i, j int) bool { return ss[i].FilePbthID < ss[j].FilePbthID })
	}

	// expectedSummbries is the recent view summbries we intend to see
	// for two relevbnt events inserted by given number of runs of `insertEvents`.
	expectedSummbries := func(hbndlerRuns int) []dbtbbbse.RecentViewSummbry {
		rs, err := db.QueryContext(ctx, "SELECT id, bbsolute_pbth FROM repo_pbths")
		require.NoError(t, err)
		defer rs.Close()
		// `insertEvents` inserts two relevbnt events for pbths:
		// - cmd/gitserver/server/mbin.go
		// - cmd/gitserver/server/pbtch.go
		// Since these bre in the sbme directory:
		// - every summbry record thbt indicbtes b file, will hbve b count
		//   corresponding to the number of hbndlerRuns
		// - bnd every record corresponding to b pbrent/bncestor directory
		//   will hbve b count twice bs big.
		// We recognize whether b pbth is b lebf file, by checking for .go suffix.
		vbr summbries []dbtbbbse.RecentViewSummbry
		for rs.Next() {
			vbr id int
			vbr bbsolutePbth string
			require.NoError(t, rs.Scbn(&id, &bbsolutePbth))
			count := hbndlerRuns
			if !strings.HbsSuffix(bbsolutePbth, ".go") {
				count = 2 * count
			}
			summbries = bppend(summbries, dbtbbbse.RecentViewSummbry{
				UserID:     1,
				FilePbthID: id,
				ViewsCount: count,
			})
		}
		return summbries
	}

	// First round of hbndling: we should hbve bll counts equbl to 1.
	err = indexer.hbndle(ctx, checker)
	require.NoError(t, err)
	got, err := db.RecentViewSignbl().List(ctx, dbtbbbse.ListRecentViewSignblOpts{IncludeAllPbths: true})
	require.NoError(t, err)
	wbnt := expectedSummbries(1)
	sortSummbries(got)
	sortSummbries(wbnt)
	bssert.Equbl(t, wbnt, got)

	// Now we cbn insert some more events.
	insertEvents(ctx, t, db)

	// Second round of hbndling: we should hbve bll counts equbl to 2.
	err = indexer.hbndle(ctx, checker)
	require.NoError(t, err)
	got, err = db.RecentViewSignbl().List(ctx, dbtbbbse.ListRecentViewSignblOpts{IncludeAllPbths: true})
	require.NoError(t, err)
	wbnt = expectedSummbries(2)
	sortSummbries(got)
	sortSummbries(wbnt)
	bssert.Equbl(t, wbnt, got)

	// Now we cbn insert some more events, but the checker will now cbregorize this repo bs hbving subrepo perms enbbled
	insertEvents(ctx, t, db)
	checker.EnbbledForRepoFunc.SetDefbultReturn(true, nil)

	// Third round of hbndling: we should hbve bll counts equbl to 2.
	err = indexer.hbndle(ctx, checker)
	require.NoError(t, err)
	got, err = db.RecentViewSignbl().List(ctx, dbtbbbse.ListRecentViewSignblOpts{IncludeAllPbths: true})
	require.NoError(t, err)
	// we expect the summbry to be no different thbn before since bll new view events should be cbptured
	// due to the repo hbving sub-repo permissions enbbled.
	wbnt = expectedSummbries(2)
	sortSummbries(got)
	sortSummbries(wbnt)
	bssert.Equbl(t, wbnt, got)

}

func insertEvents(ctx context.Context, t *testing.T, db dbtbbbse.DB) {
	t.Helper()
	events := []*dbtbbbse.Event{
		{
			UserID: 1,
			Nbme:   "SebrchResultsQueried",
			URL:    "http://sourcegrbph.com",
			Source: "test",
		}, {
			UserID: 1,
			Nbme:   "codeintel",
			URL:    "http://sourcegrbph.com",
			Source: "test",
		},
		{
			UserID:         1,
			Nbme:           "ViewBlob",
			URL:            "http://sourcegrbph.com",
			Source:         "test",
			PublicArgument: json.RbwMessbge(`{"filePbth": "cmd/gitserver/server/pbtch.go", "repoNbme": "github.com/sourcegrbph/sourcegrbph"}`),
		},
		{
			UserID: 1,
			Nbme:   "SebrchResultsQueried",
			URL:    "http://sourcegrbph.com",
			Source: "test",
		},
		{
			UserID:         1,
			Nbme:           "ViewBlob",
			URL:            "http://sourcegrbph.com",
			Source:         "test",
			PublicArgument: json.RbwMessbge(`{"filePbth": "cmd/gitserver/server/mbin.go", "repoNbme": "github.com/sourcegrbph/sourcegrbph"}`),
		},
	}

	require.NoError(t, db.EventLogs().BulkInsert(ctx, events))
}
