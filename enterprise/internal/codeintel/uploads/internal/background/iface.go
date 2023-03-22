package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	codeinteltypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Locker interface {
	Lock(ctx context.Context, key int32, blocking bool) (bool, locker.UnlockFunc, error)
}

type PolicyService interface {
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]codeinteltypes.ConfigurationPolicy, int, error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, repoName api.RepoName, policies []codeinteltypes.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

type RepoStore interface {
	Get(ctx context.Context, repo api.RepoID) (_ *types.Repo, err error)
}
