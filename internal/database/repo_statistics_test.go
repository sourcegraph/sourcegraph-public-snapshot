package database

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRepoStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	s := RepoStatisticsWith(basestore.NewWithHandle(db.Handle()))

	shards := []string{
		"shard-1",
		"shard-2",
		"shard-3",
	}
	repos := types.Repos{
		&types.Repo{Name: "repo1"},
		&types.Repo{Name: "repo2"},
		&types.Repo{Name: "repo3"},
		&types.Repo{Name: "repo4"},
		&types.Repo{Name: "repo5"},
		&types.Repo{Name: "repo6"},
	}

	createTestRepos(ctx, t, db, repos)

	assertRepoStatistics(t, ctx, s, repoStatistics{
		Total: 6, NotCloned: 6, SoftDeleted: 0,
	})
	assertGitserverReposStatistics(t, ctx, s, []gitserverReposStatistics{
		{ShardID: "", Total: 6, NotCloned: 6},
	})

	// Move to to shards[0] as cloning
	setCloneStatus(t, db, repos[0].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[1].Name, shards[0], types.CloneStatusCloning)

	assertRepoStatistics(t, ctx, s, repoStatistics{
		Total: 6, SoftDeleted: 0, NotCloned: 4, Cloning: 2,
	})
	assertGitserverReposStatistics(t, ctx, s, []gitserverReposStatistics{
		{ShardID: "", Total: 4, NotCloned: 4},
		{ShardID: shards[0], Total: 2, Cloning: 2},
	})

	// Move two repos to shards[1] as cloning
	setCloneStatus(t, db, repos[2].Name, shards[1], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[3].Name, shards[1], types.CloneStatusCloning)
	// Move two repos to shards[2] as cloning
	setCloneStatus(t, db, repos[4].Name, shards[2], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[5].Name, shards[2], types.CloneStatusCloning)

	assertRepoStatistics(t, ctx, s, repoStatistics{
		Total: 6, SoftDeleted: 0, Cloning: 6,
	})
	assertGitserverReposStatistics(t, ctx, s, []gitserverReposStatistics{
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2},
		{ShardID: shards[1], Total: 2, Cloning: 2},
		{ShardID: shards[2], Total: 2, Cloning: 2},
	})

	// Move from shards[0] to shards[2] and change status
	setCloneStatus(t, db, repos[2].Name, shards[2], types.CloneStatusCloned)
	assertRepoStatistics(t, ctx, s, repoStatistics{
		Total: 6, SoftDeleted: 0, Cloning: 5, Cloned: 1,
	})
	assertGitserverReposStatistics(t, ctx, s, []gitserverReposStatistics{
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2},
		{ShardID: shards[1], Total: 1, Cloning: 1},
		{ShardID: shards[2], Total: 3, Cloning: 2, Cloned: 1},
	})

	// Soft delete repos
	if err := db.Repos().Delete(ctx, repos[2].ID); err != nil {
		t.Fatal(err)
	}
	deletedRepoName := queryRepoName(t, ctx, s, repos[2].ID)

	// Deletion is reflected in repoStatistics
	assertRepoStatistics(t, ctx, s, repoStatistics{
		Total: 5, SoftDeleted: 1, Cloning: 5,
	})
	// But gitserverReposStatistics is unchanged
	assertGitserverReposStatistics(t, ctx, s, []gitserverReposStatistics{
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2},
		{ShardID: shards[1], Total: 1, Cloning: 1},
		{ShardID: shards[2], Total: 3, Cloning: 2, Cloned: 1},
	})
	// Until we remove it from disk in gitserver, which causes the clone status
	// to be set to not_cloned:
	setCloneStatus(t, db, deletedRepoName, shards[2], types.CloneStatusNotCloned)
	assertGitserverReposStatistics(t, ctx, s, []gitserverReposStatistics{
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2},
		{ShardID: shards[1], Total: 1, Cloning: 1},
		{ShardID: shards[2], Total: 3, Cloning: 2, NotCloned: 1},
	})
	// Global stats are unchanged
	assertRepoStatistics(t, ctx, s, repoStatistics{
		Total: 5, SoftDeleted: 1, Cloning: 5,
	})
}

func queryRepoName(t *testing.T, ctx context.Context, s *repoStatisticsStore, repoID api.RepoID) api.RepoName {
	t.Helper()
	var name api.RepoName
	err := s.QueryRow(ctx, sqlf.Sprintf("SELECT name FROM repo WHERE id = %s", repoID)).Scan(&name)
	if err != nil {
		t.Fatalf("failed to query repo name: %s", err)
	}
	return name
}

func setCloneStatus(t *testing.T, db DB, repoName api.RepoName, shard string, status types.CloneStatus) {
	t.Helper()
	if err := db.GitserverRepos().SetCloneStatus(context.Background(), repoName, status, shard); err != nil {
		t.Fatalf("failed to set clone status for repo %s: %s", repoName, err)
	}
}

func assertRepoStatistics(t *testing.T, ctx context.Context, s *repoStatisticsStore, want repoStatistics) {
	t.Helper()

	have, err := s.GetRepoStatistics(ctx)
	if err != nil {
		t.Fatalf("GetRepoStatistics failed: %s", err)
	}

	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("repoStatistics differ: %s", diff)
	}
}

func assertGitserverReposStatistics(t *testing.T, ctx context.Context, s *repoStatisticsStore, want []gitserverReposStatistics) {
	t.Helper()

	have, err := s.GetGitserverReposStatistics(ctx)
	if err != nil {
		t.Fatalf("GetRepoStatistics failed: %s", err)
	}

	sort.Slice(have, func(i, j int) bool { return have[i].ShardID < have[j].ShardID })
	sort.Slice(want, func(i, j int) bool { return want[i].ShardID < want[j].ShardID })

	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("repoStatistics differ: %s", diff)
	}
}
