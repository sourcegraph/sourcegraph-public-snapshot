package indexabilityupdater

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

type Updater struct {
	db              db.DB
	gitserverClient gitserver.Client
	interval        time.Duration
	metrics         UpdaterMetrics
	done            chan struct{}
	once            sync.Once
}

func NewUpdater(
	db db.DB,
	gitserverClient gitserver.Client,
	interval time.Duration,
	metrics UpdaterMetrics,
) *Updater {
	return &Updater{
		db:              db,
		gitserverClient: gitserverClient,
		interval:        interval,
		metrics:         metrics,
		done:            make(chan struct{}),
	}
}

func (u *Updater) Start() {
	for {
		if err := u.update(context.Background()); err != nil {
			u.metrics.Errors.Inc()
			log15.Error("Failed to update index queue", "err", err)
		}

		select {
		case <-time.After(u.interval):
		case <-u.done:
			return
		}
	}
}

func (u *Updater) Stop() {
	u.once.Do(func() {
		close(u.done)
	})
}

func (u *Updater) update(ctx context.Context) error {
	stats, err := u.db.RepoUsageStatistics(ctx)
	if err != nil {
		return errors.Wrap(err, "db.RepoUsageStatistics")
	}

	for _, stat := range stats {
		if err := u.queueRepository(ctx, stat); err != nil {
			if vcs.IsRepoNotExist(err) {
				continue
			}

			return err
		}
	}

	return nil
}

func (u *Updater) queueRepository(ctx context.Context, repoUsageStatistics db.RepoUsageStatistics) error {
	commit, err := u.gitserverClient.Head(ctx, u.db, repoUsageStatistics.RepositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	exists, err := u.gitserverClient.FileExists(ctx, u.db, repoUsageStatistics.RepositoryID, commit, "go.mod")
	if err != nil || !exists {
		return errors.Wrap(err, "gitserver.FileExists")
	}

	// TODO(efritz) - also check repo size

	indexableRepository := db.UpdateableIndexableRepository{
		RepositoryID: repoUsageStatistics.RepositoryID,
		SearchCount:  &repoUsageStatistics.SearchCount,
		PreciseCount: &repoUsageStatistics.PreciseCount,
	}

	if err := u.db.UpdateIndexableRepository(ctx, indexableRepository); err != nil {
		return errors.Wrap(err, "db.UpdateIndexableRepository")
	}

	log15.Debug("Updated indexable repository metadata", "repository_id", repoUsageStatistics.RepositoryID)
	return nil
}
