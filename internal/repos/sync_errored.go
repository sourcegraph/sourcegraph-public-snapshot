package repos

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
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

func (s *Syncer) RunSyncReposWithLastErrorsWorker(ctx context.Context, rateLimiter *ratelimit.InstrumentedLimiter) {
	for {
		s.Logger.Info("running worker for SyncReposWithLastErrors", log.Time("time", time.Now()))
		err := s.SyncReposWithLastErrors(ctx, rateLimiter)
		if err != nil {
			s.Logger.Error("Error syncing repos w/ errors", log.Error(err))
		}

		// Wait and run task again
		time.Sleep(syncInterval)
	}
}

// SyncReposWithLastErrors iterates through all repos which have a non-empty last_error column in the gitserver_repos
// table, indicating there was an issue updating the repo, and syncs each of these repos. Repos which are no longer
// visible (i.e. deleted or made private) will be deleted from the DB. Note that this is only being run in Sourcegraph
// Dot com mode.
func (s *Syncer) SyncReposWithLastErrors(ctx context.Context, rateLimiter *ratelimit.InstrumentedLimiter) error {
	erroredRepoGauge.Set(0)
	s.setTotalErroredRepos(ctx)
	err := s.Store.GitserverReposStore().IterateWithNonemptyLastError(ctx, func(repo api.RepoName) error {
		err := rateLimiter.Wait(ctx)
		if err != nil {
			return errors.Errorf("error waiting for rate limiter: %s", err)
		}
		_, err = s.SyncRepo(ctx, repo, false)
		if err != nil {
			s.Logger.Error("error syncing repo", log.String("repo", string(repo)), log.Error(err))
		}
		erroredRepoGauge.Inc()
		return nil
	})
	return err
}

func (s *Syncer) setTotalErroredRepos(ctx context.Context) {
	totalErrored, err := s.Store.GitserverReposStore().TotalErroredCloudDefaultRepos(ctx)
	if err != nil {
		s.Logger.Error("error fetching count of total errored repos", log.Error(err))
		return
	}
	totalErroredRepos.Set(float64(totalErrored))
}
