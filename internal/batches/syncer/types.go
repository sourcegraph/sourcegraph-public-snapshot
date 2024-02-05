package syncer

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type ChangesetSyncRegistry interface {
	goroutine.BackgroundRoutine

	UnarchivedChangesetSyncRegistry
	// EnqueueChangesetSyncs will queue the supplied changesets to sync ASAP.
	EnqueueChangesetSyncs(ctx context.Context, ids []int64) error
}

type UnarchivedChangesetSyncRegistry interface {
	// EnqueueChangesetSyncsForRepos will queue a sync for every changeset in
	// every given repo ASAP.
	EnqueueChangesetSyncsForRepos(ctx context.Context, repoIDs []api.RepoID) error
}
