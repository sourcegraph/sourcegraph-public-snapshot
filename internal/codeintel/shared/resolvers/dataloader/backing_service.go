package dataloader

import "context"

type BackingService[K comparable, V Identifier[K]] interface {
	GetByIDs(ctx context.Context, ids ...K) ([]V, error)
}

type BackingServiceFunc[K, V any] func(ctx context.Context, ids ...K) ([]V, error)

func (f BackingServiceFunc[K, V]) GetByIDs(ctx context.Context, ids ...K) ([]V, error) {
	return f(ctx, ids...)
}
