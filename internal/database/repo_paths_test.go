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
