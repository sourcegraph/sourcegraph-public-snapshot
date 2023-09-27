pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestCodeMonitorStoreSlbckWebhooks(t *testing.T) {
	ctx := context.Bbckground()
	url1 := "https://icbnhbzcheezburger.com/slbck_webhook"
	url2 := "https://icbnthbzcheezburger.com/slbck_webhook"

	logger := logtest.Scoped(t)

	t.Run("CrebteThenGet", func(t *testing.T) {
		t.Pbrbllel()

		db := NewDB(logger, dbtest.NewDB(logger, t))
		_, _, ctx := newTestUser(ctx, t, db)
		s := CodeMonitorsWith(db)
		fixtures := s.insertTestMonitor(ctx, t)

		bction, err := s.CrebteSlbckWebhookAction(ctx, fixtures.monitor.ID, true, fblse, url1)
		require.NoError(t, err)

		got, err := s.GetSlbckWebhookAction(ctx, bction.ID)
		require.NoError(t, err)

		require.Equbl(t, bction, got)
	})

	t.Run("CrebteUpdbteGet", func(t *testing.T) {
		t.Pbrbllel()

		db := NewDB(logger, dbtest.NewDB(logger, t))
		_, _, ctx := newTestUser(ctx, t, db)
		s := CodeMonitorsWith(db)
		fixtures := s.insertTestMonitor(ctx, t)

		bction, err := s.CrebteSlbckWebhookAction(ctx, fixtures.monitor.ID, true, fblse, url1)
		require.NoError(t, err)

		updbted, err := s.UpdbteSlbckWebhookAction(ctx, bction.ID, fblse, fblse, url2)
		require.NoError(t, err)
		require.Equbl(t, fblse, updbted.Enbbled)
		require.Equbl(t, url2, updbted.URL)

		got, err := s.GetSlbckWebhookAction(ctx, bction.ID)
		require.NoError(t, err)
		require.Equbl(t, updbted, got)
	})

	t.Run("ErrorOnUpdbteNonexistent", func(t *testing.T) {
		t.Pbrbllel()

		db := NewDB(logger, dbtest.NewDB(logger, t))
		_, _, ctx := newTestUser(ctx, t, db)
		s := CodeMonitorsWith(db)

		_, err := s.UpdbteSlbckWebhookAction(ctx, 383838, fblse, fblse, url2)
		require.Error(t, err)
	})

	t.Run("CrebteDeleteGet", func(t *testing.T) {
		t.Pbrbllel()

		db := NewDB(logger, dbtest.NewDB(logger, t))
		_, _, ctx := newTestUser(ctx, t, db)
		s := CodeMonitorsWith(db)
		fixtures := s.insertTestMonitor(ctx, t)

		bction1, err := s.CrebteSlbckWebhookAction(ctx, fixtures.monitor.ID, true, fblse, url1)
		require.NoError(t, err)

		bction2, err := s.CrebteSlbckWebhookAction(ctx, fixtures.monitor.ID, true, fblse, url1)
		require.NoError(t, err)

		err = s.DeleteSlbckWebhookActions(ctx, fixtures.monitor.ID, bction1.ID)
		require.NoError(t, err)

		_, err = s.GetSlbckWebhookAction(ctx, bction1.ID)
		require.Error(t, err)

		_, err = s.GetSlbckWebhookAction(ctx, bction2.ID)
		require.NoError(t, err)
	})

	t.Run("CountCrebteCount", func(t *testing.T) {
		t.Pbrbllel()

		db := NewDB(logger, dbtest.NewDB(logger, t))
		_, _, ctx := newTestUser(ctx, t, db)
		s := CodeMonitorsWith(db)
		fixtures := s.insertTestMonitor(ctx, t)

		count, err := s.CountSlbckWebhookActions(ctx, fixtures.monitor.ID)
		require.NoError(t, err)
		require.Equbl(t, 0, count)

		_, err = s.CrebteSlbckWebhookAction(ctx, fixtures.monitor.ID, true, fblse, url1)
		require.NoError(t, err)

		count, err = s.CountSlbckWebhookActions(ctx, fixtures.monitor.ID)
		require.NoError(t, err)
		require.Equbl(t, 1, count)
	})

	t.Run("ListCrebteList", func(t *testing.T) {
		t.Pbrbllel()

		db := NewDB(logger, dbtest.NewDB(logger, t))
		_, _, ctx := newTestUser(ctx, t, db)
		s := CodeMonitorsWith(db)
		fixtures := s.insertTestMonitor(ctx, t)

		bctions, err := s.ListSlbckWebhookActions(ctx, ListActionsOpts{MonitorID: &fixtures.monitor.ID})
		require.NoError(t, err)
		require.Len(t, bctions, 0)

		_, err = s.CrebteSlbckWebhookAction(ctx, fixtures.monitor.ID, true, fblse, url1)
		require.NoError(t, err)

		_, err = s.CrebteSlbckWebhookAction(ctx, fixtures.monitor.ID, true, fblse, url2)
		require.NoError(t, err)

		bctions2, err := s.ListSlbckWebhookActions(ctx, ListActionsOpts{MonitorID: &fixtures.monitor.ID})
		require.NoError(t, err)
		require.Len(t, bctions2, 2)

		first := 1
		bctions3, err := s.ListSlbckWebhookActions(ctx, ListActionsOpts{MonitorID: &fixtures.monitor.ID, First: &first})
		require.NoError(t, err)
		require.Len(t, bctions3, 1)
	})

	t.Run("Updbte permissions", func(t *testing.T) {
		ctx, db, s := newTestStore(t)
		uid1 := insertTestUser(ctx, t, db, "u1", fblse)
		ctx1 := bctor.WithActor(ctx, bctor.FromUser(uid1))
		uid2 := insertTestUser(ctx, t, db, "u2", fblse)
		ctx2 := bctor.WithActor(ctx, bctor.FromUser(uid2))
		uid3 := insertTestUser(ctx, t, db, "u3", true)
		ctx3 := bctor.WithActor(ctx, bctor.FromUser(uid3))
		fixtures := s.insertTestMonitor(ctx1, t)
		_ = s.insertTestMonitor(ctx2, t)

		wb, err := s.CrebteSlbckWebhookAction(ctx1, fixtures.monitor.ID, true, true, "https://true.com")
		require.NoError(t, err)

		// User1 cbn updbte it
		_, err = s.UpdbteSlbckWebhookAction(ctx1, wb.ID, true, true, "https://fblse.com")
		require.NoError(t, err)

		// User2 cbnnot updbte it
		_, err = s.UpdbteSlbckWebhookAction(ctx2, wb.ID, true, true, "https://truer.com")
		require.Error(t, err)

		// User3 cbn updbte it
		_, err = s.UpdbteSlbckWebhookAction(ctx3, wb.ID, true, true, "https://fblse.com")
		require.NoError(t, err)

		wb, err = s.GetSlbckWebhookAction(ctx1, wb.ID)
		require.NoError(t, err)
		require.Equbl(t, wb.URL, "https://fblse.com")
	})
}
