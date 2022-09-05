package graphql

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver struct {
	svc        *uploads.Service
	operations *operations
}

func newResolver(svc *uploads.Service, observationContext *observation.Context) *Resolver {
	return &Resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ gql.LSIFUploadResolver, err error) {
	ctx, _, endObservation := r.operations.lsifUploadByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_, _ = ctx, id
	return nil, errors.New("unimplemented: LSIFUploadByID")
}

func (r *Resolver) LSIFUploads(ctx context.Context, args *gql.LSIFUploadsQueryArgs) (_ gql.LSIFUploadConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.lsifUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_, _ = ctx, args
	return nil, errors.New("unimplemented: LSIFUploads")
}

func (r *Resolver) LSIFUploadsByRepo(ctx context.Context, args *gql.LSIFRepositoryUploadsQueryArgs) (_ gql.LSIFUploadConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.lsifUploadsByRepo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_, _ = ctx, args
	return nil, errors.New("unimplemented: LSIFUploadsByRepo")
}

func (r *Resolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *gql.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteLSIFUpload.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_, _ = ctx, args
	return nil, errors.New("unimplemented: DeleteLSIFUpload")
}

func (r *Resolver) CommitGraph(ctx context.Context, id graphql.ID) (_ gql.CodeIntelligenceCommitGraphResolver, err error) {
	ctx, _, endObservation := r.operations.commitGraph.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_, _ = ctx, id
	return nil, errors.New("unimplemented: CommitGraph")
}
