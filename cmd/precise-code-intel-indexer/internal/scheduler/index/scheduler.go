package indexscheduler

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

type Scheduler struct {
	db                          db.DB
	gitserverClient             gitserver.Client
	interval                    time.Duration
	batchSize                   int
	minimumTimeSinceLastEnqueue time.Duration
	minimumSearchCount          int
	minimumPreciseCount         int
	minimumSearchRatio          float64
	metrics                     SchedulerMetrics
	done                        chan struct{}
	once                        sync.Once
}

func NewScheduler(
	db db.DB,
	gitserverClient gitserver.Client,
	interval time.Duration,
	batchSize int,
	minimumTimeSinceLastEnqueue time.Duration,
	minimumSearchCount int,
	minimumPreciseCount int,
	minimumSearchRatio float64,
	metrics SchedulerMetrics,
) *Scheduler {
	return &Scheduler{
		db:                          db,
		gitserverClient:             gitserverClient,
		interval:                    interval,
		batchSize:                   batchSize,
		minimumTimeSinceLastEnqueue: minimumTimeSinceLastEnqueue,
		minimumSearchCount:          minimumSearchCount,
		minimumPreciseCount:         minimumPreciseCount,
		minimumSearchRatio:          minimumSearchRatio,
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
	indexableRepositories, err := s.db.IndexableRepositories(ctx, db.IndexableRepositoryQueryOptions{
		Limit:                       s.batchSize,
		MinimumTimeSinceLastEnqueue: s.minimumTimeSinceLastEnqueue,
		MinimumSearchCount:          s.minimumSearchCount,
		MinimumPreciseCount:         s.minimumPreciseCount,
		MinimumSearchRatio:          s.minimumSearchRatio,
	})
	if err != nil {
		return errors.Wrap(err, "db.IndexableRepositories")
	}

	for _, indexableRepository := range indexableRepositories {
		if err := s.queueIndex(ctx, indexableRepository); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scheduler) queueIndex(ctx context.Context, indexableRepository db.IndexableRepository) (err error) {
	commit, err := s.gitserverClient.Head(ctx, s.db, indexableRepository.RepositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	isQueued, err := s.db.IsQueued(ctx, indexableRepository.RepositoryID, commit)
	if err != nil {
		return errors.Wrap(err, "db.IsQueued")
	}
	if isQueued {
		return nil
	}

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "db.Transact")
	}
	defer func() {
		err = tx.Done(err)
	}()

	id, err := tx.InsertIndex(ctx, db.Index{
		Commit:       commit,
		RepositoryID: indexableRepository.RepositoryID,
		State:        "queued",
	})
	if err != nil {
		return errors.Wrap(err, "db.QueueIndex")
	}

	now := time.Now()

	if err := tx.UpdateIndexableRepository(ctx, db.UpdateableIndexableRepository{
		RepositoryID:        indexableRepository.RepositoryID,
		LastIndexEnqueuedAt: &now,
	}); err != nil {
		return errors.Wrap(err, "db.UpdateIndexableRepository")
	}

	log15.Info(
		"Enqueued index",
		"id", id,
		"repository_id", indexableRepository.RepositoryID,
		"commit", commit,
	)

	return nil
}
