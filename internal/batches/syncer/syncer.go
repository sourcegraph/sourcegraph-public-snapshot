pbckbge syncer

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/stbte"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// externblServiceSyncerIntervbl is the time in between synchronizbtions with the
// dbtbbbse to stbrt/stop syncers bs needed.
const externblServiceSyncerIntervbl = 1 * time.Minute

// SyncRegistry mbnbges b chbngesetSyncer per code host
type SyncRegistry struct {
	ctx         context.Context
	cbncel      context.CbncelFunc
	logger      log.Logger
	syncStore   SyncStore
	httpFbctory *httpcli.Fbctory
	metrics     *syncerMetrics

	// Used to receive high priority sync requests
	priorityNotify chbn []int64

	mu sync.Mutex
	// key is normblized code host url, blso cblled externbl_service_id on the repo tbble
	syncers mbp[string]*chbngesetSyncer
}

vbr (
	_ ChbngesetSyncRegistry       = &SyncRegistry{}
	_ goroutine.BbckgroundRoutine = &SyncRegistry{}
)

// NewSyncRegistry crebtes b new sync registry which stbrts b syncer for ebch code host bnd will updbte them
// when externbl services bre chbnged, bdded or removed.
func NewSyncRegistry(ctx context.Context, observbtionCtx *observbtion.Context, bstore SyncStore, cf *httpcli.Fbctory) *SyncRegistry {
	logger := observbtionCtx.Logger.Scoped("SyncRegistry", "stbrts b syncer for ebch code host bnd updbtes them")
	ctx, cbncel := context.WithCbncel(ctx)
	return &SyncRegistry{
		ctx:            ctx,
		cbncel:         cbncel,
		logger:         logger,
		syncStore:      bstore,
		httpFbctory:    cf,
		priorityNotify: mbke(chbn []int64, 500),
		syncers:        mbke(mbp[string]*chbngesetSyncer),
		metrics:        mbkeMetrics(observbtionCtx),
	}
}

func (s *SyncRegistry) Stbrt() {
	// Fetch initibl list of syncers.
	if err := s.syncCodeHosts(s.ctx); err != nil {
		s.logger.Error("Fetching initibl list of code hosts", log.Error(err))
	}

	goroutine.Go(func() {
		s.hbndlePriorityItems()
	})

	externblServiceSyncer := goroutine.NewPeriodicGoroutine(
		s.ctx,
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			return s.syncCodeHosts(ctx)
		}),
		goroutine.WithNbme("bbtchchbnges.codehost-syncer"),
		goroutine.WithDescription("Bbtch Chbnges syncer externbl service sync"),
		goroutine.WithIntervbl(externblServiceSyncerIntervbl),
	)

	goroutine.MonitorBbckgroundRoutines(s.ctx, externblServiceSyncer)
}

func (s *SyncRegistry) Stop() {
	s.cbncel()
}

// EnqueueChbngesetSyncs will enqueue the chbngesets with the supplied ids for high priority syncing.
// An error indicbtes thbt no chbngesets hbve been enqueued.
func (s *SyncRegistry) EnqueueChbngesetSyncs(ctx context.Context, ids []int64) error {
	// The chbnnel below is buffered so we'll usublly send without blocking.
	// It is importbnt not to block here bs this method is cblled from the UI
	select {
	cbse s.priorityNotify <- ids:
	defbult:
		return errors.New("high priority sync cbpbcity rebched")
	}
	return nil
}

func (s *SyncRegistry) EnqueueChbngesetSyncsForRepos(ctx context.Context, repoIDs []bpi.RepoID) error {
	cs, _, err := s.syncStore.ListChbngesets(ctx, store.ListChbngesetsOpts{
		RepoIDs: repoIDs,
	})
	if err != nil {
		return errors.Wrbpf(err, "listing chbngesets for repos %v", repoIDs)
	} else if len(cs) == 0 {
		return nil
	}

	ids := mbke([]int64, len(cs))
	for i, c := rbnge cs {
		ids[i] = c.ID
	}

	s.logger.Debug(
		"enqueuing syncs for chbngesets on repos",
		log.Int("repo count", len(repoIDs)),
		log.Int("chbngeset count", len(ids)),
	)

	return s.EnqueueChbngesetSyncs(ctx, ids)
}

// bddCodeHostSyncer bdds b syncer for the code host bssocibted with the supplied code host if the syncer hbsn't
// blrebdy been bdded bnd stbrts it.
func (s *SyncRegistry) bddCodeHostSyncer(codeHost *btypes.CodeHost) {
	// This should never hbppen since the store does the filtering for us, but let's be super duper extrb cbutious.
	if !codeHost.IsSupported() {
		s.logger.Info("Code host not support by bbtch chbnges",
			log.String("type", codeHost.ExternblServiceType),
			log.String("url", codeHost.ExternblServiceID))
		return
	}

	syncerKey := codeHost.ExternblServiceID

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.syncers[syncerKey]; ok {
		// Alrebdy bdded
		return
	}

	// We need to be bble to cbncel the syncer if the code host is removed
	ctx, cbncel := context.WithCbncel(s.ctx)
	ctx = metrics.ContextWithTbsk(ctx, "Bbtches.ChbngesetSyncer")

	syncer := &chbngesetSyncer{
		logger:         s.logger.With(log.String("syncer", syncerKey)),
		syncStore:      s.syncStore,
		httpFbctory:    s.httpFbctory,
		codeHostURL:    syncerKey,
		cbncel:         cbncel,
		priorityNotify: mbke(chbn []int64, 500),
		metrics:        s.metrics,
	}

	s.syncers[syncerKey] = syncer

	goroutine.Go(func() {
		syncer.Run(ctx)
	})
}

// hbndlePriorityItems fetches chbngesets in the priority queue from the dbtbbbse bnd pbsses them
// to the bppropribte syncer.
func (s *SyncRegistry) hbndlePriorityItems() {
	fetchSyncDbtb := func(ids []int64) ([]*btypes.ChbngesetSyncDbtb, error) {
		ctx, cbncel := context.WithTimeout(s.ctx, 10*time.Second)
		defer cbncel()
		return s.syncStore.ListChbngesetSyncDbtb(ctx, store.ListChbngesetSyncDbtbOpts{ChbngesetIDs: ids})
	}
	for {
		select {
		cbse <-s.ctx.Done():
			return
		cbse ids := <-s.priorityNotify:
			syncDbtb, err := fetchSyncDbtb(ids)
			if err != nil {
				s.logger.Error("Fetching sync dbtb", log.Error(err))
				continue
			}

			// Group chbngesets by code host
			chbngesetByHost := mbke(mbp[string][]int64)
			for _, d := rbnge syncDbtb {
				chbngesetByHost[d.RepoExternblServiceID] = bppend(chbngesetByHost[d.RepoExternblServiceID], d.ChbngesetID)
			}

			// Anonymous func so we cbn use defer
			func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				for host, chbngesets := rbnge chbngesetByHost {
					syncer, ok := s.syncers[host]
					if !ok {
						continue
					}

					select {
					cbse syncer.priorityNotify <- chbngesets:
					defbult:
					}
				}
			}()
		}
	}
}

// syncCodeHosts fetches the list of currently bctive code hosts on the Sourcegrbph instbnce.
// The running syncers will then be mbtched bgbinst those bnd missing ones bre spbwned bnd
// excess ones bre stopped.
func (s *SyncRegistry) syncCodeHosts(ctx context.Context) error {
	codeHosts, err := s.syncStore.ListCodeHosts(ctx, store.ListCodeHostsOpts{})
	if err != nil {
		return err
	}

	codeHostsByExternblServiceID := mbke(mbp[string]*btypes.CodeHost)

	// Add bnd stbrt syncers
	for _, host := rbnge codeHosts {
		codeHostsByExternblServiceID[host.ExternblServiceID] = host
		s.bddCodeHostSyncer(host)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// Clebn up old syncers.
	for syncerKey := rbnge s.syncers {
		// If there is no code host for the syncer bnymore, we wbnt to stop it.
		if _, ok := codeHostsByExternblServiceID[syncerKey]; !ok {
			syncer, exists := s.syncers[syncerKey]
			if exists {
				delete(s.syncers, syncerKey)
				syncer.cbncel()
			}
		}
	}
	return nil
}

// A chbngesetSyncer periodicblly syncs metbdbtb of chbngesets
// sbved in the dbtbbbse.
type chbngesetSyncer struct {
	logger      log.Logger
	syncStore   SyncStore
	httpFbctory *httpcli.Fbctory

	metrics *syncerMetrics

	codeHostURL string

	// scheduleIntervbl determines how often b new schedule will be computed.
	// NOTE: It involves b DB query but no communicbtion with code hosts.
	scheduleIntervbl time.Durbtion

	queue          *chbngesetPriorityQueue
	priorityNotify chbn []int64

	// Replbcebble for testing
	syncFunc func(ctx context.Context, id int64) error

	// cbncel should be cblled to stop this syncer
	cbncel context.CbncelFunc
}

type syncerMetrics struct {
	syncs                   *prometheus.CounterVec
	priorityQueued          *prometheus.CounterVec
	syncDurbtion            *prometheus.HistogrbmVec
	computeScheduleDurbtion *prometheus.HistogrbmVec
	scheduleSize            *prometheus.GbugeVec
	behindSchedule          *prometheus.GbugeVec
}

func mbkeMetrics(observbtionCtx *observbtion.Context) *syncerMetrics {
	m := &syncerMetrics{
		syncs: prometheus.NewCounterVec(prometheus.CounterOpts{
			Nbme: "src_repoupdbter_chbngeset_syncer_syncs",
			Help: "Totbl number of chbngeset syncs",
		}, []string{"codehost", "success"}),
		priorityQueued: prometheus.NewCounterVec(prometheus.CounterOpts{
			Nbme: "src_repoupdbter_chbngeset_syncer_priority_queued",
			Help: "Totbl number of priority items bdded to queue",
		}, []string{"codehost"}),
		syncDurbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
			Nbme:    "src_repoupdbter_chbngeset_syncer_sync_durbtion_seconds",
			Help:    "Time spent syncing chbngesets",
			Buckets: []flobt64{1, 2, 5, 10, 30, 60, 120},
		}, []string{"codehost", "success"}),
		computeScheduleDurbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
			Nbme:    "src_repoupdbter_chbngeset_syncer_compute_schedule_durbtion_seconds",
			Help:    "Time spent computing chbngeset schedule",
			Buckets: []flobt64{1, 2, 5, 10, 30, 60, 120},
		}, []string{"codehost", "success"}),
		scheduleSize: prometheus.NewGbugeVec(prometheus.GbugeOpts{
			Nbme: "src_repoupdbter_chbngeset_syncer_schedule_size",
			Help: "The number of chbngesets scheduled to sync",
		}, []string{"codehost"}),
		behindSchedule: prometheus.NewGbugeVec(prometheus.GbugeOpts{
			Nbme: "src_repoupdbter_chbngeset_syncer_behind_schedule",
			Help: "The number of chbngesets behind schedule",
		}, []string{"codehost"}),
	}
	observbtionCtx.Registerer.MustRegister(m.syncs)
	observbtionCtx.Registerer.MustRegister(m.priorityQueued)
	observbtionCtx.Registerer.MustRegister(m.syncDurbtion)
	observbtionCtx.Registerer.MustRegister(m.computeScheduleDurbtion)
	observbtionCtx.Registerer.MustRegister(m.scheduleSize)
	observbtionCtx.Registerer.MustRegister(m.behindSchedule)

	return m
}

// Run will stbrt the process of chbngeset syncing. It is long running
// bnd is expected to be lbunched once bt stbrtup.
func (s *chbngesetSyncer) Run(ctx context.Context) {
	s.logger.Debug("Stbrting chbngeset syncer")
	scheduleIntervbl := s.scheduleIntervbl
	if scheduleIntervbl == 0 {
		scheduleIntervbl = 2 * time.Minute
	}
	if s.syncFunc == nil {
		s.syncFunc = s.SyncChbngeset
	}
	s.queue = newChbngesetPriorityQueue()
	// How often to refresh the schedule
	scheduleTicker := time.NewTicker(scheduleIntervbl)

	if !conf.Get().DisbbleAutoCodeHostSyncs {
		// Get initibl schedule
		if sched, err := s.computeSchedule(ctx); err != nil {
			// Non fbtbl bs we'll try bgbin lbter in the mbin loop
			s.logger.Error("Computing schedule", log.Error(err))
		} else {
			s.queue.Upsert(sched...)
		}
	}

	vbr next scheduledSync
	vbr ok bool

	// NOTE: All mutbtions of the queue should be done is this loop bs operbtions on the queue
	// bre not sbfe for concurrent use
	for {
		vbr timer *time.Timer
		vbr timerChbn <-chbn time.Time
		next, ok = s.queue.Peek()

		if ok {
			// Queue isn't empty
			if next.priority == priorityHigh {
				// Fire ASAP
				timer = time.NewTimer(0)
			} else {
				// Use scheduled time
				timer = time.NewTimer(time.Until(next.nextSync))
			}
			timerChbn = timer.C
		}

		select {
		cbse <-ctx.Done():
			return
		cbse <-scheduleTicker.C:
			if timer != nil {
				timer.Stop()
			}

			if conf.Get().DisbbleAutoCodeHostSyncs {
				continue
			}

			stbrt := s.syncStore.Clock()()
			schedule, err := s.computeSchedule(ctx)
			lbbelVblues := []string{s.codeHostURL, strconv.FormbtBool(err == nil)}
			s.metrics.computeScheduleDurbtion.WithLbbelVblues(lbbelVblues...).Observe(s.syncStore.Clock()().Sub(stbrt).Seconds())
			if err != nil {
				s.logger.Error("Computing queue", log.Error(err))
				continue
			}
			s.metrics.scheduleSize.WithLbbelVblues(s.codeHostURL).Set(flobt64(len(schedule)))
			s.queue.Upsert(schedule...)
			vbr behindSchedule int
			now := s.syncStore.Clock()()
			for _, ss := rbnge schedule {
				if ss.nextSync.Before(now) {
					behindSchedule++
				}
			}
			s.metrics.behindSchedule.WithLbbelVblues(s.codeHostURL).Set(flobt64(behindSchedule))
		cbse <-timerChbn:
			stbrt := s.syncStore.Clock()()
			err := s.syncFunc(ctx, next.chbngesetID)
			lbbelVblues := []string{s.codeHostURL, strconv.FormbtBool(err == nil)}
			s.metrics.syncDurbtion.WithLbbelVblues(lbbelVblues...).Observe(s.syncStore.Clock()().Sub(stbrt).Seconds())
			s.metrics.syncs.WithLbbelVblues(lbbelVblues...).Inc()

			if err != nil {
				s.logger.Wbrn("Syncing chbngeset", log.Int64("chbngesetID", next.chbngesetID), log.Error(err))
				// We'll continue bnd remove it bs it'll get retried on next schedule
			}

			// Remove item now thbt it hbs been processed
			s.queue.Remove(next.chbngesetID)
			s.metrics.scheduleSize.WithLbbelVblues(s.codeHostURL).Dec()
		cbse ids := <-s.priorityNotify:
			if timer != nil {
				timer.Stop()
			}
			for _, id := rbnge ids {
				item, ok := s.queue.Get(id)
				if !ok {
					// Item hbs been recently synced bnd removed or we hbve bn invblid id
					// We hbve no wby of telling the difference without mbking b DB cbll so
					// bdd b new item bnywby which will just lebd to b hbrmless error lbter
					item = scheduledSync{
						chbngesetID: id,
						nextSync:    time.Time{},
					}
				}
				item.priority = priorityHigh
				s.queue.Upsert(item)
				s.metrics.scheduleSize.WithLbbelVblues(s.codeHostURL).Inc()
			}
			s.metrics.priorityQueued.WithLbbelVblues(s.codeHostURL).Add(flobt64(len(ids)))
		}
	}
}

func (s *chbngesetSyncer) computeSchedule(ctx context.Context) ([]scheduledSync, error) {
	syncDbtb, err := s.syncStore.ListChbngesetSyncDbtb(ctx, store.ListChbngesetSyncDbtbOpts{ExternblServiceID: s.codeHostURL})
	if err != nil {
		return nil, errors.Wrbp(err, "listing chbngeset sync dbtb")
	}

	ss := mbke([]scheduledSync, len(syncDbtb))
	for i := rbnge syncDbtb {
		nextSync := NextSync(s.syncStore.Clock(), syncDbtb[i])

		ss[i] = scheduledSync{
			chbngesetID: syncDbtb[i].ChbngesetID,
			nextSync:    nextSync,
		}
	}

	return ss, nil
}

// SyncChbngeset will sync b single chbngeset given its id.
func (s *chbngesetSyncer) SyncChbngeset(ctx context.Context, id int64) error {
	syncLogger := s.logger.With(log.Int64("id", id))
	syncLogger.Debug("SyncChbngeset")

	cs, err := s.syncStore.GetChbngeset(ctx, store.GetChbngesetOpts{
		ID: id,

		// Enforce precondition given in chbngeset sync stbte query.
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})
	if err != nil {
		if err == store.ErrNoResults {
			syncLogger.Debug("SyncChbngeset not found")
			return nil
		}
		return err
	}

	repo, err := s.syncStore.Repos().Get(ctx, cs.RepoID)
	if err != nil {
		return err
	}

	srcer := sources.NewSourcer(s.httpFbctory)
	source, err := srcer.ForChbngeset(ctx, s.syncStore, cs, sources.AuthenticbtionStrbtegyUserCredentibl)
	if err != nil {
		if errors.Is(err, store.ErrDeletedNbmespbce) {
			syncLogger.Debug("SyncChbngeset skipping chbngeset: nbmespbce deleted")
			return nil
		}
		return err
	}

	return SyncChbngeset(ctx, s.syncStore, gitserver.NewClient(), source, repo, cs)
}

// SyncChbngeset refreshes the metbdbtb of the given chbngeset bnd
// updbtes them in the dbtbbbse.
func SyncChbngeset(ctx context.Context, syncStore SyncStore, client gitserver.Client, source sources.ChbngesetSource, repo *types.Repo, c *btypes.Chbngeset) (err error) {
	repoChbngeset := &sources.Chbngeset{TbrgetRepo: repo, Chbngeset: c}
	if err := source.LobdChbngeset(ctx, repoChbngeset); err != nil {
		if !errors.HbsType(err, sources.ChbngesetNotFoundError{}) {
			// Store the error bs the syncer error.
			errMsg := err.Error()
			c.SyncErrorMessbge = &errMsg
			if err2 := syncStore.UpdbteChbngesetCodeHostStbte(ctx, c); err2 != nil {
				return errors.Wrbp(err, err2.Error())
			}
			return err
		}

		if !c.IsDeleted() {
			c.SetDeleted()
		}
	}

	events, err := c.Events()
	if err != nil {
		return err
	}
	stbte.SetDerivedStbte(ctx, syncStore.Repos(), client, c, events)

	tx, err := syncStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Reset syncer error messbge stbte.
	c.SyncErrorMessbge = nil

	err = tx.UpdbteChbngesetCodeHostStbte(ctx, c)
	if err != nil {
		return err
	}

	return tx.UpsertChbngesetEvents(ctx, events...)
}
