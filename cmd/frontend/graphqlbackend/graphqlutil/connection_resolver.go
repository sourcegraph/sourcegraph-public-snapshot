package graphqlutil

import (
	"sync"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type ConnectionResolverArgs struct {
	First  *int32
	Last   *int32
	After  *string
	Before *string
}

func (a *ConnectionResolverArgs) ToPaginationArgs() *database.PaginationArgs {
	if a == nil {
		return nil
	}

	paginationArgs := database.PaginationArgs(*a)

	if paginationArgs.First != nil {
		limit := *paginationArgs.First + 1
		paginationArgs.First = &limit
	} else if paginationArgs.Last != nil {
		limit := *paginationArgs.Last + 1
		paginationArgs.Last = &limit
	}

	return &paginationArgs
}

func (a *ConnectionResolverArgs) Limit() (limit int32) {
	if a == nil {
		return 0
	}

	if a.First != nil {
		limit = *a.First
	} else if a.Last != nil {
		limit = *a.Last
	}

	return
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

type ConnectionResolverStore[N ConnectionNode] interface {
	ComputeTotal() (*int32, error)
	ComputeNodes(*database.PaginationArgs) ([]*N, error)
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

func (r *ConnectionResolver[N]) TotalCount() (int32, error) {
	r.once.total.Do(func() {
		r.data.total, r.data.totalError = r.store.ComputeTotal()
	})

	if r.data.totalError != nil || r.data.total == nil {
		return 0, r.data.totalError
	}

	return *r.data.total, r.data.totalError
}

func (r *ConnectionResolver[N]) Nodes() ([]*N, error) {
	r.once.nodes.Do(func() {
		r.data.nodes, r.data.nodesError = r.store.ComputeNodes(r.args.ToPaginationArgs())
	})

	nodes := r.data.nodes
	if len(nodes) > 0 {
		nodes = nodes[:len(nodes)-1]
	}

	return nodes, r.data.totalError
}

func (r *ConnectionResolver[N]) PageInfo() (*ConnectionPageInfo[N], error) {
	_, err := r.Nodes()
	if err != nil {
		return nil, err
	}

	return &ConnectionPageInfo[N]{
		r.data.nodes,
		r.args,
	}, nil
}

type ConnectionPageInfo[N ConnectionNode] struct {
	nodes []*N
	args  *ConnectionResolverArgs
}

func (p *ConnectionPageInfo[N]) HasNextPage() bool {
	return len(p.nodes) > int(p.args.Limit())
}

func (p *ConnectionPageInfo[N]) HasPreviousPage() bool {
	return false
}

func (p *ConnectionPageInfo[N]) EndCursor() *string {
	if len(p.nodes) == 0 {
		return nil
	}

	cursor := string((*p.nodes[len(p.nodes)-1]).ID())

	return &cursor
}

func (p *ConnectionPageInfo[N]) StartCursor() *string {
	if len(p.nodes) == 0 {
		return nil
	}

	cursor := string((*p.nodes[0]).ID())

	return &cursor
}

func NewConnectionResolver[N ConnectionNode](store ConnectionResolverStore[N], connectionArgs *ConnectionResolverArgs) *ConnectionResolver[N] {
	return &ConnectionResolver[N]{
		store,
		connectionArgs,
		connectionData[N]{},
		resolveOnce{sync.Once{}, sync.Once{}},
	}
}
