package repos

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

const syncInterval = 2 * time.Minute // TODO: decide on appropriate interval. Currently set low for ease of testing

func (s *Syncer) RunSyncReposWithLastErrorsWorker(ctx context.Context, rateLimiterRegistry *ratelimit.Registry) {
	for {
		log15.Info("running worker for SyncReposWithLastErrors", "time", time.Now())
		s.SyncReposWithLastErrors(ctx, rateLimiterRegistry)

		// Wait and run task again
		time.Sleep(syncInterval)
	}
}

// SyncReposWithLastErrors iterates through all repos which have a non-empty last_error column in the gitserver_repos
// table, indicating there was an issue updating the repo, and syncs each of these repos. Repos which are no longer
// visible (i.e. deleted or made private) will be deleted from the DB. Note that this is only being run in Sourcegraph
// Dot com mode.
func (s *Syncer) SyncReposWithLastErrors(ctx context.Context, rateLimiterRegistry *ratelimit.Registry) {
	err := s.Store.GitserverReposStore.IterateWithNonemptyLastError(ctx, func(repo types.RepoGitserverStatus) error {
		codehost := extsvc.CodeHostOf(repo.Name, extsvc.PublicCodeHosts...)

		err := waitForRateLimit(ctx, rateLimiterRegistry, codehost.ServiceID, 1)
		if err != nil {
			return errors.Errorf("error waiting for rate limiter: %s", err)
		}
		_, err = s.SyncRepo(ctx, repo.Name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log15.Error("Error syncing repos w/ errors", "err", err)
	}
}

// TODO: this is copied from enterprise/cmd/repo-updater/internal/authz/perms_syncer.go, maybe this is worth putting
// in a central location?
func waitForRateLimit(ctx context.Context, registry *ratelimit.Registry, serviceID string, n int) error {
	if registry == nil {
		return nil
	}
	rl := registry.Get(serviceID)
	if err := rl.WaitN(ctx, n); err != nil {
		return err
	}
	return nil
}
