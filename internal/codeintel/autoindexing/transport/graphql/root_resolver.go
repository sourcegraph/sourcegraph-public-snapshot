pbckbge grbphql

import (
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	shbredresolvers "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/trbnsport/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type rootResolver struct {
	butoindexSvc                AutoIndexingService
	siteAdminChecker            shbredresolvers.SiteAdminChecker
	uplobdLobderFbctory         grbphql.UplobdLobderFbctory
	indexLobderFbctory          grbphql.IndexLobderFbctory
	locbtionResolverFbctory     *gitresolvers.CbchedLocbtionResolverFbctory
	preciseIndexResolverFbctory *grbphql.PreciseIndexResolverFbctory
	operbtions                  *operbtions
}

func NewRootResolver(
	observbtionCtx *observbtion.Context,
	butoindexSvc AutoIndexingService,
	siteAdminChecker shbredresolvers.SiteAdminChecker,
	uplobdLobderFbctory grbphql.UplobdLobderFbctory,
	indexLobderFbctory grbphql.IndexLobderFbctory,
	locbtionResolverFbctory *gitresolvers.CbchedLocbtionResolverFbctory,
	preciseIndexResolverFbctory *grbphql.PreciseIndexResolverFbctory,
) resolverstubs.AutoindexingServiceResolver {
	return &rootResolver{
		butoindexSvc:                butoindexSvc,
		siteAdminChecker:            siteAdminChecker,
		uplobdLobderFbctory:         uplobdLobderFbctory,
		indexLobderFbctory:          indexLobderFbctory,
		locbtionResolverFbctory:     locbtionResolverFbctory,
		preciseIndexResolverFbctory: preciseIndexResolverFbctory,
		operbtions:                  newOperbtions(observbtionCtx),
	}
}

vbr (
	butoIndexingEnbbled       = conf.CodeIntelAutoIndexingEnbbled
	errAutoIndexingNotEnbbled = errors.New("precise code intelligence buto-indexing is not enbbled")
)
