// Package redispool exports pools to specific redis instances.
package redispool

import (
	"regexp"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	// addrCache is the network address of redis cache.
	addrCache string
	// addrStore is the network address of redis store.
	addrStore string
)

func init() {
	// Set addresses. Prefer in this order:
	// * Specific envvar REDIS_${NAME}_ENDPOINT
	// * Fallback envvar REDIS_ENDPOINT
	// * Default
	//
	// Additionally keep this logic in sync with cmd/server/redis.go
	fallback := env.Get("REDIS_ENDPOINT", "", "redis endpoint. Used as fallback if REDIS_CACHE_ENDPOINT or REDIS_STORE_ENDPOINT is not specified.")

	// addrCache
	for _, addr := range []string{
		env.Get("REDIS_CACHE_ENDPOINT", "", "redis used for cache data. Default redis-cache:6379"),
		fallback,
		"redis-cache:6379",
	} {
		if addr != "" {
			addrCache = addr
			break
		}
	}

	// addrStore
	for _, addr := range []string{
		env.Get("REDIS_STORE_ENDPOINT", "", "redis used for persistent stores (eg HTTP sessions). Default redis-store:6379"),
		fallback,
		"redis-store:6379",
	} {
		if addr != "" {
			addrStore = addr
			break
		}
	}
}

var schemeMatcher = regexp.MustCompile("^[A-Za-z][A-Za-z0-9\\+\\-\\.]*://")

// redisDialer returns the appropriate Redis dial function given the raw endpoint string.
// The string can have two formats:
// 1) If it begins with "redis://", it is assumed to be of the format specified in
//    https://www.iana.org/assignments/uri-schemes/prov/redis.
// 2) Otherwise, it is assumed to be of the format $HOSTNAME:$PORT.
func redisDialer(rawEndpoint string) func() (redis.Conn, error) {
	return func() (redis.Conn, error) {
		if schemeMatcher.MatchString(rawEndpoint) { // expect "redis://"
			return redis.DialURL(rawEndpoint)
		}
		return redis.Dial("tcp", rawEndpoint)
	}
}

// Cache is a redis configured for caching. You usually want to use this. Only
// store data that can be recomputed here.
//
// In Kubernetes the service is called redis-cache.
var Cache = &redis.Pool{
	MaxIdle:     3,
	IdleTimeout: 240 * time.Second,
	Dial:        redisDialer(addrCache),
	TestOnBorrow: func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	},
}

// Store is a redis configured for persisting data. Do not abuse this pool,
// only use if you have data with a high write rate.
//
// In Kubernetes the service is called redis-store.
var Store = &redis.Pool{
	MaxIdle:     10,
	IdleTimeout: 240 * time.Second,
	TestOnBorrow: func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	},
	Dial: redisDialer(addrStore),
}
