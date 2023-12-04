package repos

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const syncInterval = 5 * time.Minute

var erroredRepoGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "src_repoupdater_syncer_sync_repos_with_last_error_total",
	Help: "Counts number of repos with non empty_last errors which have been synced.",
})

var totalErroredRepos = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "src_repoupdater_syncer_total_errored_repos",
	Help: "Total number of repos with last error currently.",
})

func (s *Syncer) NewSyncReposWithLastErrorsWorker(ctx context.Context, rateLimiter *ratelimit.InstrumentedLimiter) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(ctx),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			s.ObsvCtx.Logger.Info("running worker for SyncReposWithLastErrors", log.Time("time", time.Now()))
			err := s.SyncReposWithLastErrors(ctx, rateLimiter)
			if err != nil {
				return errors.Wrap(err, "Error syncing repos with errors")
			}
			return nil
		}),
		goroutine.WithName("repo-updater.repos-with-last-errors-syncer"),
		goroutine.WithDescription("iterates through all repos which have a non-empty last_error column in the gitserver_repos table, indicating there was an issue updating the repo, and syncs each of these repos. Repos which are no longer visible (i.e. deleted or made private) will be deleted from the DB. Sourcegraph.com only."),
		goroutine.WithInterval(syncInterval),
	)
}

// SyncReposWithLastErrors iterates through all repos which have a non-empty last_error column in the gitserver_repos
// table, indicating there was an issue updating the repo, and syncs each of these repos. Repos which are no longer
// visible (i.e. deleted or made private) will be deleted from the DB. Note that this is only being run in Sourcegraph
// Dot com mode.
func (s *Syncer) SyncReposWithLastErrors(ctx context.Context, rateLimiter *ratelimit.InstrumentedLimiter) error {
	erroredRepoGauge.Set(0)
	s.setTotalErroredRepos(ctx)
	repoNames, err := s.Store.GitserverReposStore().ListReposWithLastError(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list gitserver_repos with last_error not null")
	}

	for _, repoName := range repoNames {
		err := rateLimiter.Wait(ctx)
		if err != nil {
			return errors.Errorf("error waiting for rate limiter: %s", err)
		}
		_, err = s.SyncRepo(ctx, repoName, false)
		if err != nil {
			s.ObsvCtx.Logger.Error("error syncing repo", log.String("repo", string(repoName)), log.Error(err))
		}
		erroredRepoGauge.Inc()
	}

	return err
}

func (s *Syncer) setTotalErroredRepos(ctx context.Context) {
	totalErrored, err := s.Store.GitserverReposStore().TotalErroredCloudDefaultRepos(ctx)
	if err != nil {
		s.ObsvCtx.Logger.Error("error fetching count of total errored repos", log.Error(err))
		return
	}
	totalErroredRepos.Set(float64(totalErrored))
}
