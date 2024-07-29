package object

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type expirer struct {
	store  Storage
	prefix string
	maxAge time.Duration
}

func NewExpirer(ctx context.Context, store Storage, prefix string, maxAge time.Duration, interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&expirer{
			store:  store,
			prefix: prefix,
			maxAge: maxAge,
		},
		goroutine.WithName("codeintel.upload-store-expirer"),
		goroutine.WithDescription("expires entries in the code intel upload store"),
		goroutine.WithInterval(interval),
	)
}

func (e *expirer) Handle(ctx context.Context) error {
	return e.store.ExpireObjects(ctx, e.prefix, e.maxAge)
}
