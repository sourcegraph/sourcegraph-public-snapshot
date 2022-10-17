package policies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type GitserverClient interface {
	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)
	RefDescriptions(ctx context.Context, repositoryID int, gitOjbs ...string) (map[string][]gitdomain.RefDescription, error)
	CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error)
}
