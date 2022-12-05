package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	uploadSvc    UploadService
	autoindexSvc AutoIndexingService
	policySvc    PolicyService
	operations   *operations
}

func NewRootResolver(observationCtx *observation.Context, uploadSvc UploadService, autoindexSvc AutoIndexingService, policySvc PolicyService) resolverstubs.UploadsServiceResolver {
	return &rootResolver{
		uploadSvc:    uploadSvc,
		autoindexSvc: autoindexSvc,
		policySvc:    policySvc,
		operations:   newOperations(observationCtx),
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

// 🚨 SECURITY: dbstore layer handles authz for GetUploadByID
func (r *rootResolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ resolverstubs.LSIFUploadResolver, err error) {
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
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)

	upload, exists, err := prefetcher.GetUploadByID(ctx, int(uploadID))
	if err != nil || !exists {
		return nil, err
	}

	return sharedresolvers.NewUploadResolver(r.uploadSvc, r.autoindexSvc, r.policySvc, upload, prefetcher, traceErrs), nil
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploads
func (r *rootResolver) LSIFUploads(ctx context.Context, args *resolverstubs.LSIFUploadsQueryArgs) (_ resolverstubs.LSIFUploadConnectionResolver, err error) {
	// Delegate behavior to LSIFUploadsByRepo with no specified repository identifier
	return r.LSIFUploadsByRepo(ctx, &resolverstubs.LSIFRepositoryUploadsQueryArgs{LSIFUploadsQueryArgs: args})
}

func (r *rootResolver) LSIFUploadsByRepo(ctx context.Context, args *resolverstubs.LSIFRepositoryUploadsQueryArgs) (_ resolverstubs.LSIFUploadConnectionResolver, err error) {
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
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)
	uploadsResolver := sharedresolvers.NewUploadsResolver(r.uploadSvc, opts)

	return sharedresolvers.NewUploadConnectionResolver(r.uploadSvc, r.autoindexSvc, r.policySvc, uploadsResolver, prefetcher, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteLsifUpload.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("uploadID", string(args.ID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}

	uploadID, err := unmarshalLSIFUploadGQLID(args.ID)
	if err != nil {
		return nil, err
	}

	if _, err := r.uploadSvc.DeleteUploadByID(ctx, int(uploadID)); err != nil {
		return nil, err
	}

	return &resolverstubs.EmptyResponse{}, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) DeleteLSIFUploads(ctx context.Context, args *resolverstubs.DeleteLSIFUploadsArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteLsifUploads.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}

	opts, err := makeDeleteUploadsOptions(args)
	if err != nil {
		return nil, err
	}
	if err := r.uploadSvc.DeleteUploads(ctx, opts); err != nil {
		return nil, err
	}

	return &resolverstubs.EmptyResponse{}, nil
}
