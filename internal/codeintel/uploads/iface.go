package uploads

import (
	"context"
	"io"
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	sharedIndexes "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/gitserver"
	codeinteltypes "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	sharedUploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
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
	Head(ctx context.Context, repositoryID int) (string, bool, error)
	CommitExists(ctx context.Context, repositoryID int, commit string) (bool, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)

	ArchiveReader(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, options gitserver.ArchiveOptions) (io.ReadCloser, error)
	RequestRepoUpdate(context.Context, api.RepoName, time.Duration) (*protocol.RepoUpdateResponse, error)
}

type RepoStore interface {
	Get(ctx context.Context, repo api.RepoID) (_ *types.Repo, err error)
	ResolveRev(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error)
}

type UploadServiceForExpiration interface {
	// Uploads
	GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (uploads []codeinteltypes.Upload, totalCount int, err error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error)
	BackfillReferenceCountBatch(ctx context.Context, batchSize int) error

	// Commits
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)

	// Repositories
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error)
}

type PolicyService interface {
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]codeinteltypes.ConfigurationPolicy, int, error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, policies []codeinteltypes.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

type UploadServiceForCleanup interface {
	GetStaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]sharedUploads.SourcedCommits, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (int, int, error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, error)

	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, error)
	SoftDeleteExpiredUploads(ctx context.Context) (int, error)
	HardDeleteExpiredUploads(ctx context.Context) (int, error)

	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (int, error)

	// Workerutil
	WorkerutilStore() dbworkerstore.Store
}

type AutoIndexingService interface {
	GetStaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]sharedIndexes.SourcedCommits, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (int, error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, error)

	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
}

type RepoUpdaterClient interface {
	EnqueueRepoUpdate(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error)
}
