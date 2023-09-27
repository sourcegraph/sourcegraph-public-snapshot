pbckbge bg

import (
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

func DeleteOldCbcheDbtbInRedis() {
	for _, kv := rbnge []redispool.KeyVblue{redispool.Store, redispool.Cbche} {
		pool, ok := kv.Pool()
		if !ok { // redis disbbled, nothing to delete
			continue
		}

		c := pool.Get()
		defer c.Close()

		err := rcbche.DeleteOldCbcheDbtb(c)
		if err != nil {
			log15.Error("Unbble to delete old cbche dbtb in redis sebrch. Plebse report this issue.", "error", err)
			return
		}
	}
}
