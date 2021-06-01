package commitgraph

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

type DBStore interface {
	Lock(ctx context.Context, key int, blocking bool) (bool, dbstore.UnlockFunc, error)
	DirtyRepositories(ctx context.Context) (map[int]int, error)
	CalculateVisibleUploads(ctx context.Context, repositoryID int, graph *gitserver.CommitGraph, tipCommit string, dirtyToken int, now time.Time) error
	GetOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error)
}

type GitserverClient interface {
	Head(ctx context.Context, repositoryID int) (string, error)
	CommitGraph(ctx context.Context, repositoryID int, options gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error)
}
