package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	codeinteltypes "github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Locker interface {
	Lock(ctx context.Context, key int32, blocking bool) (bool, locker.UnlockFunc, error)
}

type CommitCache interface {
	ExistsBatch(ctx context.Context, commits []codeintelgitserver.RepositoryCommit) ([]bool, error)
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
}

type RepoStore interface {
	Get(ctx context.Context, repo api.RepoID) (_ *types.Repo, err error)
	ResolveRev(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error)
}

type UploadServiceForExpiration interface {
	// Uploads
	GetUploads(ctx context.Context, opts codeinteltypes.GetUploadsOptions) (uploads []codeinteltypes.Upload, totalCount int, err error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error)
	BackfillReferenceCountBatch(ctx context.Context, batchSize int) error

	// Commits
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)

	// Repositories
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error)
}

type PolicyService interface {
	GetConfigurationPolicies(ctx context.Context, opts codeinteltypes.GetConfigurationPoliciesOptions) ([]codeinteltypes.ConfigurationPolicy, int, error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, policies []codeinteltypes.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}
