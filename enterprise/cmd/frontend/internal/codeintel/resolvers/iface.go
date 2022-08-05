package resolvers

import (
	"context"
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	codenavShared "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	gs "github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type GitserverClient interface {
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)
	RefDescriptions(ctx context.Context, repositoryID int, gitOjbs ...string) (map[string][]gitdomain.RefDescription, error)
	CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error)
	CommitsExist(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
	CommitGraph(ctx context.Context, repositoryID int, options gs.CommitGraphOptions) (*gitdomain.CommitGraph, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
}

type DBStore interface {
	gitserver.DBStore

	GetUploadByID(ctx context.Context, id int) (dbstore.Upload, bool, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) ([]dbstore.Upload, error)
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	DeleteUploadByID(ctx context.Context, id int) (bool, error)
	FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) ([]dbstore.Dump, error)
	FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, graph *gitdomain.CommitGraph) ([]dbstore.Dump, error)

	HasRepository(ctx context.Context, repositoryID int) (bool, error)
	HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error)
	MarkRepositoryAsDirty(ctx context.Context, repositoryID int) error
	CommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, _ error)
	GetIndexByID(ctx context.Context, id int) (dbstore.Index, bool, error)
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]dbstore.Index, error)
	GetIndexes(ctx context.Context, opts dbstore.GetIndexesOptions) ([]dbstore.Index, int, error)
	DeleteIndexByID(ctx context.Context, id int) (bool, error)
	GetConfigurationPolicies(ctx context.Context, opts dbstore.GetConfigurationPoliciesOptions) ([]dbstore.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (dbstore.ConfigurationPolicy, bool, error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy dbstore.ConfigurationPolicy) (dbstore.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy dbstore.ConfigurationPolicy) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error)
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (dbstore.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) error
	RepoIDsByGlobPatterns(ctx context.Context, patterns []string, limit, offset int) ([]int, int, error)
	CommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
	RecentUploadsSummary(ctx context.Context, repositoryID int) ([]dbstore.UploadsWithRepositoryNamespace, error)
	RecentIndexesSummary(ctx context.Context, repositoryID int) ([]dbstore.IndexesWithRepositoryNamespace, error)
	LastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	LastIndexScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	RequestLanguageSupport(ctx context.Context, userID int, language string) error
	LanguagesRequestedBy(ctx context.Context, userID int) ([]string, error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) ([]dbstore.UploadLog, error)
	UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) (err error)
}

type LSIFStore interface {
	Exists(ctx context.Context, bundleID int, path string) (bool, error)
	DocumentPaths(ctx context.Context, bundleID int, path string) ([]string, int, error)
}

type IndexEnqueuer interface {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) ([]dbstore.Index, error)
	InferIndexConfiguration(ctx context.Context, repositoryID int, commit string, bypassLimit bool) (*config.IndexConfiguration, []config.IndexJobHint, error)
}

type CodeNavResolver interface {
	Definitions(ctx context.Context, args codenavShared.RequestArgs, requestState codenav.RequestState) (_ []codenavShared.UploadLocation, err error)
	Diagnostics(ctx context.Context, args codenavShared.RequestArgs, requestState codenav.RequestState) (diagnosticsAtUploads []codenavShared.DiagnosticAtUpload, _ int, err error)
	Hover(ctx context.Context, args codenavShared.RequestArgs, requestState codenav.RequestState) (string, codenavShared.Range, bool, error)
	Implementations(ctx context.Context, args codenavShared.RequestArgs, requestState codenav.RequestState) (_ []codenavShared.UploadLocation, _ string, err error)
	LSIFUploads(ctx context.Context, requestState codenav.RequestState) (uploads []codenavShared.Dump, err error)
	Ranges(ctx context.Context, args codenavShared.RequestArgs, requestState codenav.RequestState, startLine, endLine int) (adjustedRanges []codenavShared.AdjustedCodeIntelligenceRange, err error)
	References(ctx context.Context, args codenavShared.RequestArgs, requestState codenav.RequestState) (_ []codenavShared.UploadLocation, _ string, err error)
	Stencil(ctx context.Context, args codenavShared.RequestArgs, requestState codenav.RequestState) (adjustedRanges []codenavShared.Range, err error)

	GetHunkCacheSize() int
}

type (
	RepoUpdaterClient       = autoindexing.RepoUpdaterClient
	EnqueuerDBStore         = autoindexing.DBStore
	EnqueuerGitserverClient = autoindexing.GitserverClient
)
