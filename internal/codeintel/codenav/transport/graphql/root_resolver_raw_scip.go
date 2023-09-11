package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	uploadgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (r *gitBlobLSIFDataResolver) Snapshot(ctx context.Context, args *struct{ IndexID graphql.ID }) (resolvers *[]resolverstubs.SnapshotDataResolver, err error) {
	uploadID, _, err := uploadgraphql.UnmarshalPreciseIndexGQLID(args.IndexID)
	if err != nil {
		return nil, err
	}

	ctx, _, endObservation := r.operations.snapshot.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	data, err := r.codeNavSvc.SnapshotForDocument(ctx, r.requestState.RepositoryID, r.requestState.Commit, r.requestState.Path, uploadID)
	if err != nil {
		return nil, err
	}

	resolvers = new([]resolverstubs.SnapshotDataResolver)
	for _, d := range data {
		*resolvers = append(*resolvers, &snapshotDataResolver{
			data: d,
		})
	}
	return
}

type snapshotDataResolver struct {
	data shared.SnapshotData
}

func (r *snapshotDataResolver) Offset() int32 {
	return int32(r.data.DocumentOffset)
}

func (r *snapshotDataResolver) Data() string {
	return r.data.Symbol
}

func (r *snapshotDataResolver) Additional() *[]string {
	return &r.data.AdditionalData
}
