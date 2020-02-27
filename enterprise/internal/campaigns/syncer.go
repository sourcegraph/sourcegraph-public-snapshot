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

	// The number of changesets to request from the db at a time
	BatchSize        int
	ScheduleInterval time.Duration

	clock func() time.Time
}

// StartSyncing will start the process of changeset syncing. It is long running
// as is expected to be launched once at startup.
func (s *ChangesetSyncer) StartSyncing() {
	// TODO: Setup instrumentation here
	ctx := context.Background()
	scheduleInterval := s.ScheduleInterval
	if scheduleInterval == 0 {
		scheduleInterval = 2 * time.Minute
	}
	batchSize := s.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	if s.clock == nil {
		s.clock = time.Now
	}

	// Get initial queue
	var queue *changesetQueue
	var err error
	for {
		queue, err = s.computeQueue(ctx)
		if err != nil {
			log15.Error("Computing queue", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	// How often to refresh the queue
	scheduleTicker := time.NewTicker(scheduleInterval)

	// TODO: Many functions below were written when we were syncing groups of changesets
	// Use more efficient versions that work for single changeset syncing
	for {
		select {
		case <-scheduleTicker.C:
			// TODO: Retries?
			q, err := s.computeQueue(ctx)
			if err != nil {
				log15.Error("Computing queue", "err", err)
				continue
			}
			queue.cancel()
			queue = q
		case cs := <-queue.csChan:
			log15.Info("Syncing changeset", "external branch", cs.ExternalBranch)
			err := s.SyncChangesets(ctx, []*campaigns.Changeset{cs}...)
			if err != nil {
				log15.Error("Syncing changesets", "err", err)
			}
			log15.Info("Syncing done")
		}
	}
}

func nextUpdate(clock func() time.Time, ours, theirs time.Time) time.Time {
	minDelay := 2 * time.Minute
	maxDelay := 8 * time.Hour
	now := clock()

	sinceLastSync := now.Sub(ours)
	if sinceLastSync >= maxDelay {
		return now
	}
	if sinceLastSync <= minDelay {
		// Last sync was recent, push back next update
		return now.Add(maxDelay)
	}

	// Simple linear backoff for now
	diff := ours.Sub(theirs)
	if diff >= maxDelay {
		diff = maxDelay
	}
	if diff <= minDelay {
		diff = minDelay
	}
	return ours.Add(diff)
}

func (s *ChangesetSyncer) computeQueue(ctx context.Context) (*changesetQueue, error) {
	// This will happen in the db later, for now we'll grab everything and order in code
	cs, err := s.listAllNonDeletedChangesets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing all changesets")
	}

	log15.Info("listAllNonDeletedChangesets", "count", len(cs))

	csd := make([]changesetDeadline, len(cs))
	for i := range cs {
		deadline := nextUpdate(s.clock, cs[i].UpdatedAt, changed(cs[i]))

		csd[i] = changesetDeadline{
			cs:       cs[i],
			deadline: deadline,
		}
	}

	sort.Slice(csd, func(i, j int) bool {
		return csd[i].deadline.Before(csd[j].deadline)
	})

	for i := range csd {
		log15.Info("ChangesetDeadline", "external branch", csd[i].cs.ExternalBranch, "deadline", csd[i].deadline, "diff", csd[i].deadline.Sub(time.Now()))
	}

	q := newChangesetQueue(ctx, s.clock, csd)
	return q, nil

}

func changed(cs *campaigns.Changeset) time.Time {
	return cs.ExternalUpdatedAt
}

// syncAll refreshes the metadata of all changesets and updates them in the
// database
func (s *ChangesetSyncer) syncAll(ctx context.Context) error {
	cs, err := s.listAllNonDeletedChangesets(ctx)
	if err != nil {
		log15.Error("ChangesetSyncer.listAllNonDeletedChangesets", "error", err)
		return err
	}

	if err := s.SyncChangesets(ctx, cs...); err != nil {
		log15.Error("ChangesetSyncer", "error", err)
		return err
	}
	return nil
}

// EnqueueChangesetSyncs will enqueue the changesets with the supplied ids for high priority syncing.
// An error indicates that no changesets have been synced
func (s *ChangesetSyncer) EnqueueChangesetSyncs(ctx context.Context, ids []int64) error {
	// TODO(ryanslade): For now, we're not actually enqueueing but doing a blocking syncAll
	// Change this once we have a proper scheduler in place and we've decided how to deal with
	// it in places where we currently expect blocking
	cs, _, err := s.Store.ListChangesets(ctx, ListChangesetsOpts{
		Limit:          -1,
		IDs:            ids,
		WithoutDeleted: true,
	})
	if err != nil {
		return err
	}
	return s.SyncChangesets(ctx, cs...)
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
	csChan chan *campaigns.Changeset

	ctx    context.Context
	cancel context.CancelFunc
}

func newChangesetQueue(ctx context.Context, clock func() time.Time, csd []changesetDeadline) *changesetQueue {
	q := new(changesetQueue)
	q.ctx, q.cancel = context.WithCancel(ctx)
	q.csChan = make(chan *campaigns.Changeset)

	var timer *time.Timer
	go func() {
		for i := range csd {
			deadline := csd[i].deadline
			d := deadline.Sub(clock())
			timer = time.NewTimer(d)
			log15.Info("queueInnerLoop", "i", i, "of", len(csd), "deadline", deadline, "in", deadline.Sub(clock()))
			select {
			case <-ctx.Done():
				log15.Info("queueInnerLoop one Done")
				timer.Stop()
				return
			case <-timer.C:
			}

			select {
			case <-ctx.Done():
				log15.Info("queueInnerLoop two Done")
				return
			case q.csChan <- csd[i].cs:
			}
		}
	}()

	return q
}

type changesetDeadline struct {
	cs       *campaigns.Changeset
	deadline time.Time
}

// A SourceChangesets groups *repos.Changesets together with the
// repos.ChangesetSource that can be used to modify the changesets.
type SourceChangesets struct {
	repos.ChangesetSource
	Changesets []*repos.Changeset
}
