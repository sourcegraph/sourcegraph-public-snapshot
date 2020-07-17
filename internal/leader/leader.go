package leader

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

const (
	defaultAcquireInterval = 30 * time.Second
)

type Options struct {
	// AcquireInterval defines how frequently we should attempt to acquire
	// leadership when not the leader.
	AcquireInterval time.Duration
	MutexOptions    rcache.MutexOptions
}

// Do will ensure that only one instance of workFn is running globally per key at any point using a mutex
// stored in Redis.
// workFn could lose leadership at any point so it is important that the supplied context is checked before performing
// any work that should not run in parallel with another worker.
// release can be called from within workFn to explicitly release the lock.
func Do(parentCtx context.Context, key string, workFn func(ctx context.Context), options Options) {
	if options.AcquireInterval == 0 {
		options.AcquireInterval = defaultAcquireInterval
	}
	for {
		select {
		case <-parentCtx.Done():
			return
		default:
		}

		ctx, cancel, ok := rcache.TryAcquireMutex(parentCtx, key, options.MutexOptions)
		if !ok {
			time.Sleep(options.AcquireInterval)
			continue
		}

		func() {
			defer cancel()
			workFn(ctx)
		}()
	}
}
