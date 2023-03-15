package uploadstore

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type expirer struct {
	store  Store
	prefix string
	maxAge time.Duration
}

func NewExpirer(ctx context.Context, store Store, prefix string, maxAge time.Duration, interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(ctx, "codeintel.upload-store-expirer", "expires entries in the code intel upload store", interval, &expirer{
		store:  store,
		prefix: prefix,
		maxAge: maxAge,
	})
}

func (e *expirer) Handle(ctx context.Context) error {
	return e.store.ExpireObjects(ctx, e.prefix, e.maxAge)
}
