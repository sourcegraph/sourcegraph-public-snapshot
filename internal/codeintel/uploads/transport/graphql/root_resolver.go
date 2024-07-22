package graphql

import (
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	uploadSvc                   UploadsService
	autoindexSvc                AutoIndexingService
	siteAdminChecker            sharedresolvers.SiteAdminChecker
	uploadLoaderFactory         UploadLoaderFactory
	autoIndexJobLoaderFactory   AutoIndexJobLoaderFactory
	locationResolverFactory     *gitresolvers.CachedLocationResolverFactory
	preciseIndexResolverFactory *PreciseIndexResolverFactory
	operations                  *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	uploadSvc UploadsService,
	autoindexSvc AutoIndexingService,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	uploadLoaderFactory UploadLoaderFactory,
	autoIndexJobLoaderFactory AutoIndexJobLoaderFactory,
	locationResolverFactory *gitresolvers.CachedLocationResolverFactory,
	preciseIndexResolverFactory *PreciseIndexResolverFactory,
) resolverstubs.UploadsServiceResolver {
	return &rootResolver{
		uploadSvc:                   uploadSvc,
		autoindexSvc:                autoindexSvc,
		siteAdminChecker:            siteAdminChecker,
		uploadLoaderFactory:         uploadLoaderFactory,
		autoIndexJobLoaderFactory:   autoIndexJobLoaderFactory,
		locationResolverFactory:     locationResolverFactory,
		preciseIndexResolverFactory: preciseIndexResolverFactory,
		operations:                  newOperations(observationCtx),
	}
}
