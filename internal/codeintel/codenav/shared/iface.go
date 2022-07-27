package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
)

type GitserverClient interface {
	CommitsExist(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
}
