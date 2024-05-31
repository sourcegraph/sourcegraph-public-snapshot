package redis

import (
	"context"

	goredis "github.com/go-redis/redis/v8"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Client struct {
	c *goredis.Client
}

func NewClient(ctx context.Context, contract runtime.Contract) (*Client, error) {
	redisOpts, err := goredis.ParseURL(*contract.RedisEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "invalid Redis DSN")
	}
	return &Client{goredis.NewClient(redisOpts)}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	pong := c.c.Ping(ctx)
	return pong.Err()
}
