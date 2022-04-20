package live

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func GetService(db database.DB, syncer dependencies.Syncer) *dependencies.Service {
	return dependencies.GetService(db, NewGitService(db), syncer)
}

// GetServiceWithoutSyncer should be used when constructing a dependencies service that only uses the
// system-level behaviors (e.g., listing repositories to sync since the last request). This syncer value
// will issue errors on invocation indicating that gitserver/repoupdater services are not expected to
// invoke such methods.
func GetServiceWithoutSyncer(db database.DB) *dependencies.Service {
	return dependencies.GetService(db, NewGitService(db), &errorSyncer{})
}

type errorSyncer struct{}

func (s *errorSyncer) Sync(ctx context.Context, repo api.RepoName) error {
	return errors.Newf("codeintel/dependencies/Syncer should not be required by service methods called from gitserver or repoupdater")
}
