package bg

import (
	"github.com/garyburd/redigo/redis"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/pkg/redispool"
	"gopkg.in/inconshreveable/log15.v2"
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
