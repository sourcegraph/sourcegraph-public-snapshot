package database

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAssignedOwnersStore_ListAssignedOwnersForRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating 2 users.
	user1, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)
	user2, err := db.Users().Create(ctx, NewUser{Username: "user2"})
	require.NoError(t, err)

	// Creating 2 repos.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)
	err = db.Repos().Create(ctx, &types.Repo{ID: 2, Name: "github.com/sourcegraph/sourcegraph2"})
	require.NoError(t, err)

	// Inserting assigned owners.
	store := AssignedOwnersStoreWith(db, logger)
	err = store.Insert(ctx, user1.ID, 1, "src", user2.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "src/abc", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "src/def", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "", user1.ID)
	require.NoError(t, err)

	// Getting assigned owners for a non-existent repo.
	owners, err := store.ListAssignedOwnersForRepo(ctx, 1337)
	require.NoError(t, err)
	assert.Empty(t, owners)

	// Getting assigned owners for a repo without owners.
	owners, err = store.ListAssignedOwnersForRepo(ctx, 2)
	require.NoError(t, err)
	assert.Empty(t, owners)

	// Getting assigned owners for a given repo.
	owners, err = store.ListAssignedOwnersForRepo(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, owners, 4)
	sort.Slice(owners, func(i, j int) bool {
		return owners[i].FilePath < owners[j].FilePath
	})
	// We are checking everything except timestamps, non-zero check is sufficient for them.
	assert.Equal(t, owners[0], &AssignedOwnerSummary{OwnerUserID: 2, RepoID: 1, FilePath: "", WhoAssignedUserID: 1, AssignedAt: owners[0].AssignedAt})
	assert.NotZero(t, owners[0].AssignedAt)
	assert.Equal(t, owners[1], &AssignedOwnerSummary{OwnerUserID: 1, RepoID: 1, FilePath: "src", WhoAssignedUserID: 2, AssignedAt: owners[1].AssignedAt})
	assert.NotZero(t, owners[1].AssignedAt)
	assert.Equal(t, owners[2], &AssignedOwnerSummary{OwnerUserID: 2, RepoID: 1, FilePath: "src/abc", WhoAssignedUserID: 1, AssignedAt: owners[2].AssignedAt})
	assert.NotZero(t, owners[2].AssignedAt)
	assert.Equal(t, owners[3], &AssignedOwnerSummary{OwnerUserID: 2, RepoID: 1, FilePath: "src/def", WhoAssignedUserID: 1, AssignedAt: owners[3].AssignedAt})
	assert.NotZero(t, owners[3].AssignedAt)
}

func TestAssignedOwnersStore_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating a user.
	user1, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	store := AssignedOwnersStoreWith(db, logger)

	// Inserting assigned owner for non-existing repo, which led to failing to ensure
	// repo paths.
	err = store.Insert(ctx, user1.ID, 1337, "src", user1.ID)
	assert.EqualError(t, err, `cannot insert repo paths`)

	// Successfully inserting assigned owner.
	err = store.Insert(ctx, user1.ID, 1, "src", user1.ID)
	require.NoError(t, err)

	// Inserting an already existing assigned owner shouldn't error out, the update
	// is ignored due to `ON CONFLICT DO NOTHING` clause.
	err = store.Insert(ctx, user1.ID, 1, "src", user1.ID)
	require.NoError(t, err)
}

func TestAssignedOwnersStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating users.
	user1, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)
	user2, err := db.Users().Create(ctx, NewUser{Username: "user2"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	store := AssignedOwnersStoreWith(db, logger)

	// Inserting assigned owners.
	err = store.Insert(ctx, user1.ID, 1, "src", user2.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "src", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, user2.ID, 1, "src/abc", user1.ID)
	require.NoError(t, err)

	assertNumberOfOwnersForRepo := func(repoID api.RepoID, length int) {
		summaries, err := store.ListAssignedOwnersForRepo(ctx, repoID)
		require.NoError(t, err)
		assert.Len(t, summaries, length)
	}
	// Deleting an owner with non-existent path.
	err = store.DeleteOwner(ctx, user1.ID, 1, "no/way")
	assert.EqualError(t, err, `cannot delete assigned owner with ID=1 for "no/way" path for repo with ID=1`)
	assertNumberOfOwnersForRepo(1, 3)
	// Deleting an owner with a path for non-existent repo.
	err = store.DeleteOwner(ctx, user1.ID, 1337, "no/way")
	assert.EqualError(t, err, `cannot delete assigned owner with ID=1 for "no/way" path for repo with ID=1337`)
	assertNumberOfOwnersForRepo(1, 3)
	// Deleting an owner with non-existent ID.
	err = store.DeleteOwner(ctx, 1337, 1, "src/abc")
	assert.EqualError(t, err, `cannot delete assigned owner with ID=1337 for "src/abc" path for repo with ID=1`)
	assertNumberOfOwnersForRepo(1, 3)
	// Deleting an existing owner.
	err = store.DeleteOwner(ctx, user2.ID, 1, "src/abc")
	assert.NoError(t, err)
	assertNumberOfOwnersForRepo(1, 2)
}

func TestAssignedOwnersStore_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating users.
	user1, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Inserting assigned owners.
	paths := []string{"a/b/c", "", "foo/bar", "src/main/java/sourcegraph"}
	for _, path := range paths {
		err = db.AssignedOwners().Insert(ctx, user1.ID, 1, path, user1.ID)
		require.NoError(t, err)
	}
	count, err := db.AssignedOwners().CountAssignedOwners(ctx)
	require.NoError(t, err)
	assert.Equal(t, int32(len(paths)), count)
}
