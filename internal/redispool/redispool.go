// Package redispool exports pools to specific redis instances.
package redispool

import (
	"os"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Set addresses. We do it as a function closure to ensure the addresses are
// set before we create Store and Cache. Prefer in this order:
// * Specific envvar REDIS_${NAME}_ENDPOINT
// * Fallback envvar REDIS_ENDPOINT
// * Default
//
// Additionally keep this logic in sync with cmd/server/redis.go
var addresses = func() struct {
	Cache string
	Store string
} {
	redis := struct {
		Cache string
		Store string
	}{}

	fallback := func(d string) string {
		if os.Getenv("REDIS_ENDPOINT") != "" {
			return os.Getenv("REDIS_ENDPOINT")
		}
		return d
	}

	redis.Cache = env.Get("REDIS_CACHE_ENDPOINT", fallback("redis-cache:6379"), "redis used for cache data. if not set, REDIS_ENDPOINT will be considered")
	redis.Store = env.Get("REDIS_STORE_ENDPOINT", fallback("redis-store:6379"), "redis used for persistent stores (eg HTTP sessions). if not set, REDIS_ENDPOINT will be considered")

	return redis
}()

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
var Cache = NewKeyValue(addresses.Cache, &redis.Pool{
	MaxIdle:     3,
	IdleTimeout: 240 * time.Second,
	MaxActive:   1000,
})

// Store is a redis configured for persisting data. Do not abuse this pool,
// only use if you have data with a high write rate.
//
// In Kubernetes the service is called redis-store.
var Store = NewKeyValue(addresses.Store, &redis.Pool{
	MaxIdle:     10,
	IdleTimeout: 240 * time.Second,
	MaxActive:   1000,
})
