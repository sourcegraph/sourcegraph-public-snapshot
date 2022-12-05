package graphqlutil

import (
	"context"
	"math"
	"sync"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

const MAX_PAGE_SIZE int32 = 100

func applyMaxPageSize(limit int32) int {
	return int(math.Min(float64(limit), float64(MAX_PAGE_SIZE)))
}

type ConnectionResolver[N ConnectionNode] struct {
	store ConnectionResolverStore[N]
	args  *ConnectionResolverArgs
	data  connectionData[N]
	once  resolveOnce
}

type ConnectionNode interface {
	ID() graphql.ID
}

type ConnectionResolverArgs struct {
	First  *int32
	Last   *int32
	After  *string
	Before *string
}

func (a *ConnectionResolverArgs) Limit() int {
	if a == nil {
		return 0
	}

	limit := int(MAX_PAGE_SIZE)

	if a.First != nil {
		limit = applyMaxPageSize(*a.First)
	} else if a.Last != nil {
		limit = applyMaxPageSize(*a.Last)
	}

	return limit
}

type ConnectionResolverStore[N ConnectionNode] interface {
	ComputeTotal(context.Context) (*int32, error)
	ComputeNodes(context.Context, *database.PaginationArgs) ([]*N, error)
	MarshalCursor(*N) (*string, error)
	UnMarshalCursor(string) (*int32, error)
}

type connectionData[N ConnectionNode] struct {
	total      *int32
	totalError error

	nodes      []*N
	nodesError error
}

type resolveOnce struct {
	total sync.Once
	nodes sync.Once
}

func (r *ConnectionResolver[N]) paginationArgs() (*database.PaginationArgs, error) {
	if r.args == nil {
		return nil, nil
	}

	paginationArgs := database.PaginationArgs{}

	if r.args.First != nil {
		limit := int32(applyMaxPageSize(*r.args.First)) + 1
		paginationArgs.First = &limit
	} else if r.args.Last != nil {
		limit := int32(applyMaxPageSize(*r.args.Last)) + 1
		paginationArgs.Last = &limit
	} else {
		limit := MAX_PAGE_SIZE + 1
		paginationArgs.First = &limit
	}

	if r.args.After != nil {
		after, err := r.store.UnMarshalCursor(*r.args.After)
		if err != nil {
			return nil, err
		}

		paginationArgs.After = after
	}

	if r.args.Before != nil {
		before, err := r.store.UnMarshalCursor(*r.args.Before)
		if err != nil {
			return nil, err
		}

		paginationArgs.Before = before
	}

	return &paginationArgs, nil
}

func (r *ConnectionResolver[N]) TotalCount(ctx context.Context) (int32, error) {
	r.once.total.Do(func() {
		r.data.total, r.data.totalError = r.store.ComputeTotal(ctx)
	})

	if r.data.totalError != nil || r.data.total == nil {
		return 0, r.data.totalError
	}

	return *r.data.total, r.data.totalError
}

func (r *ConnectionResolver[N]) Nodes(ctx context.Context) ([]*N, error) {
	r.once.nodes.Do(func() {
		paginationArgs, err := r.paginationArgs()
		if err != nil {
			r.data.nodesError = err
			return
		}

		r.data.nodes, r.data.nodesError = r.store.ComputeNodes(ctx, paginationArgs)

		/* NOTE(naman): with `last` argument the items are sorted in opposite
		 * direction in the SQL query. Here we are reversing the list to return
		 * them in correct order, to reduce complexity */
		if r.args.Last != nil {
			for i, j := 0, len(r.data.nodes)-1; i < j; i, j = i+1, j-1 {
				r.data.nodes[i], r.data.nodes[j] = r.data.nodes[j], r.data.nodes[i]
			}
		}
	})

	nodes := r.data.nodes

	/* NOTE(naman): we pass actual_limit + 1 to SQL query so that we
	 * can check for `hasNextPage`. Here we need to remove the extra item,
	 * last item in case of `first` and first item in case of `last` as
	 * they are sorted in opposite directions in SQL query.*/
	if len(nodes) > r.args.Limit() {
		if r.args.Last != nil {
			nodes = nodes[1:]
		} else {
			nodes = nodes[:len(nodes)-1]
		}
	}

	return nodes, r.data.nodesError
}

func (r *ConnectionResolver[N]) PageInfo(ctx context.Context) (*ConnectionPageInfo[N], error) {
	nodes, err := r.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	return &ConnectionPageInfo[N]{
		len(r.data.nodes),
		nodes,
		r.store,
		r.args,
	}, nil
}

type ConnectionPageInfo[N ConnectionNode] struct {
	fetchedNodesCount int
	nodes             []*N
	store             ConnectionResolverStore[N]
	args              *ConnectionResolverArgs
}

func (p *ConnectionPageInfo[N]) HasNextPage() bool {
	if p.args.Before != nil {
		return true
	}

	return p.fetchedNodesCount > p.args.Limit()
}

func (p *ConnectionPageInfo[N]) HasPreviousPage() bool {
	if p.args.After != nil {
		return true
	}

	if p.args.Before != nil {
		return p.fetchedNodesCount > p.args.Limit()
	}

	return false
}

func (p *ConnectionPageInfo[N]) EndCursor() (cursor *string, err error) {
	if len(p.nodes) == 0 {
		return nil, nil
	}

	endNode := p.nodes[len(p.nodes)-1]

	cursor, err = p.store.MarshalCursor(endNode)

	return
}

func (p *ConnectionPageInfo[N]) StartCursor() (cursor *string, err error) {
	if len(p.nodes) == 0 {
		return nil, nil
	}

	startNode := p.nodes[0]

	cursor, err = p.store.MarshalCursor(startNode)

	return
}

func NewConnectionResolver[N ConnectionNode](store ConnectionResolverStore[N], connectionArgs *ConnectionResolverArgs) *ConnectionResolver[N] {
	return &ConnectionResolver[N]{
		store,
		connectionArgs,
		connectionData[N]{},
		resolveOnce{sync.Once{}, sync.Once{}},
	}
}
