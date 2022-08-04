package database

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		Total:       6,
		SoftDeleted: 0,
	})
	assertGitserverReposStatistics(t, ctx, s, []gitserverReposStatistics{
		{
			ShardID:   "",
			Total:     6,
			NotCloned: 6,
		},
	})

	setCloneStatus(t, db, repos[0].Name, shards[0], types.CloneStatusCloning)
	setCloneStatus(t, db, repos[1].Name, shards[1], types.CloneStatusCloning)

	assertGitserverReposStatistics(t, ctx, s, []gitserverReposStatistics{
		{
			ShardID:   "",
			Total:     6,
			NotCloned: 6,
		},
		{
			ShardID: shards[0],
			Total:   2,
			Cloning: 2,
		},
	})
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
