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

type batchSpecWorkspaceConnectionResolver struct {
	store *store.Store
	opts  store.ListBatchSpecWorkspacesOpts

	// Cache results because they are used by multiple fields.
	once       sync.Once
	workspaces []*btypes.BatchSpecWorkspace
	next       int64
	err        error
}

var _ graphqlbackend.BatchSpecWorkspaceConnectionResolver = &batchSpecWorkspaceConnectionResolver{}

func (r *batchSpecWorkspaceConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.BatchSpecWorkspaceResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return []graphqlbackend.BatchSpecWorkspaceResolver{}, nil
	}

	nodeIDs := make([]int64, 0, len(nodes))
	for _, n := range nodes {
		nodeIDs = append(nodeIDs, n.ID)
	}
	executions, err := r.store.ListBatchSpecWorkspaceExecutionJobs(ctx, store.ListBatchSpecWorkspaceExecutionJobsOpts{BatchSpecWorkspaceIDs: nodeIDs})
	if err != nil {
		return nil, err
	}
	executionsByWorkspaceID := make(map[int64]*btypes.BatchSpecWorkspaceExecutionJob)
	for _, e := range executions {
		executionsByWorkspaceID[e.BatchSpecWorkspaceID] = e
	}
	resolvers := make([]graphqlbackend.BatchSpecWorkspaceResolver, 0, len(nodes))
	for _, w := range nodes {
		res := &batchSpecWorkspaceResolver{
			store:     r.store,
			workspace: w,
		}
		if ex, ok := executionsByWorkspaceID[w.ID]; ok {
			res.execution = ex
		}
		resolvers = append(resolvers, res)
	}
	return resolvers, nil
}

func (r *batchSpecWorkspaceConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// TODO(ssbc): not implemented
	return 0, nil
}

func (r *batchSpecWorkspaceConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *batchSpecWorkspaceConnectionResolver) compute(ctx context.Context) ([]*btypes.BatchSpecWorkspace, int64, error) {
	r.once.Do(func() {
		r.workspaces, r.next, r.err = r.store.ListBatchSpecWorkspaces(ctx, r.opts)
	})
	return r.workspaces, r.next, r.err
}

func (r *batchSpecWorkspaceConnectionResolver) Stats(ctx context.Context) graphqlbackend.BatchSpecWorkspacesStatsResolver {
	// TODO(ssbc): not implemented
	return nil
}
