package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	uploadSvc               UploadService
	autoindexSvc            AutoIndexingService
	policySvc               PolicyService
	gitserverClient         gitserver.Client
	operations              *operations
	siteAdminChecker        sharedresolvers.SiteAdminChecker
	repoStore               database.RepoStore
	prefetcherFactory       *sharedresolvers.PrefetcherFactory
	locationResolverFactory *sharedresolvers.CachedLocationResolverFactory
}

func NewRootResolver(observationCtx *observation.Context, uploadSvc UploadService, policySvc PolicyService, gitserverClient gitserver.Client, siteAdminChecker sharedresolvers.SiteAdminChecker, repoStore database.RepoStore, prefetcherFactory *sharedresolvers.PrefetcherFactory, locationResolverFactory *sharedresolvers.CachedLocationResolverFactory) resolverstubs.UploadsServiceResolver {
	return &rootResolver{
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

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *rootResolver) CommitGraph(ctx context.Context, id graphql.ID) (_ resolverstubs.CodeIntelligenceCommitGraphResolver, err error) {
	ctx, _, endObservation := r.operations.commitGraph.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	repositoryID, err := unmarshalRepositoryID(id)
	if err != nil {
		return nil, err
	}

	stale, updatedAt, err := r.uploadSvc.GetCommitGraphMetadata(ctx, int(repositoryID))
	if err != nil {
		return nil, err
	}

	return NewCommitGraphResolver(stale, updatedAt), nil
}
