package graphql

import (
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers/gitresolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	autoindexSvc            AutoIndexingService
	uploadSvc               UploadsService
	policySvc               PolicyService
	gitserverClient         gitserver.Client
	operations              *operations
	siteAdminChecker        sharedresolvers.SiteAdminChecker
	repoStore               database.RepoStore
	prefetcherFactory       *PrefetcherFactory
	locationResolverFactory *gitresolvers.CachedLocationResolverFactory
}

func NewRootResolver(
	observationCtx *observation.Context,
	uploadSvc UploadsService,
	autoindexSvc AutoIndexingService,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	repoStore database.RepoStore,
	prefetcherFactory *PrefetcherFactory,
	locationResolverFactory *gitresolvers.CachedLocationResolverFactory,
) resolverstubs.UploadsServiceResolver {
	return &rootResolver{
		autoindexSvc:            autoindexSvc,
		uploadSvc:               uploadSvc,
		policySvc:               policySvc,
		gitserverClient:         gitserverClient,
		operations:              newOperations(observationCtx),
		siteAdminChecker:        siteAdminChecker,
		repoStore:               repoStore,
		prefetcherFactory:       prefetcherFactory,
		locationResolverFactory: locationResolverFactory,
	}
}
