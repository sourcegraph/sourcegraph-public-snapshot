// Package redispool exports pools to specific redis instances.
package redispool

import (
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

var schemeMatcher = lazyregexp.New(`^[A-Za-z][A-Za-z0-9\+\-\.]*://`)

// dialRedis dials Redis given the raw endpoint string. The string can have two formats:
//  1. If there is a HTTP scheme, it should be either be "redis://" or "rediss://" and the URL
//     must be of the format specified in https://www.iana.org/assignments/uri-schemes/prov/redis.
//  2. Otherwise, it is assumed to be of the format $HOSTNAME:$PORT.
func dialRedis(rawEndpoint string) (redis.Conn, error) {
	if schemeMatcher.MatchString(rawEndpoint) { // expect "redis://"
		return redis.DialURL(rawEndpoint)
	}
	if strings.Contains(rawEndpoint, "/") {
		return nil, errors.New("Redis endpoint without scheme should not contain '/'")
	}
	return redis.Dial("tcp", rawEndpoint)
}

// Cache is a redis configured for caching. You usually want to use this. Only
// store data that can be recomputed here. Although this data is treated as ephemeral,
// Sourcegraph depends on it to operate performantly, so we persist in Redis to avoid cold starts,
// rather than having it in-memory only.
//
// In Kubernetes the service is called redis-cache.
var Cache = &redis.Pool{
	MaxIdle:     3,
	IdleTimeout: 240 * time.Second,
	Dial: func() (redis.Conn, error) {
		return dialRedis(addrCache)
	},
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
	Dial: func() (redis.Conn, error) {
		return dialRedis(addrStore)
	},
}
