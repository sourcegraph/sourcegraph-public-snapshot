package bg

import (
	"github.com/gomodule/redigo/redis"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

func DeleteOldCacheDataInRedis() {
	storeConn := redispool.Store.Get()
	defer storeConn.Close()

	cacheConn := redispool.Cache.Get()
	defer cacheConn.Close()

	for _, c := range []redis.Conn{storeConn, cacheConn} {
		err := rcache.DeleteOldCacheData(c)
		if err != nil {
			log15.Error("Unable to delete old cache data in redis search. Please report this issue.", "error", err)
			return
		}
	}
}
