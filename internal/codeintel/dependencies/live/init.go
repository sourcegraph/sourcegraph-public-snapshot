// This package contains the version of the dependencies service initialization that should
// be used in all non-test uses. We need to have a separate package here because gitserver
// requires the dependencies service (abstractly), but our gitService dependency requires
// the internal/vcs/git. This breaks that circular dependency.
package live

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetService creates or returns an already-initialized dependencies service. If the service is
// new, it will use the given database handle and syncer instance. If the given syncer is nil,
/// then ErrorSyncer will be used instead.
func GetService(db database.DB, syncer dependencies.Syncer) *dependencies.Service {
	if syncer == nil {
		syncer = ErrorSyncer
	}

	return dependencies.GetService(db, NewGitService(db), syncer)
}

// TestService creates a fresh dependencies service with the given database handle and syncer
// instance. If the given syncer is nil, then ErrorSyncer will be used instead.
func TestService(db database.DB, syncer dependencies.Syncer) *dependencies.Service {
	return dependencies.TestService(db, NewGitService(db), syncer)
}

// ErrorSyncer should be used from gitserver and repoupdater code/tests.
//
// If no syncer is supplied, then we can expect to be in gitserver or repo-updater service,
// which doesn't need any of the service behaviors dependenet on the syncer. We install a
// fail-fast syncer here because we don't expect this value to ever be exercised in a production
// setup.
var ErrorSyncer = &errorSyncer{}

type errorSyncer struct{}

func (s *errorSyncer) Sync(ctx context.Context, repo api.RepoName) error {
	return errors.Newf("codeintel/dependencies/Syncer should not be required by service methods called from gitserver or repoupdater")
}
