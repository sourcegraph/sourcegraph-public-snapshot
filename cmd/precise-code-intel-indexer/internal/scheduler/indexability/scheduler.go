package indexabilityscheduler

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
	db              db.DB
	gitserverClient gitserver.Client
	interval        time.Duration
	metrics         SchedulerMetrics
	done            chan struct{}
	once            sync.Once
}

func NewScheduler(
	db db.DB,
	gitserverClient gitserver.Client,
	interval time.Duration,
	metrics SchedulerMetrics,
) *Scheduler {
	return &Scheduler{
		db:              db,
		gitserverClient: gitserverClient,
		interval:        interval,
		metrics:         metrics,
		done:            make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	for {
		if err := s.update(context.Background()); err != nil {
			s.metrics.Errors.Inc()
			log15.Error("Failed to update index queue", "err", err)
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
	stats, err := s.db.RepoUsageStatistics(ctx)
	if err != nil {
		return errors.Wrap(err, "db.RepoUsageStatistics")
	}

	for _, stat := range stats {
		if err := s.queueRepository(ctx, stat); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scheduler) queueRepository(ctx context.Context, repoUsageStatistics db.RepoUsageStatistics) error {
	commit, err := s.gitserverClient.Head(ctx, s.db, repoUsageStatistics.RepositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	exists, err := s.gitserverClient.FileExists(ctx, s.db, repoUsageStatistics.RepositoryID, commit, "go.mod")
	if err != nil || !exists {
		return errors.Wrap(err, "gitserver.FileExists")
	}

	// TODO - also check repo size

	indexableRepository := db.UpdateableIndexableRepository{
		RepositoryID: repoUsageStatistics.RepositoryID,
		SearchCount:  &repoUsageStatistics.SearchCount,
		PreciseCount: &repoUsageStatistics.PreciseCount,
	}

	if err := s.db.UpdateIndexableRepository(ctx, indexableRepository); err != nil {
		return errors.Wrap(err, "db.UpdateIndexableRepository")
	}

	log15.Debug("Updated indexable repository metadata", "repository_id", repoUsageStatistics.RepositoryID)
	return nil
}
