package repos

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

// RunRepositoryPurgeWorker is a worker which deletes repos which are present
// on gitserver, but not enabled/present in our repos table.
func RunRepositoryPurgeWorker(ctx context.Context) {
	// Temporary escape hatch if this feature proves to be dangerous
	if disabled, _ := strconv.ParseBool(os.Getenv("DISABLE_REPO_PURGE")); disabled {
		log15.Info("repository purger is disabled via env DISABLE_REPO_PURGE")
		return
	}

	for {
		err := purge(ctx)
		if err != nil {
			log15.Error("failed to run repository clone purge", "error", err)
			// Sleep for longer to avoid spamming if we have a consistent
			// failure.
			time.Sleep(10 * time.Minute)
		}
		randSleep(time.Minute, 10*time.Second)
	}
}

func purge(ctx context.Context) error {
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
	enabled := make(map[api.RepoURI]struct{})
	for _, repo := range enabledList {
		enabled[protocol.NormalizeRepo(repo)] = struct{}{}
	}

	success := 0
	failed := 0

	// remove repositories that are in cloned but not in enabled
	for _, repoStr := range cloned {
		repo := protocol.NormalizeRepo(api.RepoURI(repoStr))
		if _, ok := enabled[repo]; ok {
			continue
		}

		// Race condition: A repo can be re-enabled between our listing and
		// now. This should be very rare, so we ignore it since it will get
		// cloned again.
		err := gitserver.DefaultClient.Remove(ctx, repo)
		if err != nil {
			// Do not fail at this point, just log so we can remove other
			// repos.
			log15.Error("Failed to remove disabled repository", "repo", repo, "error", err)
			purgeFailed.Inc()
			failed++
			continue
		}
		log15.Info("removed disabled repository clone", "repo", repo)
		success++
		purgeSuccess.Inc()
	}

	log15.Debug("repository cloned purge finished", "enabled", len(enabled), "cloned", len(cloned)-success, "removed", success, "failed", failed)

	return nil
}

// randSleep will sleep for an expected d duration with a jitter in [-jitter /
// 2, jitter / 2].
func randSleep(d, jitter time.Duration) {
	delta := time.Duration(rand.Int63n(int64(jitter))) - (jitter / 2)
	time.Sleep(d + delta)
}
