// Package redispool exports pools to specific redis instances.
package redispool

import (
	"github.com/pkg/errors"
	"net/url"
	"strconv"
	"strings"
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

func connectRedis(redisUrl string) (redis.Conn, error) {
	parsedUrl, err := url.Parse(redisUrl); if err != nil {
		return nil, err
	}

	if len(parsedUrl.Scheme) > 0 && parsedUrl.Scheme != "redis" {
		return nil, errors.New(redisUrl + ", is not a valid redis protocol, redis://:password@host:port/db")
	}

	var host string
	if len(parsedUrl.Opaque) > 0 {
		host = parsedUrl.Scheme + ":" + parsedUrl.Opaque
	} else {
		host = parsedUrl.Host
	}

	conn, err := redis.Dial("tcp", host); if err != nil {
		return nil, err
	}

	password, passwordIsSet := parsedUrl.User.Password()
	if passwordIsSet {
		_, err := conn.Do("AUTH", password); if err != nil {
			conn.Close()
			return nil, err
		}
	}

	var db = 0
	dbString := strings.Trim(parsedUrl.Path, "/")
	if len(dbString) > 0 {
		db, err = strconv.Atoi(dbString); if err != nil {
			conn.Close()
			return nil, err
		}

		_, err := conn.Do("SELECT", db); if err != nil {
			conn.Close()
			return nil, err
		}
	}

	return conn, nil
}

// Cache is a redis configured for caching. You usually want to use this. Only
// store data that can be recomputed here.
//
// In Kubernetes the service is called redis-cache.
var Cache = &redis.Pool{
	MaxIdle:     3,
	IdleTimeout: 240 * time.Second,
	Dial: func() (redis.Conn, error) {
		return connectRedis(addrCache)
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
		return connectRedis(addrStore)
	},
}
