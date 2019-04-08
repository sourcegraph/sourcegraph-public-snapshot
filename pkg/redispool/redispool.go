// Package redispool exports pools to specific redis instances.
package redispool

import (
	"github.com/pkg/errors"
	"regexp"
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


type RedisConnectionConfiguration struct {
	Host string
	Port string
	Password string
	Db string
}


func ParseRedisConnectionUrl(redisUrl string) (*RedisConnectionConfiguration, error) {
	if strings.HasPrefix(redisUrl,"redis") {
		return parseRedisConnectionUrlWithScheme(redisUrl)
	} else {
		return parseRedisConnectionUrlWithoutScheme(redisUrl)
	}
}

func parseRedisConnectionUrlWithScheme(redisUrl string) (*RedisConnectionConfiguration, error) {
	redisUrlRegexp := regexp.MustCompile(`redis://((\w+)?:(\w+)@)?([^:]+)(:(\d+))?(/(\d+))?`)
	matchedGroups := redisUrlRegexp.FindStringSubmatch(redisUrl)
	matchedLength := len(matchedGroups)
	if matchedLength <= 0 {
		return nil, errors.New("invalid redis url, correct is redis://:pasword@host:port/db")
	}
	var redisConnectionConfiguration RedisConnectionConfiguration
	redisConnectionConfiguration.Password = matchedGroups[3]
	redisConnectionConfiguration.Host = matchedGroups[4]
	if len(matchedGroups[6]) <= 0 {
		redisConnectionConfiguration.Port = "6379"
	} else {
		redisConnectionConfiguration.Port = matchedGroups[6]
	}
	redisConnectionConfiguration.Db = matchedGroups[8]

	return &redisConnectionConfiguration, nil
}

func parseRedisConnectionUrlWithoutScheme(redisUrl string) (*RedisConnectionConfiguration, error) {
	redisUrlRegexp := regexp.MustCompile(`([^:]+)(:(\d+))?(/(\d+))?`)
	matchedGroups := redisUrlRegexp.FindStringSubmatch(redisUrl)
	matchedLength := len(matchedGroups)
	if matchedLength <= 0 {
		return nil, errors.New("error when parsing redis url, it should be host:port")
	}

	var redisConnectionConfiguration RedisConnectionConfiguration
	redisConnectionConfiguration.Host = matchedGroups[1]
	if len(matchedGroups[3]) == 0 {
		redisConnectionConfiguration.Port = "6379"
	} else {
		redisConnectionConfiguration.Port = matchedGroups[3]
	}
	redisConnectionConfiguration.Db = matchedGroups[5]

	return &redisConnectionConfiguration, nil
}

func connectRedis(redisUrl string) (redis.Conn, error) {
	redisConnectionConfiguration, err := ParseRedisConnectionUrl(redisUrl); if err != nil {
		return nil, err
	}

	hostAndPort := redisConnectionConfiguration.Host + ":" + redisConnectionConfiguration.Port
	conn, err := redis.Dial("tcp", hostAndPort); if err != nil {
		return nil, err
	}

	if len(redisConnectionConfiguration.Password) > 0 {
		_, err := conn.Do("AUTH", redisConnectionConfiguration.Password); if err != nil {
			conn.Close()
			return nil, err
		}
	}

	if len(redisConnectionConfiguration.Db) > 0 {
		_, err := conn.Do("SELECT", redisConnectionConfiguration.Db); if err != nil {
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
