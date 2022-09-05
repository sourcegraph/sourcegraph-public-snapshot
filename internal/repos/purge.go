package repos

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RunRepositoryPurgeWorker is a worker which deletes repos which are present on
// gitserver, but not enabled/present in our repos table. ttl, should be >= 0 and
// specifies how long ago a repo must be deleted before it is purged.
func RunRepositoryPurgeWorker(ctx context.Context, logger log.Logger, db database.DB, ttl time.Duration) {
	sglogError := log.Error

	limiter := ratelimit.NewInstrumentedLimiter("PurgeRepoWorker", rate.NewLimiter(10, 1))

	// Temporary escape hatch if this feature proves to be dangerous
	if disabled, _ := strconv.ParseBool(os.Getenv("DISABLE_REPO_PURGE")); disabled {
		logger.Info("repository purger is disabled via env DISABLE_REPO_PURGE")
		return
	}

	for {
		// We only run in a 1-hour period on the weekend. During normal working hours a
		// migration or admin could accidentally remove all repositories. Recloning all
		// of them is slow, so we drastically reduce the chance of this happening by only
		// purging at a weird time to be configuring Sourcegraph.
		now := time.Now()
		if !isSaturdayNight(now) {
			randSleep(10*time.Minute, 1*time.Minute)
			continue
		}
		if err := purge(ctx, logger, db, database.IteratePurgableReposOptions{
			Limit:         5000,
			Limiter:       limiter,
			DeletedBefore: now.Add(-ttl),
		}); err != nil {
			logger.Error("failed to run repository clone purge", sglogError(err))
		}
		randSleep(1*time.Minute, 10*time.Second)
	}
}

// PurgeOldestRepos will start a go routine to purge the oldest repos limited by
// limit. The repos are ordered by when they were deleted. limit must be greater
// than zero.
func PurgeOldestRepos(logger log.Logger, db database.DB, limit int, perSecond float64) error {
	if limit <= 0 {
		return errors.Errorf("limit must be greater than zero, got %d", limit)
	}
	sglogError := log.Error

	go func() {
		limiter := ratelimit.NewInstrumentedLimiter("PurgeOldestRepos", rate.NewLimiter(rate.Limit(perSecond), 1))
		// Use a background routine so that we don't time out based on the http context.
		if err := purge(context.Background(), logger, db, database.IteratePurgableReposOptions{
			Limit:   limit,
			Limiter: limiter,
		}); err != nil {
			logger.Error("Purging old repos", sglogError(err))
		}
	}()
	return nil
}

// purge purges repos, returning the number of repos that were successfully purged
func purge(ctx context.Context, logger log.Logger, db database.DB, options database.IteratePurgableReposOptions) error {
	start := time.Now()
	gitserverClient := gitserver.NewClient(db)
	var (
		total   int
		success int
		failed  int
	)

	err := db.GitserverRepos().IteratePurgeableRepos(ctx, options, func(repo api.RepoName) error {
		if options.Limiter != nil {
			if err := options.Limiter.Wait(ctx); err != nil {
				// A rate limit failure is fatal
				return errors.Wrap(err, "waiting for rate limiter")
			}
		}
		total++
		if err := gitserverClient.Remove(ctx, repo); err != nil {
			// Do not fail at this point, just log so we can remove other repos.
			logger.Warn("failed to remove repository", log.String("repo", string(repo)), log.Error(err))
			purgeFailed.Inc()
			failed++
			return nil
		}
		success++
		purgeSuccess.Inc()
		return nil
	})
	// If we did something we log with a higher level.
	statusLogger := logger.Info
	if failed > 0 {
		statusLogger = logger.Warn
	}
	statusLogger("repository purge finished", log.Int("total", total), log.Int("removed", success), log.Int("failed", failed), log.Duration("duration", time.Since(start)))
	return errors.Wrap(err, "iterating purgeable repos")
}

func isSaturdayNight(t time.Time) bool {
	// According to The Cure, 10:15 Saturday Night you should be sitting in your
	// kitchen sink, not adjusting your external service configuration.
	return t.Format("Mon 15") == "Sat 22"
}

// randSleep will sleep for an expected d duration with a jitter in [-jitter /
// 2, jitter / 2].
func randSleep(d, jitter time.Duration) {
	delta := time.Duration(rand.Int63n(int64(jitter))) - (jitter / 2)
	time.Sleep(d + delta)
}
