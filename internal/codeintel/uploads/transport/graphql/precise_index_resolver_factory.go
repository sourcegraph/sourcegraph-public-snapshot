pbckbge grbphql

import (
	"context"

	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	shbredresolvers "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type PreciseIndexResolverFbctory struct {
	uplobdsSvc       UplobdsService
	policySvc        PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker shbredresolvers.SiteAdminChecker
	repoStore        dbtbbbse.RepoStore
}

func NewPreciseIndexResolverFbctory(
	uplobdsSvc UplobdsService,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	siteAdminChecker shbredresolvers.SiteAdminChecker,
	repoStore dbtbbbse.RepoStore,
) *PreciseIndexResolverFbctory {
	return &PreciseIndexResolverFbctory{
		uplobdsSvc:       uplobdsSvc,
		policySvc:        policySvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		repoStore:        repoStore,
	}
}

func (f *PreciseIndexResolverFbctory) Crebte(
	ctx context.Context,
	uplobdLobder UplobdLobder,
	indexLobder IndexLobder,
	locbtionResolver *gitresolvers.CbchedLocbtionResolver,
	trbceErrs *observbtion.ErrCollector,
	uplobd *shbred.Uplobd,
	index *uplobdsshbred.Index,
) (resolverstubs.PreciseIndexResolver, error) {
	return newPreciseIndexResolver(
		ctx,
		f.uplobdsSvc,
		f.policySvc,
		f.gitserverClient,
		uplobdLobder,
		indexLobder,
		f.siteAdminChecker,
		f.repoStore,
		locbtionResolver,
		trbceErrs,
		uplobd,
		index,
	)
}
