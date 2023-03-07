package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	codeinteltypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type GitserverClient interface {
	CommitGraph(ctx context.Context, repositoryID int, opts gitserver.CommitGraphOptions) (_ *gitdomain.CommitGraph, err error)
	RefDescriptions(ctx context.Context, repositoryID int, pointedAt ...string) (_ map[string][]gitdomain.RefDescription, err error)

	DirectoryChildren(ctx context.Context, repositoryID int, commit string, dirnames []string) (map[string][]string, error)
	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
	DefaultBranchContains(ctx context.Context, repositoryID int, commit string) (bool, error)
}

type Locker interface {
	Lock(ctx context.Context, key int32, blocking bool) (bool, locker.UnlockFunc, error)
}

type PolicyService interface {
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]codeinteltypes.ConfigurationPolicy, int, error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, policies []codeinteltypes.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

type RepoStore interface {
	Get(ctx context.Context, repo api.RepoID) (_ *types.Repo, err error)
	ResolveRev(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error)
}
