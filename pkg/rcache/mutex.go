package rcache

import (
	"fmt"
	"time"

	"context"

	"github.com/garyburd/redigo/redis"
	"github.com/hjr265/redsync.go/redsync"
	log15 "gopkg.in/inconshreveable/log15.v2"
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
	// mutexPrefix allows us to run concurrent tests which try to lock the
	// same resources.
	mutexPrefix = "mutex"
)

// TryAcquireMutex tries to Lock a distributed mutex. If the mutex is already
// locked, it will return `ctx, nil, false`. Otherwise it returns `ctx,
// release, true`. Release must be called to free the lock. When the lock is
// free the returned context is cancelled.
func TryAcquireMutex(ctx context.Context, name string) (context.Context, func(), bool) {
	name = fmt.Sprintf("%s:mutex:%s", globalPrefix, name)
	mu, err := redsync.NewMutexWithPool(name, []*redis.Pool{pool})
	if err != nil {
		log15.Warn("failed to create redis mutex", "name", name, "error", err)
		return ctx, nil, false
	}
	mu.Expiry = mutexExpiry
	mu.Tries = mutexTries
	mu.Delay = mutexDelay
	err = mu.Lock()
	if err != nil {
		return ctx, nil, false
	}
	ctx, cancel := context.WithCancel(ctx)
	unlockedC := make(chan interface{})
	go func() {
		ticker := time.NewTicker(mu.Expiry / 2)
		for {
			select {
			case <-ctx.Done():
				// TODO handle error
				mu.Unlock()
				ticker.Stop()
				close(unlockedC)
				return
			case <-ticker.C:
				// TODO simple retry
				if !mu.Touch() {
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
