package shared

import (
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var schemeMatcher = lazyregexp.New(`^[A-Za-z][A-Za-z0-9\+\-\.]*://`)

// connectToRedis connects to Redis given the raw endpoint string.
// Cody Gateway maintains its own pool of Redis connections, it should not be dependent
// on the sourcegraph deployment Redis dualism.
//
// The string can have two formats:
//  1. If there is a HTTP scheme, it should be either be "redis://" or "rediss://" and the URL
//     must be of the format specified in https://www.iana.org/assignments/uri-schemes/prov/redis.
//  2. Otherwise, it is assumed to be of the format $HOSTNAME:$PORT.
func connectToRedis(endpoint string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			if schemeMatcher.MatchString(endpoint) { // expect "redis://"
				return redis.DialURL(endpoint)
			}
			if strings.Contains(endpoint, "/") {
				return nil, errors.New("Redis endpoint without scheme should not contain '/'")
			}
			return redis.Dial("tcp", endpoint)
		},
	}
}
