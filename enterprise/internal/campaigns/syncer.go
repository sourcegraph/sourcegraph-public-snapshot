package campaigns

import (
	"context"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"gopkg.in/inconshreveable/log15.v2"
)

// A ChangesetSyncer periodically sync the metadata of the changesets
// saved in the database
type ChangesetSyncer struct {
	Store       *Store
	ReposStore  repos.Store
	HTTPFactory *httpcli.Factory

	ScheduleInterval time.Duration

	clock func() time.Time
	queue *changesetQueue
}

// StartSyncing will start the process of changeset syncing. It is long running
// and is expected to be launched once at startup.
func (s *ChangesetSyncer) StartSyncing() {
	// TODO: Setup instrumentation here
	ctx := context.Background()
	scheduleInterval := s.ScheduleInterval
	if scheduleInterval == 0 {
		scheduleInterval = 2 * time.Minute
	}
	if s.clock == nil {
		s.clock = time.Now
	}

	queue := newChangesetQueue(100)
	// Get initial schedule
	for {
		sched, err := s.computeSchedule(ctx)
		if err != nil {
			log15.Error("Computing queue", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		queue.updateSchedule(sched)
		break
	}
	s.queue = queue

	// How often to refresh the schedule
	scheduleTicker := time.NewTicker(scheduleInterval)

	for {
		select {
		case <-scheduleTicker.C:
			// TODO: Retries?
			sched, err := s.computeSchedule(ctx)
			if err != nil {
				log15.Error("Computing queue", "err", err)
				continue
			}
			queue.updateSchedule(sched)
		case id := <-queue.scheduled:
			err := s.SyncChangesetByID(ctx, id)
			if err != nil {
				log15.Error("Syncing changeset", "err", err)
			}
		case id := <-queue.priority:
			err := s.SyncChangesetByID(ctx, id)
			if err != nil {
				log15.Error("Syncing changeset", "err", err)
			}
		}
	}
}

var (
	minSyncDelay = 2 * time.Minute
	maxSyncDelay = 8 * time.Hour
)

// nextSync computes the time we want the next sync to happen
func nextSync(clock func() time.Time, lastSync, lastChange time.Time) time.Time {
	now := clock()

	sinceLastSync := now.Sub(lastSync)
	if sinceLastSync >= maxSyncDelay {
		// TODO: We may want to add some jitter so that we don't have a large cluster
		// of old repos all syncing around the same time
		return now
	}

	// Simple linear backoff for now
	diff := lastSync.Sub(lastChange)
	if diff >= maxSyncDelay {
		diff = maxSyncDelay
	}
	if diff <= minSyncDelay {
		diff = minSyncDelay
	}
	return lastSync.Add(diff)
}

func (s *ChangesetSyncer) computeSchedule(ctx context.Context) ([]syncSchedule, error) {
	hs, err := s.Store.ListChangesetSyncHeuristics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing changeset heuristics")
	}

	ss := make([]syncSchedule, len(hs))
	for i := range hs {
		nextSync := nextSync(s.clock, hs[i].UpdatedAt, maxTime(hs[i].ExternalUpdatedAt, hs[i].LatestEvent))

		ss[i] = syncSchedule{
			changesetID: hs[i].ChangesetID,
			nextSync:    nextSync,
		}
	}

	// This will happen in the db later, for now we'll grab everything and order in code
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].nextSync.Before(ss[j].nextSync)
	})

	return ss, nil
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// EnqueueChangesetSyncs will enqueue the changesets with the supplied ids for high priority syncing.
// An error indicates that no changesets have been synced
func (s *ChangesetSyncer) EnqueueChangesetSyncs(ctx context.Context, ids []int64) error {
	if s.queue == nil {
		return errors.New("background syncing not initialised")
	}
	return s.queue.enqueuePriority(ids)
}

// SyncChangesetByID will sync a single changeset given its id
func (s *ChangesetSyncer) SyncChangesetByID(ctx context.Context, id int64) error {
	cs, err := s.Store.GetChangeset(ctx, GetChangesetOpts{
		ID: id,
	})
	if err != nil {
		return err
	}
	return s.SyncChangesets(ctx, []*campaigns.Changeset{cs}...)
}

// SyncChangesets refreshes the metadata of the given changesets and
// updates them in the database
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

			events = append(events, c.Events()...)
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

func (s *ChangesetSyncer) listAllNonDeletedChangesets(ctx context.Context) (all []*campaigns.Changeset, err error) {
	for cursor := int64(-1); cursor != 0; {
		opts := ListChangesetsOpts{
			Cursor:         cursor,
			Limit:          1000,
			WithoutDeleted: true,
		}
		cs, next, err := s.Store.ListChangesets(ctx, opts)
		if err != nil {
			return nil, err
		}
		all, cursor = append(all, cs...), next
	}

	return all, err
}

type changesetQueue struct {
	scheduled chan int64
	priority  chan int64

	// ordered slice of scheduled syncs
	cancel context.CancelFunc
}

func newChangesetQueue(priorityCapacity int) *changesetQueue {
	return &changesetQueue{
		scheduled: make(chan int64),
		priority:  make(chan int64, priorityCapacity),
	}
}

func (q *changesetQueue) updateSchedule(schedule []syncSchedule) {
	// cancel existing goroutine if running
	if q.cancel != nil {
		// cancel closes the done chan on the existing context
		// which will cause the existing worker routine to exit
		q.cancel()
	}

	ctx := context.Background()
	ctx, q.cancel = context.WithCancel(ctx)
	var timer *time.Timer
	go func() {
		for i := range schedule {
			// Get most urgent changeset and sleep until it should be synced
			now := time.Now()
			nextSync := schedule[i].nextSync
			d := nextSync.Sub(now)
			timer = time.NewTimer(d)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				// Timer ready, try and send sync instruction
			}

			select {
			case <-ctx.Done():
				return
			case q.scheduled <- schedule[i].changesetID:
			}
		}
	}()
}

func (q *changesetQueue) enqueuePriority(ids []int64) error {
	// q.priority will be buffered so that we allow
	// a fixed depth in the high priority queue and
	// can perform a non blocking add here
	for _, id := range ids {
		select {
		case q.priority <- id:
		default:
			return errors.New("queue capacity for priority syncing reached")
		}
	}
	return nil
}

type syncSchedule struct {
	changesetID int64
	nextSync    time.Time
}

// A SourceChangesets groups *repos.Changesets together with the
// repos.ChangesetSource that can be used to modify the changesets.
type SourceChangesets struct {
	repos.ChangesetSource
	Changesets []*repos.Changeset
}
