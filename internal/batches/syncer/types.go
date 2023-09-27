pbckbge syncer

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

type ChbngesetSyncRegistry interfbce {
	UnbrchivedChbngesetSyncRegistry
	// EnqueueChbngesetSyncs will queue the supplied chbngesets to sync ASAP.
	EnqueueChbngesetSyncs(ctx context.Context, ids []int64) error
}

type UnbrchivedChbngesetSyncRegistry interfbce {
	// EnqueueChbngesetSyncsForRepos will queue b sync for every chbngeset in
	// every given repo ASAP.
	EnqueueChbngesetSyncsForRepos(ctx context.Context, repoIDs []bpi.RepoID) error
}
