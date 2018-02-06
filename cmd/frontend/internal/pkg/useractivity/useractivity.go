// Package useractivity provides an interface to update and access information about
// individual and aggregate Sourcegraph Server users' activity levels.
//
// Note that this package should not be used on sourcegraph.com, only on self-hosted
// deployments.
package useractivity

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	gcOnce sync.Once // ensures we only have 1 redis gc goroutine

	keyPrefix = "user_activity:"
	pool      *redis.Pool
)

const (
	fPageViews     = "pageviews"
	fLastActive    = "lastactive"
	fSearchQueries = "searchqueries"
)

func init() {
	address := env.Get("SRC_STORE_REDIS", "", "redis used for storing persistent data")

	// This logic below is the behaviour used by our session store. We
	// fallback to it since we will always be using the same redis (for
	// now). In future we will hopefully have migrated to just using
	// SRC_STORE_REDIS.
	if address == "" {
		var ok bool
		address, ok = os.LookupEnv("SRC_SESSION_STORE_REDIS")
		if !ok {
			address = "redis-store:6379"
		}
	}
	if address == "" {
		address = ":6379"
	}

	pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			// Setup our GC of active key goroutine
			gcOnce.Do(func() {
				go gc()
			})
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// MigrateUserActivityData moves all old user activity data from the DB to Redis.
// Should only ever happen one time.
func MigrateUserActivityData(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log15.Error("panic in useractivity.MigrateUserActivityData", "error", err)
		}
	}()

	c := pool.Get()
	defer c.Close()

	migrateKey := keyPrefix + "dbmigrate"
	migrated, err := redis.Bool(c.Do("GET", migrateKey))
	if err != nil && err != redis.ErrNil {
		log15.Error("Failed to check if useractivity is migrated", "error", err)
		return
	}
	if migrated {
		return
	}

	allUserActivity, err := db.UserActivity.GetAll(ctx)
	if err != nil {
		log15.Error("Error migrating user_activity data to persistent redis cache", "error", err)
		return
	}

	for _, userActivity := range allUserActivity {
		userIDStr := strconv.Itoa(int(userActivity.UserID))
		key := keyPrefix + userIDStr
		c.Send("HMSET", key,
			fPageViews, userActivity.PageViews,
			fSearchQueries, userActivity.SearchQueries)
	}

	c.Send("SET", migrateKey, strconv.FormatBool(true))
}

// GetByUserID returns a single user's UserActivity.
func GetByUserID(userID int32) (*types.UserActivity, error) {
	userIDStr := strconv.Itoa(int(userID))
	key := keyPrefix + userIDStr

	c := pool.Get()
	values, err := redis.Values(c.Do("HMGET", key, fPageViews, fSearchQueries, fLastActive))
	c.Close()
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	var lastActiveStr string
	a := &types.UserActivity{
		UserID: userID,
	}
	_, err = redis.Scan(values, &a.PageViews, &a.SearchQueries, &lastActiveStr)
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	if lastActiveStr != "" {
		t, err := time.Parse(time.RFC3339, lastActiveStr)
		if err != nil {
			return nil, err
		}
		a.LastPageViewTime = &t
	}

	return a, nil
}

// GetUsersActiveTodayCount returns a count of users that have had pageviews
// so far today.
func GetUsersActiveTodayCount() (int, error) {
	c := pool.Get()
	defer c.Close()

	count, err := redis.Int(c.Do("SCARD", usersActiveKey()))
	if err == redis.ErrNil {
		err = nil
	}
	return count, err
}

// LogPageView increments a user's pageview count.
func LogPageView(isAuthenticated bool, userID int32, userCookieID string) error {
	c := pool.Get()
	defer c.Close()

	uniqueID := userCookieID

	// If the user is authenticated, store their activity in the appropriate user ID-keyed cache.
	if isAuthenticated {
		userIDStr := strconv.Itoa(int(userID))
		uniqueID = userIDStr
		key := keyPrefix + uniqueID

		// Increment the user's pageview count.
		if err := c.Send("HINCRBY", key, fPageViews, 1); err != nil {
			return err
		}
		if err := c.Send("HSET", key, fLastActive, time.Now().Format(time.RFC3339)); err != nil {
			return err
		}
	}

	// Regardless of authenicatation status, add the user's unique ID to the set of active users.
	return c.Send("SADD", usersActiveKey(), uniqueID)
}

// LogSearchQuery increments a user's search query count.
func LogSearchQuery(userID int32) error {
	key := keyPrefix + strconv.Itoa(int(userID))
	c := pool.Get()
	defer c.Close()

	return c.Send("HINCRBY", key, fSearchQueries, 1)
}

func usersActiveKey() string {
	return keyPrefix + ":usersactive:" + time.Now().Format("2006-01-02")
}

// gc expires active user sets after a week. We only use the current day, but
// a weeks worth of data may be useful in the future for analysis.
func gc() {
	for {
		key := usersActiveKey()

		c := pool.Get()
		err := c.Send("EXPIRE", key, 60*60*24*7) // 1 week
		c.Close()

		if err != nil {
			log15.Warn("EXPIRE failed", "key", key, "error", err)
		}

		jitter := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(time.Hour + jitter)
	}
}
