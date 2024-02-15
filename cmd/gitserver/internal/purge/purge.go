package purge

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	purgeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_purge", // TODO: Adjust dashboard to new name
		Help: "Incremented each time we remove a repository clone.",
	}, []string{"status"})
)

// NewRepositoryPurgeWorker is a worker which deletes repos which are present on
// gitserver, but not enabled/present in our repos table. ttl, should be >= 0 and
// specifies how long ago a repo must be deleted before it is purged.
func NewRepositoryPurgeWorker(ctx context.Context, logger log.Logger, db database.DB, currentLastRun time.Time, conf conftypes.SiteConfigQuerier) (lastRun time.Time, err error) {
	ctx = actor.WithInternalActor(ctx)
	var timeToNextPurge time.Duration

	purgeConfig := getPurgeConfig(conf)

	if purgeConfig.IntervalMinutes <= 0 {
		logger.Debug("purge worker disabled via site config", log.Int("repoPurgeWorker.interval", purgeConfig.IntervalMinutes))
		return time.Time{}, nil
	}

	if time.Since(currentLastRun) < time.Duration(purgeConfig.IntervalMinutes)*time.Minute {
		logger.Debug("purge worker not due to run yet", log.Int("repoPurgeWorker.interval", purgeConfig.IntervalMinutes), log.Time("lastRun", currentLastRun), log.Time("nextRun", currentLastRun.Add(time.Duration(purgeConfig.IntervalMinutes)*time.Minute)))
		return currentLastRun, nil
	}

	deletedBefore := time.Now().Add(-time.Duration(purgeConfig.DeletedTTLMinutes) * time.Minute)
	purgeLogger := logger.With(log.Time("deletedBefore", deletedBefore))

	timeToNextPurge = time.Duration(purgeConfig.IntervalMinutes) * time.Minute
	purgeLogger.Debug("running repository purge", log.Duration("timeToNextPurge", timeToNextPurge))
	if err := purge(ctx, purgeLogger, db, database.ListPurgableReposOptions{
		Limit:         5000,
		DeletedBefore: deletedBefore,
	}); err != nil {
		// We ran an attempt, so we return time.Now() although it failed, to prevent
		// another run right in the next loop.
		return time.Now(), errors.Wrap(err, "failed to run repository clone purge")
	}

	return time.Now(), nil
}

// PurgeOldestRepos will start a go routine to purge the oldest repos limited by
// limit. The repos are ordered by when they were deleted. limit must be greater
// than zero.
func PurgeOldestRepos(logger log.Logger, db database.DB, limit int) error {
	if limit <= 0 {
		return errors.Errorf("limit must be greater than zero, got %d", limit)
	}
	sglogError := log.Error

	go func() {
		// Use a background routine so that we don't time out based on the http context.
		if err := purge(context.Background(), logger, db, database.ListPurgableReposOptions{
			Limit: limit,
		}); err != nil {
			logger.Error("Purging old repos", sglogError(err))
		}
	}()
	return nil
}

func purge(ctx context.Context, logger log.Logger, db database.DB, conf conftypes.SiteConfigQuerier, options database.ListPurgableReposOptions) error {
	// Delete at most 10 repos per second. We artificially limit throughput here
	// to avoid IO starvation for other more critical parts of gitserver.
	limiter := ratelimit.NewInstrumentedLimiter("PurgeRepoWorker", rate.NewLimiter(10, 1))

	start := time.Now()
	var (
		total   int
		success int
		failed  int
	)

	repos, err := db.GitserverRepos().ListPurgeableRepos(ctx, options)
	if err != nil {
		return errors.Wrap(err, "listing purgeable repos")
	}

	for _, repo := range repos {
		if limiter != nil {
			gitServerAddrs := gitserver.NewGitserverAddresses(conf.Get())
			addrs := gitServerAddrs.Addresses

			var found bool
			for _, a := range addrs {
				if hostnameMatch(shardID, a) {
					found = true
					break
				}
			}
			if !found {
				return errors.Errorf("gitserver hostname, %q, not found in list", shardID)
			}

			// We may have a deleted repo, we need to extract the original name both to
			// ensure that the shard check is correct and also so that we can find the
			// directory.
			repo = api.UndeletedRepoName(repo)

			// Ensure we're only dealing with repos we are responsible for.
			addr := gitServerAddrs.AddrForRepo(ctx, repo)
			if !hostnameMatch(shardID, addr) {
				continue
			}

			if err := limiter.Wait(ctx); err != nil {
				// A rate limit failure is fatal
				return errors.Wrap(err, "waiting for rate limiter")
			}
		}
		total++
		err := gitserverfs.RemoveRepoDirectory(ctx, logger, db, shardID, reposDir, gitDir, true)
		if err != nil {
			// Do not fail at this point, just log so we can remove other repos.
			logger.Warn("failed to remove repository", log.String("repo", string(repo)), log.Error(err))
			purgeCounter.WithLabelValues("failure").Inc()
			failed++
			continue
		}
		success++
		purgeCounter.WithLabelValues("success").Inc()
	}

	// If we did something we log with a higher level.
	statusLogger := logger.Info
	if failed > 0 {
		statusLogger = logger.Warn
	}
	statusLogger("repository purge finished", log.Int("total", total), log.Int("removed", success), log.Int("failed", failed), log.Duration("duration", time.Since(start)))
	return nil
}

func getPurgeConfig(conf conftypes.SiteConfigQuerier) *schema.RepoPurgeWorker {
	purgeConfig := conf.SiteConfig().RepoPurgeWorker
	if purgeConfig == nil {
		purgeConfig = &schema.RepoPurgeWorker{
			// Defaults - align with documentation
			IntervalMinutes:   15,
			DeletedTTLMinutes: 60,
		}
	}
	return purgeConfig
}
