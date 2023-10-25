package database

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	teamName  = "a-team"
	teamName2 = "a2-team"
)

func TestAssignedTeamsStore_ListAssignedTeamsForRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating a user and 2 teams.
	user1, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)
	team1 := createTeam(t, ctx, db, teamName)
	team2 := createTeam(t, ctx, db, teamName2)

	// Creating 2 repos.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)
	err = db.Repos().Create(ctx, &types.Repo{ID: 2, Name: "github.com/sourcegraph/sourcegraph2"})
	require.NoError(t, err)

	// Inserting assigned teams.
	store := AssignedTeamsStoreWith(db, logger)
	err = store.Insert(ctx, team1.ID, 1, "src", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, team2.ID, 1, "src/abc", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, team2.ID, 1, "src/def", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, team1.ID, 1, "", user1.ID)
	require.NoError(t, err)

	// Getting assigned teams for a non-existent repo.
	teams, err := store.ListAssignedTeamsForRepo(ctx, 1337)
	require.NoError(t, err)
	assert.Empty(t, teams)

	// Getting assigned teams for a repo without owners.
	teams, err = store.ListAssignedTeamsForRepo(ctx, 2)
	require.NoError(t, err)
	assert.Empty(t, teams)

	// Getting assigned teams for a given repo.
	teams, err = store.ListAssignedTeamsForRepo(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, teams, 4)
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].FilePath < teams[j].FilePath
	})
	// We are checking everything except timestamps, non-zero check is sufficient for them.
	assert.Equal(t, teams[0], &AssignedTeamSummary{OwnerTeamID: team1.ID, RepoID: 1, FilePath: "", WhoAssignedUserID: 1, AssignedAt: teams[0].AssignedAt})
	assert.NotZero(t, teams[0].AssignedAt)
	assert.Equal(t, teams[1], &AssignedTeamSummary{OwnerTeamID: team1.ID, RepoID: 1, FilePath: "src", WhoAssignedUserID: 1, AssignedAt: teams[1].AssignedAt})
	assert.NotZero(t, teams[1].AssignedAt)
	assert.Equal(t, teams[2], &AssignedTeamSummary{OwnerTeamID: team2.ID, RepoID: 1, FilePath: "src/abc", WhoAssignedUserID: 1, AssignedAt: teams[2].AssignedAt})
	assert.NotZero(t, teams[2].AssignedAt)
	assert.Equal(t, teams[3], &AssignedTeamSummary{OwnerTeamID: team2.ID, RepoID: 1, FilePath: "src/def", WhoAssignedUserID: 1, AssignedAt: teams[3].AssignedAt})
	assert.NotZero(t, teams[3].AssignedAt)
}

func TestAssignedTeamsStore_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating a user and a team.
	user1, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)
	team := createTeam(t, ctx, db, teamName)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	store := AssignedTeamsStoreWith(db, logger)

	// Inserting assigned team for non-existing repo, which led to failing to ensure
	// repo paths.
	err = store.Insert(ctx, team.ID, 1337, "src", user1.ID)
	assert.EqualError(t, err, `cannot insert repo paths`)

	// Successfully inserting assigned team.
	err = store.Insert(ctx, team.ID, 1, "src", user1.ID)
	require.NoError(t, err)

	// Inserting an already existing assigned team shouldn't error out, the update
	// is ignored due to `ON CONFLICT DO NOTHING` clause.
	err = store.Insert(ctx, team.ID, 1, "src", user1.ID)
	require.NoError(t, err)
}

func TestAssignedTeamsStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating a user and 2 teams.
	user1, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)
	team1 := createTeam(t, ctx, db, teamName)
	team2 := createTeam(t, ctx, db, teamName2)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	store := AssignedTeamsStoreWith(db, logger)

	// Inserting assigned owners.
	err = store.Insert(ctx, team1.ID, 1, "src", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, team2.ID, 1, "src", user1.ID)
	require.NoError(t, err)
	err = store.Insert(ctx, team2.ID, 1, "src/abc", user1.ID)
	require.NoError(t, err)

	assertNumberOfTeamsForRepo := func(repoID api.RepoID, length int) {
		summaries, err := store.ListAssignedTeamsForRepo(ctx, repoID)
		require.NoError(t, err)
		assert.Len(t, summaries, length)
	}
	// Deleting an owner team with non-existent path.
	err = store.DeleteOwnerTeam(ctx, user1.ID, 1, "no/way")
	assert.EqualError(t, err, `cannot delete assigned owner team with ID=1 for "no/way" path for repo with ID=1`)
	assertNumberOfTeamsForRepo(1, 3)
	// Deleting an owner with a path for non-existent repo.
	err = store.DeleteOwnerTeam(ctx, user1.ID, 1337, "no/way")
	assert.EqualError(t, err, `cannot delete assigned owner team with ID=1 for "no/way" path for repo with ID=1337`)
	assertNumberOfTeamsForRepo(1, 3)
	// Deleting an owner with non-existent ID.
	err = store.DeleteOwnerTeam(ctx, 1337, 1, "src/abc")
	assert.EqualError(t, err, `cannot delete assigned owner team with ID=1337 for "src/abc" path for repo with ID=1`)
	assertNumberOfTeamsForRepo(1, 3)
	// Deleting an existing owner.
	err = store.DeleteOwnerTeam(ctx, team2.ID, 1, "src/abc")
	assert.NoError(t, err)
	assertNumberOfTeamsForRepo(1, 2)
}

func createTeam(t *testing.T, ctx context.Context, db DB, teamName string) *types.Team {
	t.Helper()
	team, err := db.Teams().CreateTeam(ctx, &types.Team{Name: teamName})
	require.NoError(t, err)
	return team
}
