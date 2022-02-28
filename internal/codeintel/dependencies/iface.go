package dependencies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Syncer interface {
	// Sync will lazily sync the repos that have been inserted into the database but have not yet been
	// cloned. See repos.Syncer.SyncRepo.
	Sync(ctx context.Context, repo api.RepoName) error
}
