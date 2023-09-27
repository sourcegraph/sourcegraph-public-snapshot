pbckbge grbphql

import (
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	shbredresolvers "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type rootResolver struct {
	uplobdSvc                   UplobdsService
	butoindexSvc                AutoIndexingService
	siteAdminChecker            shbredresolvers.SiteAdminChecker
	uplobdLobderFbctory         UplobdLobderFbctory
	indexLobderFbctory          IndexLobderFbctory
	locbtionResolverFbctory     *gitresolvers.CbchedLocbtionResolverFbctory
	preciseIndexResolverFbctory *PreciseIndexResolverFbctory
	operbtions                  *operbtions
}

func NewRootResolver(
	observbtionCtx *observbtion.Context,
	uplobdSvc UplobdsService,
	butoindexSvc AutoIndexingService,
	siteAdminChecker shbredresolvers.SiteAdminChecker,
	uplobdLobderFbctory UplobdLobderFbctory,
	indexLobderFbctory IndexLobderFbctory,
	locbtionResolverFbctory *gitresolvers.CbchedLocbtionResolverFbctory,
	preciseIndexResolverFbctory *PreciseIndexResolverFbctory,
) resolverstubs.UplobdsServiceResolver {
	return &rootResolver{
		uplobdSvc:                   uplobdSvc,
		butoindexSvc:                butoindexSvc,
		siteAdminChecker:            siteAdminChecker,
		uplobdLobderFbctory:         uplobdLobderFbctory,
		indexLobderFbctory:          indexLobderFbctory,
		locbtionResolverFbctory:     locbtionResolverFbctory,
		preciseIndexResolverFbctory: preciseIndexResolverFbctory,
		operbtions:                  newOperbtions(observbtionCtx),
	}
}
