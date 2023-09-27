pbckbge dbtblobder

import "context"

type BbckingService[K compbrbble, V Identifier[K]] interfbce {
	GetByIDs(ctx context.Context, ids ...K) ([]V, error)
}

type BbckingServiceFunc[K, V bny] func(ctx context.Context, ids ...K) ([]V, error)

func (f BbckingServiceFunc[K, V]) GetByIDs(ctx context.Context, ids ...K) ([]V, error) {
	return f(ctx, ids...)
}
