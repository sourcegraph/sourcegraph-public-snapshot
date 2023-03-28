package graphql

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	policySvc        PoliciesService
	repoStore        database.RepoStore
	operations       *operations
	siteAdminChecker sharedresolvers.SiteAdminChecker
}

func NewRootResolver(observationCtx *observation.Context, policySvc *policies.Service, repoStore database.RepoStore, siteAdminChecker sharedresolvers.SiteAdminChecker) resolverstubs.PoliciesServiceResolver {
	return &rootResolver{
		policySvc:        policySvc,
		repoStore:        repoStore,
		operations:       newOperations(observationCtx),
		siteAdminChecker: siteAdminChecker,
	}
}
