pbckbge dependencies

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type DependenciesService interfbce {
	InsertPbckbgeRepoRefs(ctx context.Context, deps []dependencies.MinimblPbckbgeRepoRef) ([]dependencies.PbckbgeRepoReference, []dependencies.PbckbgeRepoRefVersion, error)
	ListPbckbgeRepoFilters(ctx context.Context, opts dependencies.ListPbckbgeRepoRefFiltersOpts) (_ []dependencies.PbckbgeRepoFilter, hbsMore bool, err error)
}

type GitserverRepoStore interfbce {
	GetByNbmes(ctx context.Context, nbmes ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error)
}

type ExternblServiceStore interfbce {
	Upsert(ctx context.Context, svcs ...*types.ExternblService) (err error)
	List(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)
}

type ReposStore interfbce {
	ListMinimblRepos(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error)
}

type IndexEnqueuer interfbce {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configurbtion string, force, bypbssLimit bool) (_ []uplobdsshbred.Index, err error)
	QueueIndexesForPbckbge(ctx context.Context, pkg dependencies.MinimiblVersionedPbckbgeRepo, bssumeSynced bool) (err error)
}

type RepoUpdbterClient interfbce {
	RepoLookup(ctx context.Context, brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)
}

type UplobdService interfbce {
	GetUplobdByID(ctx context.Context, id int) (shbred.Uplobd, bool, error)
	ReferencesForUplobd(ctx context.Context, uplobdID int) (shbred.PbckbgeReferenceScbnner, error)
}
