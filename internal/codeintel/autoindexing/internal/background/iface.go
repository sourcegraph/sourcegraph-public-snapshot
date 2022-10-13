package background

import (
	"context"
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	autoindexingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	codeinteltypes "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type DependenciesService interface {
	UpsertDependencyRepos(ctx context.Context, deps []dependencies.Repo) ([]dependencies.Repo, error)
}

type GitserverRepoStore interface {
	GetByNames(ctx context.Context, names ...api.RepoName) (map[api.RepoName]*types.GitserverRepo, error)
}

type ExternalServiceStore interface {
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)
	List(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

type ReposStore interface {
	ListMinimalRepos(context.Context, database.ReposListOptions) ([]types.MinimalRepo, error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicyInternal(ctx context.Context, repositoryID int, policies []codeinteltypes.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

type PoliciesService interface {
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]codeinteltypes.ConfigurationPolicy, int, error)
}

type AutoIndexingService interface {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []codeinteltypes.Index, err error)
	QueueIndexesForPackage(ctx context.Context, pkg precise.Package) (err error)
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)

	GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []autoindexingshared.SourcedCommits, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (indexesUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration) (indexesDeleted int, err error)

	ProcessRepoRevs(ctx context.Context, batchSize int) (err error)
}

type RepoUpdaterClient interface {
	RepoLookup(ctx context.Context, args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)
	EnqueueRepoUpdate(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error)
}

type GitserverClient interface {
	Head(ctx context.Context, repositoryID int) (string, bool, error)
	CommitExists(ctx context.Context, repositoryID int, commit string) (bool, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
	ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)

	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)
	RefDescriptions(ctx context.Context, repositoryID int, gitOjbs ...string) (map[string][]gitdomain.RefDescription, error)
	CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error)
}

type InferenceService interface {
	InferIndexJobs(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJob, error)
	InferIndexJobHints(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJobHint, error)
}

type UploadService interface {
	GetRepoName(ctx context.Context, repositoryID int) (_ string, err error)                // upload service
	GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error)                    // upload service
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []codeinteltypes.Upload, err error) // upload service
	GetUploadByID(ctx context.Context, id int) (codeinteltypes.Upload, bool, error)
	ReferencesForUpload(ctx context.Context, uploadID int) (shared.PackageReferenceScanner, error)
	GetRepositoriesForIndexScan(ctx context.Context, table, column string, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) (_ []int, err error)
}
