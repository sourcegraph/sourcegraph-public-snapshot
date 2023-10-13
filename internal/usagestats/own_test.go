package usagestats

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestGetOwnershipUsageStatsReposCount(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	if err := db.Repos().Create(ctx, &types.Repo{Name: "does-not-have-codeowners"}); err != nil {
		t.Fatalf("failed to create test repo: %s", err)
	}
	if err := db.Repos().Create(ctx, &types.Repo{Name: "has-codeowners"}); err != nil {
		t.Fatalf("failed to create test repo: %s", err)
	}
	repo, err := db.Repos().GetByName(ctx, "has-codeowners")
	if err != nil {
		t.Fatalf("failed to get test repo: %s", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO codeowners (repo_id, contents, contents_proto)
		VALUES ($1, $2, $3)
	`, repo.ID, `test-file @test-owner`, []byte{}).Err(); err != nil {
		t.Fatalf("failed to create codeowners data: %s", err)
	}
	if err := db.RepoStatistics().CompactRepoStatistics(ctx); err != nil {
		t.Fatalf("failed to recompute stats: %s", err)
	}
	stats, err := GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	want := &types.OwnershipUsageReposCounts{
		Total:                 pointers.Ptr(int32(2)),
		WithIngestedOwnership: pointers.Ptr(int32(1)),
	}
	if diff := cmp.Diff(want, stats.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsageStates.ReposCount, +want,-got:\n%s", diff)
	}
}

func TestGetOwnershipUsageStatsReposCountNoCodeowners(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	if err := db.Repos().Create(ctx, &types.Repo{Name: "does-not-have-codeowners"}); err != nil {
		t.Fatalf("failed to create test repo: %s", err)
	}
	if err := db.RepoStatistics().CompactRepoStatistics(ctx); err != nil {
		t.Fatalf("failed to recompute stats: %s", err)
	}
	stats, err := GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	want := &types.OwnershipUsageReposCounts{
		Total:                 pointers.Ptr(int32(1)),
		WithIngestedOwnership: pointers.Ptr(int32(0)),
	}
	if diff := cmp.Diff(want, stats.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsageStates.ReposCount, +want,-got:\n%s", diff)
	}
}

func TestGetOwnershipUsageStatsReposCountNoRepos(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	if err := db.RepoStatistics().CompactRepoStatistics(ctx); err != nil {
		t.Fatalf("failed to compact repo stats: %s", err)
	}
	if err := db.RepoStatistics().CompactRepoStatistics(ctx); err != nil {
		t.Fatalf("failed to recompute stats: %s", err)
	}
	stats, err := GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	want := &types.OwnershipUsageReposCounts{
		Total:                 pointers.Ptr(int32(0)),
		WithIngestedOwnership: pointers.Ptr(int32(0)),
	}
	if diff := cmp.Diff(want, stats.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsageStates.ReposCount, -want+got:\n%s", diff)
	}
}

func TestGetOwnershipUsageStatsReposCountStatsNotCompacted(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	if err := db.Repos().Create(ctx, &types.Repo{Name: "does-not-have-codeowners"}); err != nil {
		t.Fatalf("failed to create test repo: %s", err)
	}
	if err := db.Repos().Create(ctx, &types.Repo{Name: "has-codeowners"}); err != nil {
		t.Fatalf("failed to create test repo: %s", err)
	}
	repo, err := db.Repos().GetByName(ctx, "has-codeowners")
	if err != nil {
		t.Fatalf("failed to get test repo: %s", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO codeowners (repo_id, contents, contents_proto)
		VALUES ($1, $2, $3)
	`, repo.ID, `test-file @test-owner`, []byte{}).Err(); err != nil {
		t.Fatalf("failed to create codeowners data: %s", err)
	}
	// No repo stats computation.
	stats, err := GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	want := &types.OwnershipUsageReposCounts{
		// Can have zero repos and one ingested ownership then.
		Total:                 pointers.Ptr(int32(2)),
		WithIngestedOwnership: pointers.Ptr(int32(1)),
	}
	if diff := cmp.Diff(want, stats.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsageStates.ReposCount, -want,+got:\n%s", diff)
	}
}

func TestGetOwnershipUsageStatsAggregatedStats(t *testing.T) {
	// Not parallel as we're replacing timeNow.
	now := time.Date(2020, 10, 13, 12, 0, 0, 0, time.UTC) // Tuesday
	backupTimeNow := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() { timeNow = backupTimeNow })
	logger := logtest.Scoped(t)
	// Event names are different, so the same database can be reused.
	for eventName, lens := range map[string]func(*types.OwnershipUsageStatistics) *types.OwnershipUsageStatisticsActiveUsers{
		"SelectFileOwnersSearch": func(stats *types.OwnershipUsageStatistics) *types.OwnershipUsageStatisticsActiveUsers {
			return stats.SelectFileOwnersSearch
		},
		"FileHasOwnerSearch": func(stats *types.OwnershipUsageStatistics) *types.OwnershipUsageStatisticsActiveUsers {
			return stats.FileHasOwnerSearch
		},
		"OwnershipPanelOpened": func(stats *types.OwnershipUsageStatistics) *types.OwnershipUsageStatisticsActiveUsers {
			return stats.OwnershipPanelOpened
		},
	} {
		t.Run(eventName, func(t *testing.T) {
			t.Parallel()
			db := database.NewDB(logger, dbtest.NewDB(t))
			ctx := context.Background()
			if err := db.EventLogs().Insert(ctx, &database.Event{
				UserID: 1,
				Name:   eventName,
				Source: "BACKEND",
				// Monday, same week & month as now: MAU+1, WAU+1, DAU - no change.
				Timestamp: time.Date(2020, 10, 12, 12, 0, 0, 0, time.UTC),
			}); err != nil {
				t.Fatal(err)
			}
			if err := db.EventLogs().Insert(ctx, &database.Event{
				UserID: 2,
				Name:   eventName,
				Source: "BACKEND",
				// Saturday, week before, same month, different user: MAU+1, WAU, DAU - no change.
				Timestamp: time.Date(2020, 10, 10, 12, 0, 0, 0, time.UTC),
			}); err != nil {
				t.Fatal(err)
			}
			stats, err := GetOwnershipUsageStats(ctx, db)
			if err != nil {
				t.Fatalf("GetOwnershipUsageStats err: %s", err)
			}
			want := &types.OwnershipUsageStatisticsActiveUsers{
				MAU: pointers.Ptr(int32(2)),
				WAU: pointers.Ptr(int32(1)),
				DAU: pointers.Ptr(int32(0)),
			}
			if diff := cmp.Diff(want, lens(stats)); diff != "" {
				t.Errorf("GetOwnershipUsageStats().%s -want+got: %s", eventName, diff)
			}
		})
	}
}

func TestGetOwnershipUsageStatsAssignedOwnersCount(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	var repoID api.RepoID = 1
	require.NoError(t, db.Repos().Create(ctx, &types.Repo{
		ID:   repoID,
		Name: "github.com/sourcegraph/sourcegraph",
	}))
	user, err := db.Users().Create(ctx, database.NewUser{Username: "foo"})
	require.NoError(t, err)
	paths := []string{"src", "test", "docs/README.md"}
	for _, p := range paths {
		require.NoError(t, db.AssignedOwners().Insert(ctx, user.ID, repoID, p, user.ID))
	}
	stats, err := GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	wantCount := int32(len(paths))
	assert.Equal(t, &wantCount, stats.AssignedOwnersCount)
}
