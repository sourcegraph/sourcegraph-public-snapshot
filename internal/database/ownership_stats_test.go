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
	_, err := ensureRepoPaths(ctx, d.(*db).Store, []string{"file1", "file2"}, repo.ID)
	require.NoError(t, err)
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
	// 3. Query back counts for file:
	opts := TreeLocationOpts{
		RepoID: repo.ID,
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
	// 4. Query back counts for repo root:
	opts = TreeLocationOpts{
		RepoID: repo.ID,
	}
	got, err = d.OwnershipStats().QueryIndividualCounts(ctx, opts, limitOffset)
	require.NoError(t, err)
	want = []TreeCounts{
		{CodeownersReference: "ownerA", CodeownedFileCount: 2},
		{CodeownersReference: "ownerB", CodeownedFileCount: 1},
	}
	assert.DeepEqual(t, want, got)
}
