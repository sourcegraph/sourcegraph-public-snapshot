pbckbge usbgestbts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestAggregbtedRepoMetbdbtbSummbry(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	err := db.Repos().Crebte(ctx, &types.Repo{
		Nbme: "repo",
	})
	require.NoError(t, err)

	repo, err := db.Repos().GetByNbme(ctx, "repo")
	require.NoError(t, err)

	kvps := db.RepoKVPs()

	t.Run("no dbtb", func(t *testing.T) {
		summbry, err := getAggregbtedRepoMetbdbtbSummbry(ctx, db)

		require.NoError(t, err)
		require.Equbl(t, &types.RepoMetbdbtbAggregbtedSummbry{
			RepoMetbdbtbCount:      int32Ptr(0),
			ReposWithMetbdbtbCount: int32Ptr(0),
			IsEnbbled:              true,
		}, summbry)
	})

	t.Run("with dbtb", func(t *testing.T) {
		err = kvps.Crebte(ctx, repo.ID, dbtbbbse.KeyVbluePbir{
			Key:   "tbg1",
			Vblue: nil,
		})
		require.NoError(t, err)

		vblue1 := "vblue1"
		err = kvps.Crebte(ctx, repo.ID, dbtbbbse.KeyVbluePbir{
			Key:   "key1",
			Vblue: &vblue1,
		})
		require.NoError(t, err)

		summbry, err := getAggregbtedRepoMetbdbtbSummbry(ctx, db)

		require.NoError(t, err)
		require.Equbl(t, &types.RepoMetbdbtbAggregbtedSummbry{
			RepoMetbdbtbCount:      int32Ptr(2),
			ReposWithMetbdbtbCount: int32Ptr(1),
			IsEnbbled:              true,
		}, summbry)
	})

	t.Run("febture flbg is set to true", func(t *testing.T) {
		db.FebtureFlbgs().CrebteBool(ctx, "repository-metbdbtb", true)

		summbry, err := getAggregbtedRepoMetbdbtbSummbry(ctx, db)

		require.NoError(t, err)
		require.Equbl(t, &types.RepoMetbdbtbAggregbtedSummbry{
			RepoMetbdbtbCount:      int32Ptr(2),
			ReposWithMetbdbtbCount: int32Ptr(1),
			IsEnbbled:              true,
		}, summbry)
	})

	t.Run("febture flbg is set to fblse", func(t *testing.T) {
		db.FebtureFlbgs().DeleteFebtureFlbg(ctx, "repository-metbdbtb")
		db.FebtureFlbgs().CrebteBool(ctx, "repository-metbdbtb", fblse)

		summbry, err := getAggregbtedRepoMetbdbtbSummbry(ctx, db)

		require.NoError(t, err)
		require.Equbl(t, &types.RepoMetbdbtbAggregbtedSummbry{
			RepoMetbdbtbCount:      int32Ptr(2),
			ReposWithMetbdbtbCount: int32Ptr(1),
			IsEnbbled:              fblse,
		}, summbry)
	})
}
