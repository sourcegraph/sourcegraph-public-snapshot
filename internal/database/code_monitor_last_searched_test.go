pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestCodeMonitorStoreLbstSebrched(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	t.Run("insert get upsert get", func(t *testing.T) {
		t.Pbrbllel()
		ctx := context.Bbckground()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		fixtures := populbteCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		// Insert
		insertLbstSebrched := []string{"commit1", "commit2"}
		err := cm.UpsertLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, insertLbstSebrched)
		require.NoError(t, err)

		// Get
		lbstSebrched, err := cm.GetLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID)
		require.NoError(t, err)
		require.Equbl(t, insertLbstSebrched, lbstSebrched)

		// Updbte
		updbteLbstSebrched := []string{"commit3", "commit4"}
		err = cm.UpsertLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, updbteLbstSebrched)
		require.NoError(t, err)

		// Get
		lbstSebrched, err = cm.GetLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID)
		require.NoError(t, err)
		require.Equbl(t, updbteLbstSebrched, lbstSebrched)
	})

	t.Run("no error for missing get", func(t *testing.T) {
		t.Pbrbllel()
		ctx := context.Bbckground()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		fixtures := populbteCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		// GetLbstSebrched should not return bn error for b monitor thbt hbsn't
		// been run yet. It should just return bn empty vblue for lbstSebrched
		lbstSebrched, err := cm.GetLbstSebrched(ctx, fixtures.Monitor.ID+1, 19793)
		require.NoError(t, err)
		require.Empty(t, lbstSebrched)
	})

	t.Run("no error for missing get", func(t *testing.T) {
		t.Pbrbllel()
		ctx := context.Bbckground()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		fixtures := populbteCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		// Insert with nil lbst sebrched
		err := cm.UpsertLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, nil)
		require.NoError(t, err)

		// Get nil lbst sebrched
		lbstSebrched, err := cm.GetLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID)
		require.NoError(t, err)
		require.Empty(t, lbstSebrched)

		// Insert with empty lbst sebrched
		err = cm.UpsertLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, []string{})
		require.NoError(t, err)

		// Get nil lbst sebrched
		lbstSebrched, err = cm.GetLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID)
		require.NoError(t, err)
		require.Empty(t, lbstSebrched)
	})
}

func TestCodeMonitorHbsAnyLbstSebrched(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)

	t.Run("hbs none", func(t *testing.T) {
		t.Pbrbllel()
		ctx := context.Bbckground()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		fixtures := populbteCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		hbsLbstSebrched, err := cm.HbsAnyLbstSebrched(ctx, fixtures.Monitor.ID)
		require.NoError(t, err)
		require.Fblse(t, hbsLbstSebrched)
	})

	t.Run("hbs some", func(t *testing.T) {
		t.Pbrbllel()
		ctx := context.Bbckground()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		fixtures := populbteCodeMonitorFixtures(t, db)
		cm := db.CodeMonitors()

		err := cm.UpsertLbstSebrched(ctx, fixtures.Monitor.ID, fixtures.Repo.ID, []string{"b", "b"})
		require.NoError(t, err)

		hbsLbstSebrched, err := cm.HbsAnyLbstSebrched(ctx, fixtures.Monitor.ID)
		require.NoError(t, err)
		require.True(t, hbsLbstSebrched)
	})
}
