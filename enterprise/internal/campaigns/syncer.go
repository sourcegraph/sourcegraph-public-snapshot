package campaigns

import (
	"container/heap"
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"golang.org/x/time/rate"
)

// SyncRegistry manages a ChangesetSyncer per external service.
type SyncRegistry struct {
	Ctx                 context.Context
	SyncStore           SyncStore
	RepoStore           RepoStore
	HTTPFactory         *httpcli.Factory
	RateLimiterRegistry *repos.RateLimiterRegistry

	priorityNotify chan []int64

	mu      sync.Mutex
	syncers map[int64]*ChangesetSyncer
}

type RepoStore interface {
	ListExternalServices(context.Context, repos.StoreListExternalServicesArgs) ([]*repos.ExternalService, error)
	ListRepos(context.Context, repos.StoreListReposArgs) ([]*repos.Repo, error)
}

// NewSycnRegistry creates a new sync registry which starts a syncer for each external service and will update them
// when external services are changed, added or removed.
func NewSyncRegistry(ctx context.Context, store SyncStore, repoStore RepoStore, cf *httpcli.Factory, rateLimiterRegistry *repos.RateLimiterRegistry) *SyncRegistry {
	r := &SyncRegistry{
		Ctx:                 ctx,
		SyncStore:           store,
		RepoStore:           repoStore,
		HTTPFactory:         cf,
		RateLimiterRegistry: rateLimiterRegistry,
		priorityNotify:      make(chan []int64, 500),
		syncers:             make(map[int64]*ChangesetSyncer),
	}

	services, err := repoStore.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{})
	if err != nil {
		log15.Error("Fetching initial external services", "err", err)
	}

	// Add and start syncers
	for _, service := range services {
		r.Add(service.ID)
	}

	go r.handlePriorityItems()

	return r
}

// Add adds a syncer for the supplied external service if the syncer hasn't already been added and starts it.
func (s *SyncRegistry) Add(extServiceID int64) {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()
	services, err := s.RepoStore.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{
		IDs: []int64{extServiceID},
	})
	if err != nil {
		log15.Error("Listing external services", "err", err)
		return
	}
	if len(services) < 1 {
		return
	}

	service := services[0]

	switch service.Kind {
	case "GITHUB", "BITBUCKETSERVER":
	// Supported by campaigns
	default:
		log15.Debug("Campaigns syncer not started for unsupported code host", "kind", service.Kind)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.syncers[extServiceID]; ok {
		// Already added
		return
	}

	// We need to be able to cancel the syncer if the service is removed
	ctx, cancel = context.WithCancel(s.Ctx)

	syncer := &ChangesetSyncer{
		SyncStore:         s.SyncStore,
		ReposStore:        s.RepoStore,
		HTTPFactory:       s.HTTPFactory,
		externalServiceID: extServiceID,
		cancel:            cancel,
		priorityNotify:    make(chan []int64, 500),
		rateLimitRegistry: s.RateLimiterRegistry,
	}

	s.syncers[extServiceID] = syncer

	go syncer.Run(ctx)
}

// handlePriorityItems fetches changesets in the priority queue from the db and passes them
// to the appropriate syncer.
func (s *SyncRegistry) handlePriorityItems() {
	fetchSyncData := func(ids []int64) ([]campaigns.ChangesetSyncData, error) {
		ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
		defer cancel()
		return s.SyncStore.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{ChangesetIDs: ids})
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

			// Assign changesets to external services
			changesetsByService := make(map[int64][]int64)
			for _, d := range syncData {
				svcID := shardChangeset(d.ChangesetID, d.ExternalServiceIDs)
				changesetsByService[svcID] = append(changesetsByService[svcID], d.ChangesetID)
			}

			// Anonymous func so we can use defer
			func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				for svcID, changesets := range changesetsByService {
					syncer, ok := s.syncers[svcID]
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
	s.mu.Lock()
	syncer, exists := s.syncers[es.ID]
	s.mu.Unlock()

	if timeIsNilOrZero(es.DeletedAt) && !exists {
		s.Add(es.ID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if es.DeletedAt != nil && exists {
		delete(s.syncers, es.ID)
		syncer.cancel()
	}
}

func timeIsNilOrZero(t *time.Time) bool {
	if t == nil {
		return true
	}
	return t.IsZero()
}

// shardChangeset assigns an external service to the supplied changeset.
// each changeset can belong to multiple external services but we only want one syncer to be
// assigned to a changeset.
// externalServices should be sorted in ascending order.
func shardChangeset(changesetID int64, externalServices []int64) (externalServiceID int64) {
	if len(externalServices) == 0 {
		return 0
	}
	// This will consistently return the same index into exteralServices given the same
	// changeset and list of external services. It's random, but deterministic.
	i := int(changesetID) % len(externalServices)
	return externalServices[i]
}

// A ChangesetSyncer periodically syncs metadata of changesets
// saved in the database.
type ChangesetSyncer struct {
	SyncStore   SyncStore
	ReposStore  RepoStore
	HTTPFactory *httpcli.Factory

	externalServiceID int64

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

	// rateLimitRegistry should be used fetch the current rate limiter for an external service
	rateLimitRegistry *repos.RateLimiterRegistry
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
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "changeset_syncer_syncs",
		Help:      "Total number of changeset syncs",
	}, []string{"extsvc", "success"})
	syncerMetrics.priorityQueued = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "changeset_syncer_priority_queued",
		Help:      "Total number of priority items added to queue",
	}, []string{"extsvc"})
	syncerMetrics.syncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "changeset_syncer_sync_duration_seconds",
		Help:      "Time spent syncing changesets",
		Buckets:   []float64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"extsvc", "success"})
	syncerMetrics.computeScheduleDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "changeset_syncer_compute_schedule_duration_seconds",
		Help:      "Time spent computing changeset schedule",
		Buckets:   []float64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"extsvc", "success"})
	syncerMetrics.scheduleSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "changeset_syncer_schedule_size",
		Help:      "The number of changesets scheduled to sync",
	}, []string{"extsvc"})
	syncerMetrics.behindSchedule = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "changeset_syncer_behind_schedule",
		Help:      "The number of changesets behind schedule",
	}, []string{"extsvc"})
}

type SyncStore interface {
	ListChangesetSyncData(context.Context, ListChangesetSyncDataOpts) ([]campaigns.ChangesetSyncData, error)
	GetChangeset(context.Context, GetChangesetOpts) (*campaigns.Changeset, error)
	ListChangesets(context.Context, ListChangesetsOpts) ([]*campaigns.Changeset, int64, error)
	UpdateChangesets(ctx context.Context, cs ...*campaigns.Changeset) error
	UpsertChangesetEvents(ctx context.Context, cs ...*campaigns.ChangesetEvent) error
	Transact(context.Context) (*Store, error)
}

// Run will start the process of changeset syncing. It is long running
// and is expected to be launched once at startup.
func (s *ChangesetSyncer) Run(ctx context.Context) {
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

	svcID := strconv.FormatInt(s.externalServiceID, 10)

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
			start := time.Now()
			schedule, err := s.computeSchedule(ctx)
			labelValues := []string{svcID, strconv.FormatBool(err == nil)}
			syncerMetrics.computeScheduleDuration.WithLabelValues(labelValues...).Observe(time.Since(start).Seconds())
			if err != nil {
				log15.Error("Computing queue", "err", err)
				continue
			}
			syncerMetrics.scheduleSize.WithLabelValues(svcID).Set(float64(len(schedule)))
			s.queue.Upsert(schedule...)
			var behindSchedule int
			now := time.Now()
			for _, ss := range schedule {
				if ss.nextSync.Before(now) {
					behindSchedule++
				}
			}
			syncerMetrics.behindSchedule.WithLabelValues(svcID).Set(float64(behindSchedule))
		case <-timerChan:
			start := time.Now()
			err := s.syncFunc(ctx, next.changesetID)
			labelValues := []string{svcID, strconv.FormatBool(err == nil)}
			syncerMetrics.syncDuration.WithLabelValues(labelValues...).Observe(time.Since(start).Seconds())
			syncerMetrics.syncs.WithLabelValues(labelValues...).Add(1)

			if err != nil {
				log15.Error("Syncing changeset", "err", err)
				// We'll continue and remove it as it'll get retried on next schedule
			}

			// Remove item now that it has been processed
			s.queue.Remove(next.changesetID)
			syncerMetrics.scheduleSize.WithLabelValues(svcID).Dec()
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
				syncerMetrics.scheduleSize.WithLabelValues(svcID).Inc()
			}
			syncerMetrics.priorityQueued.WithLabelValues(svcID).Add(float64(len(ids)))
		}
	}
}

var (
	minSyncDelay = 2 * time.Minute
	maxSyncDelay = 8 * time.Hour
)

// NextSync computes the time we want the next sync to happen.
func NextSync(clock func() time.Time, h campaigns.ChangesetSyncData) time.Time {
	lastSync := h.UpdatedAt

	if lastSync.IsZero() {
		// Edge case where we've never synced
		return clock()
	}

	var lastChange time.Time
	// When we perform a sync, event timestamps are all updated even if nothing has changed.
	// We should fall back to h.ExternalUpdated if the diff is small
	// TODO: This is a workaround while we try to implement syncing without always updating events. See: https://github.com/sourcegraph/sourcegraph/pull/8771
	// Once the above issue is fixed we can simply use maxTime(h.ExternalUpdatedAt, h.LatestEvent)
	if diff := h.LatestEvent.Sub(lastSync); !h.LatestEvent.IsZero() && absDuration(diff) < minSyncDelay {
		lastChange = h.ExternalUpdatedAt
	} else {
		lastChange = maxTime(h.ExternalUpdatedAt, h.LatestEvent)
	}

	// Simple linear backoff for now
	diff := lastSync.Sub(lastChange)

	// If the last change has happened AFTER our last sync this indicates a webhook
	// has arrived. In this case, we should check again in minSyncDelay after
	// the hook arrived. If multiple webhooks arrive in close succession this will
	// cause us to wait for a quiet period of at least minSyncDelay
	if diff < 0 {
		return lastChange.Add(minSyncDelay)
	}

	if diff > maxSyncDelay {
		diff = maxSyncDelay
	}
	if diff < minSyncDelay {
		diff = minSyncDelay
	}
	return lastSync.Add(diff)
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func absDuration(d time.Duration) time.Duration {
	if d >= 0 {
		return d
	}
	return -1 * d
}

func (s *ChangesetSyncer) computeSchedule(ctx context.Context) ([]scheduledSync, error) {
	allSyncData, err := s.SyncStore.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "listing changeset sync data")
	}

	syncData := filterSyncData(s.externalServiceID, allSyncData)

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
func (s *ChangesetSyncer) SyncChangeset(ctx context.Context, id int64) error {
	log15.Debug("SyncChangeset", "id", id)
	cs, err := s.SyncStore.GetChangeset(ctx, GetChangesetOpts{
		ID: id,
	})
	if err != nil {
		return err
	}
	return syncChangesets(ctx, s.ReposStore, s.SyncStore, s.HTTPFactory, s.rateLimitRegistry, cs)
}

// SyncChangesets refreshes the metadata of the given changesets and
// updates them in the database.
func SyncChangesets(ctx context.Context, repoStore RepoStore, syncStore SyncStore, cf *httpcli.Factory, cs ...*campaigns.Changeset) (err error) {
	return syncChangesets(ctx, repoStore, syncStore, cf, nil, cs...)
}

func syncChangesets(ctx context.Context, repoStore RepoStore, syncStore SyncStore, cf *httpcli.Factory, rlr *repos.RateLimiterRegistry, cs ...*campaigns.Changeset) (err error) {
	if len(cs) == 0 {
		return nil
	}

	bySource, err := GroupChangesetsBySource(ctx, repoStore, cf, rlr, cs...)
	if err != nil {
		return err
	}

	return SyncChangesetsWithSources(ctx, syncStore, bySource)
}

// SyncChangesetsWithSources refreshes the metadata of the given changesets
// with the given ChangesetSources and updates them in the database.
func SyncChangesetsWithSources(ctx context.Context, store SyncStore, bySource []*SourceChangesets) (err error) {
	var (
		events []*campaigns.ChangesetEvent
		cs     []*campaigns.Changeset
	)

	for _, s := range bySource {
		var notFound []*repos.Changeset

		err := s.LoadChangesets(ctx, s.Changesets...)
		if err != nil {
			notFoundErr, ok := err.(repos.ChangesetsNotFoundError)
			if !ok {
				return err
			}
			notFound = notFoundErr.Changesets
		}

		notFoundById := make(map[int64]*repos.Changeset, len(notFound))
		for _, c := range notFound {
			notFoundById[c.Changeset.ID] = c
		}

		for _, c := range s.Changesets {
			_, notFound := notFoundById[c.Changeset.ID]
			if notFound && !c.Changeset.IsDeleted() {
				c.Changeset.SetDeleted()
			}

			csEvents := c.Events()
			SetDerivedState(c.Changeset, csEvents)

			events = append(events, csEvents...)
			cs = append(cs, c.Changeset)
		}
	}

	tx, err := store.Transact(ctx)
	if err != nil {
		return err
	}
	defer tx.Done(&err)

	if err = tx.UpdateChangesets(ctx, cs...); err != nil {
		return err
	}

	return tx.UpsertChangesetEvents(ctx, events...)
}

// GroupChangesetsBySource returns a slice of SourceChangesets in which the
// given *campaigns.Changesets are grouped together as repos.Changesets with the
// repos.Source that can modify them.
// rlr is optional
func GroupChangesetsBySource(ctx context.Context, reposStore RepoStore, cf *httpcli.Factory, rlr *repos.RateLimiterRegistry, cs ...*campaigns.Changeset) ([]*SourceChangesets, error) {
	var repoIDs []api.RepoID
	repoSet := map[api.RepoID]*repos.Repo{}

	for _, c := range cs {
		id := c.RepoID
		if _, ok := repoSet[id]; !ok {
			repoSet[id] = nil
			repoIDs = append(repoIDs, id)
		}
	}

	rs, err := reposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
	if err != nil {
		return nil, err
	}

	for _, r := range rs {
		repoSet[r.ID] = r
	}

	for _, c := range cs {
		repo := repoSet[c.RepoID]
		if repo == nil {
			log15.Warn("changeset not synced, repo not in database", "changeset_id", c.ID, "repo_id", c.RepoID)
		}
	}

	es, err := reposStore.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{RepoIDs: repoIDs})
	if err != nil {
		return nil, err
	}

	byRepo := make(map[api.RepoID]int64, len(rs))
	for _, r := range rs {
		eids := r.ExternalServiceIDs()
		for _, id := range eids {
			if _, ok := byRepo[r.ID]; !ok {
				byRepo[r.ID] = id
				break
			}
		}
	}

	bySource := make(map[int64]*SourceChangesets, len(es))
	for _, e := range es {
		var rl *rate.Limiter
		if rlr != nil {
			rl = rlr.GetRateLimiter(e.ID)
		}
		css, err := repos.NewChangesetSource(e, cf, rl)
		if err != nil {
			return nil, err
		}

		bySource[e.ID] = &SourceChangesets{ChangesetSource: css}
	}

	for _, c := range cs {
		repoID := c.RepoID
		s := bySource[byRepo[repoID]]
		if s == nil {
			continue
		}
		s.Changesets = append(s.Changesets, &repos.Changeset{
			Changeset: c,
			Repo:      repoSet[repoID],
		})
	}

	res := make([]*SourceChangesets, 0, len(bySource))
	for _, s := range bySource {
		res = append(res, s)
	}

	return res, nil
}

// filterSyncData filters to changesets that serviceID is responsible for.
func filterSyncData(serviceID int64, allSyncData []campaigns.ChangesetSyncData) []campaigns.ChangesetSyncData {
	syncData := make([]campaigns.ChangesetSyncData, 0, len(allSyncData))
	for _, d := range allSyncData {
		svcID := shardChangeset(d.ChangesetID, d.ExternalServiceIDs)
		if svcID == serviceID {
			syncData = append(syncData, d)
		}
	}
	return syncData
}

type scheduledSync struct {
	changesetID int64
	nextSync    time.Time
	priority    priority
}

// changesetPriorityQueue is a min heap that sorts syncs by priority
// and time of next sync. It is not safe for concurrent use.
type changesetPriorityQueue struct {
	items []scheduledSync
	index map[int64]int // changesetID -> index
}

// newChangesetPriorityQueue creates a new queue for holding changeset sync instructions in chronological order.
// items with a high priority will always appear at the front of the queue.
func newChangesetPriorityQueue() *changesetPriorityQueue {
	q := &changesetPriorityQueue{
		items: make([]scheduledSync, 0),
		index: make(map[int64]int),
	}
	heap.Init(q)
	return q
}

// The following methods implement heap.Interface based on the priority queue example:
// https://golang.org/pkg/container/heap/#example__priorityQueue

func (pq *changesetPriorityQueue) Len() int { return len(pq.items) }

func (pq *changesetPriorityQueue) Less(i, j int) bool {
	// We want items ordered by priority, then NextSync
	// Order by priority and then NextSync
	a := pq.items[i]
	b := pq.items[j]

	if a.priority != b.priority {
		// Greater than here since we want high priority items to be ranked before low priority
		return a.priority > b.priority
	}
	if !a.nextSync.Equal(b.nextSync) {
		return a.nextSync.Before(b.nextSync)
	}
	return a.changesetID < b.changesetID
}

func (pq *changesetPriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.index[pq.items[i].changesetID] = i
	pq.index[pq.items[j].changesetID] = j
}

// Push is here to implement the Heap interface, please use Upsert
func (pq *changesetPriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(scheduledSync)
	pq.index[item.changesetID] = n
	pq.items = append(pq.items, item)
}

// Pop is not to be used directly, use heap.Pop(pq)
func (pq *changesetPriorityQueue) Pop() interface{} {
	item := pq.items[len(pq.items)-1]
	delete(pq.index, item.changesetID)
	pq.items = pq.items[:len(pq.items)-1]
	return item
}

// End of heap methods

// Peek fetches the highest priority item without removing it.
func (pq *changesetPriorityQueue) Peek() (scheduledSync, bool) {
	if len(pq.items) == 0 {
		return scheduledSync{}, false
	}
	return pq.items[0], true
}

// SyncWebhooks modifies at item if it exists or adds a new item if not.
// NOTE: If an existing item is high priority, it will not be changed back
// to normal. This allows high priority items to stay that way through reschedules.
func (pq *changesetPriorityQueue) Upsert(ss ...scheduledSync) {
	for _, s := range ss {
		i, ok := pq.index[s.changesetID]
		if !ok {
			heap.Push(pq, s)
			continue
		}
		oldPriority := pq.items[i].priority
		pq.items[i] = s
		if oldPriority == priorityHigh {
			pq.items[i].priority = priorityHigh
		}
		heap.Fix(pq, i)
	}
}

// Get fetches the item with the supplied id without removing it.
func (pq *changesetPriorityQueue) Get(id int64) (scheduledSync, bool) {
	i, ok := pq.index[id]
	if !ok {
		return scheduledSync{}, false
	}
	item := pq.items[i]
	return item, true
}

func (pq *changesetPriorityQueue) Remove(id int64) {
	i, ok := pq.index[id]
	if !ok {
		return
	}
	heap.Remove(pq, i)
}

type priority int

const (
	priorityNormal priority = iota
	priorityHigh
)

// A SourceChangesets groups *repos.Changesets together with the
// repos.ChangesetSource that can be used to modify the changesets.
type SourceChangesets struct {
	repos.ChangesetSource
	Changesets []*repos.Changeset
}
