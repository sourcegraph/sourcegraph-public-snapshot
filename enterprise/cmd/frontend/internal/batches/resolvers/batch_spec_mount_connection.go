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

var _ graphqlbackend.BatchSpecMountConnectionResolver = &batchSpecMountConnectionResolver{}

type batchSpecMountConnectionResolver struct {
	store *store.Store
	opts  store.ListBatchSpecMountsOpts

	// Cache results because they are used by multiple fields.
	once   sync.Once
	mounts []*btypes.BatchSpecMount
	next   int64
	err    error
}

func (r *batchSpecMountConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBatchSpecMounts(ctx, r.opts)
	return int32(count), err
}

func (r *batchSpecMountConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *batchSpecMountConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.BatchSpecMountResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return []graphqlbackend.BatchSpecMountResolver{}, nil
	}

	resolvers := make([]graphqlbackend.BatchSpecMountResolver, len(nodes))
	for i, node := range nodes {
		resolvers[i] = &batchSpecMountResolver{
			store:           r.store,
			batchSpecRandID: r.opts.BatchSpecRandID,
			mount:           node,
		}
	}
	return resolvers, nil
}

func (r *batchSpecMountConnectionResolver) compute(ctx context.Context) ([]*btypes.BatchSpecMount, int64, error) {
	r.once.Do(func() {
		r.mounts, r.next, r.err = r.store.ListBatchSpecMounts(ctx, r.opts)
	})
	return r.mounts, r.next, r.err
}
