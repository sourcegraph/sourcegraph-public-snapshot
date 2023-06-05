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

type fakeCodeownersWalk map[string][]TreeCounts

func (w fakeCodeownersWalk) Iterate(f func(string, TreeCounts) error) error {
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
	walk := fakeCodeownersWalk{
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
	updatedRows, err := d.OwnershipStats().UpdateIndividualCounts(ctx, repo.ID, walk, timestamp)
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
	walk1 := fakeCodeownersWalk{
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
	_, err := d.OwnershipStats().UpdateIndividualCounts(ctx, repo1.ID, walk1, timestamp)
	require.NoError(t, err)
	walk2 := fakeCodeownersWalk{
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
	_, err = d.OwnershipStats().UpdateIndividualCounts(ctx, repo2.ID, walk2, timestamp)
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
		want := []TreeCounts{
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
		want := []TreeCounts{
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
		want := []TreeCounts{
			{CodeownersReference: "ownerA", CodeownedFileCount: 22}, // from both repos
			{CodeownersReference: "ownerC", CodeownedFileCount: 10}, // only repo2
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},  // only repo1
		}
		assert.DeepEqual(t, want, got)
	})
}
