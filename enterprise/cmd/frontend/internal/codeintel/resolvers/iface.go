package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type GitserverClient interface {
	CommitExists(ctx context.Context, repositoryID int, commit string) (bool, error)
	CommitGraph(ctx context.Context, repositoryID int, options gitserver.CommitGraphOptions) (*gitserver.CommitGraph, error)
}

type DBStore interface {
	gitserver.DBStore

	GetUploadByID(ctx context.Context, id int) (dbstore.Upload, bool, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) ([]dbstore.Upload, error)
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	DeleteUploadByID(ctx context.Context, id int) (bool, error)
	GetDumpsByIDs(ctx context.Context, ids []int) ([]dbstore.Dump, error)
	FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) ([]dbstore.Dump, error)
	FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, graph *gitserver.CommitGraph) ([]dbstore.Dump, error)
	DefinitionDumps(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []dbstore.Dump, err error)
	ReferenceIDsAndFilters(ctx context.Context, repositoryID int, commit string, monikers []precise.QualifiedMonikerData, limit, offset int) (_ dbstore.PackageReferenceScanner, _ int, err error)
	HasRepository(ctx context.Context, repositoryID int) (bool, error)
	HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error)
	MarkRepositoryAsDirty(ctx context.Context, repositoryID int) error
	CommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, _ error)
	GetIndexByID(ctx context.Context, id int) (dbstore.Index, bool, error)
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]dbstore.Index, error)
	GetIndexes(ctx context.Context, opts dbstore.GetIndexesOptions) ([]dbstore.Index, int, error)
	DeleteIndexByID(ctx context.Context, id int) (bool, error)
	GetConfigurationPolicies(ctx context.Context, opts store.GetConfigurationPoliciesOptions) ([]store.ConfigurationPolicy, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (store.ConfigurationPolicy, bool, error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy store.ConfigurationPolicy) (store.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy store.ConfigurationPolicy) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error)
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (store.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) error
}

type LSIFStore interface {
	Exists(ctx context.Context, bundleID int, path string) (bool, error)
	Stencil(ctx context.Context, bundelID int, path string) ([]lsifstore.Range, error)
	Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) ([]lsifstore.CodeIntelligenceRange, error)
	Definitions(ctx context.Context, bundleID int, path string, line, character, limit, offset int) ([]lsifstore.Location, int, error)
	References(ctx context.Context, bundleID int, path string, line, character, limit, offset int) ([]lsifstore.Location, int, error)
	Hover(ctx context.Context, bundleID int, path string, line, character int) (string, lsifstore.Range, bool, error)
	Diagnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) ([]lsifstore.Diagnostic, int, error)
	MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) ([][]precise.MonikerData, error)
	BulkMonikerResults(ctx context.Context, tableName string, ids []int, args []precise.MonikerData, limit, offset int) (_ []lsifstore.Location, _ int, err error)
	PackageInformation(ctx context.Context, bundleID int, path string, packageInformationID string) (precise.PackageInformationData, bool, error)
	DocumentationPage(ctx context.Context, bundleID int, pathID string) (*precise.DocumentationPageData, error)
	DocumentationPathInfo(ctx context.Context, bundleID int, pathID string) (*precise.DocumentationPathInfoData, error)
	DocumentationDefinitions(ctx context.Context, bundleID int, pathID string, limit, offset int) ([]lsifstore.Location, int, error)
	DocumentationAtPosition(ctx context.Context, bundleID int, path string, line, character int) ([]string, error)
	DocumentationSearch(ctx context.Context, table, query string, repos []string) ([]precise.DocumentationSearchResult, error)
}

type IndexEnqueuer interface {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force bool) ([]dbstore.Index, error)
	InferIndexConfiguration(ctx context.Context, repositoryID int) (*config.IndexConfiguration, error)
}

type RepoUpdaterClient = enqueuer.RepoUpdaterClient
type EnqueuerDBStore = enqueuer.DBStore
type EnqueuerGitserverClient = enqueuer.GitserverClient
