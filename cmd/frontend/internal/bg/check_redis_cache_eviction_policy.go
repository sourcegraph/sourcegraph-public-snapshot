package bg

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/sourcegraph/sourcegraph/pkg/redispool"
	"gopkg.in/inconshreveable/log15.v2"
)

const recommendedPolicy = "allkeys-lru"

func CheckRedisCacheEvictionPolicy() {
	c := redispool.Cache.Get()
	defer c.Close()

	vals, err := redis.Strings(c.Do("CONFIG", "GET", "maxmemory-policy"))
	if err != nil {
		log15.Error("Reading `maxmemory-policy` from Redis failed", "error", err)
		return
	}

	if len(vals) == 2 && vals[1] != recommendedPolicy {
		msg := fmt.Sprintf("ATTENTION: Your Redis cache instance does not have the recommended `maxmemory-policy` set. The current value is '%s'. Recommend for the cache is '%s'.", vals[1], recommendedPolicy)
		log15.Warn("****************************")
		log15.Warn(msg)
		log15.Warn("****************************")
	}
}
