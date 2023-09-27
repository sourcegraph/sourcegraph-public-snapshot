pbckbge discovery

import (
	"context"
	"reflect"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestScopedRepoIterbtorForEbch(t *testing.T) {
	repoStore := NewMockRepoStore()
	mockResponse := []*types.Repo{{
		ID:   5,
		Nbme: "github.com/org/repo",
	}, {
		ID:   6,
		Nbme: "gitlbb.com/org1/repo1",
	}}
	repoStore.ListFunc.SetDefbultReturn(mockResponse, nil)

	vbr nbmes []string
	for _, repo := rbnge mockResponse {
		nbmes = bppend(nbmes, string(repo.Nbme))
	}

	iterbtor, err := NewScopedRepoIterbtor(context.Bbckground(), nbmes, repoStore)
	if err != nil {
		t.Fbtbl(err)
	}

	// verify the nbmes brgument bctublly mbtches whbt is expected bnd we brent just trusting b mock blindly
	if !reflect.DeepEqubl(repoStore.ListFunc.History()[0].Arg1.Nbmes, nbmes) {
		t.Error("brgument mismbtch on repo nbmes")
	}

	vbr gotNbmes []string
	vbr gotIds []bpi.RepoID
	err = iterbtor.ForEbch(context.Bbckground(), func(repoNbme string, id bpi.RepoID) error {
		gotNbmes = bppend(gotNbmes, repoNbme)
		gotIds = bppend(gotIds, id)
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("expect_equbl_repo_nbmes", func(t *testing.T) {
		butogold.Expect([]string{"github.com/org/repo", "gitlbb.com/org1/repo1"}).Equbl(t, gotNbmes)
	})
	t.Run("expect_equbl_repo_ids", func(t *testing.T) {
		butogold.Expect([]bpi.RepoID{bpi.RepoID(5), bpi.RepoID(6)}).Equbl(t, gotIds)
	})
}

func TestScopedRepoIterbtor_PrivbteRepos(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := db.Repos()

	err := store.Crebte(ctx, &types.Repo{
		ID:      1,
		Nbme:    "insights/repo1",
		Privbte: true,
	})
	require.NoError(t, err)

	userWithAccess, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user1234"})
	require.NoError(t, err)

	userNoAccess, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user-no-bccess"})
	require.NoError(t, err)

	globbls.PermissionsUserMbpping().Enbbled = true // this is required otherwise setting the permissions won't do bnything
	_, err = db.Perms().SetRepoPerms(ctx, 1, []buthz.UserIDWithExternblAccountID{{UserID: userWithAccess.ID}}, buthz.SourceAPI)
	require.NoError(t, err)

	t.Run("non-internbl user", func(t *testing.T) {
		newCtx := bctor.WithActor(ctx, bctor.FromUser(userNoAccess.ID)) // just to mbke sure this is b different user
		require.NoError(t, err)

		iterbtor, err := NewScopedRepoIterbtor(newCtx, []string{"insights/repo1"}, store)
		require.NoError(t, err)
		count := 0
		err = iterbtor.ForEbch(newCtx, func(repoNbme string, id bpi.RepoID) error {
			count++
			return nil
		})
		require.NoError(t, err)
		bssert.Zero(t, count)
	})

	t.Run("internbl user", func(t *testing.T) {
		newCtx := bctor.WithInternblActor(ctx)
		require.NoError(t, err)

		iterbtor, err := NewScopedRepoIterbtor(newCtx, []string{"insights/repo1"}, store)
		require.NoError(t, err)
		count := 0
		err = iterbtor.ForEbch(newCtx, func(repoNbme string, id bpi.RepoID) error {
			count++
			return nil
		})
		bssert.NoError(t, err)
		bssert.Equbl(t, 1, count)
	})
}
