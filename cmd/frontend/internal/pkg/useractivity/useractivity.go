// Package useractivity provides an interface to update and access information about
// individual and aggregate Sourcegraph Server users' activity levels.
//
// Note that this package should not be used on sourcegraph.com, only on self-hosted
// deployments.
package useractivity

import (
	"context"
	"fmt"
	"math"
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

	timeNow = time.Now
)

const (
	fPageViews                     = "pageviews"
	fLastActive                    = "lastactive"
	fSearchQueries                 = "searchqueries"
	fCodeIntelActions              = "codeintelactions"
	fLastActiveCodeHostIntegration = "lastactivecodehostintegration"

	daysOfHistory   = 14
	weeksOfHistory  = 10
	monthsOfHistory = 3
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
	values, err := redis.Values(c.Do("HMGET", key, fPageViews, fSearchQueries, fLastActive, fCodeIntelActions, fLastActiveCodeHostIntegration))
	c.Close()
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	var lastActiveStr, lastActiveCodeHostStr string
	a := &types.UserActivity{
		UserID: userID,
	}
	_, err = redis.Scan(values, &a.PageViews, &a.SearchQueries, &lastActiveStr, &a.CodeIntelligenceActions, &lastActiveCodeHostStr)
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	if lastActiveStr != "" {
		t, err := time.Parse(time.RFC3339, lastActiveStr)
		if err != nil {
			return nil, err
		}
		a.LastActiveTime = &t
	}

	if lastActiveCodeHostStr != "" {
		t, err := time.Parse(time.RFC3339, lastActiveCodeHostStr)
		if err != nil {
			return nil, err
		}
		a.LastCodeHostIntegrationTime = &t
	}

	return a, nil
}

type SiteActivityOptions struct {
	Days   int
	Weeks  int
	Months int
}

func minIntOrZero(a, b int) int {
	min := b
	if a < b {
		min = a
	}
	if min < 0 {
		return 0
	}
	return min
}

// GetSiteActivity returns the current site's SiteActivity
func GetSiteActivity(opt *SiteActivityOptions) (*types.SiteActivity, error) {
	var (
		days   = daysOfHistory
		weeks  = weeksOfHistory
		months = monthsOfHistory
	)

	if opt != nil {
		days = minIntOrZero(opt.Days, daysOfHistory)
		weeks = minIntOrZero(opt.Weeks, weeksOfHistory)
		months = minIntOrZero(opt.Months, monthsOfHistory)
	}

	daus, err := DAUs(days)
	if err != nil {
		return nil, err
	}
	waus, err := WAUs(weeks)
	if err != nil {
		return nil, err
	}
	maus, err := MAUs(months)
	if err != nil {
		return nil, err
	}
	return &types.SiteActivity{
		DAUs: daus,
		WAUs: waus,
		MAUs: maus,
	}, nil
}

// GetUsersActiveTodayCount returns a count of users that have been active today.
func GetUsersActiveTodayCount() (int, error) {
	c := pool.Get()
	defer c.Close()

	count, err := redis.Int(c.Do("SCARD", usersActiveKeyFromDaysAgo(0)))
	if err == redis.ErrNil {
		err = nil
	}
	return count, err
}

// calcUniques calculates the number of unique users starting at 00:00:00 on a given UTC date over a
// period of time (years, months, and days).
func calcUniques(c redis.Conn, dayStart time.Time, years int, months int, days int) (*types.SiteActivityPeriod, error) {
	allUniqueUserIds := map[string]bool{}
	registeredUserIds := map[string]bool{}

	dayStart = time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC)
	d := dayStart.AddDate(years, months, days).AddDate(0, 0, -1)

	// Start at the end date, and loop backwards until reaching the start
	for d.After(dayStart) || d.Equal(dayStart) {
		values, err := redis.Values(c.Do("SMEMBERS", usersActiveKeyFromDate(d)))
		if err != nil && err != redis.ErrNil {
			return nil, err
		}
		for _, id := range values {
			bid := id.([]byte)
			sid := string(bid)
			allUniqueUserIds[sid] = true
			if len(bid) != 36 { // id is a numerical Sourcegraph user id, not an anonymous user's Telligent cookie id
				registeredUserIds[sid] = true
			}
		}

		d = d.AddDate(0, 0, -1)
	}

	return &types.SiteActivityPeriod{
		StartTime:           dayStart,
		UserCount:           int32(len(allUniqueUserIds)),
		RegisteredUserCount: int32(len(registeredUserIds)),
		AnonymousUserCount:  int32(len(allUniqueUserIds)) - int32(len(registeredUserIds)),
	}, nil
}

// DAUs returns a count of daily active users for the last daysCount days (including the current, partial day).
func DAUs(daysCount int) ([]*types.SiteActivityPeriod, error) {
	c := pool.Get()
	defer c.Close()

	var daus []*types.SiteActivityPeriod
	now := timeNow().UTC()
	for daysAgo := 0; daysAgo < daysCount; daysAgo++ {
		uniques, err := calcUniques(c, now.AddDate(0, 0, -daysAgo), 0, 0, 1)
		if err != nil {
			return nil, err
		}
		daus = append(daus, uniques)
	}
	return daus, nil
}

// WAUs returns a count of daily active users for the last weeksCount calendar weeks (including the current, partial week).
func WAUs(weeksCount int) ([]*types.SiteActivityPeriod, error) {
	c := pool.Get()
	defer c.Close()

	var waus []*types.SiteActivityPeriod

	for w := 0; w < weeksCount; w++ {
		weekStartDate := startOfWeek(w)
		uniques, err := calcUniques(c, weekStartDate, 0, 0, 7)
		if err != nil {
			return nil, err
		}
		waus = append(waus, uniques)
	}
	return waus, nil
}

// MAUs returns a count of daily active users for the last monthsCount calendar months (including the current, partial month).
func MAUs(monthsCount int) ([]*types.SiteActivityPeriod, error) {
	c := pool.Get()
	defer c.Close()

	var maus []*types.SiteActivityPeriod

	for m := 0; m < monthsCount; m++ {
		monthStartDate := startOfMonth(m)
		uniques, err := calcUniques(c, monthStartDate, 0, 1, 0)
		if err != nil {
			return nil, err
		}
		maus = append(maus, uniques)
	}
	return maus, nil
}

// logPageView increments a user's pageview count.
func logPageView(userID int32) error {
	key := keyPrefix + strconv.Itoa(int(userID))
	c := pool.Get()
	defer c.Close()

	return c.Send("HINCRBY", key, fPageViews, 1)
}

// logSearchQuery increments a user's search query count.
func logSearchQuery(userID int32) error {
	key := keyPrefix + strconv.Itoa(int(userID))
	c := pool.Get()
	defer c.Close()

	return c.Send("HINCRBY", key, fSearchQueries, 1)
}

// logCodeIntel increments a user's code intelligence usage count.
func logCodeIntelAction(userID int32) error {
	key := keyPrefix + strconv.Itoa(int(userID))
	c := pool.Get()
	defer c.Close()

	return c.Send("HINCRBY", key, fCodeIntelActions, 1)
}

// logCodeHostIntegrationUsage logs the last time a user was active on a code host integration
func logCodeHostIntegrationUsage(userID int32) error {
	key := keyPrefix + strconv.Itoa(int(userID))
	c := pool.Get()
	defer c.Close()

	now := timeNow().UTC()
	return c.Send("HSET", key, fLastActiveCodeHostIntegration, now.Format(time.RFC3339))
}

// LogActivity logs any user activity (page view, integration usage, etc) to their "last active" time, and
// adds their unique ID to the set of active users
func LogActivity(isAuthenticated bool, userID int32, userCookieID string, event string) error {
	c := pool.Get()
	defer c.Close()

	uniqueID := userCookieID

	// If the user is authenticated, set uniqueID to their user ID, and store their "last active time" in the
	// appropriate user ID-keyed cache.
	if isAuthenticated {
		userIDStr := strconv.Itoa(int(userID))
		uniqueID = userIDStr
		key := keyPrefix + uniqueID

		// Set the user's last active time
		now := timeNow().UTC()
		if err := c.Send("HSET", key, fLastActive, now.Format(time.RFC3339)); err != nil {
			return err
		}
	}

	if uniqueID == "" {
		log15.Warn("useractivity.LogActivity: no user ID provided")
		return nil
	}

	// Regardless of authenicatation status, add the user's unique ID to the set of active users.
	if err := c.Send("SADD", usersActiveKeyFromDaysAgo(0), uniqueID); err != nil {
		return err
	}

	// If the user isn't authenticated, return at this point and don't record user-level properties.
	if !isAuthenticated {
		return nil
	}

	switch event {
	case "SEARCHQUERY":
		return logSearchQuery(userID)
	case "PAGEVIEW":
		return logPageView(userID)
	case "CODEINTEL":
		return logCodeIntelAction(userID)
	case "CODEINTELINTEGRATION":
		if err := logCodeHostIntegrationUsage(userID); err != nil {
			return err
		}
		return logCodeIntelAction(userID)
	}
	return fmt.Errorf("unknown user event %s", event)
}

func usersActiveKeyFromDate(date time.Time) string {
	return keyPrefix + ":usersactive:" + time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

func usersActiveKeyFromDaysAgo(daysAgo int) string {
	now := timeNow().UTC()
	return usersActiveKeyFromDate(now.AddDate(0, 0, -daysAgo))
}

func startOfWeek(weeksAgo int) time.Time {
	if weeksAgo > 0 {
		return startOfWeek(0).AddDate(0, 0, -7*weeksAgo)
	}

	// If weeksAgo == 0, start at timeNow(), and loop back by day until we hit a Sunday
	now := timeNow().UTC()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}

func startOfMonth(monthsAgo int) time.Time {
	now := timeNow().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -monthsAgo, 0)
}

// gc expires active user sets after the max of daysOfHistory, weeksOfHistory, and monthsOfHistory have passed.
func gc() {
	for {
		key := usersActiveKeyFromDaysAgo(0)

		c := pool.Get()
		err := c.Send("EXPIRE", key, 60*60*24*int(math.Max(31*monthsOfHistory, math.Max(7*weeksOfHistory, daysOfHistory))))
		c.Close()

		if err != nil {
			log15.Warn("EXPIRE failed", "key", key, "error", err)
		}

		jitter := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(time.Hour + jitter)
	}
}
