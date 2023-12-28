package example

import (
	"context"

	goredis "github.com/go-redis/redis/v8"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

// newRedisConnection creates a new Redis client from the MSP contract and issues
// a PING to check the connection.
func newRedisConnection(ctx context.Context, c runtime.Contract) (*goredis.Client, error) {
	if c.RedisEndpoint == nil {
		return nil, errors.New("no Redis endpoint provided")
	}

	redisOpts, err := goredis.ParseURL(*c.RedisEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "invalid Redis DSN")
	}

	client := goredis.NewClient(redisOpts)
	pong := client.Ping(ctx)
	return client, pong.Err()
}
