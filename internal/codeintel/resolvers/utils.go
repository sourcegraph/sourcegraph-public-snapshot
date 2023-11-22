package resolvers

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type ConnectionResolver[T any] interface {
	Nodes(ctx context.Context) ([]T, error)
}

type connectionResolver[T any] struct {
	nodes []T
}

func NewConnectionResolver[T any](nodes []T) ConnectionResolver[T] {
	return &connectionResolver[T]{
		nodes: nodes,
	}
}

func (r *connectionResolver[T]) Nodes(ctx context.Context) ([]T, error) {
	return r.nodes, nil
}

//
//

type PagedConnectionResolver[T any] interface {
	ConnectionResolver[T]
	PageInfo() PageInfo
}

type cursorConnectionResolver[T any] struct {
	*connectionResolver[T]
	cursor string
}

func NewCursorConnectionResolver[T any](nodes []T, cursor string) PagedConnectionResolver[T] {
	return &cursorConnectionResolver[T]{
		connectionResolver: &connectionResolver[T]{
			nodes: nodes,
		},
		cursor: cursor,
	}
}

func (r *cursorConnectionResolver[T]) PageInfo() PageInfo {
	return NewPageInfoFromCursor(r.cursor)
}

//
//

type lazyConnectionResolver[T any] struct {
	resolveFunc func(ctx context.Context) ([]T, error)
	cursor      string
}

func NewLazyConnectionResolver[T any](resolveFunc func(ctx context.Context) ([]T, error), cursor string) PagedConnectionResolver[T] {
	return &lazyConnectionResolver[T]{
		resolveFunc: resolveFunc,
		cursor:      cursor,
	}
}

func (r *lazyConnectionResolver[T]) Nodes(ctx context.Context) ([]T, error) {
	return r.resolveFunc(ctx)
}

func (r *lazyConnectionResolver[T]) PageInfo() PageInfo {
	return NewPageInfoFromCursor(r.cursor)
}

//
//

type PagedConnectionWithTotalCountResolver[T any] interface {
	PagedConnectionResolver[T]
	TotalCount() *int32
}

type cursorConnectionWithTotalCountResolver[T any] struct {
	*cursorConnectionResolver[T]
	totalCount int32
}

func NewCursorWithTotalCountConnectionResolver[T any](nodes []T, cursor string, totalCount int32) PagedConnectionWithTotalCountResolver[T] {
	return &cursorConnectionWithTotalCountResolver[T]{
		cursorConnectionResolver: &cursorConnectionResolver[T]{
			connectionResolver: &connectionResolver[T]{
				nodes: nodes,
			},
			cursor: cursor,
		},
		totalCount: totalCount,
	}
}

func (r *cursorConnectionWithTotalCountResolver[T]) TotalCount() *int32 {
	return &r.totalCount
}

//
//

type totalCountConnectionResolver[T any] struct {
	*connectionResolver[T]
	offset     int32
	totalCount int32
}

func NewTotalCountConnectionResolver[T any](nodes []T, offset, totalCount int32) PagedConnectionWithTotalCountResolver[T] {
	return &totalCountConnectionResolver[T]{
		connectionResolver: &connectionResolver[T]{
			nodes: nodes,
		},
		offset:     offset,
		totalCount: totalCount,
	}
}

func (r *totalCountConnectionResolver[T]) TotalCount() *int32 {
	return &r.totalCount
}

func (r *totalCountConnectionResolver[T]) PageInfo() PageInfo {
	return NewSimplePageInfo(r.offset+int32(len(r.nodes)) < r.totalCount)
}

//
//

type PageInfo interface {
	HasNextPage() bool
	EndCursor() *string
}

type pageInfo struct {
	endCursor   *string
	hasNextPage bool
}

func NewSimplePageInfo(hasNextPage bool) PageInfo {
	return &pageInfo{
		hasNextPage: hasNextPage,
	}
}

func NewPageInfoFromCursor(endCursor string) PageInfo {
	if endCursor == "" {
		return &pageInfo{}
	}

	return &pageInfo{
		hasNextPage: true,
		endCursor:   &endCursor,
	}
}

func (r *pageInfo) EndCursor() *string { return r.endCursor }
func (r *pageInfo) HasNextPage() bool  { return r.hasNextPage }

type ConnectionArgs struct {
	First *int32
}

func (a *ConnectionArgs) Limit(defaultValue int32) int32 {
	return pointers.Deref(a.First, defaultValue)
}

type PagedConnectionArgs struct {
	ConnectionArgs
	After *string
}

func (a *PagedConnectionArgs) ParseOffset() (int32, error) {
	if a.After == nil {
		return 0, nil
	}

	v, err := strconv.ParseInt(*a.After, 10, 32)
	return int32(v), err
}

func (a *PagedConnectionArgs) ParseLimitOffset(defaultValue int32) (limit, offset int32, _ error) {
	offset, err := a.ParseOffset()
	return a.Limit(defaultValue), offset, err
}

type EmptyResponse struct{}

var Empty = &EmptyResponse{}

func (er *EmptyResponse) AlwaysNil() *string {
	return nil
}

func UnmarshalID[T any](id graphql.ID) (val T, err error) {
	err = relay.UnmarshalSpec(id, &val)
	return
}

func MarshalID[T any](kind string, id T) graphql.ID {
	return relay.MarshalID(kind, id)
}
