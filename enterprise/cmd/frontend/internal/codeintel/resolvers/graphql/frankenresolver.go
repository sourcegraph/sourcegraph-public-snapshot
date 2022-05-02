package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type frankenResolver struct {
	*Resolver
	gql.UploadsServiceResolver
}

func (r *frankenResolver) getUploadsServiceResolver() gql.UploadsServiceResolver {
	return r.Resolver

	// Uncomment after https://github.com/sourcegraph/sourcegraph/issues/33375
	// return r.UploadsServiceResolver
}

func (r *frankenResolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ gql.LSIFUploadResolver, err error) {
	return r.getUploadsServiceResolver().LSIFUploadByID(ctx, id)
}

func (r *frankenResolver) LSIFUploads(ctx context.Context, args *gql.LSIFUploadsQueryArgs) (_ gql.LSIFUploadConnectionResolver, err error) {
	return r.getUploadsServiceResolver().LSIFUploads(ctx, args)
}

func (r *frankenResolver) LSIFUploadsByRepo(ctx context.Context, args *gql.LSIFRepositoryUploadsQueryArgs) (_ gql.LSIFUploadConnectionResolver, err error) {
	return r.getUploadsServiceResolver().LSIFUploadsByRepo(ctx, args)
}

func (r *frankenResolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *gql.EmptyResponse, err error) {
	return r.getUploadsServiceResolver().DeleteLSIFUpload(ctx, args)
}

func (r *frankenResolver) CommitGraph(ctx context.Context, id graphql.ID) (_ gql.CodeIntelligenceCommitGraphResolver, err error) {
	return r.getUploadsServiceResolver().CommitGraph(ctx, id)
}
