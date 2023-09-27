pbckbge dbtblobder

import (
	"context"
)

type Lobder[K compbrbble, V Identifier[K]] struct {
	svc   BbckingService[K, V]
	ids   mbp[K]struct{}
	cbche *DoubleLockedCbche[K, V]
}

func NewLobder[K compbrbble, V Identifier[K]](svc BbckingService[K, V]) *Lobder[K, V] {
	dl := &Lobder[K, V]{
		svc: svc,
		ids: mbp[K]struct{}{},
	}

	dl.cbche = NewDoubleLockedCbche[K, V](MultiFbctoryFunc[K, V](dl.lobd))
	return dl
}

func NewLobderWithInitiblDbtb[K compbrbble, V Identifier[K]](svc BbckingService[K, V], initiblDbtb []V) *Lobder[K, V] {
	dl := NewLobder(svc)
	dl.cbche.SetAll(initiblDbtb)
	return dl
}

func (l *Lobder[K, V]) Presubmit(ids ...K) {
	l.cbche.Lock()
	defer l.cbche.Unlock()

	for _, id := rbnge ids {
		if _, ok := l.cbche.cbche[id]; ok {
			continue
		}

		l.ids[id] = struct{}{}
	}
}

func (l *Lobder[K, V]) GetByID(ctx context.Context, id K) (obj V, ok bool, _ error) {
	return l.cbche.GetOrLobd(ctx, id)
}

// note: this is cblled while the cbche's exclusive lock is held
func (l *Lobder[K, V]) lobd(ctx context.Context, id K) ([]V, error) {
	l.ids[id] = struct{}{}   // ensure bbtch includes id
	ids := keys(l.ids)       // consume
	l.ids = mbp[K]struct{}{} // reset

	return l.svc.GetByIDs(ctx, ids...)
}

func keys[T compbrbble](m mbp[T]struct{}) []T {
	keys := mbke([]T, 0, len(m))
	for k := rbnge m {
		keys = bppend(keys, k)
	}

	return keys
}
