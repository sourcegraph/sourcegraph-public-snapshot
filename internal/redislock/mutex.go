package redislock

import (
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// OnlyOne runs the given function only if the lock can be acquired. If the lock
// is already held, it will wait until the lock is released or returns an error
// when the expiry time is reached.
func OnlyOne(logger log.Logger, client *redis.Client, name string, expiry time.Duration, fn func() error) (err error) {
	logger = logger.With(log.String("name", name))
	logger.Debug("acquiring lock")
	mutex := redsync.New(goredis.NewPool(client)).NewMutex(name, redsync.WithExpiry(expiry))
	if err = mutex.Lock(); err != nil {
		return errors.Wrap(err, "acquire lock")
	}

	defer func() {
		logger.Debug("releasing lock")
		_, unlockErr := mutex.Unlock()
		// Only overwrite the error if there wasn't already an error.
		if err == nil && unlockErr != nil {
			err = errors.Wrap(unlockErr, "release lock")
		}
	}()
	return fn()
}
