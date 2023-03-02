package usagestats_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

func TestGetOwnershipUsageStatsReposCount(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
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
	stats, err := usagestats.GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	iptr := func(i int32) *int32 { return &i }
	want := &types.OwnershipUsageReposCounts{
		Total:                 iptr(2),
		WithIngestedOwnership: iptr(1),
	}
	if diff := cmp.Diff(want, stats.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsageStates.ReposCount, +want,-got:\n%s", diff)
	}
}

func TestGetOwnershipUsageStatsReposCountNoCodeowners(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	if err := db.Repos().Create(ctx, &types.Repo{Name: "does-not-have-codeowners"}); err != nil {
		t.Fatalf("failed to create test repo: %s", err)
	}
	if err := db.RepoStatistics().CompactRepoStatistics(ctx); err != nil {
		t.Fatalf("failed to recompute stats: %s", err)
	}
	stats, err := usagestats.GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	iptr := func(i int32) *int32 { return &i }
	want := &types.OwnershipUsageReposCounts{
		Total:                 iptr(1),
		WithIngestedOwnership: iptr(0),
	}
	if diff := cmp.Diff(want, stats.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsageStates.ReposCount, +want,-got:\n%s", diff)
	}
}

func TestGetOwnershipUsageStatsReposCountNoRepos(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	if err := db.RepoStatistics().CompactRepoStatistics(ctx); err != nil {
		t.Fatalf("failed to compact repo stats: %s", err)
	}
	if err := db.RepoStatistics().CompactRepoStatistics(ctx); err != nil {
		t.Fatalf("failed to recompute stats: %s", err)
	}
	stats, err := usagestats.GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	iptr := func(i int32) *int32 { return &i }
	want := &types.OwnershipUsageReposCounts{
		Total:                 iptr(0),
		WithIngestedOwnership: iptr(0),
	}
	if diff := cmp.Diff(want, stats.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsageStates.ReposCount, -want+got:\n%s", diff)
	}
}

func TestGetOwnershipUsageStatsReposCountNoStats(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
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
	stats, err := usagestats.GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	iptr := func(i int32) *int32 { return &i }
	want := &types.OwnershipUsageReposCounts{
		// Can have zero repos and one ingested ownership then.
		Total:                 iptr(0),
		WithIngestedOwnership: iptr(1),
	}
	if diff := cmp.Diff(want, stats.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsageStates.ReposCount, +want,-got:\n%s", diff)
	}
}

func TestGetOwnershipUsageStatsFeatureFlagOn(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	if _, err := db.FeatureFlags().CreateBool(ctx, "search-ownership", true); err != nil {
		t.Fatal(err)
	}
	stats, err := usagestats.GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	want := true
	if diff := cmp.Diff(&want, stats.FeatureFlagOn); diff != "" {
		t.Errorf("GetOwnershipUsageStats.FeatureFlagOn, -want+got: %s", diff)
	}
}

func TestGetOwnershipUsageStatsFeatureFlagOff(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	if _, err := db.FeatureFlags().CreateBool(ctx, "search-ownership", false); err != nil {
		t.Fatal(err)
	}
	stats, err := usagestats.GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	want := false
	if diff := cmp.Diff(&want, stats.FeatureFlagOn); diff != "" {
		t.Errorf("GetOwnershipUsageStats.FeatureFlagOn, -want+got: %s", diff)
	}
}

func TestGetOwnershipUsageStatsFeatureFlagAbsent(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	stats, err := usagestats.GetOwnershipUsageStats(ctx, db)
	if err != nil {
		t.Fatalf("GetOwnershipUsageStats err: %s", err)
	}
	want := false
	if diff := cmp.Diff(&want, stats.FeatureFlagOn); diff != "" {
		t.Errorf("GetOwnershipUsageStats.FeatureFlagOn, -want+got: %s", diff)
	}
}

// TODO: Event logs stats tests pending landing #48493
