package indexabilityupdater

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"golang.org/x/time/rate"
)

const MaxGitserverRequestsPerSecond = 20

type Updater struct {
	store           store.Store
	gitserverClient gitserver.Client
	metrics         UpdaterMetrics
	limiter         *rate.Limiter
}

var _ goroutine.Handler = &Updater{}

func NewUpdater(
	store store.Store,
	gitserverClient gitserver.Client,
	interval time.Duration,
	metrics UpdaterMetrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &Updater{
		store:           store,
		gitserverClient: gitserverClient,
		metrics:         metrics,
		limiter:         rate.NewLimiter(MaxGitserverRequestsPerSecond, 1),
	})
}

func (u *Updater) Handle(ctx context.Context) error {
	start := time.Now().UTC()

	stats, err := u.store.RepoUsageStatistics(ctx)
	if err != nil {
		return errors.Wrap(err, "store.RepoUsageStatistics")
	}

	for _, stat := range stats {
		if err := u.queueRepository(ctx, stat); err != nil {
			if isRepoNotExist(err) {
				continue
			}

			return err
		}
	}

	// Anything we didn't update hasn't had any activity and didn't come back
	// from RepoUsageStatitsics. Ensure we don't retain the last usage count
	// for these repositories indefinitely.
	if err := u.store.ResetIndexableRepositories(ctx, start); err != nil {
		return errors.Wrap(err, "store.ResetIndexableRepositories")
	}

	return nil
}

func (u *Updater) HandleError(err error) {
	u.metrics.Errors.Inc()
	log15.Error("Failed to update index queue", "err", err)
}

func (u *Updater) queueRepository(ctx context.Context, repoUsageStatistics store.RepoUsageStatistics) error {
	if err := u.limiter.Wait(ctx); err != nil {
		return err
	}

	commit, err := u.gitserverClient.Head(ctx, u.store, repoUsageStatistics.RepositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	exists, err := u.gitserverClient.FileExists(ctx, u.store, repoUsageStatistics.RepositoryID, commit, "go.mod")
	if err != nil || !exists {
		return errors.Wrap(err, "gitserver.FileExists")
	}

	// TODO(efritz) - also check repo size

	indexableRepository := store.UpdateableIndexableRepository{
		RepositoryID: repoUsageStatistics.RepositoryID,
		SearchCount:  &repoUsageStatistics.SearchCount,
		PreciseCount: &repoUsageStatistics.PreciseCount,
	}

	if err := u.store.UpdateIndexableRepository(ctx, indexableRepository, time.Now().UTC()); err != nil {
		return errors.Wrap(err, "store.UpdateIndexableRepository")
	}

	log15.Debug("Updated indexable repository metadata", "repository_id", repoUsageStatistics.RepositoryID)
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
