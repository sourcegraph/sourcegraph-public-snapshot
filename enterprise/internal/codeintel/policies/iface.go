package policies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
)

type GitserverClient interface {
	CommitDate(ctx context.Context, repositoryID int, commit string) (time.Time, error)
	RefDescriptions(ctx context.Context, repositoryID int) (map[string][]gitserver.RefDescription, error)
	CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error)
}
