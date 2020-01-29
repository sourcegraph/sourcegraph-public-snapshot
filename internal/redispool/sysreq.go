package redispool

import (
	"context"
	"time"

	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/sysreq"
)

var timeout, _ = time.ParseDuration(env.Get("SRC_REDIS_WAIT_FOR", "90s", "Duration to wait for Redis to become ready before quitting"))

func init() {
	sysreq.AddCheck("Redis Store", redisCheck("Store", addrStore, timeout, Store))
	sysreq.AddCheck("Redis Cache", redisCheck("Cache", addrCache, timeout, Cache))
}

func redisCheck(name, addr string, timeout time.Duration, pool *redis.Pool) sysreq.CheckFunc {
	return func(ctx context.Context) (problem, fix string, err error) {
		deadline := time.Now().Add(timeout)

		for time.Now().Before(deadline) {
			c := pool.Get()
			_, err = c.Do("PING")
			c.Close()

			if err == nil {
				// Success
				return "", "", nil
			}
			// Try again
			time.Sleep(250 * time.Millisecond)
		}

		return fmt.Sprintf("Redis %q is unavailable or misconfigured", name),
			fmt.Sprintf("Start a Redis server listening at port %s", addr),
			err
	}
}
