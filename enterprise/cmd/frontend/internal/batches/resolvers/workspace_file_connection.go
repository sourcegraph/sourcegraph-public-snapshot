package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

var _ graphqlbackend.WorkspaceFileConnectionResolver = &workspaceFileConnectionResolver{}

type workspaceFileConnectionResolver struct {
	store *store.Store
	opts  store.ListBatchSpecWorkspaceFileOpts

	// Cache results to save on hit to the database.
	once  sync.Once
	files []*btypes.BatchSpecWorkspaceFile
	next  int64
	err   error
}

func (r *workspaceFileConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBatchSpecWorkspaceFiles(ctx, r.opts)
	return int32(count), err
}

func (r *workspaceFileConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *workspaceFileConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.BatchWorkspaceFileResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return []graphqlbackend.BatchWorkspaceFileResolver{}, nil
	}

	resolvers := make([]graphqlbackend.BatchWorkspaceFileResolver, len(nodes))
	for i, node := range nodes {
		resolvers[i] = &workspaceFileResolver{
			batchSpecRandID: r.opts.BatchSpecRandID,
			file:            node,
		}
	}
	return resolvers, nil
}

func (r *workspaceFileConnectionResolver) compute(ctx context.Context) ([]*btypes.BatchSpecWorkspaceFile, int64, error) {
	r.once.Do(func() {
		r.files, r.next, r.err = r.store.ListBatchSpecWorkspaceFiles(ctx, r.opts)
	})
	return r.files, r.next, r.err
}
