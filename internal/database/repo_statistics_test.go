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
	s := &repoStatisticsStore{Store: basestore.NewWithHandle(db.Handle())}

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

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 6, NotCloned: 6, SoftDeleted: 0,
	}, []GitserverReposStatistic{
		{ShardID: "", Total: 6, NotCloned: 6},
	})

	// Move to to shards[0] as cloning
	setCloneStatus(t, db, repos[0].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[1].Name, shards[0], types.CloneStatusCloning)

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 6, SoftDeleted: 0, NotCloned: 4, Cloning: 2,
	}, []GitserverReposStatistic{
		{ShardID: "", Total: 4, NotCloned: 4},
		{ShardID: shards[0], Total: 2, Cloning: 2},
	})

	// Move two repos to shards[1] as cloning
	setCloneStatus(t, db, repos[2].Name, shards[1], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[3].Name, shards[1], types.CloneStatusCloning)
	// Move two repos to shards[2] as cloning
	setCloneStatus(t, db, repos[4].Name, shards[2], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[5].Name, shards[2], types.CloneStatusCloning)

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 6, SoftDeleted: 0, Cloning: 6,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2},
		{ShardID: shards[1], Total: 2, Cloning: 2},
		{ShardID: shards[2], Total: 2, Cloning: 2},
	})

	// Move from shards[0] to shards[2] and change status
	setCloneStatus(t, db, repos[2].Name, shards[2], types.CloneStatusCloned)
	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 6, SoftDeleted: 0, Cloning: 5, Cloned: 1,
	}, []GitserverReposStatistic{
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
	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 5, SoftDeleted: 1, Cloning: 5,
	}, []GitserverReposStatistic{
		// But gitserverReposStatistics is unchanged
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2},
		{ShardID: shards[1], Total: 1, Cloning: 1},
		{ShardID: shards[2], Total: 3, Cloning: 2, Cloned: 1},
	})

	// Until we remove it from disk in gitserver, which causes the clone status
	// to be set to not_cloned:
	setCloneStatus(t, db, deletedRepoName, shards[2], types.CloneStatusNotCloned)

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		// Global stats are unchanged
		Total: 5, SoftDeleted: 1, Cloning: 5,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2},
		{ShardID: shards[1], Total: 1, Cloning: 1},
		// But now it's reflected as NotCloned
		{ShardID: shards[2], Total: 3, Cloning: 2, NotCloned: 1},
	})

	// Now we set errors on 2 non-deleted repositories
	setLastError(t, db, repos[0].Name, shards[0], "internet broke repo-1")
	setLastError(t, db, repos[4].Name, shards[2], "internet broke repo-3")
	assertRepoStatistics(t, ctx, s, RepoStatistics{
		// Only FailedFetch changed
		Total: 5, SoftDeleted: 1, Cloning: 5, FailedFetch: 2,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2, FailedFetch: 1},
		{ShardID: shards[1], Total: 1, Cloning: 1, FailedFetch: 0},
		{ShardID: shards[2], Total: 3, Cloning: 2, NotCloned: 1, FailedFetch: 1},
	})
	// Now we move a repo and set an error
	setLastError(t, db, repos[1].Name, shards[1], "internet broke repo-2")
	assertRepoStatistics(t, ctx, s, RepoStatistics{
		// Only FailedFetch changed
		Total: 5, SoftDeleted: 1, Cloning: 5, FailedFetch: 3,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 1, Cloning: 1, FailedFetch: 1},
		{ShardID: shards[1], Total: 2, Cloning: 2, FailedFetch: 1},
		{ShardID: shards[2], Total: 3, Cloning: 2, NotCloned: 1, FailedFetch: 1},
	})
}

func TestRepoStatistics_DeleteAndUndelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	s := &repoStatisticsStore{Store: basestore.NewWithHandle(db.Handle())}

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

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 6, NotCloned: 6, SoftDeleted: 0,
	}, []GitserverReposStatistic{
		{ShardID: "", Total: 6, NotCloned: 6},
	})

	// Move to to shards[0] as cloning
	setCloneStatus(t, db, repos[0].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[1].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[2].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[3].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[4].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[5].Name, shards[0], types.CloneStatusCloning)

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 6, SoftDeleted: 0, Cloning: 6,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 6, Cloning: 6},
	})

	// Soft delete repos
	if err := db.Repos().Delete(ctx, repos[2].ID); err != nil {
		t.Fatal(err)
	}
	deletedRepoName := queryRepoName(t, ctx, s, repos[2].ID)

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		// correct
		Total: 5, SoftDeleted: 1, Cloning: 5,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 6, Cloning: 6},
	})

	// Until we remove it from disk in gitserver, which causes the clone status
	// to be set to not_cloned:
	setCloneStatus(t, db, deletedRepoName, shards[0], types.CloneStatusNotCloned)

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 5, SoftDeleted: 1, Cloning: 5,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 6, Cloning: 5, NotCloned: 1},
	})

	// Undelete it
	err := s.Exec(ctx, sqlf.Sprintf("UPDATE repo SET deleted_at = NULL WHERE name = %s;", deletedRepoName))
	if err != nil {
		t.Fatalf("failed to query repo name: %s", err)
	}
	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 6, SoftDeleted: 0, Cloning: 5, NotCloned: 1,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 6, Cloning: 5, NotCloned: 1},
	})

	// reshard and clone
	setCloneStatus(t, db, deletedRepoName, shards[1], types.CloneStatusCloning)

	assertRepoStatistics(t, ctx, s, RepoStatistics{
		Total: 6, SoftDeleted: 0, Cloning: 6, NotCloned: 0,
	}, []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 5, Cloning: 5, NotCloned: 0},
		{ShardID: shards[1], Total: 1, Cloning: 1, NotCloned: 0},
	})
}

func TestRepoStatistics_Compaction(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	s := &repoStatisticsStore{Store: basestore.NewWithHandle(db.Handle())}

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

	// Trigger 9 insertions into repo_statistics table:
	createTestRepos(ctx, t, db, repos)
	setCloneStatus(t, db, repos[0].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[1].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[2].Name, shards[1], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[3].Name, shards[1], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[4].Name, shards[2], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[5].Name, shards[2], types.CloneStatusCloning)
	setLastError(t, db, repos[0].Name, shards[0], "internet broke repo-1")
	setLastError(t, db, repos[4].Name, shards[2], "internet broke repo-3")
	// Safety check that the counts are right:
	wantRepoStatistics := RepoStatistics{
		Total: 6, Cloning: 6, FailedFetch: 2,
	}
	wantGitserverReposStatistics := []GitserverReposStatistic{
		{ShardID: ""},
		{ShardID: shards[0], Total: 2, Cloning: 2, FailedFetch: 1},
		{ShardID: shards[1], Total: 2, Cloning: 2},
		{ShardID: shards[2], Total: 2, Cloning: 2, FailedFetch: 1},
	}
	assertRepoStatistics(t, ctx, s, wantRepoStatistics, wantGitserverReposStatistics)

	// The initial insert in the migration also added a row, which means we want:
	wantCount := 10
	count := queryRepoStatisticsCount(t, ctx, s)
	if count != wantCount {
		t.Fatalf("wrong statistics count. have=%d, want=%d", count, wantCount)
	}

	// Now we compact the rows into a single row:
	if err := s.CompactRepoStatistics(ctx); err != nil {
		t.Fatalf("GetRepoStatistics failed: %s", err)
	}

	// We should be left with 1 row
	wantCount = 1
	count = queryRepoStatisticsCount(t, ctx, s)
	if count != wantCount {
		t.Fatalf("wrong statistics count. have=%d, want=%d", count, wantCount)
	}

	// And counts should still be the same
	assertRepoStatistics(t, ctx, s, wantRepoStatistics, wantGitserverReposStatistics)

	// Safety check: add another event and make sure row count goes up again
	setCloneStatus(t, db, repos[5].Name, shards[2], types.CloneStatusCloned)
	wantCount = 2
	count = queryRepoStatisticsCount(t, ctx, s)
	if count != wantCount {
		t.Fatalf("wrong statistics count. have=%d, want=%d", count, wantCount)
	}
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

func queryRepoStatisticsCount(t *testing.T, ctx context.Context, s *repoStatisticsStore) int {
	t.Helper()
	var count int
	err := s.QueryRow(ctx, sqlf.Sprintf("SELECT COUNT(*) FROM repo_statistics;")).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query repo name: %s", err)
	}
	return count
}

func setCloneStatus(t *testing.T, db DB, repoName api.RepoName, shard string, status types.CloneStatus) {
	t.Helper()
	if err := db.GitserverRepos().SetCloneStatus(context.Background(), repoName, status, shard); err != nil {
		t.Fatalf("failed to set clone status for repo %s: %s", repoName, err)
	}
}

func setLastError(t *testing.T, db DB, repoName api.RepoName, shard string, msg string) {
	t.Helper()
	if err := db.GitserverRepos().SetLastError(context.Background(), repoName, msg, shard); err != nil {
		t.Fatalf("failed to set clone status for repo %s: %s", repoName, err)
	}
}

func assertRepoStatistics(t *testing.T, ctx context.Context, s RepoStatisticsStore, wantRepoStats RepoStatistics, wantGitserverStats []GitserverReposStatistic) {
	t.Helper()

	haveRepoStats, err := s.GetRepoStatistics(ctx)
	if err != nil {
		t.Fatalf("GetRepoStatistics failed: %s", err)
	}

	if diff := cmp.Diff(haveRepoStats, wantRepoStats); diff != "" {
		t.Errorf("repoStatistics differ: %s", diff)
	}

	haveGitserverStats, err := s.GetGitserverReposStatistics(ctx)
	if err != nil {
		t.Fatalf("GetRepoStatistics failed: %s", err)
	}

	sort.Slice(haveGitserverStats, func(i, j int) bool { return haveGitserverStats[i].ShardID < haveGitserverStats[j].ShardID })
	sort.Slice(wantGitserverStats, func(i, j int) bool { return wantGitserverStats[i].ShardID < wantGitserverStats[j].ShardID })

	if diff := cmp.Diff(haveGitserverStats, wantGitserverStats); diff != "" {
		t.Fatalf("gitserverReposStatistics differ: %s", diff)
	}
}
