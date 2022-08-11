package policies

import (
	"context"
	"time"

	policiesEnterprise "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
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

type PolicyMatcher interface {
	CommitsDescribedByPolicyInternal(ctx context.Context, repositoryID int, policies []shared.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policiesEnterprise.PolicyMatch, error)
}
