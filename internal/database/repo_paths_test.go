package database

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type fakeRepoTreeCounts map[string]int

func (f fakeRepoTreeCounts) Iterate(fn func(string, int) error) error {
	for path, count := range f {
		if err := fn(path, count); err != nil {
			return err
		}
	}
	return nil
}

func TestUpdateFileCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	d := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	// Create repo
	repo := mustCreate(ctx, t, d, &types.Repo{Name: "a/b"})

	// Insert new path
	counts := fakeRepoTreeCounts{"new_path": 10}
	timestamp := time.Now()
	updatedRows, err := d.RepoPaths().UpdateFileCounts(ctx, repo.ID, counts, timestamp)
	require.NoError(t, err)
	assert.Equal(t, updatedRows, 1)

	// Update existing path
	counts = fakeRepoTreeCounts{"new_path": 20}
	updatedRows, err = d.RepoPaths().UpdateFileCounts(ctx, repo.ID, counts, timestamp)
	require.NoError(t, err)
	assert.Equal(t, updatedRows, 1)
}

func TestAggregateFileCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	d := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create repos
	repo1 := mustCreate(ctx, t, d, &types.Repo{Name: "a/b"})
	repo2 := mustCreate(ctx, t, d, &types.Repo{Name: "c/d"})

	// Check counts without data.
	count, err := d.RepoPaths().AggregateFileCount(ctx, TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, int32(0), count)

	counts1 := fakeRepoTreeCounts{
		"":      30,
		"path1": 10,
		"path2": 20,
	}
	timestamp := time.Now()
	_, err = d.RepoPaths().UpdateFileCounts(ctx, repo1.ID, counts1, timestamp)
	require.NoError(t, err)
	counts2 := fakeRepoTreeCounts{
		"":      50,
		"path3": 50,
	}
	_, err = d.RepoPaths().UpdateFileCounts(ctx, repo2.ID, counts2, timestamp)
	require.NoError(t, err)

	// Aggregate counts for single path in repo1
	count, err = d.RepoPaths().AggregateFileCount(ctx, TreeLocationOpts{
		Path:   "path1",
		RepoID: repo1.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(counts1["path1"]), count)

	// Aggregate counts for root path in repo1
	count, err = d.RepoPaths().AggregateFileCount(ctx, TreeLocationOpts{
		RepoID: repo1.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(counts1[""]), count)

	// Aggregate counts for all repos
	count, err = d.RepoPaths().AggregateFileCount(ctx, TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, int32(counts1[""]+counts2[""]), count)
}
