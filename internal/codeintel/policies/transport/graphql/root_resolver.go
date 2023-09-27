pbckbge grbphql

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	shbredresolvers "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type rootResolver struct {
	policySvc        PoliciesService
	repoStore        dbtbbbse.RepoStore
	siteAdminChecker shbredresolvers.SiteAdminChecker
	operbtions       *operbtions
}

func NewRootResolver(
	observbtionCtx *observbtion.Context,
	policySvc *policies.Service,
	repoStore dbtbbbse.RepoStore,
	siteAdminChecker shbredresolvers.SiteAdminChecker,
) resolverstubs.PoliciesServiceResolver {
	return &rootResolver{
		policySvc:        policySvc,
		repoStore:        repoStore,
		siteAdminChecker: siteAdminChecker,
		operbtions:       newOperbtions(observbtionCtx),
	}
}
