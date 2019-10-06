package diagnostics

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (GraphQLResolver) ThreadDiagnostics(ctx context.Context, arg *graphqlbackend.ThreadDiagnosticConnectionArgs) (graphqlbackend.ThreadDiagnosticConnection, error) {
	opt, err := threadDiagnosticConnectionArgsToListOptions(arg)
	if err != nil {
		return nil, err
	}
	return &threadDiagnosticConnection{opt: *opt}, nil
}

func threadDiagnosticConnectionArgsToListOptions(arg *graphqlbackend.ThreadDiagnosticConnectionArgs) (*dbThreadDiagnosticEdgesListOptions, error) {
	var opt dbThreadDiagnosticEdgesListOptions
	arg.Set(&opt.LimitOffset)
	if arg.Thread != nil {
		var err error
		opt.ThreadID, err = graphqlbackend.UnmarshalThreadID(*arg.Thread)
		if err != nil {
			return nil, err
		}
	}
	if arg.Campaign != nil {
		var err error
		opt.CampaignID, err = graphqlbackend.UnmarshalCampaignID(*arg.Campaign)
		if err != nil {
			return nil, err
		}
	}
	return &opt, nil
}

type threadDiagnosticConnection struct {
	opt dbThreadDiagnosticEdgesListOptions

	once              sync.Once
	threadDiagnostics []*dbThreadDiagnostic
	err               error
}

func (r *threadDiagnosticConnection) compute(ctx context.Context) ([]*dbThreadDiagnostic, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.threadDiagnostics, r.err = dbThreadDiagnosticEdges{}.List(ctx, opt2)
	})
	return r.threadDiagnostics, r.err
}

func (r *threadDiagnosticConnection) Edges(ctx context.Context) ([]graphqlbackend.ThreadDiagnosticEdge, error) {
	dbThreadDiagnosticEdges, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.opt.LimitOffset != nil && len(dbThreadDiagnosticEdges) > r.opt.LimitOffset.Limit {
		dbThreadDiagnosticEdges = dbThreadDiagnosticEdges[:r.opt.LimitOffset.Limit]
	}

	edges := make([]graphqlbackend.ThreadDiagnosticEdge, len(dbThreadDiagnosticEdges))
	for i, dbThreadDiagnostic := range dbThreadDiagnosticEdges {
		edges[i] = &gqlThreadDiagnosticEdge{db: dbThreadDiagnostic}
	}
	return edges, nil
}

func (r *threadDiagnosticConnection) Nodes(ctx context.Context) ([]graphqlbackend.Diagnostic, error) {
	edges, err := r.Edges(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]graphqlbackend.Diagnostic, len(edges))
	for i, edge := range edges {
		var err error
		nodes[i], err = edge.Diagnostic()
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (r *threadDiagnosticConnection) TotalCount(ctx context.Context) (int32, error) {
	count, err := dbThreadDiagnosticEdges{}.Count(ctx, r.opt)
	return int32(count), err
}

func (r *threadDiagnosticConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	threads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(threads) > r.opt.Limit), nil
}
