package featureflag

import (
	"fmt"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	pool = redispool.Store
)

func GetEvaluatedFlagSetFromCache(flags []*FeatureFlag, visitorID string) FlagSet {
	flagSet := FlagSet{}

	c := pool.Get()
	defer c.Close()

	for _, flag := range flags {
		value, _ := redis.Bool(c.Do("HGET", flag.CacheKey(), visitorID))

		flagSet[flag.Name] = value
	}

	return flagSet
}

func SetEvaluatedFlagToCache(f *FeatureFlag, visitorID string, value bool) {
	c := pool.Get()
	defer c.Close()

	c.Do("HSET", f.CacheKey(), visitorID, fmt.Sprintf("%v", value))
}

func ClearFlagFromCache(name string) {
	c := pool.Get()
	defer c.Close()

	c.Do("DEL", GetFlagCacheKey(name))
}

func GetVisitorIDForUser(userID int32) string {
	return fmt.Sprintf("uid_%v", userID)
}

func GetVisitorIDForAnonymousUser(anonymousUID string) string {
	return "auid_" + anonymousUID
}

func GetFlagCacheKey(name string) string {
	return "ff_" + name
}
