package indexing

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	codeinteltypes "github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type DBStore interface {
	GetUploadByID(ctx context.Context, id int) (codeinteltypes.Upload, bool, error)
	ReferencesForUpload(ctx context.Context, uploadID int) (dbstore.PackageReferenceScanner, error)
	InsertCloneableDependencyRepo(ctx context.Context, dependency precise.Package) (bool, error)
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (int, error)
}

type RepoUpdaterClient interface {
	RepoLookup(ctx context.Context, name api.RepoName) (info *protocol.RepoInfo, err error)
}

type ReposStore interface {
	ListMinimalRepos(context.Context, database.ReposListOptions) ([]types.MinimalRepo, error)
}

type ExternalServiceStore interface {
	List(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)
}

type GitserverRepoStore interface {
	GetByNames(ctx context.Context, names ...api.RepoName) (map[api.RepoName]*types.GitserverRepo, error)
}

type IndexEnqueuer interface {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) ([]codeinteltypes.Index, error)
	QueueIndexesForPackage(ctx context.Context, pkg precise.Package) error
}

// For mocking in tests
var autoIndexingEnabled = conf.CodeIntelAutoIndexingEnabled
