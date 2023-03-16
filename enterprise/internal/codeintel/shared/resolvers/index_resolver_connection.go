package sharedresolvers

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type IndexConnectionResolver struct {
	uploadsSvc       UploadsService
	policySvc        PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker SiteAdminChecker
	repoStore        database.RepoStore
	indexesResolver  *IndexesResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	errTracer        *observation.ErrCollector
}

func NewIndexConnectionResolver(uploadsSvc UploadsService, policySvc PolicyService, gitserverClient gitserver.Client, siteAdminChecker SiteAdminChecker, repoStore database.RepoStore, indexesResolver *IndexesResolver, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, errTracer *observation.ErrCollector) resolverstubs.LSIFIndexConnectionResolver {
	return &IndexConnectionResolver{
		uploadsSvc:       uploadsSvc,
		policySvc:        policySvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		repoStore:        repoStore,
		indexesResolver:  indexesResolver,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		errTracer:        errTracer,
	}
}

func (r *IndexConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.LSIFIndexResolver, error) {
	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.LSIFIndexResolver, 0, len(r.indexesResolver.Indexes))
	for i := range r.indexesResolver.Indexes {

		resolvers = append(resolvers, NewIndexResolver(r.uploadsSvc, r.policySvc, r.gitserverClient, r.siteAdminChecker, r.repoStore, r.indexesResolver.Indexes[i], r.prefetcher, r.locationResolver, r.errTracer))
	}
	return resolvers, nil
}

func (r *IndexConnectionResolver) TotalCount(ctx context.Context) (_ *int32, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConnectionResolver.field", "totalCount"))

	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return toInt32(&r.indexesResolver.TotalCount), nil
}

func (r *IndexConnectionResolver) PageInfo(ctx context.Context) (_ resolverstubs.PageInfo, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConnectionResolver.field", "pageInfo"))

	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return EncodeIntCursor(r.indexesResolver.NextOffset), nil
}
