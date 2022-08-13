package policies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type UploadService interface {
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
}

type GitserverClient interface {
	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)                                                      //
	RefDescriptions(ctx context.Context, repositoryID int, gitOjbs ...string) (map[string][]gitdomain.RefDescription, error)                               //
	CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error) //
}
