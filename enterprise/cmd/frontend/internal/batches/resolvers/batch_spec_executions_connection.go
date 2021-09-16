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

var _ graphqlbackend.BatchSpecExecutionConnectionResolver = &batchSpecExecutionConnectionResolver{}

type batchSpecExecutionConnectionResolver struct {
	store *store.Store
	opts  store.ListBatchSpecExecutionsOpts

	once                sync.Once
	batchSpecExecutions []*btypes.BatchSpecExecution
	next                int64
	err                 error
}

func (r *batchSpecExecutionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.BatchSpecExecutionResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.BatchSpecExecutionResolver, 0, len(nodes))
	for _, c := range nodes {
		resolvers = append(resolvers, &batchSpecExecutionResolver{store: r.store, exec: c})
	}
	return resolvers, nil
}

func (r *batchSpecExecutionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := store.CountBatchSpecExecutionsOpts{}
	count, err := r.store.CountBatchSpecExecutions(ctx, opts)
	return int32(count), err
}

func (r *batchSpecExecutionConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *batchSpecExecutionConnectionResolver) compute(ctx context.Context) ([]*btypes.BatchSpecExecution, int64, error) {
	r.once.Do(func() {
		r.batchSpecExecutions, r.next, r.err = r.store.ListBatchSpecExecutions(ctx, r.opts)
	})
	return r.batchSpecExecutions, r.next, r.err
}
