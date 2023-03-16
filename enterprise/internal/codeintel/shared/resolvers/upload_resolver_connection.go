package sharedresolvers

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type UploadConnectionResolver struct {
	uploadsSvc       UploadsService
	policySvc        PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker SiteAdminChecker
	repoStore        database.RepoStore
	uploadsResolver  *UploadsResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewUploadConnectionResolver(uploadsSvc UploadsService, policySvc PolicyService, gitserverClient gitserver.Client, siteAdminChecker SiteAdminChecker, repoStore database.RepoStore, uploadsResolver *UploadsResolver, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, traceErrs *observation.ErrCollector) resolverstubs.LSIFUploadConnectionResolver {
	return &UploadConnectionResolver{
		uploadsSvc:       uploadsSvc,
		policySvc:        policySvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		uploadsResolver:  uploadsResolver,
		repoStore:        repoStore,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		traceErrs:        traceErrs,
	}
}

func (r *UploadConnectionResolver) Nodes(ctx context.Context) (_ []resolverstubs.LSIFUploadResolver, err error) {
	defer r.traceErrs.Collect(&err, log.String("uploadConnectionResolver.field", "nodes"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.LSIFUploadResolver, 0, len(r.uploadsResolver.Uploads))
	for i := range r.uploadsResolver.Uploads {
		resolvers = append(resolvers, NewUploadResolver(r.uploadsSvc, r.policySvc, r.gitserverClient, r.siteAdminChecker, r.repoStore, r.uploadsResolver.Uploads[i], r.prefetcher, r.locationResolver, r.traceErrs))
	}
	return resolvers, nil
}

func (r *UploadConnectionResolver) TotalCount(ctx context.Context) (_ *int32, err error) {
	defer r.traceErrs.Collect(&err, log.String("uploadConnectionResolver.field", "totalCount"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return toInt32(&r.uploadsResolver.TotalCount), nil
}

func (r *UploadConnectionResolver) PageInfo(ctx context.Context) (_ resolverstubs.PageInfo, err error) {
	defer r.traceErrs.Collect(&err, log.String("uploadConnectionResolver.field", "pageInfo"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return EncodeIntCursor(toInt32(r.uploadsResolver.NextOffset)), nil
}
