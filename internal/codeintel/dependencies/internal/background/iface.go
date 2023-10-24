package background

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ExternalServiceStore interface {
	List(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)
	GetByID(ctx context.Context, id int64) (*types.ExternalService, error)
}

type AutoIndexingService interface {
	QueueIndexesForPackage(ctx context.Context, pkg shared.MinimialVersionedPackageRepo) (err error)
}

type DependenciesService interface {
	InsertPackageRepoRefs(ctx context.Context, deps []shared.MinimalPackageRepoRef) (_ []shared.PackageRepoReference, _ []shared.PackageRepoRefVersion, err error)
}
