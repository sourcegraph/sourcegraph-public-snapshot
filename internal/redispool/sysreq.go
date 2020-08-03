package redispool

import (
	"context"
	"strings"
	"time"

	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshreveable/log15"
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
		check := func() error {
			// Instead of just a PING, we also use this hook point to force a rewrite of
			// the AOF file on startup of the frontend as a way to ensure it doesn't
			// grow out of bounds which slows down future startups.
			// See https://github.com/sourcegraph/sourcegraph/issues/3300 for more context

			c := pool.Get()
			defer func() { _ = c.Close() }()

			if err := c.Send("PING"); err != nil {
				return err
			}

			if err := c.Send("BGREWRITEAOF"); err != nil {
				return err
			}

			if err := c.Flush(); err != nil {
				return err
			}

			_, err := c.Receive()
			if err != nil && strings.HasPrefix(err.Error(), "MISCONF") {
				// This is best effort. The BGREWRITEAOF can fail if the target redis instance
				// is not able to persist on disk. In this case we want to warn the operator,
				// but not stop the frontend from starting up as all data we store in redis is
				// considered flushable.
				log15.Warn(err.Error())
				return nil
			}
			return err
		}

		deadline := time.Now().Add(timeout)

		for time.Now().Before(deadline) {
			err = check()
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
