package graphql

import (
	"context"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type PreciseIndexResolverFactory struct {
	uploadsSvc       UploadsService
	policySvc        PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker sharedresolvers.SiteAdminChecker
	repoStore        database.RepoStore
}

func NewPreciseIndexResolverFactory(
	uploadsSvc UploadsService,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	repoStore database.RepoStore,
) *PreciseIndexResolverFactory {
	return &PreciseIndexResolverFactory{
		uploadsSvc:       uploadsSvc,
		policySvc:        policySvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		repoStore:        repoStore,
	}
}

func (f *PreciseIndexResolverFactory) Create(
	ctx context.Context,
	uploadLoader UploadLoader,
	autoIndexJobLoader AutoIndexJobLoader,
	locationResolver *gitresolvers.CachedLocationResolver,
	traceErrs *observation.ErrCollector,
	upload *shared.Upload,
	index *uploadsshared.AutoIndexJob,
) (resolverstubs.PreciseIndexResolver, error) {
	return newPreciseIndexResolver(
		ctx,
		f.uploadsSvc,
		f.policySvc,
		f.gitserverClient,
		uploadLoader,
		autoIndexJobLoader,
		f.siteAdminChecker,
		f.repoStore,
		locationResolver,
		traceErrs,
		upload,
		index,
	)
}
