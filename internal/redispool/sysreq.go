package redispool

import (
	"context"
	"time"

	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/sysreq"
)

var timeout, _ = time.ParseDuration(env.Get("SRC_REDIS_WAIT_FOR", "90s", "Duration to wait for Redis to become ready before quitting"))

func init() {
	sysreq.AddCheck("Redis", func(ctx context.Context) (problem, fix string, err error) {
		c := Store.Get()
		defer c.Close()

		deadline := time.Now().Add(timeout)

		for time.Now().Before(deadline) {
			_, err = c.Do("PING")
			if err == nil {
				// Success
				return "", "", nil
			}
			// Try again
			time.Sleep(250 * time.Millisecond)
		}
		return "Redis is unavailable or misconfigured",
			fmt.Sprintf("Start a Redis server listening at port %s", addrStore),
			err
	})
}
