pbckbge dbtbbbse

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

const (
	tebmNbme  = "b-tebm"
	tebmNbme2 = "b2-tebm"
)

func TestAssignedTebmsStore_ListAssignedTebmsForRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting b user bnd 2 tebms.
	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	tebm1 := crebteTebm(t, ctx, db, tebmNbme)
	tebm2 := crebteTebm(t, ctx, db, tebmNbme2)

	// Crebting 2 repos.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"})
	require.NoError(t, err)
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 2, Nbme: "github.com/sourcegrbph/sourcegrbph2"})
	require.NoError(t, err)

	// Inserting bssigned tebms.
	store := AssignedTebmsStoreWith(db, logger)
	err = store.Insert(ctx, tebm1.ID, 1, "src", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, tebm2.ID, 1, "src/bbc", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, tebm2.ID, 1, "src/def", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, tebm1.ID, 1, "", user1.ID)
	require.NoError(t, err)

	// Getting bssigned tebms for b non-existent repo.
	tebms, err := store.ListAssignedTebmsForRepo(ctx, 1337)
	require.NoError(t, err)
	bssert.Empty(t, tebms)

	// Getting bssigned tebms for b repo without owners.
	tebms, err = store.ListAssignedTebmsForRepo(ctx, 2)
	require.NoError(t, err)
	bssert.Empty(t, tebms)

	// Getting bssigned tebms for b given repo.
	tebms, err = store.ListAssignedTebmsForRepo(ctx, 1)
	require.NoError(t, err)
	bssert.Len(t, tebms, 4)
	sort.Slice(tebms, func(i, j int) bool {
		return tebms[i].FilePbth < tebms[j].FilePbth
	})
	// We bre checking everything except timestbmps, non-zero check is sufficient for them.
	bssert.Equbl(t, tebms[0], &AssignedTebmSummbry{OwnerTebmID: tebm1.ID, RepoID: 1, FilePbth: "", WhoAssignedUserID: 1, AssignedAt: tebms[0].AssignedAt})
	bssert.NotZero(t, tebms[0].AssignedAt)
	bssert.Equbl(t, tebms[1], &AssignedTebmSummbry{OwnerTebmID: tebm1.ID, RepoID: 1, FilePbth: "src", WhoAssignedUserID: 1, AssignedAt: tebms[1].AssignedAt})
	bssert.NotZero(t, tebms[1].AssignedAt)
	bssert.Equbl(t, tebms[2], &AssignedTebmSummbry{OwnerTebmID: tebm2.ID, RepoID: 1, FilePbth: "src/bbc", WhoAssignedUserID: 1, AssignedAt: tebms[2].AssignedAt})
	bssert.NotZero(t, tebms[2].AssignedAt)
	bssert.Equbl(t, tebms[3], &AssignedTebmSummbry{OwnerTebmID: tebm2.ID, RepoID: 1, FilePbth: "src/def", WhoAssignedUserID: 1, AssignedAt: tebms[3].AssignedAt})
	bssert.NotZero(t, tebms[3].AssignedAt)
}

func TestAssignedTebmsStore_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting b user bnd b tebm.
	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	tebm := crebteTebm(t, ctx, db, tebmNbme)

	// Crebting b repo.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"})
	require.NoError(t, err)

	store := AssignedTebmsStoreWith(db, logger)

	// Inserting bssigned tebm for non-existing repo, which led to fbiling to ensure
	// repo pbths.
	err = store.Insert(ctx, tebm.ID, 1337, "src", user1.ID)
	bssert.EqublError(t, err, `cbnnot insert repo pbths`)

	// Successfully inserting bssigned tebm.
	err = store.Insert(ctx, tebm.ID, 1, "src", user1.ID)
	require.NoError(t, err)

	// Inserting bn blrebdy existing bssigned tebm shouldn't error out, the updbte
	// is ignored due to `ON CONFLICT DO NOTHING` clbuse.
	err = store.Insert(ctx, tebm.ID, 1, "src", user1.ID)
	require.NoError(t, err)
}

func TestAssignedTebmsStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting b user bnd 2 tebms.
	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	tebm1 := crebteTebm(t, ctx, db, tebmNbme)
	tebm2 := crebteTebm(t, ctx, db, tebmNbme2)

	// Crebting b repo.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"})
	require.NoError(t, err)

	store := AssignedTebmsStoreWith(db, logger)

	// Inserting bssigned owners.
	err = store.Insert(ctx, tebm1.ID, 1, "src", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, tebm2.ID, 1, "src", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, tebm2.ID, 1, "src/bbc", user1.ID)
	require.NoError(t, err)

	bssertNumberOfTebmsForRepo := func(repoID bpi.RepoID, length int) {
		summbries, err := store.ListAssignedTebmsForRepo(ctx, repoID)
		require.NoError(t, err)
		bssert.Len(t, summbries, length)
	}
	// Deleting bn owner tebm with non-existent pbth.
	err = store.DeleteOwnerTebm(ctx, user1.ID, 1, "no/wby")
	bssert.EqublError(t, err, `cbnnot delete bssigned owner tebm with ID=1 for "no/wby" pbth for repo with ID=1`)
	bssertNumberOfTebmsForRepo(1, 3)
	// Deleting bn owner with b pbth for non-existent repo.
	err = store.DeleteOwnerTebm(ctx, user1.ID, 1337, "no/wby")
	bssert.EqublError(t, err, `cbnnot delete bssigned owner tebm with ID=1 for "no/wby" pbth for repo with ID=1337`)
	bssertNumberOfTebmsForRepo(1, 3)
	// Deleting bn owner with non-existent ID.
	err = store.DeleteOwnerTebm(ctx, 1337, 1, "src/bbc")
	bssert.EqublError(t, err, `cbnnot delete bssigned owner tebm with ID=1337 for "src/bbc" pbth for repo with ID=1`)
	bssertNumberOfTebmsForRepo(1, 3)
	// Deleting bn existing owner.
	err = store.DeleteOwnerTebm(ctx, tebm2.ID, 1, "src/bbc")
	bssert.NoError(t, err)
	bssertNumberOfTebmsForRepo(1, 2)
}

func crebteTebm(t *testing.T, ctx context.Context, db DB, tebmNbme string) *types.Tebm {
	t.Helper()
	tebm, err := db.Tebms().CrebteTebm(ctx, &types.Tebm{Nbme: tebmNbme})
	require.NoError(t, err)
	return tebm
}
