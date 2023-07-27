package graphqlutil

import (
	"context"
)

type SliceConnectionResolver[T any] interface {
	Nodes(ctx context.Context) []T
	TotalCount(ctx context.Context) int32
	PageInfo(ctx context.Context) *PageInfo
}

// NewSliceConnectionResolver creates a new sliceConnectionResolver that implements
// the SliceConnectionResolver interface. This is simply a convenience helper to return
// paginated slice in graphql-compliant way.
//
// data is the slice of nodes for this connection.
// total is the total number of nodes available.
// currentEnd is the current end index of the nodes slice.
//
// Returns a new sliceConnectionResolver that provides resolver methods for
// connection fields.
func NewSliceConnectionResolver[T any](data []T, total, currentEnd int) SliceConnectionResolver[T] {
	return &sliceConnectionResolver[T]{
		data:       data,
		total:      total,
		currentEnd: currentEnd,
	}
}

// sliceConnectionResolver implements the SliceConnectionResolver interface
// to provide resolver functions for a connection backed by a slice.
//
// data is the slice of nodes for this connection.
// total is the total number of nodes available.
// currentEnd is the current end index of the nodes slice.
type sliceConnectionResolver[T any] struct {
	data       []T
	total      int
	currentEnd int
}

func (c *sliceConnectionResolver[T]) Nodes(ctx context.Context) []T {
	return c.data
}

func (c *sliceConnectionResolver[T]) TotalCount(ctx context.Context) int32 {
	return int32(c.total)
}

func (c *sliceConnectionResolver[T]) PageInfo(ctx context.Context) *PageInfo {
	return HasNextPage(c.total > c.currentEnd)
}
