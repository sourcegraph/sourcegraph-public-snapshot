package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	resolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/sharedresolvers"
	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type RootResolver interface {
	CommitGraph(ctx context.Context, id graphql.ID) (CodeIntelligenceCommitGraphResolver, error)
	LSIFUploadByID(ctx context.Context, id graphql.ID) (resolvers.LSIFUploadResolver, error)
	LSIFUploads(ctx context.Context, args *LSIFUploadsQueryArgs) (resolvers.LSIFUploadConnectionResolver, error)
	LSIFUploadsByRepo(ctx context.Context, args *LSIFRepositoryUploadsQueryArgs) (resolvers.LSIFUploadConnectionResolver, error)
	DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
}

type rootResolver struct {
	svc            *uploads.Service
	autoindexer    AutoIndexingService
	policyResolver PolicyResolver
	operations     *operations
}

func NewRootResolver(svc *uploads.Service, autoindexer AutoIndexingService, policyResolver PolicyResolver, observationContext *observation.Context) RootResolver {
	return &rootResolver{
		svc:            svc,
		autoindexer:    autoindexer,
		policyResolver: policyResolver,
		operations:     newOperations(observationContext),
	}
}

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *rootResolver) CommitGraph(ctx context.Context, id graphql.ID) (_ CodeIntelligenceCommitGraphResolver, err error) {
	ctx, _, endObservation := r.operations.commitGraph.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	repositoryID, err := unmarshalRepositoryID(id)
	if err != nil {
		return nil, err
	}

	// commitGraphResolver := r.resolver.UploadsResolver().CommitGraphResolverFromFactory(ctx, int(repositoryID))
	stale, updatedAt, err := r.svc.GetCommitGraphMetadata(ctx, int(repositoryID))
	if err != nil {
		return nil, err
	}

	return NewCommitGraphResolver(stale, updatedAt), nil

	// return commitGraphResolver,
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploadByID
func (r *rootResolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ resolvers.LSIFUploadResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.lsifUploadByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("uploadID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	uploadID, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := resolvers.NewPrefetcher(r.autoindexer, r.svc)

	upload, exists, err := prefetcher.GetUploadByID(ctx, int(uploadID))
	if err != nil || !exists {
		return nil, err
	}

	return resolvers.NewUploadResolver(r.svc, r.autoindexer, r.policyResolver, upload, prefetcher, traceErrs), nil
}

type LSIFUploadsQueryArgs struct {
	ConnectionArgs
	Query           *string
	State           *string
	IsLatestForRepo *bool
	DependencyOf    *graphql.ID
	DependentOf     *graphql.ID
	After           *string
	IncludeDeleted  *bool
}

type LSIFRepositoryUploadsQueryArgs struct {
	*LSIFUploadsQueryArgs
	RepositoryID graphql.ID
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploads
func (r *rootResolver) LSIFUploads(ctx context.Context, args *LSIFUploadsQueryArgs) (_ resolvers.LSIFUploadConnectionResolver, err error) {
	// ctx, _, endObservation := r.observationContext.lsifUploads.With(ctx, &err, observation.Args{})
	// endObservation.EndOnCancel(ctx, 1, observation.Args{})

	// Delegate behavior to LSIFUploadsByRepo with no specified repository identifier
	return r.LSIFUploadsByRepo(ctx, &LSIFRepositoryUploadsQueryArgs{LSIFUploadsQueryArgs: args})
}

func (r *rootResolver) LSIFUploadsByRepo(ctx context.Context, args *LSIFRepositoryUploadsQueryArgs) (_ resolvers.LSIFUploadConnectionResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.lsifUploadsByRepo.WithErrors(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("repoID", string(args.RepositoryID)),
		},
	})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	opts, err := makeGetUploadsOptions(args)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := resolvers.NewPrefetcher(r.autoindexer, r.svc)
	// uploadConnectionResolver := r.resolver.UploadsResolver().UploadsConnectionResolverFromFactory(opts)
	uploadsResolver := resolvers.NewUploadsResolver(r.svc, opts)

	return resolvers.NewUploadConnectionResolver(r.svc, r.autoindexer, r.policyResolver, uploadsResolver, prefetcher, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteLsifUpload.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("uploadID", string(args.ID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexer.GetUnsafeDB()); err != nil {
		return nil, err
	}

	uploadID, err := unmarshalLSIFUploadGQLID(args.ID)
	if err != nil {
		return nil, err
	}

	if _, err := r.svc.DeleteUploadByID(ctx, int(uploadID)); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
