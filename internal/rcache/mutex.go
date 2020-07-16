package rcache

import (
	"fmt"
	"time"

	"context"

	"github.com/go-redsync/redsync"
)

var (
	// mutexExpiry is relatively long since we currently are only using
	// locks for co-ordinating longer running processes. If we want short
	// lived granular locks, we should switch away from Redis.
	mutexExpiry = time.Minute
	// mutexTries is how many tries we have before we give up acquiring a
	// lock. We make it low since we want to give up quickly + we only
	// have a single node. So failing to acquire the lock will be
	// unrelated to failing to reach quoram. var to allow tests to
	// override.
	mutexTries = 3
	// mutexDelay is how long to sleep between attempts to lock. We use
	// the default delay.
	mutexDelay = 512 * time.Millisecond
)

// TryAcquireMutex tries to Lock a distributed mutex. If the mutex is already
// locked, it will return `ctx, nil, false`. Otherwise it returns `ctx,
// release, true`. Release must be called to free the lock.
// The lock has a 1 minute lifetime, but a background routine extends it every 30 seconds.
// If, on release, we are unable to unlock the mutex it will continue to be locked until
// it is expired by Redis.
// The returned context will be cancelled if any of the following occur:
// * The parent context in cancelled
// * The release function is called
// * There is an error extending the lock expiry or the expiry can't be extended because
//   they key no longer exists in Redis
// A caller can therefore assume that they are the sole holder of the lock as long as the
// context has not been cancelled.
func TryAcquireMutex(ctx context.Context, name string) (context.Context, func(), bool) {
	// We return a canceled context if we fail, so create the context here
	ctx, cancel := context.WithCancel(ctx)

	name = fmt.Sprintf("%s:mutex:%s", globalPrefix, name)
	mu := redsync.New([]redsync.Pool{pool}).NewMutex(
		name,
		redsync.SetExpiry(mutexExpiry),
		redsync.SetTries(mutexTries),
		redsync.SetRetryDelay(mutexDelay),
	)

	err := mu.Lock()
	if err != nil {
		cancel()
		return ctx, nil, false
	}
	unlockedC := make(chan struct{})
	go func() {
		ticker := time.NewTicker(mutexExpiry / 2)
		for {
			select {
			case <-ctx.Done():
				// An error here means we may not have released the lock.
				// It's OK to ignore as we'll stop extending the lock anyway
				// and it will expire.
				_, _ = mu.Unlock()
				ticker.Stop()
				close(unlockedC)
				return
			case <-ticker.C:
				// We do not retry on error as we should cancel the context
				// as soon as we are not 100% sure we hold the lock. This minimises
				// the chance of more than one instance thinking they hold it.
				if ok, err := mu.Extend(); !ok || err != nil {
					cancel()
				}
			}
		}
	}()
	return ctx, func() {
		cancel()
		<-unlockedC
	}, true
}
