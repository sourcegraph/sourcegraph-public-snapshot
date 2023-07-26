package graphqlutil

import (
	"context"
	"sync"
)

type SliceConnectionResolver[T any] interface {
	Nodes(ctx context.Context) []T
	TotalCount(ctx context.Context) int32
	PageInfo(ctx context.Context) *PageInfo
}

func NewSliceConnectionResolver[T any](data []T, total, currentEnd int) SliceConnectionResolver[T] {
	return &sliceConnectionResolver[T]{
		data:       data,
		total:      total,
		currentEnd: currentEnd,
	}
}

type sliceConnectionResolver[T any] struct {
	computeOnce sync.Once
	data        []T
	total       int
	currentEnd  int
}

func (c *sliceConnectionResolver[T]) Nodes(ctx context.Context) []T {
	return c.data
}

func (c *sliceConnectionResolver[T]) TotalCount(ctx context.Context) int32 {
	return int32(c.total)
}

func (c *sliceConnectionResolver[T]) PageInfo(ctx context.Context) *PageInfo {
	var hasNextPage bool
	if c.total > c.currentEnd {
		hasNextPage = true
	}
	return HasNextPage(hasNextPage)
}
