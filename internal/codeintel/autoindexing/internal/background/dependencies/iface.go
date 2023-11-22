package dependencies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type DependenciesService interface {
	InsertPackageRepoRefs(ctx context.Context, deps []dependencies.MinimalPackageRepoRef) ([]dependencies.PackageRepoReference, []dependencies.PackageRepoRefVersion, error)
	ListPackageRepoFilters(ctx context.Context, opts dependencies.ListPackageRepoRefFiltersOpts) (_ []dependencies.PackageRepoFilter, hasMore bool, err error)
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

type IndexEnqueuer interface {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []uploadsshared.Index, err error)
	QueueIndexesForPackage(ctx context.Context, pkg dependencies.MinimialVersionedPackageRepo) (err error)
}

type RepoUpdaterClient interface {
	RepoLookup(ctx context.Context, args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)
}

type UploadService interface {
	GetUploadByID(ctx context.Context, id int) (shared.Upload, bool, error)
	ReferencesForUpload(ctx context.Context, uploadID int) (shared.PackageReferenceScanner, error)
}
