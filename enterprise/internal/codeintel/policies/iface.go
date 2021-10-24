package policies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type GitserverClient interface {
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)
	RefDescriptions(ctx context.Context, repositoryID int) (map[string][]gitserver.RefDescription, error)
	CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error)
}
