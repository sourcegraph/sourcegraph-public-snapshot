package rcache

import (
	"context"
	"time"

	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"
)

func init() {
	sysreq.AddCheck("Redis", func(ctx context.Context) (problem, fix string, err error) {
		c := pool.Get()
		defer c.Close()

		timeout := 5 * time.Second
		deadline := time.Now().Add(timeout)

		for {
			time.Sleep(100 * time.Millisecond)
			if _, err := c.Do("PING"); err != nil && time.Now().After(deadline) {
				return "Redis is unavailable or misconfigured",
					fmt.Sprintf("Start a Redis server listening at port %s", redisMasterEndpoint),
					err
			} else if err == nil && !time.Now().After(deadline) {
				break
			}
		}

		return "", "", nil
	})
}
