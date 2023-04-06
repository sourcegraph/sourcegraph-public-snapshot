package graphql

import (
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers/gitresolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	uploadSvc                   UploadsService
	autoindexSvc                AutoIndexingService
	siteAdminChecker            sharedresolvers.SiteAdminChecker
	prefetcherFactory           *PrefetcherFactory
	locationResolverFactory     *gitresolvers.CachedLocationResolverFactory
	preciseIndexResolverFactory *PreciseIndexResolverFactory
	operations                  *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	uploadSvc UploadsService,
	autoindexSvc AutoIndexingService,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	prefetcherFactory *PrefetcherFactory,
	locationResolverFactory *gitresolvers.CachedLocationResolverFactory,
	preciseIndexResolverFactory *PreciseIndexResolverFactory,
) resolverstubs.UploadsServiceResolver {
	return &rootResolver{
		uploadSvc:                   uploadSvc,
		autoindexSvc:                autoindexSvc,
		siteAdminChecker:            siteAdminChecker,
		prefetcherFactory:           prefetcherFactory,
		locationResolverFactory:     locationResolverFactory,
		preciseIndexResolverFactory: preciseIndexResolverFactory,
		operations:                  newOperations(observationCtx),
	}
}
