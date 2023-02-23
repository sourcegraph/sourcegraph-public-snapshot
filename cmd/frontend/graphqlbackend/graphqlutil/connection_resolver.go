package graphqlutil

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DefaultMaxPageSize = 100

type ConnectionResolver[N any] struct {
	store   ConnectionResolverStore[N]
	args    *ConnectionResolverArgs
	options *ConnectionResolverOptions
	data    connectionData[N]
	once    resolveOnce
}

type ConnectionResolverStore[N any] interface {
	// ComputeTotal returns the total count of all the items in the connection, independent of pagination arguments.
	ComputeTotal(context.Context) (*int32, error)
	// ComputeNodes returns the list of nodes based on the pagination args.
	ComputeNodes(context.Context, *database.PaginationArgs) ([]N, error)
	// MarshalCursor returns cursor for a node and is called for generating start and end cursors.
	MarshalCursor(N, database.OrderBy) (*string, error)
	// UnmarshalCursor returns node id from after/before cursor string.
	UnmarshalCursor(string, database.OrderBy) (*string, error)
}

type ConnectionResolverArgs struct {
	First  *int32
	Last   *int32
	After  *string
	Before *string
}

// Limit returns max nodes limit based on resolver arguments.
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
	// The maximum number of nodes that can be returned in a single page.
	MaxPageSize *int
	// Used to enable or disable the automatic reversal of nodes in backward
	// pagination mode.
	//
	// Setting this to `false` is useful when the data is not fetched via a SQL
	// index.
	//
	// Defaults to `true` when not set.
	Reverse *bool
	// Columns to order by.
	OrderBy database.OrderBy
	// Order direction.
	Ascending bool

	// If set to true, the resolver won't throw an error when `first` or `last` isn't provided
	// in `ConnectionResolverArgs`. Be careful when setting this to true, as this could cause
	// performance issues when fetching large data.
	AllowNoLimit bool
}

// MaxPageSizeOrDefault returns the configured max page limit for the connection.
func (o *ConnectionResolverOptions) MaxPageSizeOrDefault() int {
	if o.MaxPageSize != nil {
		return *o.MaxPageSize
	}

	return DefaultMaxPageSize
}

// ApplyMaxPageSize return max page size by applying the configured max limit to the first, last arguments.
func (o *ConnectionResolverOptions) ApplyMaxPageSize(limit *int32) int {
	maxPageSize := o.MaxPageSizeOrDefault()

	if limit == nil {
		return maxPageSize
	}

	if int(*limit) < maxPageSize {
		return int(*limit)
	}

	return maxPageSize
}

type connectionData[N any] struct {
	total      *int32
	totalError error

	nodes      []N
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

	paginationArgs := database.PaginationArgs{
		OrderBy:   r.options.OrderBy,
		Ascending: r.options.Ascending,
	}

	limit := r.pageSize() + 1
	if r.args.First != nil {
		paginationArgs.First = &limit
	} else if r.args.Last != nil {
		paginationArgs.Last = &limit
	} else if !r.options.AllowNoLimit {
		return nil, errors.New("you must provide a `first` or `last` value to properly paginate")
	}

	if r.args.After != nil {
		after, err := r.store.UnmarshalCursor(*r.args.After, r.options.OrderBy)
		if err != nil {
			return nil, err
		}

		paginationArgs.After = after
	}

	if r.args.Before != nil {
		before, err := r.store.UnmarshalCursor(*r.args.Before, r.options.OrderBy)
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

	if r.data.total != nil {
		return *r.data.total, r.data.totalError
	}

	return 0, r.data.totalError
}

// Nodes returns value for connection.Nodes and is called by the graphql api.
func (r *ConnectionResolver[N]) Nodes(ctx context.Context) ([]N, error) {
	r.once.nodes.Do(func() {
		paginationArgs, err := r.paginationArgs()
		if err != nil {
			r.data.nodesError = err
			return
		}

		r.data.nodes, r.data.nodesError = r.store.ComputeNodes(ctx, paginationArgs)

		if r.options.Reverse != nil && !*r.options.Reverse {
			return
		}

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
		orderBy:           r.options.OrderBy,
	}, nil
}

type ConnectionPageInfo[N any] struct {
	pageSize          int
	fetchedNodesCount int
	nodes             []N
	store             ConnectionResolverStore[N]
	args              *ConnectionResolverArgs
	orderBy           database.OrderBy
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

	cursor, err = p.store.MarshalCursor(p.nodes[len(p.nodes)-1], p.orderBy)

	return
}

// StartCursor returns value for connection.pageInfo.startCursor and is called by the graphql api.
func (p *ConnectionPageInfo[N]) StartCursor() (cursor *string, err error) {
	if len(p.nodes) == 0 {
		return nil, nil
	}

	cursor, err = p.store.MarshalCursor(p.nodes[0], p.orderBy)

	return
}

// NewConnectionResolver returns a new connection resolver built using the store and connection args.
func NewConnectionResolver[N any](store ConnectionResolverStore[N], args *ConnectionResolverArgs, options *ConnectionResolverOptions) (*ConnectionResolver[N], error) {
	if options == nil {
		options = &ConnectionResolverOptions{OrderBy: database.OrderBy{{Field: "id"}}}
	}

	if len(options.OrderBy) == 0 {
		options.OrderBy = database.OrderBy{{Field: "id"}}
	}

	return &ConnectionResolver[N]{
		store:   store,
		args:    args,
		options: options,
		data:    connectionData[N]{},
	}, nil
}
