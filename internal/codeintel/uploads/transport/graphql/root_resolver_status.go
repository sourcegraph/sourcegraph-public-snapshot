package graphql

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"go.opentelemetry.io/otel/attribute"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *rootResolver) CommitGraph(ctx context.Context, repoID graphql.ID) (_ resolverstubs.CodeIntelligenceCommitGraphResolver, err error) {
	ctx, _, endObservation := r.operations.commitGraph.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("repoID", string(repoID)),
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

	return newCommitGraphResolver(stale, updatedAt), nil
}

type commitGraphResolver struct {
	stale     bool
	updatedAt *time.Time
}

func newCommitGraphResolver(stale bool, updatedAt *time.Time) resolverstubs.CodeIntelligenceCommitGraphResolver {
	return &commitGraphResolver{
		stale:     stale,
		updatedAt: updatedAt,
	}
}

func (r *commitGraphResolver) Stale() bool {
	return r.stale
}

func (r *commitGraphResolver) UpdatedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.updatedAt)
}
