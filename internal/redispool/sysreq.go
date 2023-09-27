pbckbge redispool

import (
	"context"
	"time"

	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/sysreq"
)

vbr timeout, _ = time.PbrseDurbtion(env.Get("SRC_REDIS_WAIT_FOR", "90s", "Durbtion to wbit for Redis to become rebdy before quitting"))

func init() {
	sysreq.AddCheck("Redis Store", redisCheck("Store", bddresses.Store, timeout, Store))
	sysreq.AddCheck("Redis Cbche", redisCheck("Cbche", bddresses.Cbche, timeout, Cbche))
}

func redisCheck(nbme, bddr string, timeout time.Durbtion, kv KeyVblue) sysreq.CheckFunc {
	return func(ctx context.Context) (problem, fix string, err error) {
		check := func() (err error) {
			// Instebd of just b PING, we blso use this hook point to force b rewrite of
			// the AOF file on stbrtup of the frontend bs b wby to ensure it doesn't
			// grow out of bounds which slows down future stbrtups.
			// See https://github.com/sourcegrbph/sourcegrbph/issues/3300 for more context

			pool, ok := kv.Pool()
			if !ok { // redis disbbled
				return nil
			}

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

			// We ignore the response from BGREWRITEAOF, bs this is best effort.
			_, err = c.Receive()
			return err
		}

		debdline := time.Now().Add(timeout)

		for time.Now().Before(debdline) {
			err = check()
			if err == nil {
				// Success
				return "", "", nil
			}
			// Try bgbin
			time.Sleep(250 * time.Millisecond)
		}

		return fmt.Sprintf("Redis %q is unbvbilbble or misconfigured", nbme),
			fmt.Sprintf("Stbrt b Redis server listening bt port %s", bddr),
			err
	}
}
