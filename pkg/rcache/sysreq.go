package rcache

import (
	"golang.org/x/net/context"

	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"
)

func init() {
	sysreq.AddCheck("Redis", func(ctx context.Context) (problem, fix string, err error) {
		if _, err := redisPool(); err != nil {
			return "Redis is unavailable or misconfigured",
				fmt.Sprintf("Start a Redis server listening at port %s", conf.GetenvOrDefault("REDIS_MASTER_ENDPOINT", ":6379")),
				err
		}
		return
	})
}
