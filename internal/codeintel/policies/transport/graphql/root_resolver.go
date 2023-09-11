package graphql

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	policySvc        PoliciesService
	repoStore        database.RepoStore
	siteAdminChecker sharedresolvers.SiteAdminChecker
	operations       *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	policySvc *policies.Service,
	repoStore database.RepoStore,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
) resolverstubs.PoliciesServiceResolver {
	return &rootResolver{
		policySvc:        policySvc,
		repoStore:        repoStore,
		siteAdminChecker: siteAdminChecker,
		operations:       newOperations(observationCtx),
	}
}
