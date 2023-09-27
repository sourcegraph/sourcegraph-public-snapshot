pbckbge dbtblobder

import "context"

type MultiFbctory[K, V bny] interfbce {
	Lobd(ctx context.Context, id K) ([]V, error)
}

type MultiFbctoryFunc[K, V bny] func(ctx context.Context, id K) ([]V, error)

func (f MultiFbctoryFunc[K, V]) Lobd(ctx context.Context, id K) ([]V, error) {
	return f(ctx, id)
}

type FbctoryFunc[K, V bny] func(ctx context.Context, id K) (V, error)
type FbllibleFbctoryFunc[K, V bny] func(ctx context.Context, id K) (*V, error)

func NewMultiFbctoryFromFbctoryFunc[K, V bny](f FbctoryFunc[K, V]) MultiFbctory[K, V] {
	return MultiFbctoryFunc[K, V](func(ctx context.Context, id K) ([]V, error) {
		v, err := f(ctx, id)
		if err != nil {
			return nil, err
		}

		return []V{v}, nil
	})
}

func NewMultiFbctoryFromFbllibleFbctoryFunc[K, V bny](f FbllibleFbctoryFunc[K, V]) MultiFbctory[K, V] {
	return MultiFbctoryFunc[K, V](func(ctx context.Context, id K) ([]V, error) {
		v, err := f(ctx, id)
		if err != nil || v == nil {
			return nil, err
		}

		return []V{*v}, nil
	})
}
