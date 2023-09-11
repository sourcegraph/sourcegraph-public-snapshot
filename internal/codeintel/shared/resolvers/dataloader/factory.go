package dataloader

import "context"

type MultiFactory[K, V any] interface {
	Load(ctx context.Context, id K) ([]V, error)
}

type MultiFactoryFunc[K, V any] func(ctx context.Context, id K) ([]V, error)

func (f MultiFactoryFunc[K, V]) Load(ctx context.Context, id K) ([]V, error) {
	return f(ctx, id)
}

type FactoryFunc[K, V any] func(ctx context.Context, id K) (V, error)
type FallibleFactoryFunc[K, V any] func(ctx context.Context, id K) (*V, error)

func NewMultiFactoryFromFactoryFunc[K, V any](f FactoryFunc[K, V]) MultiFactory[K, V] {
	return MultiFactoryFunc[K, V](func(ctx context.Context, id K) ([]V, error) {
		v, err := f(ctx, id)
		if err != nil {
			return nil, err
		}

		return []V{v}, nil
	})
}

func NewMultiFactoryFromFallibleFactoryFunc[K, V any](f FallibleFactoryFunc[K, V]) MultiFactory[K, V] {
	return MultiFactoryFunc[K, V](func(ctx context.Context, id K) ([]V, error) {
		v, err := f(ctx, id)
		if err != nil || v == nil {
			return nil, err
		}

		return []V{*v}, nil
	})
}
