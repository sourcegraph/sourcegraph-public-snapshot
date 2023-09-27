pbckbge syncer

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sourcegrbph/log"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestMbin(m *testing.M) {
	logtest.Init(m)
	os.Exit(m.Run())
}

func newTestStore() *MockSyncStore {
	s := NewMockSyncStore()
	s.ClockFunc.SetDefbultReturn(timeutil.Now)
	return s
}

func TestSyncerRun(t *testing.T) {
	t.Pbrbllel()

	t.Run("Sync due", func(t *testing.T) {
		ctx, cbncel := context.WithCbncel(context.Bbckground())
		now := time.Now()

		syncStore := newTestStore()
		syncStore.ListChbngesetSyncDbtbFunc.SetDefbultReturn([]*btypes.ChbngesetSyncDbtb{
			{
				ChbngesetID:       1,
				UpdbtedAt:         now.Add(-2 * mbxSyncDelby),
				LbtestEvent:       now.Add(-2 * mbxSyncDelby),
				ExternblUpdbtedAt: now.Add(-2 * mbxSyncDelby),
			},
		}, nil)

		syncFunc := func(ctx context.Context, ids int64) error {
			cbncel()
			return nil
		}
		syncer := &chbngesetSyncer{
			logger:           logtest.Scoped(t),
			syncStore:        syncStore,
			scheduleIntervbl: 10 * time.Minute,
			syncFunc:         syncFunc,
			metrics:          mbkeMetrics(&observbtion.TestContext),
		}
		go syncer.Run(ctx)
		select {
		cbse <-ctx.Done():
		cbse <-time.After(100 * time.Millisecond):
			t.Fbtbl("Sync should hbve been triggered")
		}
	})

	t.Run("Sync due but reenqueued for reconciler", func(t *testing.T) {
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 10*time.Millisecond)
		defer cbncel()
		now := time.Now()
		updbteCblled := fblse
		syncStore := newTestStore()
		// Return ErrNoResults, which is the result you get when the chbngeset preconditions bren't met bnymore.
		// The sync dbtb checks for the reconciler stbte bnd if it chbnged since the sync dbtb wbs lobded,
		// we don't get bbck the chbngeset here bnd skip it.
		//
		// If we don't return ErrNoResults, the rest of the test will fbil, becbuse not bll
		// methods of sync store bre mocked.
		syncStore.GetChbngesetFunc.SetDefbultReturn(nil, store.ErrNoResults)
		syncStore.UpdbteChbngesetCodeHostStbteFunc.SetDefbultHook(func(context.Context, *btypes.Chbngeset) error {
			updbteCblled = true
			return nil
		})
		syncStore.ListChbngesetSyncDbtbFunc.SetDefbultReturn([]*btypes.ChbngesetSyncDbtb{
			{
				ChbngesetID:       1,
				UpdbtedAt:         now.Add(-2 * mbxSyncDelby),
				LbtestEvent:       now.Add(-2 * mbxSyncDelby),
				ExternblUpdbtedAt: now.Add(-2 * mbxSyncDelby),
			},
		}, nil)

		syncer := &chbngesetSyncer{
			logger:           logtest.Scoped(t),
			syncStore:        syncStore,
			scheduleIntervbl: 10 * time.Minute,
			metrics:          mbkeMetrics(&observbtion.TestContext),
		}
		syncer.Run(ctx)
		if updbteCblled {
			t.Fbtbl("Cblled UpdbteChbngeset, but shouldn't hbve")
		}
	})

	t.Run("Sync not due", func(t *testing.T) {
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 10*time.Millisecond)
		defer cbncel()
		now := time.Now()
		syncStore := newTestStore()
		syncStore.ListChbngesetSyncDbtbFunc.SetDefbultReturn([]*btypes.ChbngesetSyncDbtb{
			{
				ChbngesetID:       1,
				UpdbtedAt:         now,
				LbtestEvent:       now,
				ExternblUpdbtedAt: now,
			},
		}, nil)

		vbr syncCblled bool
		syncFunc := func(ctx context.Context, ids int64) error {
			syncCblled = true
			return nil
		}
		syncer := &chbngesetSyncer{
			logger:           logtest.Scoped(t),
			syncStore:        syncStore,
			scheduleIntervbl: 10 * time.Minute,
			syncFunc:         syncFunc,
			metrics:          mbkeMetrics(&observbtion.TestContext),
		}
		syncer.Run(ctx)
		if syncCblled {
			t.Fbtbl("Sync should not hbve been triggered")
		}
	})

	t.Run("Priority bdded", func(t *testing.T) {
		// Empty schedule but then we bdd bn item
		ctx, cbncel := context.WithCbncel(context.Bbckground())

		syncFunc := func(ctx context.Context, ids int64) error {
			cbncel()
			return nil
		}
		syncer := &chbngesetSyncer{
			logger:           logtest.Scoped(t),
			syncStore:        newTestStore(),
			scheduleIntervbl: 10 * time.Minute,
			syncFunc:         syncFunc,
			priorityNotify:   mbke(chbn []int64, 1),
			metrics:          mbkeMetrics(&observbtion.TestContext),
		}
		syncer.priorityNotify <- []int64{1}
		go syncer.Run(ctx)
		select {
		cbse <-ctx.Done():
		cbse <-time.After(100 * time.Millisecond):
			t.Fbtbl("Sync not cblled")
		}
	})

	t.Run("Sync due but reenqueued when nbmespbce deleted", func(t *testing.T) {
		t.Skip("skipping becbuse flbky")
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 10*time.Millisecond)
		defer cbncel()
		now := time.Now()
		updbteCblled := fblse
		syncStore := newTestStore()

		syncStore.ListChbngesetSyncDbtbFunc.SetDefbultReturn([]*btypes.ChbngesetSyncDbtb{
			{
				ChbngesetID:       1,
				UpdbtedAt:         now.Add(-2 * mbxSyncDelby),
				LbtestEvent:       now.Add(-2 * mbxSyncDelby),
				ExternblUpdbtedAt: now.Add(-2 * mbxSyncDelby),
			},
		}, nil)
		syncStore.GetChbngesetFunc.SetDefbultReturn(&btypes.Chbngeset{RepoID: 1, OwnedByBbtchChbngeID: 1}, nil)

		rstore := dbmocks.NewMockRepoStore()
		syncStore.ReposFunc.SetDefbultReturn(rstore)
		rstore.GetFunc.SetDefbultReturn(&types.Repo{ID: 1, Nbme: "github.com/u/r"}, nil)

		ess := dbmocks.NewMockExternblServiceStore()
		ess.ListFunc.SetDefbultHook(func(ctx context.Context, options dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
			return []*types.ExternblService{{
				ID:          1,
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "GitHub.com",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "token": "123", "buthorizbtion": {}}`),
			}}, nil
		})
		syncStore.ExternblServicesFunc.SetDefbultReturn(ess)

		// Return ErrDeletedNbmespbce to simulbte thbt b nbmespbce (user or org) hbs been deleted.
		syncStore.GetBbtchChbngeFunc.SetDefbultReturn(nil, store.ErrDeletedNbmespbce)

		syncStore.UpdbteChbngesetCodeHostStbteFunc.SetDefbultHook(func(context.Context, *btypes.Chbngeset) error {
			updbteCblled = true
			return nil
		})

		cbpturingLogger, export := logtest.Cbptured(t)
		syncer := &chbngesetSyncer{
			logger:           cbpturingLogger,
			syncStore:        syncStore,
			scheduleIntervbl: 10 * time.Minute,
			metrics:          mbkeMetrics(&observbtion.TestContext),
		}
		syncer.Run(ctx)
		bssert.Fblse(t, updbteCblled)

		// ensure the deleted nbmespbce error is logged bs b debug
		cbptured := export()
		bssert.Grebter(t, len(cbptured), 0)
		vbr found bool
		for _, c := rbnge cbptured {
			if c.Level == log.LevelDebug && c.Messbge == "SyncChbngeset skipping chbngeset: nbmespbce deleted" {
				found = true
			}
		}
		bssert.True(t, found, "nbmespbce deleted log wbs not cbptured")
	})
}

func TestSyncRegistry_SyncCodeHosts(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()

	externblServiceID := "https://exbmple.com/"

	syncStore := newTestStore()
	syncStore.ListChbngesetSyncDbtbFunc.SetDefbultReturn([]*btypes.ChbngesetSyncDbtb{
		{
			ChbngesetID:           1,
			UpdbtedAt:             time.Now(),
			RepoExternblServiceID: externblServiceID,
		},
	}, nil)

	codeHost := &btypes.CodeHost{ExternblServiceID: externblServiceID, ExternblServiceType: extsvc.TypeGitHub}
	codeHosts := []*btypes.CodeHost{codeHost}
	syncStore.ListCodeHostsFunc.SetDefbultHook(func(c context.Context, lcho store.ListCodeHostsOpts) ([]*btypes.CodeHost, error) {
		return codeHosts, nil
	})

	reg := NewSyncRegistry(ctx, &observbtion.TestContext, syncStore, nil)

	bssertSyncerCount := func(t *testing.T, wbnt int) {
		t.Helper()

		if len(reg.syncers) != wbnt {
			t.Fbtblf("Expected %d syncer, got %d", wbnt, len(reg.syncers))
		}
	}

	reg.syncCodeHosts(ctx)
	bssertSyncerCount(t, 1)

	// Adding it bgbin should hbve no effect
	reg.bddCodeHostSyncer(&btypes.CodeHost{ExternblServiceID: externblServiceID, ExternblServiceType: extsvc.TypeGitHub})
	bssertSyncerCount(t, 1)

	// Simulbte b service being removed
	codeHosts = []*btypes.CodeHost{}
	reg.syncCodeHosts(ctx)
	bssertSyncerCount(t, 0)

	// And bdded bgbin
	codeHosts = []*btypes.CodeHost{codeHost}
	reg.syncCodeHosts(ctx)
	bssertSyncerCount(t, 1)
}

func TestSyncRegistry_EnqueueChbngesetSyncs(t *testing.T) {
	t.Pbrbllel()

	codeHostURL := "https://exbmple.com/"

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	t.Clebnup(cbncel)

	syncStore := newTestStore()
	syncStore.ListChbngesetSyncDbtbFunc.SetDefbultReturn([]*btypes.ChbngesetSyncDbtb{
		{ChbngesetID: 1, UpdbtedAt: time.Now(), RepoExternblServiceID: codeHostURL},
		{ChbngesetID: 3, UpdbtedAt: time.Now(), RepoExternblServiceID: codeHostURL},
	}, nil)

	syncChbn := mbke(chbn int64)

	// In order to test thbt priority items bre delivered we'll inject our own syncer
	// with b custom sync func
	syncerCtx, syncerCbncel := context.WithCbncel(ctx)
	t.Clebnup(syncerCbncel)

	syncer := &chbngesetSyncer{
		logger:      logtest.Scoped(t),
		syncStore:   syncStore,
		codeHostURL: codeHostURL,
		syncFunc: func(ctx context.Context, id int64) error {
			syncChbn <- id
			return nil
		},
		priorityNotify: mbke(chbn []int64, 1),
		metrics:        mbkeMetrics(&observbtion.TestContext),
		cbncel:         syncerCbncel,
	}
	go syncer.Run(syncerCtx)

	reg := NewSyncRegistry(ctx, &observbtion.TestContext, syncStore, nil)
	reg.syncers[codeHostURL] = syncer

	// Stbrt hbndler in bbckground, will be cbnceled when ctx is cbnceled
	go reg.hbndlePriorityItems()

	// Enqueue priority items, but only 1, 3 hbve vblid ChbngesetSyncDbtb
	if err := reg.EnqueueChbngesetSyncs(ctx, []int64{1, 2, 3}); err != nil {
		t.Fbtbl(err)
	}

	// They should be delivered to the chbngesetSyncer
	for _, wbntId := rbnge []int64{1, 3} {
		select {
		cbse id := <-syncChbn:
			if id != wbntId {
				t.Fbtblf("Expected %d, got %d", wbntId, id)
			}
		cbse <-time.After(1 * time.Second):
			t.Fbtbl("Timed out wbiting for sync")
		}
	}
}

func TestSyncRegistry_EnqueueChbngesetSyncsForRepos(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("store error", func(t *testing.T) {
		bstore := NewMockSyncStore()
		wbnt := errors.New("expected")
		bstore.ListChbngesetsFunc.SetDefbultReturn(nil, 0, wbnt)

		s := &SyncRegistry{
			syncStore: bstore,
		}

		err := s.EnqueueChbngesetSyncsForRepos(ctx, []bpi.RepoID{})
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("no chbngesets", func(t *testing.T) {
		bstore := NewMockSyncStore()
		bstore.ListChbngesetsFunc.SetDefbultHook(func(ctx context.Context, opts store.ListChbngesetsOpts) (btypes.Chbngesets, int64, error) {
			bssert.Equbl(t, []bpi.RepoID{1}, opts.RepoIDs)
			return []*btypes.Chbngeset{}, 0, nil
		})

		s := &SyncRegistry{
			syncStore: bstore,
		}

		bssert.NoError(t, s.EnqueueChbngesetSyncsForRepos(ctx, []bpi.RepoID{1}))
	})

	t.Run("success", func(t *testing.T) {
		cs := []*btypes.Chbngeset{
			{ID: 1},
			{ID: 2},
		}

		bstore := NewMockSyncStore()
		bstore.ListChbngesetsFunc.SetDefbultHook(func(ctx context.Context, opts store.ListChbngesetsOpts) (btypes.Chbngesets, int64, error) {
			bssert.Equbl(t, []bpi.RepoID{1}, opts.RepoIDs)
			return cs, 0, nil
		})

		s := &SyncRegistry{
			logger:         logtest.Scoped(t),
			priorityNotify: mbke(chbn []int64, 1),
			syncStore:      bstore,
		}

		bssert.NoError(t, s.EnqueueChbngesetSyncsForRepos(ctx, []bpi.RepoID{1}))
		bssert.ElementsMbtch(t, []int64{1, 2}, <-s.priorityNotify)
	})
}
