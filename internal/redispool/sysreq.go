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
	sysreq.AddCheck("Redis Store", redisCheck("Store", addresses.Store, timeout, Store))
	sysreq.AddCheck("Redis Cache", redisCheck("Cache", addresses.Cache, timeout, Cache))
}

func redisCheck(name, addr string, timeout time.Duration, kv KeyValue) sysreq.CheckFunc {
	return func(ctx context.Context) (problem, fix string, err error) {
		check := func() (err error) {
			// Instead of just a PING, we also use this hook point to force a rewrite of
			// the AOF file on startup of the frontend as a way to ensure it doesn't
			// grow out of bounds which slows down future startups.
			// See https://github.com/sourcegraph/sourcegraph/issues/3300 for more context

			pool := kv.Pool()
			c := pool.Get()
			defer func() { _ = c.Close() }()

			if err = c.Send("PING"); err != nil {
				return err
			}

			if err = c.Send("BGREWRITEAOF"); err != nil {
				return err
			}

			if err = c.Flush(); err != nil {
				return err
			}

			// We ignore the response from BGREWRITEAOF, as this is best effort.
			_, err = c.Receive()
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
