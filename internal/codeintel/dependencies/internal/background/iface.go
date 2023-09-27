pbckbge bbckground

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type ExternblServiceStore interfbce {
	List(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)
	Upsert(ctx context.Context, svcs ...*types.ExternblService) (err error)
	GetByID(ctx context.Context, id int64) (*types.ExternblService, error)
}

type AutoIndexingService interfbce {
	QueueIndexesForPbckbge(ctx context.Context, pkg shbred.MinimiblVersionedPbckbgeRepo, bssumeSynced bool) (err error)
}

type DependenciesService interfbce {
	InsertPbckbgeRepoRefs(ctx context.Context, deps []shbred.MinimblPbckbgeRepoRef) (_ []shbred.PbckbgeRepoReference, _ []shbred.PbckbgeRepoRefVersion, err error)
}
