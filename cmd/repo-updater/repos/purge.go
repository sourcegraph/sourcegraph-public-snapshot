package repos

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// RunRepositoryPurgeWorker is a worker which deletes repos which are present
// on gitserver, but not enabled/present in our repos table.
func RunRepositoryPurgeWorker(ctx context.Context) {
	log := log15.Root().New("worker", "repo-purge")

	// Temporary escape hatch if this feature proves to be dangerous
	if disabled, _ := strconv.ParseBool(os.Getenv("DISABLE_REPO_PURGE")); disabled {
		log.Info("repository purger is disabled via env DISABLE_REPO_PURGE")
		return
	}

	for {
		// We only run in a 1 hour period on the weekend. During normal
		// working hours a migration or admin could accidentally remove all
		// repositories. Recloning all of them is slow, so we drastically
		// reduce the chance of this happening by only purging at a weird time
		// to be configuring Sourcegraph.
		if isSaturdayNight(time.Now()) {
			err := purge(ctx, log)
			if err != nil {
				log.Error("failed to run repository clone purge", "error", err)
			}
		}
		randSleep(10*time.Minute, time.Minute)
	}
}

func purge(ctx context.Context, log log15.Logger) error {
	// If we fetched enabled first we have the following race condition:
	//
	// 1. Fetched enabled list without repo X.
	// 2. repo X is enabled and cloned.
	// 3. Fetched cloned list with repo X.
	//
	// However, if we fetch cloned first the only race is we may miss a
	// repository that got disabled. The next time purge runs we will remove
	// it though.
	cloned, err := gitserver.DefaultClient.ListCloned(ctx)
	if err != nil {
		return err
	}

	enabledList, err := api.InternalClient.ReposListEnabled(ctx)
	if err != nil {
		return err
	}
	enabled := make(map[api.RepoName]struct{})
	for _, repo := range enabledList {
		enabled[protocol.NormalizeRepo(repo)] = struct{}{}
	}

	success := 0
	failed := 0

	// remove repositories that are in cloned but not in enabled
	for _, repoStr := range cloned {
		repo := protocol.NormalizeRepo(api.RepoName(repoStr))
		if _, ok := enabled[repo]; ok {
			continue
		}

		// Race condition: A repo can be re-enabled between our listing and
		// now. This should be very rare, so we ignore it since it will get
		// cloned again.
		if err = gitserver.DefaultClient.Remove(ctx, repo); err != nil {
			// Do not fail at this point, just log so we can remove other
			// repos.
			log.Error("failed to remove disabled repository", "repo", repo, "error", err)
			purgeFailed.Inc()
			failed++
			continue
		}
		log.Info("removed disabled repository clone", "repo", repo)
		success++
		purgeSuccess.Inc()
	}

	// If we did something we log with a higher level.
	statusLogger := log.Debug
	if success > 0 || failed > 0 {
		statusLogger = log.Info
	}
	statusLogger("repository cloned purge finished", "enabled", len(enabled), "cloned", len(cloned)-success, "removed", success, "failed", failed)

	return nil
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
