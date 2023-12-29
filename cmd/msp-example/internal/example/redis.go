package example

import (
	"context"

	goredis "github.com/go-redis/redis/v8"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

// testRedisConnection creates a new Redis client from the MSP contract and issues
// a PING to check the connection.
func testRedisConnection(ctx context.Context, c runtime.Contract) error {
	if c.RedisEndpoint == nil {
		return errors.New("no Redis endpoint provided")
	}

	redisOpts, err := goredis.ParseURL(*c.RedisEndpoint)
	if err != nil {
		return errors.Wrap(err, "invalid Redis DSN")
	}

	client := goredis.NewClient(redisOpts)
	pong := client.Ping(ctx)
	return pong.Err()
}
