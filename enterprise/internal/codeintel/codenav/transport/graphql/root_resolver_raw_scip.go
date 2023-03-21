package graphql

import (
	"context"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

func (r *gitBlobLSIFDataResolver) Snapshot(ctx context.Context) (resolvers []resolverstubs.SnapshotDataResolver, err error) {
	data, err := r.codeNavSvc.SnapshotForDocument(ctx, r.requestState.RepositoryID, r.requestState.Commit, r.requestState.Path)
	if err != nil {
		return nil, err
	}

	for _, d := range data {
		resolvers = append(resolvers, &snapshotDataResolver{
			data: d,
		})
	}
	return
}
