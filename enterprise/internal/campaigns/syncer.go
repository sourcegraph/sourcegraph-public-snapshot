package campaigns

import (
	"container/heap"
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"gopkg.in/inconshreveable/log15.v2"
)

// A ChangesetSyncer periodically syncs metadata of changesets
// saved in the database.
type ChangesetSyncer struct {
	Store       SyncStore
	ReposStore  repos.Store
	HTTPFactory *httpcli.Factory

	// ComputeScheduleInterval determines how often a new schedule will be computed.
	// Note that it involves a DB query but no communication with codehosts
	ComputeScheduleInterval time.Duration

	queue          *changesetPriorityQueue
	priorityNotify chan []int64

	// Replaceable fo testing
	syncFunc func(ctx context.Context, id int64) error
	clock    func() time.Time
}

type SyncStore interface {
	ListChangesetSyncData(context.Context) ([]campaigns.ChangesetSyncData, error)
	GetChangeset(context.Context, GetChangesetOpts) (*campaigns.Changeset, error)
	ListChangesets(context.Context, ListChangesetsOpts) ([]*campaigns.Changeset, int64, error)
	Transact(context.Context) (*Store, error)
}

// Run will start the process of changeset syncing. It is long running
// and is expected to be launched once at startup.
func (s *ChangesetSyncer) Run(ctx context.Context) {
	// TODO: Setup instrumentation here
	scheduleInterval := s.ComputeScheduleInterval
	if scheduleInterval == 0 {
		scheduleInterval = 2 * time.Minute
	}
	if s.syncFunc == nil {
		s.syncFunc = s.SyncChangesetByID
	}
	if s.clock == nil {
		s.clock = time.Now
	}
	if s.priorityNotify == nil {
		s.priorityNotify = make(chan []int64, 500)
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
			schedule, err := s.computeSchedule(ctx)
			if err != nil {
				log15.Error("Computing queue", "err", err)
				continue
			}
			s.queue.Upsert(schedule...)
		case <-timerChan:
			err := s.syncFunc(ctx, next.changesetID)
			if err != nil {
				log15.Error("Syncing changeset", "err", err)
				// We'll continue and remove it as it'll get retried on next schedule
			}
			// Remove item now that it has been processed
			s.queue.Remove(next.changesetID)
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
			}
		}
	}
}

var (
	minSyncDelay = 2 * time.Minute
	maxSyncDelay = 8 * time.Hour
)

// nextSync computes the time we want the next sync to happen.
func nextSync(clock func() time.Time, h campaigns.ChangesetSyncData) time.Time {
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
	hs, err := s.Store.ListChangesetSyncData(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing changeset sync data")
	}

	ss := make([]scheduledSync, len(hs))
	for i := range hs {
		nextSync := nextSync(s.clock, hs[i])

		ss[i] = scheduledSync{
			changesetID: hs[i].ChangesetID,
			nextSync:    nextSync,
		}
	}

	return ss, nil
}

// EnqueueChangesetSyncs will enqueue the changesets with the supplied ids for high priority syncing.
// An error indicates that no changesets have been enqueued.
func (s *ChangesetSyncer) EnqueueChangesetSyncs(ctx context.Context, ids []int64) error {
	if s.queue == nil {
		return errors.New("background syncing not initialised")
	}
	// The channel below is buffered so we'll usually send without blocking.
	// It is important not to block here as this method is called from the UI
	select {
	case s.priorityNotify <- ids:
	default:
		return errors.New("high priority sync capacity reached")
	}
	return nil
}

// SyncChangesetByID will sync a single changeset given its id.
func (s *ChangesetSyncer) SyncChangesetByID(ctx context.Context, id int64) error {
	log15.Debug("SyncChangesetByID", "id", id)
	cs, err := s.Store.GetChangeset(ctx, GetChangesetOpts{
		ID: id,
	})
	if err != nil {
		return err
	}
	return s.SyncChangesets(ctx, cs)
}

// SyncChangesets refreshes the metadata of the given changesets and
// updates them in the database.
func (s *ChangesetSyncer) SyncChangesets(ctx context.Context, cs ...*campaigns.Changeset) (err error) {
	if len(cs) == 0 {
		return nil
	}

	bySource, err := s.GroupChangesetsBySource(ctx, cs...)
	if err != nil {
		return err
	}

	return s.SyncChangesetsWithSources(ctx, bySource)
}

// SyncChangesetsWithSources refreshes the metadata of the given changesets
// with the given ChangesetSources and updates them in the database.
func (s *ChangesetSyncer) SyncChangesetsWithSources(ctx context.Context, bySource []*SourceChangesets) (err error) {
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
			c.Changeset.SetDerivedState(csEvents)

			events = append(events, csEvents...)
			cs = append(cs, c.Changeset)
		}
	}

	tx, err := s.Store.Transact(ctx)
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
func (s *ChangesetSyncer) GroupChangesetsBySource(ctx context.Context, cs ...*campaigns.Changeset) ([]*SourceChangesets, error) {
	var repoIDs []api.RepoID
	repoSet := map[api.RepoID]*repos.Repo{}

	for _, c := range cs {
		id := c.RepoID
		if _, ok := repoSet[id]; !ok {
			repoSet[id] = nil
			repoIDs = append(repoIDs, id)
		}
	}

	rs, err := s.ReposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
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

	es, err := s.ReposStore.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{RepoIDs: repoIDs})
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
		src, err := repos.NewSource(e, s.HTTPFactory)
		if err != nil {
			return nil, err
		}

		css, ok := src.(repos.ChangesetSource)
		if !ok {
			return nil, errors.Errorf("unsupported repo type %q", e.Kind)
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
	// We want items ordered by priority, then nextSync
	// Order by priority and then nextSync
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

// Upsert modifies at item if it exists or adds a new item if not.
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
