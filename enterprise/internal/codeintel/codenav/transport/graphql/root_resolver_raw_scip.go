package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	uploadgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/transport/graphql"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

func (r *gitBlobLSIFDataResolver) Snapshot(ctx context.Context, args *struct{ IndexID graphql.ID }) (resolvers *[]resolverstubs.SnapshotDataResolver, err error) {
	uploadID, _, err := uploadgraphql.UnmarshalPreciseIndexGQLID(args.IndexID)
	if err != nil {
		return nil, err
	}

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
