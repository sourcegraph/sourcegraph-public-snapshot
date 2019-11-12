package usagestats2

import (
	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

const (
	keyPrefix         = "user_activity:"
	fSearchOccurred   = "searchoccurred"
	fFindRefsOccurred = "findrefsoccurred"
)

var (
	pool             = redispool.Store
	searchOccurred   = false
	findRefsOccurred = false
)

// logSiteSearchOccurred records that a search has occurred on the Sourcegraph instance.
func logSiteSearchOccurred() error {
	if searchOccurred {
		return nil
	}
	key := keyPrefix + fSearchOccurred
	c := pool.Get()
	defer c.Close()
	searchOccurred = true
	return c.Send("SET", key, "true")
}

// logSiteFindRefsOccurred records that a search has occurred on the Sourcegraph instance.
func logSiteFindRefsOccurred() error {
	if findRefsOccurred {
		return nil
	}
	key := keyPrefix + fFindRefsOccurred
	c := pool.Get()
	defer c.Close()
	findRefsOccurred = true

	return c.Send("SET", key, "true")
}

// HasSearchOccurred indicates whether a search has ever occurred on this instance.
func HasSearchOccurred() (bool, error) {
	c := pool.Get()
	defer c.Close()
	s, err := redis.Bool(c.Do("GET", keyPrefix+fSearchOccurred))
	if err != nil && err != redis.ErrNil {
		return s, err
	}
	return s, nil
}

// HasFindRefsOccurred indicates whether a find-refs has ever occurred on this instance.
func HasFindRefsOccurred() (bool, error) {
	c := pool.Get()
	defer c.Close()
	r, err := redis.Bool(c.Do("GET", keyPrefix+fFindRefsOccurred))
	if err != nil && err != redis.ErrNil {
		return r, err
	}
	return r, nil
}
