package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

var _ graphqlbackend.BatchSpecWorkspaceFileConnectionResolver = &batchSpecWorkspaceFileConnectionResolver{}

type batchSpecWorkspaceFileConnectionResolver struct {
	store *store.Store
	opts  store.ListBatchSpecWorkspaceFileOpts

	// Cache results to save on hit to the database.
	once  sync.Once
	files []*btypes.BatchSpecWorkspaceFile
	next  int64
	err   error
}

func (r *batchSpecWorkspaceFileConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBatchSpecWorkspaceFiles(ctx, r.opts)
	return int32(count), err
}

func (r *batchSpecWorkspaceFileConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return gqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return gqlutil.HasNextPage(false), nil
}

func (r *batchSpecWorkspaceFileConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.BatchWorkspaceFileResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return []graphqlbackend.BatchWorkspaceFileResolver{}, nil
	}

	resolvers := make([]graphqlbackend.BatchWorkspaceFileResolver, len(nodes))
	for i, node := range nodes {
		resolvers[i] = newBatchSpecWorkspaceFileResolver(r.opts.BatchSpecRandID, node)
	}
	return resolvers, nil
}

func (r *batchSpecWorkspaceFileConnectionResolver) compute(ctx context.Context) ([]*btypes.BatchSpecWorkspaceFile, int64, error) {
	r.once.Do(func() {
		r.files, r.next, r.err = r.store.ListBatchSpecWorkspaceFiles(ctx, r.opts)
	})
	return r.files, r.next, r.err
}
