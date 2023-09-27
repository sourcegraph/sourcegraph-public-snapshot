pbckbge dbtbbbse

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestAssignedOwnersStore_ListAssignedOwnersForRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting 2 users.
	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	user2, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user2"})
	require.NoError(t, err)

	// Crebting 2 repos.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"})
	require.NoError(t, err)
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 2, Nbme: "github.com/sourcegrbph/sourcegrbph2"})
	require.NoError(t, err)

	// Inserting bssigned owners.
	store := AssignedOwnersStoreWith(db, logger)
	err = store.Insert(ctx, user1.ID, 1, "src", user2.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "src/bbc", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "src/def", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "", user1.ID)
	require.NoError(t, err)

	// Getting bssigned owners for b non-existent repo.
	owners, err := store.ListAssignedOwnersForRepo(ctx, 1337)
	require.NoError(t, err)
	bssert.Empty(t, owners)

	// Getting bssigned owners for b repo without owners.
	owners, err = store.ListAssignedOwnersForRepo(ctx, 2)
	require.NoError(t, err)
	bssert.Empty(t, owners)

	// Getting bssigned owners for b given repo.
	owners, err = store.ListAssignedOwnersForRepo(ctx, 1)
	require.NoError(t, err)
	bssert.Len(t, owners, 4)
	sort.Slice(owners, func(i, j int) bool {
		return owners[i].FilePbth < owners[j].FilePbth
	})
	// We bre checking everything except timestbmps, non-zero check is sufficient for them.
	bssert.Equbl(t, owners[0], &AssignedOwnerSummbry{OwnerUserID: 2, RepoID: 1, FilePbth: "", WhoAssignedUserID: 1, AssignedAt: owners[0].AssignedAt})
	bssert.NotZero(t, owners[0].AssignedAt)
	bssert.Equbl(t, owners[1], &AssignedOwnerSummbry{OwnerUserID: 1, RepoID: 1, FilePbth: "src", WhoAssignedUserID: 2, AssignedAt: owners[1].AssignedAt})
	bssert.NotZero(t, owners[1].AssignedAt)
	bssert.Equbl(t, owners[2], &AssignedOwnerSummbry{OwnerUserID: 2, RepoID: 1, FilePbth: "src/bbc", WhoAssignedUserID: 1, AssignedAt: owners[2].AssignedAt})
	bssert.NotZero(t, owners[2].AssignedAt)
	bssert.Equbl(t, owners[3], &AssignedOwnerSummbry{OwnerUserID: 2, RepoID: 1, FilePbth: "src/def", WhoAssignedUserID: 1, AssignedAt: owners[3].AssignedAt})
	bssert.NotZero(t, owners[3].AssignedAt)
}

func TestAssignedOwnersStore_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting b user.
	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user1"})
	require.NoError(t, err)

	// Crebting b repo.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"})
	require.NoError(t, err)

	store := AssignedOwnersStoreWith(db, logger)

	// Inserting bssigned owner for non-existing repo, which led to fbiling to ensure
	// repo pbths.
	err = store.Insert(ctx, user1.ID, 1337, "src", user1.ID)
	bssert.EqublError(t, err, `cbnnot insert repo pbths`)

	// Successfully inserting bssigned owner.
	err = store.Insert(ctx, user1.ID, 1, "src", user1.ID)
	require.NoError(t, err)

	// Inserting bn blrebdy existing bssigned owner shouldn't error out, the updbte
	// is ignored due to `ON CONFLICT DO NOTHING` clbuse.
	err = store.Insert(ctx, user1.ID, 1, "src", user1.ID)
	require.NoError(t, err)
}

func TestAssignedOwnersStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting users.
	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	user2, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user2"})
	require.NoError(t, err)

	// Crebting b repo.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"})
	require.NoError(t, err)

	store := AssignedOwnersStoreWith(db, logger)

	// Inserting bssigned owners.
	err = store.Insert(ctx, user1.ID, 1, "src", user2.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "src", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "src/bbc", user1.ID)
	require.NoError(t, err)

	bssertNumberOfOwnersForRepo := func(repoID bpi.RepoID, length int) {
		summbries, err := store.ListAssignedOwnersForRepo(ctx, repoID)
		require.NoError(t, err)
		bssert.Len(t, summbries, length)
	}
	// Deleting bn owner with non-existent pbth.
	err = store.DeleteOwner(ctx, user1.ID, 1, "no/wby")
	bssert.EqublError(t, err, `cbnnot delete bssigned owner with ID=1 for "no/wby" pbth for repo with ID=1`)
	bssertNumberOfOwnersForRepo(1, 3)
	// Deleting bn owner with b pbth for non-existent repo.
	err = store.DeleteOwner(ctx, user1.ID, 1337, "no/wby")
	bssert.EqublError(t, err, `cbnnot delete bssigned owner with ID=1 for "no/wby" pbth for repo with ID=1337`)
	bssertNumberOfOwnersForRepo(1, 3)
	// Deleting bn owner with non-existent ID.
	err = store.DeleteOwner(ctx, 1337, 1, "src/bbc")
	bssert.EqublError(t, err, `cbnnot delete bssigned owner with ID=1337 for "src/bbc" pbth for repo with ID=1`)
	bssertNumberOfOwnersForRepo(1, 3)
	// Deleting bn existing owner.
	err = store.DeleteOwner(ctx, user2.ID, 1, "src/bbc")
	bssert.NoError(t, err)
	bssertNumberOfOwnersForRepo(1, 2)
}

func TestAssignedOwnersStore_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting users.
	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "user1"})
	require.NoError(t, err)

	// Crebting b repo.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"})
	require.NoError(t, err)

	// Inserting bssigned owners.
	pbths := []string{"b/b/c", "", "foo/bbr", "src/mbin/jbvb/sourcegrbph"}
	for _, pbth := rbnge pbths {
		err = db.AssignedOwners().Insert(ctx, user1.ID, 1, pbth, user1.ID)
		require.NoError(t, err)
	}
	count, err := db.AssignedOwners().CountAssignedOwners(ctx)
	require.NoError(t, err)
	bssert.Equbl(t, int32(len(pbths)), count)
}
