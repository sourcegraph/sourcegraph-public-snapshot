package graphql

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	autoindexing "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver struct {
	svc        *autoindexing.Service
	operations *operations
}

func newResolver(svc *autoindexing.Service, observationContext *observation.Context) *Resolver {
	return &Resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (r *Resolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (_ gql.LSIFIndexResolver, err error) {
	ctx, _, endObservation := r.operations.lsifIndexByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_, _ = ctx, id
	return nil, errors.New("unimplemented: LSIFIndexByID")
}

func (r *Resolver) LSIFIndexes(ctx context.Context, args *gql.LSIFIndexesQueryArgs) (_ gql.LSIFIndexConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.lsifIndexes.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_, _ = ctx, args
	return nil, errors.New("unimplemented: LSIFIndexes")
}

func (r *Resolver) LSIFIndexesByRepo(ctx context.Context, args *gql.LSIFRepositoryIndexesQueryArgs) (_ gql.LSIFIndexConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.lsifIndexesByRepo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_, _ = ctx, args
	return nil, errors.New("unimplemented: LSIFIndexesByRepo")
}

func (r *Resolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *gql.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteLSIFIndex.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_, _ = ctx, args
	return nil, errors.New("unimplemented: DeleteLSIFIndex")
}

func (r *Resolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *gql.QueueAutoIndexJobsForRepoArgs) (_ []gql.LSIFIndexResolver, err error) {
	ctx, _, endObservation := r.operations.queueAutoIndexJobsForRepo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_, _ = ctx, args
	return nil, errors.New("unimplemented: QueueAutoIndexJobsForRepo")
}

func (r *Resolver) IndexConfiguration(ctx context.Context, id graphql.ID) (_ gql.IndexConfigurationResolver, err error) {
	ctx, _, endObservation := r.operations.indexConfiguration.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_, _ = ctx, id
	return nil, errors.New("unimplemented: IndexConfiguration")
}

func (r *Resolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *gql.UpdateRepositoryIndexConfigurationArgs) (_ *gql.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateRepositoryIndexConfiguration.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_, _ = ctx, args
	return nil, errors.New("unimplemented: UpdateRepositoryIndexConfiguration")
}
