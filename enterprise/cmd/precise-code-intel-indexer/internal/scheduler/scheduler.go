package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

type Scheduler struct {
	store                       store.Store
	gitserverClient             gitserver.Client
	interval                    time.Duration
	batchSize                   int
	minimumTimeSinceLastEnqueue time.Duration
	minimumSearchCount          int
	minimumSearchRatio          float64
	minimumPreciseCount         int
	metrics                     SchedulerMetrics
	done                        chan struct{}
	once                        sync.Once
}

func NewScheduler(
	store store.Store,
	gitserverClient gitserver.Client,
	interval time.Duration,
	batchSize int,
	minimumTimeSinceLastEnqueue time.Duration,
	minimumSearchCount int,
	minimumSearchRatio float64,
	minimumPreciseCount int,
	metrics SchedulerMetrics,
) *Scheduler {
	return &Scheduler{
		store:                       store,
		gitserverClient:             gitserverClient,
		interval:                    interval,
		batchSize:                   batchSize,
		minimumTimeSinceLastEnqueue: minimumTimeSinceLastEnqueue,
		minimumSearchCount:          minimumSearchCount,
		minimumSearchRatio:          minimumSearchRatio,
		minimumPreciseCount:         minimumPreciseCount,
		metrics:                     metrics,
		done:                        make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	for {
		if err := s.update(context.Background()); err != nil {
			s.metrics.Errors.Inc()
			log15.Error("Failed to update indexable repositories", "err", err)
		}

		select {
		case <-time.After(s.interval):
		case <-s.done:
			return
		}
	}
}

func (s *Scheduler) Stop() {
	s.once.Do(func() {
		close(s.done)
	})
}

func (s *Scheduler) update(ctx context.Context) error {
	indexableRepositories, err := s.store.IndexableRepositories(ctx, store.IndexableRepositoryQueryOptions{
		Limit:                       s.batchSize,
		MinimumTimeSinceLastEnqueue: s.minimumTimeSinceLastEnqueue,
		MinimumSearchCount:          s.minimumSearchCount,
		MinimumPreciseCount:         s.minimumPreciseCount,
		MinimumSearchRatio:          s.minimumSearchRatio,
	})
	if err != nil {
		return errors.Wrap(err, "store.IndexableRepositories")
	}

	for _, indexableRepository := range indexableRepositories {
		if err := s.queueIndex(ctx, indexableRepository); err != nil {
			if isRepoNotExist(err) {
				continue
			}

			return err
		}
	}

	return nil
}

func (s *Scheduler) queueIndex(ctx context.Context, indexableRepository store.IndexableRepository) (err error) {
	commit, err := s.gitserverClient.Head(ctx, s.store, indexableRepository.RepositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	isQueued, err := s.store.IsQueued(ctx, indexableRepository.RepositoryID, commit)
	if err != nil {
		return errors.Wrap(err, "store.IsQueued")
	}
	if isQueued {
		return nil
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "store.Transact")
	}
	defer func() {
		err = tx.Done(err)
	}()

	id, err := tx.InsertIndex(ctx, store.Index{
		Commit:       commit,
		RepositoryID: indexableRepository.RepositoryID,
		State:        "queued",
	})
	if err != nil {
		return errors.Wrap(err, "store.QueueIndex")
	}

	now := time.Now().UTC()
	update := store.UpdateableIndexableRepository{
		RepositoryID:        indexableRepository.RepositoryID,
		LastIndexEnqueuedAt: &now,
	}

	if err := tx.UpdateIndexableRepository(ctx, update, now); err != nil {
		return errors.Wrap(err, "store.UpdateIndexableRepository")
	}

	log15.Info(
		"Enqueued index",
		"id", id,
		"repository_id", indexableRepository.RepositoryID,
		"commit", commit,
	)

	return nil
}

func isRepoNotExist(err error) bool {
	for err != nil {
		if vcs.IsRepoNotExist(err) {
			return true
		}

		err = errors.Unwrap(err)
	}

	return false
}
