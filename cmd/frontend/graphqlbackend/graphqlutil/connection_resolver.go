package graphqlutil

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DEFAULT_MAX_PAGE_SIZE = 100

type ConnectionResolver[N ConnectionNode] struct {
	store   ConnectionResolverStore[N]
	args    *ConnectionResolverArgs
	options *ConnectionResolverOptions
	data    connectionData[N]
	once    resolveOnce
}

type ConnectionNode interface {
	ID() graphql.ID
}

type ConnectionResolverStore[N ConnectionNode] interface {
	// ComputeTotal returns the total count of all the items in the connection, independent of pagination arguments.
	ComputeTotal(context.Context) (*int32, error)
	// ComputeNodes returns the list of nodes based on the pagination args.
	ComputeNodes(context.Context, *database.PaginationArgs) ([]*N, error)
	// MarshalCursor returns cursor for a node and is called for generating start and end cursors.
	MarshalCursor(*N) (*string, error)
	// UnmarshalCursor returns node id from after/before cursor string.
	UnmarshalCursor(string) (*int, error)
}

type ConnectionResolverArgs struct {
	First  *int32
	Last   *int32
	After  *string
	Before *string
}

// Limit returns max nodes limit based on resolver arguments
func (a *ConnectionResolverArgs) Limit(options *ConnectionResolverOptions) int {
	var limit *int32

	if a.First != nil {
		limit = a.First
	} else {
		limit = a.Last
	}

	return options.ApplyMaxPageSize(limit)
}

type ConnectionResolverOptions struct {
	maxPageSize *int
}

// MaxPageSize returns the configured max page limit for the connection
func (o *ConnectionResolverOptions) MaxPageSize() int {
	if o.maxPageSize != nil {
		return *o.maxPageSize
	}

	return DEFAULT_MAX_PAGE_SIZE
}

// ApplyMaxPageSize return max page size by applying the configured max limit to the first, last arguments
func (o *ConnectionResolverOptions) ApplyMaxPageSize(limit *int32) int {
	maxPageSize := o.MaxPageSize()

	if limit == nil {
		return maxPageSize
	}

	if int(*limit) < maxPageSize {
		return int(*limit)
	}

	return maxPageSize
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

	limit := r.pageSize() + 1
	if r.args.First != nil {
		paginationArgs.First = &limit
	} else if r.args.Last != nil {
		paginationArgs.Last = &limit
	} else {
		return nil, errors.New("you must provide a `first` or `last` value to properly paginate")
	}

	if r.args.After != nil {
		after, err := r.store.UnmarshalCursor(*r.args.After)
		if err != nil {
			return nil, err
		}

		paginationArgs.After = after
	}

	if r.args.Before != nil {
		before, err := r.store.UnmarshalCursor(*r.args.Before)
		if err != nil {
			return nil, err
		}

		paginationArgs.Before = before
	}

	return &paginationArgs, nil
}

func (r *ConnectionResolver[N]) pageSize() int {
	return r.args.Limit(r.options)
}

// TotalCount returns value for connection.totalCount and is called by the graphql api.
func (r *ConnectionResolver[N]) TotalCount(ctx context.Context) (int32, error) {
	r.once.total.Do(func() {
		r.data.total, r.data.totalError = r.store.ComputeTotal(ctx)
	})

	if r.data.totalError != nil || r.data.total == nil {
		return 0, r.data.totalError
	}

	return *r.data.total, r.data.totalError
}

// Nodes returns value for connection.Nodes and is called by the graphql api.
func (r *ConnectionResolver[N]) Nodes(ctx context.Context) ([]*N, error) {
	r.once.nodes.Do(func() {
		paginationArgs, err := r.paginationArgs()
		if err != nil {
			r.data.nodesError = err
			return
		}

		r.data.nodes, r.data.nodesError = r.store.ComputeNodes(ctx, paginationArgs)

		// NOTE(naman): with `last` argument the items are sorted in opposite
		// direction in the SQL query. Here we are reversing the list to return
		// them in correct order, to reduce complexity.
		if r.args.Last != nil {
			for i, j := 0, len(r.data.nodes)-1; i < j; i, j = i+1, j-1 {
				r.data.nodes[i], r.data.nodes[j] = r.data.nodes[j], r.data.nodes[i]
			}
		}
	})

	nodes := r.data.nodes

	// NOTE(naman): we pass actual_limit + 1 to SQL query so that we
	// can check for `hasNextPage`. Here we need to remove the extra item,
	// last item in case of `first` and first item in case of `last` as
	// they are sorted in opposite directions in SQL query.
	if len(nodes) > r.pageSize() {
		if r.args.Last != nil {
			nodes = nodes[1:]
		} else {
			nodes = nodes[:len(nodes)-1]
		}
	}

	return nodes, r.data.nodesError
}

// PageInfo returns value for connection.pageInfo and is called by the graphql api.
func (r *ConnectionResolver[N]) PageInfo(ctx context.Context) (*ConnectionPageInfo[N], error) {
	nodes, err := r.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	return &ConnectionPageInfo[N]{
		pageSize:          r.pageSize(),
		fetchedNodesCount: len(r.data.nodes),
		nodes:             nodes,
		store:             r.store,
		args:              r.args,
	}, nil
}

type ConnectionPageInfo[N ConnectionNode] struct {
	pageSize          int
	fetchedNodesCount int
	nodes             []*N
	store             ConnectionResolverStore[N]
	args              *ConnectionResolverArgs
}

// HasNextPage returns value for connection.pageInfo.hasNextPage and is called by the graphql api.
func (p *ConnectionPageInfo[N]) HasNextPage() bool {
	if p.args.First != nil {
		return p.fetchedNodesCount > p.pageSize
	}

	if p.fetchedNodesCount == 0 {
		return false
	}

	return p.args.Before != nil
}

// HasPreviousPage returns value for connection.pageInfo.hasPreviousPage and is called by the graphql api.
func (p *ConnectionPageInfo[N]) HasPreviousPage() bool {
	if p.args.Last != nil {
		return p.fetchedNodesCount > p.pageSize
	}

	if p.fetchedNodesCount == 0 {
		return false
	}

	return p.args.After != nil
}

// EndCursor returns value for connection.pageInfo.endCursor and is called by the graphql api.
func (p *ConnectionPageInfo[N]) EndCursor() (cursor *string, err error) {
	if len(p.nodes) == 0 {
		return nil, nil
	}

	cursor, err = p.store.MarshalCursor(p.nodes[len(p.nodes)-1])

	return
}

// StartCursor returns value for connection.pageInfo.startCursor and is called by the graphql api.
func (p *ConnectionPageInfo[N]) StartCursor() (cursor *string, err error) {
	if len(p.nodes) == 0 {
		return nil, nil
	}

	cursor, err = p.store.MarshalCursor(p.nodes[0])

	return
}

// NewConnectionResolver returns a new connection resolver built using the store and connection args.
func NewConnectionResolver[N ConnectionNode](store ConnectionResolverStore[N], args *ConnectionResolverArgs, options *ConnectionResolverOptions) (*ConnectionResolver[N], error) {
	if args == nil || (args.First == nil && args.Last == nil) {
		return nil, errors.New("you must provide a `first` or `last` value to properly paginate")
	}

	if options == nil {
		options = &ConnectionResolverOptions{}
	}

	return &ConnectionResolver[N]{
		store:   store,
		args:    args,
		options: options,
		data:    connectionData[N]{},
	}, nil
}
