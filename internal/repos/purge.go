package repos

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/inconshreveable/log15"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RunRepositoryPurgeWorker is a worker which deletes repos which are present
// on gitserver, but not enabled/present in our repos table.
func RunRepositoryPurgeWorker(ctx context.Context, db database.DB) {
	log := log15.Root().New("worker", "repo-purge")

	// Temporary escape hatch if this feature proves to be dangerous
	if disabled, _ := strconv.ParseBool(os.Getenv("DISABLE_REPO_PURGE")); disabled {
		log.Info("repository purger is disabled via env DISABLE_REPO_PURGE")
		return
	}

	for {
		// We only run in a 1-hour period on the weekend. During normal
		// working hours a migration or admin could accidentally remove all
		// repositories. Recloning all of them is slow, so we drastically
		// reduce the chance of this happening by only purging at a weird time
		// to be configuring Sourcegraph.
		if isSaturdayNight(time.Now()) {
			err := purge(ctx, db, log, database.IteratePurgableReposOptions{})
			if err != nil {
				log.Error("failed to run repository clone purge", "error", err)
			}
		}
		randSleep(10*time.Minute, time.Minute)
	}
}

// PurgeOldestRepos will start a go routine to purge the oldest repos limited by
// limit. The repos are ordered by when they were deleted. limit must be greater
// than zero.
func PurgeOldestRepos(db database.DB, limit int, perSecond float64) error {
	if limit <= 0 {
		return errors.Errorf("limit must be greater than zero, got %d", limit)
	}
	log := log15.Root().New("request", "repo-purge")
	go func() {
		limiter := rate.NewLimiter(rate.Limit(perSecond), 1)
		// Use a background routine so that we don't time out based on the http context.
		if err := purge(context.Background(), db, log, database.IteratePurgableReposOptions{
			Limit:   limit,
			Limiter: limiter,
		}); err != nil {
			log.Error("Purging old repos", "error", err)
		}
	}()
	return nil
}

func purge(ctx context.Context, db database.DB, log log15.Logger, options database.IteratePurgableReposOptions) error {
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
			log.Warn("failed to remove repository", "repo", repo, "error", err)
			purgeFailed.Inc()
			failed++
			return nil
		}
		success++
		purgeSuccess.Inc()
		return nil
	})
	// If we did something we log with a higher level.
	statusLogger := log.Info
	if failed > 0 {
		statusLogger = log.Warn
	}
	statusLogger("repository purge finished", "total", total, "removed", success, "failed", failed, "duration", time.Since(start))
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
