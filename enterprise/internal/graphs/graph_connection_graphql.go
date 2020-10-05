package graphs

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/graphs"
)

var _ graphqlbackend.GraphConnectionResolver = &graphConnectionResolver{}

type graphConnectionResolver struct {
	store *Store
	opts  ListGraphsOpts

	// cache results because they are used by multiple fields
	once   sync.Once
	graphs []*graphs.Graph
	next   int64
	err    error
}

func (r *graphConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.GraphResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.GraphResolver, 0, len(nodes))
	for _, g := range nodes {
		resolvers = append(resolvers, &graphResolver{Graph: g})
	}
	return resolvers, nil
}

func (r *graphConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := CountGraphsOpts{
		OwnerUserID: r.opts.OwnerUserID,
		OwnerOrgID:  r.opts.OwnerOrgID,
	}
	count, err := r.store.CountGraphs(ctx, opts)
	return int32(count), err
}

func (r *graphConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *graphConnectionResolver) compute(ctx context.Context) ([]*graphs.Graph, int64, error) {
	r.once.Do(func() {
		r.graphs, r.next, r.err = r.store.ListGraphs(ctx, r.opts)
	})
	return r.graphs, r.next, r.err
}
