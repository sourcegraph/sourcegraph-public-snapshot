package database

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type fakeCodeownersWalk map[string][]CodeownedTreeCounts

func (w fakeCodeownersWalk) Walk(f func(path string, ownerCounts []CodeownedTreeCounts) error) error {
	for path, owners := range w {
		if err := f(path, owners); err != nil {
			return err
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
	repo := mustCreate(ctx, t, d, &types.Repo{Name: "a/b"})
	_, err := ensureRepoPaths(ctx, d.(*db).Store, []string{"file1", "file2"}, repo.ID)
	require.NoError(t, err)
	walk := fakeCodeownersWalk{
		"": []CodeownedTreeCounts{
			{Reference: "ownerA", FileCount: 2},
			{Reference: "ownerB", FileCount: 1},
		},
		"file1": []CodeownedTreeCounts{
			{Reference: "ownerA", FileCount: 1},
			{Reference: "ownerB", FileCount: 1},
		},
		"file2": []CodeownedTreeCounts{
			{Reference: "ownerA", FileCount: 1},
		},
	}
	timestamp := time.Now()
	updatedRows, err := d.OwnershipStats().UpdateIndividualCounts(ctx, repo.ID, walk, timestamp)
	require.NoError(t, err)
	if got, want := updatedRows, 5; got != want {
		t.Errorf("UpdateIndividualCounts, updated rows, got %d, want %d", got, want)
	}
}
