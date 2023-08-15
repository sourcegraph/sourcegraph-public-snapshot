package graphql

import (
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type rootResolver struct {
	autoindexSvc                AutoIndexingService
	siteAdminChecker            sharedresolvers.SiteAdminChecker
	uploadLoaderFactory         graphql.UploadLoaderFactory
	indexLoaderFactory          graphql.IndexLoaderFactory
	locationResolverFactory     *gitresolvers.CachedLocationResolverFactory
	preciseIndexResolverFactory *graphql.PreciseIndexResolverFactory
	operations                  *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	autoindexSvc AutoIndexingService,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	uploadLoaderFactory graphql.UploadLoaderFactory,
	indexLoaderFactory graphql.IndexLoaderFactory,
	locationResolverFactory *gitresolvers.CachedLocationResolverFactory,
	preciseIndexResolverFactory *graphql.PreciseIndexResolverFactory,
) resolverstubs.AutoindexingServiceResolver {
	return &rootResolver{
		autoindexSvc:                autoindexSvc,
		siteAdminChecker:            siteAdminChecker,
		uploadLoaderFactory:         uploadLoaderFactory,
		indexLoaderFactory:          indexLoaderFactory,
		locationResolverFactory:     locationResolverFactory,
		preciseIndexResolverFactory: preciseIndexResolverFactory,
		operations:                  newOperations(observationCtx),
	}
}

var (
	autoIndexingEnabled       = conf.CodeIntelAutoIndexingEnabled
	errAutoIndexingNotEnabled = errors.New("precise code intelligence auto-indexing is not enabled")
)
