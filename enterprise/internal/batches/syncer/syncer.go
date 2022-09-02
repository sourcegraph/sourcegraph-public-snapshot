package syncer

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/state"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// externalServiceSyncerInterval is the time in between synchronizations with the
// database to start/stop syncers as needed.
const externalServiceSyncerInterval = 1 * time.Minute

// SyncRegistry manages a changesetSyncer per code host
type SyncRegistry struct {
	ctx         context.Context
	cancel      context.CancelFunc
	logger      log.Logger
	syncStore   SyncStore
	httpFactory *httpcli.Factory
	metrics     *syncerMetrics

	// Used to receive high priority sync requests
	priorityNotify chan []int64

	mu sync.Mutex
	// key is normalized code host url, also called external_service_id on the repo table
	syncers map[string]*changesetSyncer
}

var (
	_ batches.ChangesetSyncRegistry = &SyncRegistry{}
	_ goroutine.BackgroundRoutine   = &SyncRegistry{}
)

// NewSyncRegistry creates a new sync registry which starts a syncer for each code host and will update them
// when external services are changed, added or removed.
func NewSyncRegistry(ctx context.Context, bstore SyncStore, cf *httpcli.Factory, observationContext *observation.Context) *SyncRegistry {
	logger := observationContext.Logger.Scoped("SyncRegistry", "starts a syncer for each code host and updates them")
	ctx, cancel := context.WithCancel(ctx)
	return &SyncRegistry{
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
		syncStore:      bstore,
		httpFactory:    cf,
		priorityNotify: make(chan []int64, 500),
		syncers:        make(map[string]*changesetSyncer),
		metrics:        makeMetrics(observationContext),
	}
}

func (s *SyncRegistry) Start() {
	// Fetch initial list of syncers.
	if err := s.syncCodeHosts(s.ctx); err != nil {
		s.logger.Error("Fetching initial list of code hosts", log.Error(err))
	}

	goroutine.Go(func() {
		s.handlePriorityItems()
	})

	externalServiceSyncer := goroutine.NewPeriodicGoroutine(
		s.ctx,
		externalServiceSyncerInterval,
		goroutine.NewHandlerWithErrorMessage("Batch Changes syncer external service sync", func(ctx context.Context) error {
			return s.syncCodeHosts(ctx)
		}),
	)

	goroutine.MonitorBackgroundRoutines(s.ctx, externalServiceSyncer)
}

func (s *SyncRegistry) Stop() {
	s.cancel()
}

// EnqueueChangesetSyncs will enqueue the changesets with the supplied ids for high priority syncing.
// An error indicates that no changesets have been enqueued.
func (s *SyncRegistry) EnqueueChangesetSyncs(ctx context.Context, ids []int64) error {
	// The channel below is buffered so we'll usually send without blocking.
	// It is important not to block here as this method is called from the UI
	select {
	case s.priorityNotify <- ids:
	default:
		return errors.New("high priority sync capacity reached")
	}
	return nil
}

func (s *SyncRegistry) EnqueueChangesetSyncsForRepos(ctx context.Context, repoIDs []api.RepoID) error {
	cs, _, err := s.syncStore.ListChangesets(ctx, store.ListChangesetsOpts{
		RepoIDs: repoIDs,
	})
	if err != nil {
		return errors.Wrapf(err, "listing changesets for repos %v", repoIDs)
	} else if len(cs) == 0 {
		return nil
	}

	ids := make([]int64, len(cs))
	for i, c := range cs {
		ids[i] = c.ID
	}

	s.logger.Debug(
		"enqueuing syncs for changesets on repos",
		log.Int("repo count", len(repoIDs)),
		log.Int("changeset count", len(ids)),
	)

	return s.EnqueueChangesetSyncs(ctx, ids)
}

// addCodeHostSyncer adds a syncer for the code host associated with the supplied code host if the syncer hasn't
// already been added and starts it.
func (s *SyncRegistry) addCodeHostSyncer(codeHost *btypes.CodeHost) {
	// This should never happen since the store does the filtering for us, but let's be super duper extra cautious.
	if !codeHost.IsSupported() {
		s.logger.Info("Code host not support by batch changes",
			log.String("type", codeHost.ExternalServiceType),
			log.String("url", codeHost.ExternalServiceID))
		return
	}

	syncerKey := codeHost.ExternalServiceID

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.syncers[syncerKey]; ok {
		// Already added
		return
	}

	// We need to be able to cancel the syncer if the code host is removed
	ctx, cancel := context.WithCancel(s.ctx)
	ctx = metrics.ContextWithTask(ctx, "Batches.ChangesetSyncer")

	syncer := &changesetSyncer{
		logger:         s.logger.With(log.String("syncer", syncerKey)),
		syncStore:      s.syncStore,
		httpFactory:    s.httpFactory,
		codeHostURL:    syncerKey,
		cancel:         cancel,
		priorityNotify: make(chan []int64, 500),
		metrics:        s.metrics,
	}

	s.syncers[syncerKey] = syncer

	goroutine.Go(func() {
		syncer.Run(ctx)
	})
}

// handlePriorityItems fetches changesets in the priority queue from the database and passes them
// to the appropriate syncer.
func (s *SyncRegistry) handlePriorityItems() {
	fetchSyncData := func(ids []int64) ([]*btypes.ChangesetSyncData, error) {
		ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
		defer cancel()
		return s.syncStore.ListChangesetSyncData(ctx, store.ListChangesetSyncDataOpts{ChangesetIDs: ids})
	}
	for {
		select {
		case <-s.ctx.Done():
			return
		case ids := <-s.priorityNotify:
			syncData, err := fetchSyncData(ids)
			if err != nil {
				s.logger.Error("Fetching sync data", log.Error(err))
				continue
			}

			// Group changesets by code host
			changesetByHost := make(map[string][]int64)
			for _, d := range syncData {
				changesetByHost[d.RepoExternalServiceID] = append(changesetByHost[d.RepoExternalServiceID], d.ChangesetID)
			}

			// Anonymous func so we can use defer
			func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				for host, changesets := range changesetByHost {
					syncer, ok := s.syncers[host]
					if !ok {
						continue
					}

					select {
					case syncer.priorityNotify <- changesets:
					default:
					}
				}
			}()
		}
	}
}

// syncCodeHosts fetches the list of currently active code hosts on the Sourcegraph instance.
// The running syncers will then be matched against those and missing ones are spawned and
// excess ones are stopped.
func (s *SyncRegistry) syncCodeHosts(ctx context.Context) error {
	codeHosts, err := s.syncStore.ListCodeHosts(ctx, store.ListCodeHostsOpts{})
	if err != nil {
		return err
	}

	codeHostsByExternalServiceID := make(map[string]*btypes.CodeHost)

	// Add and start syncers
	for _, host := range codeHosts {
		codeHostsByExternalServiceID[host.ExternalServiceID] = host
		s.addCodeHostSyncer(host)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// Clean up old syncers.
	for syncerKey := range s.syncers {
		// If there is no code host for the syncer anymore, we want to stop it.
		if _, ok := codeHostsByExternalServiceID[syncerKey]; !ok {
			syncer, exists := s.syncers[syncerKey]
			if exists {
				delete(s.syncers, syncerKey)
				syncer.cancel()
			}
		}
	}
	return nil
}

// A changesetSyncer periodically syncs metadata of changesets
// saved in the database.
type changesetSyncer struct {
	logger      log.Logger
	syncStore   SyncStore
	httpFactory *httpcli.Factory

	metrics *syncerMetrics

	codeHostURL string

	// scheduleInterval determines how often a new schedule will be computed.
	// NOTE: It involves a DB query but no communication with code hosts.
	scheduleInterval time.Duration

	queue          *changesetPriorityQueue
	priorityNotify chan []int64

	// Replaceable for testing
	syncFunc func(ctx context.Context, id int64) error

	// cancel should be called to stop this syncer
	cancel context.CancelFunc
}

type syncerMetrics struct {
	syncs                   *prometheus.CounterVec
	priorityQueued          *prometheus.CounterVec
	syncDuration            *prometheus.HistogramVec
	computeScheduleDuration *prometheus.HistogramVec
	scheduleSize            *prometheus.GaugeVec
	behindSchedule          *prometheus.GaugeVec
}

func makeMetrics(observationContext *observation.Context) *syncerMetrics {
	metrics := &syncerMetrics{
		syncs: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "src_repoupdater_changeset_syncer_syncs",
			Help: "Total number of changeset syncs",
		}, []string{"codehost", "success"}),
		priorityQueued: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "src_repoupdater_changeset_syncer_priority_queued",
			Help: "Total number of priority items added to queue",
		}, []string{"codehost"}),
		syncDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "src_repoupdater_changeset_syncer_sync_duration_seconds",
			Help:    "Time spent syncing changesets",
			Buckets: []float64{1, 2, 5, 10, 30, 60, 120},
		}, []string{"codehost", "success"}),
		computeScheduleDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "src_repoupdater_changeset_syncer_compute_schedule_duration_seconds",
			Help:    "Time spent computing changeset schedule",
			Buckets: []float64{1, 2, 5, 10, 30, 60, 120},
		}, []string{"codehost", "success"}),
		scheduleSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "src_repoupdater_changeset_syncer_schedule_size",
			Help: "The number of changesets scheduled to sync",
		}, []string{"codehost"}),
		behindSchedule: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "src_repoupdater_changeset_syncer_behind_schedule",
			Help: "The number of changesets behind schedule",
		}, []string{"codehost"}),
	}
	observationContext.Registerer.MustRegister(metrics.syncs)
	observationContext.Registerer.MustRegister(metrics.priorityQueued)
	observationContext.Registerer.MustRegister(metrics.syncDuration)
	observationContext.Registerer.MustRegister(metrics.computeScheduleDuration)
	observationContext.Registerer.MustRegister(metrics.scheduleSize)
	observationContext.Registerer.MustRegister(metrics.behindSchedule)

	return metrics
}

// Run will start the process of changeset syncing. It is long running
// and is expected to be launched once at startup.
func (s *changesetSyncer) Run(ctx context.Context) {
	s.logger.Debug("Starting changeset syncer")
	scheduleInterval := s.scheduleInterval
	if scheduleInterval == 0 {
		scheduleInterval = 2 * time.Minute
	}
	if s.syncFunc == nil {
		s.syncFunc = s.SyncChangeset
	}
	s.queue = newChangesetPriorityQueue()
	// How often to refresh the schedule
	scheduleTicker := time.NewTicker(scheduleInterval)

	if !conf.Get().DisableAutoCodeHostSyncs {
		// Get initial schedule
		if sched, err := s.computeSchedule(ctx); err != nil {
			// Non fatal as we'll try again later in the main loop
			s.logger.Error("Computing schedule", log.Error(err))
		} else {
			s.queue.Upsert(sched...)
		}
	}

	var next scheduledSync
	var ok bool

	// NOTE: All mutations of the queue should be done is this loop as operations on the queue
	// are not safe for concurrent use
	for {
		var timer *time.Timer
		var timerChan <-chan time.Time
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
			timerChan = timer.C
		}

		select {
		case <-ctx.Done():
			return
		case <-scheduleTicker.C:
			if timer != nil {
				timer.Stop()
			}

			if conf.Get().DisableAutoCodeHostSyncs {
				continue
			}

			start := s.syncStore.Clock()()
			schedule, err := s.computeSchedule(ctx)
			labelValues := []string{s.codeHostURL, strconv.FormatBool(err == nil)}
			s.metrics.computeScheduleDuration.WithLabelValues(labelValues...).Observe(s.syncStore.Clock()().Sub(start).Seconds())
			if err != nil {
				s.logger.Error("Computing queue", log.Error(err))
				continue
			}
			s.metrics.scheduleSize.WithLabelValues(s.codeHostURL).Set(float64(len(schedule)))
			s.queue.Upsert(schedule...)
			var behindSchedule int
			now := s.syncStore.Clock()()
			for _, ss := range schedule {
				if ss.nextSync.Before(now) {
					behindSchedule++
				}
			}
			s.metrics.behindSchedule.WithLabelValues(s.codeHostURL).Set(float64(behindSchedule))
		case <-timerChan:
			start := s.syncStore.Clock()()
			err := s.syncFunc(ctx, next.changesetID)
			labelValues := []string{s.codeHostURL, strconv.FormatBool(err == nil)}
			s.metrics.syncDuration.WithLabelValues(labelValues...).Observe(s.syncStore.Clock()().Sub(start).Seconds())
			s.metrics.syncs.WithLabelValues(labelValues...).Inc()

			if err != nil {
				s.logger.Error("Syncing changeset", log.Error(err))
				// We'll continue and remove it as it'll get retried on next schedule
			}

			// Remove item now that it has been processed
			s.queue.Remove(next.changesetID)
			s.metrics.scheduleSize.WithLabelValues(s.codeHostURL).Dec()
		case ids := <-s.priorityNotify:
			if timer != nil {
				timer.Stop()
			}
			for _, id := range ids {
				item, ok := s.queue.Get(id)
				if !ok {
					// Item has been recently synced and removed or we have an invalid id
					// We have no way of telling the difference without making a DB call so
					// add a new item anyway which will just lead to a harmless error later
					item = scheduledSync{
						changesetID: id,
						nextSync:    time.Time{},
					}
				}
				item.priority = priorityHigh
				s.queue.Upsert(item)
				s.metrics.scheduleSize.WithLabelValues(s.codeHostURL).Inc()
			}
			s.metrics.priorityQueued.WithLabelValues(s.codeHostURL).Add(float64(len(ids)))
		}
	}
}

func (s *changesetSyncer) computeSchedule(ctx context.Context) ([]scheduledSync, error) {
	syncData, err := s.syncStore.ListChangesetSyncData(ctx, store.ListChangesetSyncDataOpts{ExternalServiceID: s.codeHostURL})
	if err != nil {
		return nil, errors.Wrap(err, "listing changeset sync data")
	}

	ss := make([]scheduledSync, len(syncData))
	for i := range syncData {
		nextSync := NextSync(s.syncStore.Clock(), syncData[i])

		ss[i] = scheduledSync{
			changesetID: syncData[i].ChangesetID,
			nextSync:    nextSync,
		}
	}

	return ss, nil
}

// SyncChangeset will sync a single changeset given its id.
func (s *changesetSyncer) SyncChangeset(ctx context.Context, id int64) error {
	syncLogger := s.logger.With(log.Int64("id", id))
	syncLogger.Debug("SyncChangeset")

	cs, err := s.syncStore.GetChangeset(ctx, store.GetChangesetOpts{
		ID: id,

		// Enforce precondition given in changeset sync state query.
		ReconcilerState:  btypes.ReconcilerStateCompleted,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})
	if err != nil {
		if err == store.ErrNoResults {
			syncLogger.Debug("SyncChangeset not found")
			return nil
		}
		return err
	}

	repo, err := s.syncStore.Repos().Get(ctx, cs.RepoID)
	if err != nil {
		return err
	}

	srcer := sources.NewSourcer(s.httpFactory)
	source, err := srcer.ForChangeset(ctx, s.syncStore, cs)
	if err != nil {
		if errors.Is(err, store.ErrDeletedNamespace) {
			syncLogger.Debug("SyncChangeset skipping changeset: namespace deleted")
			return nil
		}
		return err
	}

	return SyncChangeset(ctx, s.syncStore, source, repo, cs)
}

// SyncChangeset refreshes the metadata of the given changeset and
// updates them in the database.
func SyncChangeset(ctx context.Context, syncStore SyncStore, source sources.ChangesetSource, repo *types.Repo, c *btypes.Changeset) (err error) {
	repoChangeset := &sources.Changeset{TargetRepo: repo, Changeset: c}
	if err := source.LoadChangeset(ctx, repoChangeset); err != nil {
		if !errors.HasType(err, sources.ChangesetNotFoundError{}) {
			// Store the error as the syncer error.
			errMsg := err.Error()
			c.SyncErrorMessage = &errMsg
			if err2 := syncStore.UpdateChangesetCodeHostState(ctx, c); err2 != nil {
				return errors.Wrap(err, err2.Error())
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
	state.SetDerivedState(ctx, syncStore.Repos(), c, events)

	tx, err := syncStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Reset syncer error message state.
	c.SyncErrorMessage = nil

	err = tx.UpdateChangesetCodeHostState(ctx, c)
	if err != nil {
		return err
	}

	return tx.UpsertChangesetEvents(ctx, events...)
}
