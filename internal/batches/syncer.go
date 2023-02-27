// Package batches specifies interfaces that are called by Sourcegraph OSS, but
// implemented in enterprise code in enterprise builds.
//
// No actual Batch Changes functionality is provided in this package.
package batches

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type ChangesetSyncRegistry interface {
	UnarchivedChangesetSyncRegistry
	// EnqueueChangesetSyncs will queue the supplied changesets to sync ASAP.
	EnqueueChangesetSyncs(ctx context.Context, ids []int64) error
}

type UnarchivedChangesetSyncRegistry interface {
	// EnqueueChangesetSyncsForRepos will queue a sync for every changeset in
	// every given repo ASAP.
	EnqueueChangesetSyncsForRepos(ctx context.Context, repoIDs []api.RepoID) error
}
