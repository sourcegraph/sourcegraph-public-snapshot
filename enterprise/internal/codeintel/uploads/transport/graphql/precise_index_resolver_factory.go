package graphql

import (
	"context"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers/dataloader"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers/gitresolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
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
	uploadLoader *dataloader.DataLoader[int, shared.Upload],
	indexLoader *dataloader.DataLoader[int, shared.Index],
	locationResolver *gitresolvers.CachedLocationResolver,
	traceErrs *observation.ErrCollector,
	upload *shared.Upload,
	index *uploadsshared.Index,
) (resolverstubs.PreciseIndexResolver, error) {
	return newPreciseIndexResolver(
		ctx,
		f.uploadsSvc,
		f.policySvc,
		f.gitserverClient,
		uploadLoader,
		indexLoader,
		f.siteAdminChecker,
		f.repoStore,
		locationResolver,
		traceErrs,
		upload,
		index,
	)
}
