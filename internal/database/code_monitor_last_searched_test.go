package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestCodeMonitorStoreLastSearched(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	t.Run("insert get upsert get", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db := NewDB(logger, dbtest.NewDB(t))
		fixtures := populateCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		// Insert
		insertLastSearched := []string{"commit1", "commit2"}
		err := cm.UpsertLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, insertLastSearched)
		require.NoError(t, err)

		// Get
		lastSearched, err := cm.GetLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID)
		require.NoError(t, err)
		require.Equal(t, insertLastSearched, lastSearched)

		// Update
		updateLastSearched := []string{"commit3", "commit4"}
		err = cm.UpsertLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, updateLastSearched)
		require.NoError(t, err)

		// Get
		lastSearched, err = cm.GetLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID)
		require.NoError(t, err)
		require.Equal(t, updateLastSearched, lastSearched)
	})

	t.Run("no error for missing get", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db := NewDB(logger, dbtest.NewDB(t))
		fixtures := populateCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		// GetLastSearched should not return an error for a monitor that hasn't
		// been run yet. It should just return an empty value for lastSearched
		lastSearched, err := cm.GetLastSearched(ctx, fixtures.Monitor.ID+1, 19793)
		require.NoError(t, err)
		require.Empty(t, lastSearched)
	})

	t.Run("no error for missing get", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db := NewDB(logger, dbtest.NewDB(t))
		fixtures := populateCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		// Insert with nil last searched
		err := cm.UpsertLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, nil)
		require.NoError(t, err)

		// Get nil last searched
		lastSearched, err := cm.GetLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID)
		require.NoError(t, err)
		require.Empty(t, lastSearched)

		// Insert with empty last searched
		err = cm.UpsertLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, []string{})
		require.NoError(t, err)

		// Get nil last searched
		lastSearched, err = cm.GetLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID)
		require.NoError(t, err)
		require.Empty(t, lastSearched)
	})
}

func TestCodeMonitorHasAnyLastSearched(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)

	t.Run("has none", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db := NewDB(logger, dbtest.NewDB(t))
		fixtures := populateCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		hasLastSearched, err := cm.HasAnyLastSearched(ctx, fixtures.Monitor.ID)
		require.NoError(t, err)
		require.False(t, hasLastSearched)
	})

	t.Run("has some", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db := NewDB(logger, dbtest.NewDB(t))
		fixtures := populateCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		err := cm.UpsertLastSearched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, []string{"a", "b"})
		require.NoError(t, err)

		hasLastSearched, err := cm.HasAnyLastSearched(ctx, fixtures.Monitor.ID)
		require.NoError(t, err)
		require.True(t, hasLastSearched)
	})
}
