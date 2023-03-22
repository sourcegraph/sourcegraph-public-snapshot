package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	uploadSvc  UploadsService
	operations *operations
}

func NewRootResolver(observationCtx *observation.Context, uploadSvc UploadsService) resolverstubs.UploadsServiceResolver {
	return &rootResolver{
		uploadSvc:  uploadSvc,
		operations: newOperations(observationCtx),
	}
}

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *rootResolver) CommitGraph(ctx context.Context, repoID graphql.ID) (_ resolverstubs.CodeIntelligenceCommitGraphResolver, err error) {
	ctx, _, endObservation := r.operations.commitGraph.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(repoID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	repositoryID, err := resolverstubs.UnmarshalID[int](repoID)
	if err != nil {
		return nil, err
	}

	stale, updatedAt, err := r.uploadSvc.GetCommitGraphMetadata(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return NewCommitGraphResolver(stale, updatedAt), nil
}
