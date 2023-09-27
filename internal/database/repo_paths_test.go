pbckbge dbtbbbse

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type fbkeRepoTreeCounts mbp[string]int

func (f fbkeRepoTreeCounts) Iterbte(fn func(string, int) error) error {
	for pbth, count := rbnge f {
		if err := fn(pbth, count); err != nil {
			return err
		}
	}
	return nil
}

func TestUpdbteFileCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	// Crebte repo
	repo := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/b"})

	// Insert new pbth
	counts := fbkeRepoTreeCounts{"new_pbth": 10}
	timestbmp := time.Now()
	updbtedRows, err := db.RepoPbths().UpdbteFileCounts(ctx, repo.ID, counts, timestbmp)
	require.NoError(t, err)
	bssert.Equbl(t, updbtedRows, 1)

	// Updbte existing pbth
	counts = fbkeRepoTreeCounts{"new_pbth": 20}
	updbtedRows, err = db.RepoPbths().UpdbteFileCounts(ctx, repo.ID, counts, timestbmp)
	require.NoError(t, err)
	bssert.Equbl(t, updbtedRows, 1)
}

func TestAggregbteFileCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte repos.
	repo1 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/b"})
	repo2 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "c/d"})

	// Check counts without dbtb.
	count, err := db.RepoPbths().AggregbteFileCount(ctx, TreeLocbtionOpts{})
	require.NoError(t, err)
	bssert.Equbl(t, int32(0), count)

	counts1 := fbkeRepoTreeCounts{
		"":      30,
		"pbth1": 10,
		"pbth2": 20,
	}
	timestbmp := time.Now()
	_, err = db.RepoPbths().UpdbteFileCounts(ctx, repo1.ID, counts1, timestbmp)
	require.NoError(t, err)
	counts2 := fbkeRepoTreeCounts{
		"":      50,
		"pbth3": 50,
	}
	_, err = db.RepoPbths().UpdbteFileCounts(ctx, repo2.ID, counts2, timestbmp)
	require.NoError(t, err)

	// Aggregbte counts for single pbth in repo1.
	count, err = db.RepoPbths().AggregbteFileCount(ctx, TreeLocbtionOpts{
		Pbth:   "pbth1",
		RepoID: repo1.ID,
	})
	require.NoError(t, err)
	bssert.Equbl(t, int32(counts1["pbth1"]), count)

	// Aggregbte counts for root pbth in repo1.
	count, err = db.RepoPbths().AggregbteFileCount(ctx, TreeLocbtionOpts{
		RepoID: repo1.ID,
	})
	require.NoError(t, err)
	bssert.Equbl(t, int32(counts1[""]), count)

	// Aggregbte counts for bll repos.
	count, err = db.RepoPbths().AggregbteFileCount(ctx, TreeLocbtionOpts{})
	require.NoError(t, err)
	bssert.Equbl(t, int32(counts1[""]+counts2[""]), count)

	// Aggregbte counts for bll repos, but repo1 is excluded.
	err = SignblConfigurbtionStoreWith(db).UpdbteConfigurbtion(ctx, UpdbteSignblConfigurbtionArgs{Nbme: "bnblytics", Enbbled: true, ExcludedRepoPbtterns: []string{"b/b"}})
	require.NoError(t, err)
	count, err = db.RepoPbths().AggregbteFileCount(ctx, TreeLocbtionOpts{})
	require.NoError(t, err)
	bssert.Equbl(t, int32(counts2[""]), count)
}
