package graphqlutil

import (
	"context"
	"sync"
)

type SliceConnectionResolver[T, V any] interface {
	Nodes(ctx context.Context) ([]V, error)
	TotalCount(ctx context.Context) int32
	PageInfo(ctx context.Context) (*PageInfo, error)
}

type SliceConnectionTransformer[T, V any] func(item T) (V, error)

func NewSliceConnectionResolver[T, V any](data []T, limit, offset int, transformer SliceConnectionTransformer[T, V]) SliceConnectionResolver[T, V] {
	return &sliceConnectionResolver[T, V]{
		data:        data,
		limit:       limit,
		offset:      offset,
		transformer: transformer,
	}
}

type sliceConnectionResolver[T, V any] struct {
	once sync.Once

	data []T

	paginatedData []V
	hasNextPage   bool
	transformer   SliceConnectionTransformer[T, V]
	err           error

	limit  int
	offset int
}

func (c *sliceConnectionResolver[T, V]) compute(ctx context.Context) ([]V, bool, error) {
	c.once.Do(func() {
		if c.offset < 0 || c.offset >= len(c.data) {
			c.paginatedData = make([]V, 0)
			return
		}

		start := c.offset
		end := start + c.limit

		if end > len(c.data) {
			end = len(c.data)
		}

		// rawPaginatedData is a slice containing the current page of items.
		// It slices the full items slice from the calculated start to end offsets.
		// hasNextPage is a boolean indicating if there are more pages after this one.
		// It is true if the length of the full items slice is greater than the end offset.
		rawPaginatedData := c.data[start:end]
		var hasNextPage bool
		if len(c.data) > (start + len(rawPaginatedData)) {
			hasNextPage = true
		}

		c.paginatedData = make([]V, len(rawPaginatedData))
		for i, raw := range rawPaginatedData {
			c.paginatedData[i], c.err = c.transformer(raw)
		}

		c.hasNextPage = hasNextPage
	})

	return c.paginatedData, c.hasNextPage, c.err
}

func (c *sliceConnectionResolver[T, V]) Nodes(ctx context.Context) ([]V, error) {
	data, _, err := c.compute(ctx)
	return data, err
}

func (c *sliceConnectionResolver[T, V]) TotalCount(ctx context.Context) int32 {
	return int32(len(c.data))
}

func (c *sliceConnectionResolver[T, V]) PageInfo(ctx context.Context) (*PageInfo, error) {
	_, hasNextPage, err := c.compute(ctx)
	return HasNextPage(hasNextPage), err
}
