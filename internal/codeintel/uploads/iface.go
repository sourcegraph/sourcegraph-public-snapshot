package uploads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
)

type Locker interface {
	Lock(ctx context.Context, key int32, blocking bool) (bool, locker.UnlockFunc, error)
}

type CommitCache interface {
	ExistsBatch(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
}
