package syncer

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/state"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// SyncRegistry manages a ChangesetSyncer per code host
type SyncRegistry struct {
	Ctx                  context.Context
	SyncStore            SyncStore
	RepoStore            RepoStore
	ExternalServiceStore ExternalServiceStore
	HTTPFactory          *httpcli.Factory

	// Used to receive high priority sync requests
	priorityNotify chan []int64

	mu sync.Mutex
	// key is normalised code host url, also called external_service_id on the repo table
	syncers map[string]*changesetSyncer
}

type RepoStore interface {
	Get(ctx context.Context, id api.RepoID) (*types.Repo, error)
}

type ExternalServiceStore interface {
	List(context.Context, db.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

// NewSyncRegistry creates a new sync registry which starts a syncer for each code host and will update them
// when external services are changed, added or removed.
func NewSyncRegistry(ctx context.Context, store SyncStore, repoStore RepoStore, esStore ExternalServiceStore, cf *httpcli.Factory) *SyncRegistry {
	r := &SyncRegistry{
		Ctx:                  ctx,
		SyncStore:            store,
		RepoStore:            repoStore,
		ExternalServiceStore: esStore,
		HTTPFactory:          cf,
		priorityNotify:       make(chan []int64, 500),
		syncers:              make(map[string]*changesetSyncer),
	}

	services, err := esStore.List(ctx, db.ExternalServicesListOptions{})
	if err != nil {
		log15.Error("Fetching initial external services", "err", err)
	}

	// Add and start syncers
	for _, service := range services {
		r.Add(service)
	}

	go r.handlePriorityItems()

	return r
}

// Add adds a syncer for the code host associated with the supplied external service if the syncer hasn't
// already been added and starts it.
func (s *SyncRegistry) Add(extSvc *types.ExternalService) {
	if !campaigns.IsKindSupported(extSvc.Kind) {
		log15.Info("External service not support by campaigns", "kind", extSvc.Kind)
		return
	}

	normalised, err := externalServiceSyncerKey(extSvc.Kind, extSvc.Config)
	if err != nil {
		log15.Error(err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.syncers[normalised]; ok {
		// Already added
		return
	}

	// We need to be able to cancel the syncer if the service is removed
	ctx, cancel := context.WithCancel(s.Ctx)

	syncer := &changesetSyncer{
		syncStore:            s.SyncStore,
		httpFactory:          s.HTTPFactory,
		reposStore:           s.RepoStore,
		externalServiceStore: s.ExternalServiceStore,
		codeHostURL:          normalised,
		cancel:               cancel,
		priorityNotify:       make(chan []int64, 500),
	}

	s.syncers[normalised] = syncer

	go syncer.Run(ctx)
}

// handlePriorityItems fetches changesets in the priority queue from the db and passes them
// to the appropriate syncer.
func (s *SyncRegistry) handlePriorityItems() {
	fetchSyncData := func(ids []int64) ([]*campaigns.ChangesetSyncData, error) {
		ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
		defer cancel()
		return s.SyncStore.ListChangesetSyncData(ctx, store.ListChangesetSyncDataOpts{ChangesetIDs: ids})
	}
	for {
		select {
		case <-s.Ctx.Done():
			return
		case ids := <-s.priorityNotify:
			syncData, err := fetchSyncData(ids)
			if err != nil {
				log15.Error("Fetching sync data", "err", err)
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

// HandleExternalServiceSync handles changes to external services.
func (s *SyncRegistry) HandleExternalServiceSync(es api.ExternalService) {
	normalised, err := externalServiceSyncerKey(es.Kind, es.Config)
	if err != nil {
		log15.Error(err.Error())
		return
	}

	s.mu.Lock()
	syncer, exists := s.syncers[normalised]
	s.mu.Unlock()

	if es.DeletedAt.IsZero() && !exists {
		res := (types.ExternalService)(es)
		s.Add(&res)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !es.DeletedAt.IsZero() && exists {
		delete(s.syncers, normalised)
		syncer.cancel()
	}
}

func externalServiceSyncerKey(kind, config string) (string, error) {
	baseURL, err := extsvc.ExtractBaseURL(kind, config)
	if err != nil {
		return "", errors.Wrap(err, "getting normalized URL from service")
	}
	return baseURL.String(), nil
}

// A changesetSyncer periodically syncs metadata of changesets
// saved in the database.
type changesetSyncer struct {
	syncStore            SyncStore
	httpFactory          *httpcli.Factory
	reposStore           RepoStore
	externalServiceStore ExternalServiceStore

	codeHostURL string

	// scheduleInterval determines how often a new schedule will be computed.
	// NOTE: It involves a DB query but no communication with code hosts.
	scheduleInterval time.Duration

	queue          *changesetPriorityQueue
	priorityNotify chan []int64

	// Replaceable for testing
	syncFunc func(ctx context.Context, id int64) error
	clock    func() time.Time

	// cancel should be called to stop this syncer
	cancel context.CancelFunc
}

var syncerMetrics = struct {
	syncs                   *prometheus.CounterVec
	priorityQueued          *prometheus.CounterVec
	syncDuration            *prometheus.HistogramVec
	computeScheduleDuration *prometheus.HistogramVec
	scheduleSize            *prometheus.GaugeVec
	behindSchedule          *prometheus.GaugeVec
}{}

func init() {
	syncerMetrics.syncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_changeset_syncer_syncs",
		Help: "Total number of changeset syncs",
	}, []string{"codehost", "success"})
	syncerMetrics.priorityQueued = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_changeset_syncer_priority_queued",
		Help: "Total number of priority items added to queue",
	}, []string{"codehost"})
	syncerMetrics.syncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_repoupdater_changeset_syncer_sync_duration_seconds",
		Help:    "Time spent syncing changesets",
		Buckets: []float64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"codehost", "success"})
	syncerMetrics.computeScheduleDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_repoupdater_changeset_syncer_compute_schedule_duration_seconds",
		Help:    "Time spent computing changeset schedule",
		Buckets: []float64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"codehost", "success"})
	syncerMetrics.scheduleSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_changeset_syncer_schedule_size",
		Help: "The number of changesets scheduled to sync",
	}, []string{"codehost"})
	syncerMetrics.behindSchedule = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_changeset_syncer_behind_schedule",
		Help: "The number of changesets behind schedule",
	}, []string{"codehost"})
}

type SyncStore interface {
	ListChangesetSyncData(context.Context, store.ListChangesetSyncDataOpts) ([]*campaigns.ChangesetSyncData, error)
	GetChangeset(context.Context, store.GetChangesetOpts) (*campaigns.Changeset, error)
	ListChangesets(context.Context, store.ListChangesetsOpts) (campaigns.Changesets, int64, error)
	UpdateChangeset(ctx context.Context, cs *campaigns.Changeset) error
	UpsertChangesetEvents(ctx context.Context, cs ...*campaigns.ChangesetEvent) error
	Transact(context.Context) (*store.Store, error)
}

// Run will start the process of changeset syncing. It is long running
// and is expected to be launched once at startup.
func (s *changesetSyncer) Run(ctx context.Context) {
	scheduleInterval := s.scheduleInterval
	if scheduleInterval == 0 {
		scheduleInterval = 2 * time.Minute
	}
	if s.syncFunc == nil {
		s.syncFunc = s.SyncChangeset
	}
	if s.clock == nil {
		s.clock = time.Now
	}
	s.queue = newChangesetPriorityQueue()
	// How often to refresh the schedule
	scheduleTicker := time.NewTicker(scheduleInterval)

	// Get initial schedule
	if sched, err := s.computeSchedule(ctx); err != nil {
		// Non fatal as we'll try again later in the main loop
		log15.Error("Computing schedule", "err", err)
	} else {
		s.queue.Upsert(sched...)
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
			start := s.clock()
			schedule, err := s.computeSchedule(ctx)
			labelValues := []string{s.codeHostURL, strconv.FormatBool(err == nil)}
			syncerMetrics.computeScheduleDuration.WithLabelValues(labelValues...).Observe(s.clock().Sub(start).Seconds())
			if err != nil {
				log15.Error("Computing queue", "err", err)
				continue
			}
			syncerMetrics.scheduleSize.WithLabelValues(s.codeHostURL).Set(float64(len(schedule)))
			s.queue.Upsert(schedule...)
			var behindSchedule int
			now := s.clock()
			for _, ss := range schedule {
				if ss.nextSync.Before(now) {
					behindSchedule++
				}
			}
			syncerMetrics.behindSchedule.WithLabelValues(s.codeHostURL).Set(float64(behindSchedule))
		case <-timerChan:
			start := s.clock()
			err := s.syncFunc(ctx, next.changesetID)
			labelValues := []string{s.codeHostURL, strconv.FormatBool(err == nil)}
			syncerMetrics.syncDuration.WithLabelValues(labelValues...).Observe(s.clock().Sub(start).Seconds())
			syncerMetrics.syncs.WithLabelValues(labelValues...).Add(1)

			if err != nil {
				log15.Error("Syncing changeset", "err", err)
				// We'll continue and remove it as it'll get retried on next schedule
			}

			// Remove item now that it has been processed
			s.queue.Remove(next.changesetID)
			syncerMetrics.scheduleSize.WithLabelValues(s.codeHostURL).Dec()
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
				syncerMetrics.scheduleSize.WithLabelValues(s.codeHostURL).Inc()
			}
			syncerMetrics.priorityQueued.WithLabelValues(s.codeHostURL).Add(float64(len(ids)))
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
		nextSync := NextSync(s.clock, syncData[i])

		ss[i] = scheduledSync{
			changesetID: syncData[i].ChangesetID,
			nextSync:    nextSync,
		}
	}

	return ss, nil
}

// SyncChangeset will sync a single changeset given its id.
func (s *changesetSyncer) SyncChangeset(ctx context.Context, id int64) error {
	log15.Debug("SyncChangeset", "id", id)
	cs, err := s.syncStore.GetChangeset(ctx, store.GetChangesetOpts{
		ID: id,
	})
	if err != nil {
		return err
	}
	repo, err := loadRepo(ctx, s.reposStore, cs.RepoID)
	if err != nil {
		return err
	}

	externalService, err := loadExternalService(ctx, s.externalServiceStore, repo)
	if err != nil {
		return err
	}

	sourcer := repos.NewSourcer(s.httpFactory)
	source, err := buildChangesetSource(sourcer, externalService)
	if err != nil {
		return err
	}
	return SyncChangeset(ctx, s.syncStore, source, repo, cs)
}

// SyncChangeset refreshes the metadata of the given changeset and
// updates them in the database.
func SyncChangeset(ctx context.Context, syncStore SyncStore, source repos.ChangesetSource, repo *types.Repo, c *campaigns.Changeset) (err error) {
	repoChangeset := &repos.Changeset{Repo: repo, Changeset: c}
	if err := source.LoadChangeset(ctx, repoChangeset); err != nil {
		_, ok := err.(repos.ChangesetNotFoundError)
		if !ok {
			return err
		}

		if !c.IsDeleted() {
			c.SetDeleted()
		}
	}

	events := c.Events()
	state.SetDerivedState(ctx, c, events)

	tx, err := syncStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.UpdateChangeset(ctx, c); err != nil {
		return err
	}

	return tx.UpsertChangesetEvents(ctx, events...)
}

// buildChangesetSource returns a ChangesetSource for the given external service.
func buildChangesetSource(
	sourcer repos.Sourcer,
	extSvc *types.ExternalService,
) (repos.ChangesetSource, error) {
	sources, err := sourcer(extSvc)
	if err != nil {
		return nil, err
	}

	source, ok := sources[0].(repos.ChangesetSource)
	if !ok {
		return nil, fmt.Errorf("ChangesetSource cannot be created from external service %q", extSvc.Kind)
	}

	return source, nil
}
