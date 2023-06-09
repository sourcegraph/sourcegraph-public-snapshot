package database

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type fakeCodeownersStats map[string][]PathCodeownersCounts

func (w fakeCodeownersStats) Iterate(f func(string, PathCodeownersCounts) error) error {
	for path, owners := range w {
		for _, o := range owners {
			if err := f(path, o); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestUpdateIndividualCountsSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	d := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	// 1. Setup repo and paths:
	repo := mustCreate(ctx, t, d, &types.Repo{Name: "a/b"})
	// 2. Insert countsg:
	iter := fakeCodeownersStats{
		"": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 2},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		},
		"file1": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		},
		"file2": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
		},
	}
	timestamp := time.Now()
	updatedRows, err := d.OwnershipStats().UpdateIndividualCounts(ctx, repo.ID, iter, timestamp)
	require.NoError(t, err)
	if got, want := updatedRows, 5; got != want {
		t.Errorf("UpdateIndividualCounts, updated rows, got %d, want %d", got, want)
	}
}

func TestQueryIndividualCountsAggregation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	d := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	// 1. Setup repos and paths:
	repo1 := mustCreate(ctx, t, d, &types.Repo{Name: "a/b"})
	repo2 := mustCreate(ctx, t, d, &types.Repo{Name: "a/c"})
	// 2. Insert counts:
	timestamp := time.Now()
	iter1 := fakeCodeownersStats{
		"": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 2},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		},
		"file1": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		},
		"file2": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
		},
	}
	_, err := d.OwnershipStats().UpdateIndividualCounts(ctx, repo1.ID, iter1, timestamp)
	require.NoError(t, err)
	iter2 := fakeCodeownersStats{
		"": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 20},
			{CodeownersReference: "ownerC", CodeownedFileCount: 10},
		},
		"file3": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 10},
			{CodeownersReference: "ownerC", CodeownedFileCount: 10},
		},
		"file4": {
			{CodeownersReference: "ownerC", CodeownedFileCount: 10},
		},
	}
	_, err = d.OwnershipStats().UpdateIndividualCounts(ctx, repo2.ID, iter2, timestamp)
	require.NoError(t, err)
	// 3. Query with or without aggregation:
	t.Run("query single file", func(t *testing.T) {
		opts := TreeLocationOpts{
			RepoID: repo1.ID,
			Path:   "file1",
		}
		var limitOffset *LimitOffset
		got, err := d.OwnershipStats().QueryIndividualCounts(ctx, opts, limitOffset)
		require.NoError(t, err)
		want := []PathCodeownersCounts{
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		}
		assert.DeepEqual(t, want, got)
	})
	t.Run("query single repo", func(t *testing.T) {
		opts := TreeLocationOpts{
			RepoID: repo1.ID,
		}
		var limitOffset *LimitOffset
		got, err := d.OwnershipStats().QueryIndividualCounts(ctx, opts, limitOffset)
		require.NoError(t, err)
		want := []PathCodeownersCounts{
			{CodeownersReference: "ownerA", CodeownedFileCount: 2},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		}
		assert.DeepEqual(t, want, got)
	})
	t.Run("query whole instance", func(t *testing.T) {
		opts := TreeLocationOpts{}
		var limitOffset *LimitOffset
		got, err := d.OwnershipStats().QueryIndividualCounts(ctx, opts, limitOffset)
		require.NoError(t, err)
		want := []PathCodeownersCounts{
			{CodeownersReference: "ownerA", CodeownedFileCount: 22}, // from both repos
			{CodeownersReference: "ownerC", CodeownedFileCount: 10}, // only repo2
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},  // only repo1
		}
		assert.DeepEqual(t, want, got)
	})
}

// fakeAggregateStatsIterator contains aggregate counts by file path.
type fakeAggregateStatsIterator map[string]PathAggregateCounts

func (w fakeAggregateStatsIterator) Iterate(f func(string, PathAggregateCounts) error) error {
	for path, counts := range w {
		if err := f(path, counts); err != nil {
			return err
		}
	}
	return nil
}

func TestUpdateAggregateCountsSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	d := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	// 1. Setup repo and paths:
	repo := mustCreate(ctx, t, d, &types.Repo{Name: "a/b"})
	// 2. Insert aggregate counts:
	iter := fakeAggregateStatsIterator{
		"":              {CodeownedFileCount: 3},
		"dir1":          {CodeownedFileCount: 2},
		"dir2/file1.go": {CodeownedFileCount: 1},
	}
	timestamp := time.Now()
	updatedRows, err := d.OwnershipStats().UpdateAggregateCounts(ctx, repo.ID, iter, timestamp)
	require.NoError(t, err)
	if got, want := updatedRows, len(iter); got != want {
		t.Errorf("UpdateAggregateCounts, updated rows, got %d, want %d", got, want)
	}
}

func TestQueryAggregateCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	d := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	// 1. Setup repo and paths:
	repo1 := mustCreate(ctx, t, d, &types.Repo{Name: "a/b"})
	repo2 := mustCreate(ctx, t, d, &types.Repo{Name: "a/c"})

	t.Run("no data - query single repo", func(t *testing.T) {
		opts := TreeLocationOpts{
			RepoID: repo1.ID,
		}
		got, err := d.OwnershipStats().QueryAggregateCounts(ctx, opts)
		require.NoError(t, err)
		want := []PathAggregateCounts{
			{CodeownedFileCount: 0},
		}
		assert.DeepEqual(t, want, got)
	})

	t.Run("no data - query all", func(t *testing.T) {
		opts := TreeLocationOpts{}
		got, err := d.OwnershipStats().QueryAggregateCounts(ctx, opts)
		require.NoError(t, err)
		want := []PathAggregateCounts{
			{CodeownedFileCount: 0},
		}
		assert.DeepEqual(t, want, got)
	})

	// 2. Insert aggregate counts:
	timestamp := time.Now()
	repo1Counts := fakeAggregateStatsIterator{
		"":              {CodeownedFileCount: 3},
		"dir1":          {CodeownedFileCount: 2},
		"dir2/file1.go": {CodeownedFileCount: 1},
	}
	_, err := d.OwnershipStats().UpdateAggregateCounts(ctx, repo1.ID, repo1Counts, timestamp)
	require.NoError(t, err)
	repo2Counts := fakeAggregateStatsIterator{
		"": {CodeownedFileCount: 10}, // Just the root data
	}
	_, err = d.OwnershipStats().UpdateAggregateCounts(ctx, repo2.ID, repo2Counts, timestamp)
	require.NoError(t, err)

	// 3. Query aggregate counts:
	t.Run("query single file", func(t *testing.T) {
		opts := TreeLocationOpts{
			RepoID: repo1.ID,
			Path:   "dir2/file1.go",
		}
		got, err := d.OwnershipStats().QueryAggregateCounts(ctx, opts)
		require.NoError(t, err)
		want := []PathAggregateCounts{
			{CodeownedFileCount: 1},
		}
		assert.DeepEqual(t, want, got)
	})

	t.Run("query single dir", func(t *testing.T) {
		opts := TreeLocationOpts{
			RepoID: repo1.ID,
			Path:   "dir1",
		}
		got, err := d.OwnershipStats().QueryAggregateCounts(ctx, opts)
		require.NoError(t, err)
		want := []PathAggregateCounts{
			{CodeownedFileCount: 2},
		}
		assert.DeepEqual(t, want, got)
	})

	t.Run("query repo root", func(t *testing.T) {
		opts := TreeLocationOpts{
			RepoID: repo1.ID,
		}
		got, err := d.OwnershipStats().QueryAggregateCounts(ctx, opts)
		require.NoError(t, err)
		want := []PathAggregateCounts{
			{CodeownedFileCount: 3},
		}
		assert.DeepEqual(t, want, got)
	})

	t.Run("query whole instance", func(t *testing.T) {
		opts := TreeLocationOpts{}
		got, err := d.OwnershipStats().QueryAggregateCounts(ctx, opts)
		require.NoError(t, err)
		want := []PathAggregateCounts{
			{CodeownedFileCount: 13},
		}
		assert.DeepEqual(t, want, got)
	})
}
