package live

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func GetService(db database.DB, syncer dependencies.Syncer) *dependencies.Service {
	if syncer == nil {
		// If no syncer is supplied, then we can expect to be in gitserver or repo-updater
		// service, which doesn't need any of the service behaviors dependenet on the syncer.
		// We install a fail-fast syncer here because we don't expect this value to ever be
		// exercised in a production setup.

		syncer = &errorSyncer{}
	}

	return dependencies.GetService(db, NewGitService(db), syncer)
}

type errorSyncer struct{}

func (s *errorSyncer) Sync(ctx context.Context, repo api.RepoName) error {
	return errors.Newf("codeintel/dependencies/Syncer should not be required by service methods called from gitserver or repoupdater")
}
