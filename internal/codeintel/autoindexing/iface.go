package autoindexing

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type DependenciesService interface {
	UpsertDependencyRepos(ctx context.Context, deps []dependencies.Repo) ([]dependencies.Repo, error)
}

type ReposStore interface {
	ListMinimalRepos(context.Context, database.ReposListOptions) ([]types.MinimalRepo, error)
}

type GitserverRepoStore interface {
	GetByNames(ctx context.Context, names ...api.RepoName) (map[api.RepoName]*types.GitserverRepo, error)
}

type ExternalServiceStore interface {
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)
	List(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

type IndexEnqueuer interface {
	QueueIndexesForPackage(ctx context.Context, pkg precise.Package) error
}

type AutoIndexingServiceForDepScheduling interface {
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error)
}
