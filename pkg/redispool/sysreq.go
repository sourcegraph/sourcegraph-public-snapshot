package redispool

import (
	"context"
	"time"

	"fmt"

	"sourcegraph.com/pkg/sysreq"
)

func init() {
	sysreq.AddCheck("Redis", func(ctx context.Context) (problem, fix string, err error) {
		c := Store.Get()
		defer c.Close()

		timeout := 5 * time.Second
		deadline := time.Now().Add(timeout)

		for time.Now().Before(deadline) {
			_, err = c.Do("PING")
			if err == nil {
				// Success
				return "", "", nil
			}
			// Try again
			time.Sleep(100 * time.Millisecond)
		}
		return "Redis is unavailable or misconfigured",
			fmt.Sprintf("Start a Redis server listening at port %s", addrStore),
			err
	})
}
