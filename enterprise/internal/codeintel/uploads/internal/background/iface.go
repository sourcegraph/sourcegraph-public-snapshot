package background

import (
	"context"
	"io"
	"time"

	"github.com/grafana/regexp"

	policies "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/enterprise"
	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	codeinteltypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type UploadService interface {
	SerializeRankingGraph(ctx context.Context, numRankingRoutines int) error
	VacuumRankingGraph(ctx context.Context) error
}

type GitserverClient interface {
	CommitGraph(ctx context.Context, repositoryID int, opts gitserver.CommitGraphOptions) (_ *gitdomain.CommitGraph, err error)
	RefDescriptions(ctx context.Context, repositoryID int, pointedAt ...string) (_ map[string][]gitdomain.RefDescription, err error)
	ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)

	DirectoryChildren(ctx context.Context, repositoryID int, commit string, dirnames []string) (map[string][]string, error)
	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
	DefaultBranchContains(ctx context.Context, repositoryID int, commit string) (bool, error)

	CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error)
	Head(ctx context.Context, repositoryID int) (string, bool, error)
	CommitExists(ctx context.Context, repositoryID int, commit string) (bool, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)

	ArchiveReader(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, options gitserver.ArchiveOptions) (io.ReadCloser, error)
	RequestRepoUpdate(context.Context, api.RepoName, time.Duration) (*protocol.RepoUpdateResponse, error)
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
