package redislock

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// TryAcquire attempts to acquire a Redis-based lock with the given key in a
// single pass. It does not block if the lock is already held by someone else.
//
// The locking algorithm is based on https://redis.io/commands/setnx/ for
// resolving deadlocks. While it provides less semantic guarantees and features
// than a more sophisticated distributed locking algorithm like Redlock, it is
// best suited when the number of contenders is unbounded and non-deterministic,
// which avoids the need for pre-allocating mutexes for all possible contenders
// and managing lifecycles of those mutexes. Please see the Redlock documentation
// (https://redis.io/docs/manual/patterns/distributed-locks/) for more details,
// in particular the "Why Failover-based Implementations Are Not Enough" section
// regarding when it's not a good choice to use this locking algorithm if your
// use case concerns about the drawback (i.e. it is _absolutely critical_ that
// only one contender should get the lock at any given time).
//
// CAUTION: To avoid releasing someone else's lock, the duration of the entire
// operation should be well-below the lock timeout.
func TryAcquire(ctx context.Context, rs redispool.KeyValue, lockKey string, lockTimeout time.Duration) (acquired bool, release func(), err error) {
	// Returned ctx is not yet used anywhere, discard to avoid the linter
	tr, _ := trace.New(ctx, "redislock.TryAcquire",
		attribute.String("lockKey", lockKey),
		attribute.Float64("lockTimeoutSeconds", lockTimeout.Seconds()))
	defer func() {
		tr.SetAttributes(attribute.Bool("acquired", acquired))
		tr.EndWithErr(&err)
	}()

	timeout := time.Now().Add(lockTimeout).UnixNano()
	// Encode UUID as part of the token to eliminate the chance of multiple processes
	// falsely believing they have the lock at the same time.
	lockToken := fmt.Sprintf("%d,%s", timeout, uuid.New().String())

	release = func() {
		// Best effort to check we're releasing the lock we think we have. Note that it
		// is still technically possible the lock token has changed between the GET and
		// DEL since these are two separate operations, i.e. when the current lock happen
		// to be expired at this very moment.
		get, _ := rs.Get(lockKey).String()
		if get == lockToken {
			_ = rs.Del(lockKey)
		}
	}

	set, err := rs.SetNx(lockKey, lockToken)
	if err != nil {
		return false, nil, err
	} else if set {
		return true, release, nil
	}

	// We didn't get the lock, but we can check if the lock is expired.
	currentLockToken, err := rs.Get(lockKey).String()
	if err == redis.ErrNil {
		// Someone else got the lock and released it already.
		return false, nil, nil
	} else if err != nil {
		return false, nil, err
	}

	// Extract the lock deadline.
	now := time.Now()
	currentTimeout, _ := strconv.ParseInt(strings.SplitN(currentLockToken, ",", 2)[0], 10, 64)
	tr.SetAttributes(
		attribute.Float64("secondsRemaining", time.Unix(0, currentTimeout).Sub(now).Seconds()))
	if currentTimeout > now.UnixNano() {
		// The lock is still valid.
		return false, nil, nil
	}

	// The lock has expired, try to acquire it.
	get, err := rs.GetSet(lockKey, lockToken).String()
	if err != nil {
		return false, nil, err
	} else if get != currentLockToken {
		// Someone else got the lock
		return false, nil, nil
	}

	// We got the lock.
	return true, release, nil
}
