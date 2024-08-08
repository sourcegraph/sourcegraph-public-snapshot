package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type batchSpecWorkspaceConnectionResolver struct {
	store  *store.Store
	logger log.Logger
	opts   store.ListBatchSpecWorkspacesOpts

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

	batchSpec, err := r.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: r.opts.BatchSpecID})
	if err != nil {
		return nil, err
	}

	repoIDs := make([]api.RepoID, len(nodes))
	for _, w := range nodes {
		repoIDs = append(repoIDs, w.RepoID)
	}
	repos, err := r.store.Repos().GetReposSetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.BatchSpecWorkspaceResolver, 0, len(nodes))
	for _, w := range nodes {
		res := newBatchSpecWorkspaceResolverWithRepo(r.store, r.logger, w, executionsByWorkspaceID[w.ID], batchSpec.Spec, repos[w.RepoID])
		resolvers = append(resolvers, res)
	}

	return resolvers, nil
}

func (r *batchSpecWorkspaceConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBatchSpecWorkspaces(ctx, r.opts)
	return int32(count), err
}

func (r *batchSpecWorkspaceConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return gqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return gqlutil.HasNextPage(false), nil
}

func (r *batchSpecWorkspaceConnectionResolver) compute(ctx context.Context) ([]*btypes.BatchSpecWorkspace, int64, error) {
	r.once.Do(func() {
		r.workspaces, r.next, r.err = r.store.ListBatchSpecWorkspaces(ctx, r.opts)
	})
	return r.workspaces, r.next, r.err
}

func (r *batchSpecWorkspaceConnectionResolver) Stats(ctx context.Context) (graphqlbackend.BatchSpecWorkspacesStatsResolver, error) {
	stats, err := r.store.GetBatchSpecStats(ctx, []int64{r.opts.BatchSpecID})
	if err != nil {
		return nil, err
	}
	stat, ok := stats[r.opts.BatchSpecID]
	if !ok {
		return nil, errors.New("stats not found")
	}
	return &batchSpecWorkspacesStatsResolver{
		errored: int32(stat.Failed),
		// We count cached workspaces as completed as well.
		completed:  int32(stat.Completed + stat.CachedWorkspaces),
		processing: int32(stat.Processing),
		queued:     int32(stat.Queued),
		// TODO: Handle more ignored cases here.
		// Cached workspaces should not be considered ignored, although they
		// were skipped for execution.
		ignored: int32(stat.SkippedWorkspaces - stat.CachedWorkspaces),
	}, nil
}
