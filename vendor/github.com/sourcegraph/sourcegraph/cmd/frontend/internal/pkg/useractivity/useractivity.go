// Package useractivity provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
//
// Note that this package should not be used on sourcegraph.com, only on self-hosted
// deployments.
package useractivity

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/redispool"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	gcOnce sync.Once // ensures we only have 1 redis gc goroutine

	keyPrefix = "user_activity:"
	pool      = redispool.Store

	timeNow = time.Now
)

const (
	fPageViews                     = "pageviews"
	fLastActive                    = "lastactive"
	fSearchQueries                 = "searchqueries"
	fCodeIntelActions              = "codeintelactions"
	fLastActiveCodeHostIntegration = "lastactivecodehostintegration"

	defaultDays   = 14
	defaultWeeks  = 10
	defaultMonths = 3

	maxStorageDays = 93
)

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

// SiteActivityOptions contains options for the number of daily, weekly, and monthly periods in
// which to calculate the number of unique users (i.e., how many days of Daily Active Users, or DAUs,
// how many weeks of Weekly Active Users, or WAUs, and how many months of Monthly Active Users, or MAUs).
type SiteActivityOptions struct {
	DayPeriods   *int
	WeekPeriods  *int
	MonthPeriods *int
}

// UsageDuration in aggregate represents a duration of time over which to calculate a set of unique users.
type UsageDuration struct {
	Days   int
	Months int
}

// ActiveUsers contains sets of unique user IDs.
type ActiveUsers struct {
	All              []string
	Registered       []string
	Anonymous        []string
	UsedIntegrations []string
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

// GetSiteActivity returns the current site's SiteActivity.
func GetSiteActivity(opt *SiteActivityOptions) (*types.SiteActivity, error) {
	var (
		dayPeriods   = defaultDays
		weekPeriods  = defaultWeeks
		monthPeriods = defaultMonths
	)

	if opt != nil {
		if opt.DayPeriods != nil {
			dayPeriods = minIntOrZero(maxStorageDays, *opt.DayPeriods)
		}
		if opt.WeekPeriods != nil {
			weekPeriods = minIntOrZero(maxStorageDays/7, *opt.WeekPeriods)
		}
		if opt.MonthPeriods != nil {
			monthPeriods = minIntOrZero(maxStorageDays/31, *opt.MonthPeriods)
		}
	}

	daus, err := daus(dayPeriods)
	if err != nil {
		return nil, err
	}
	waus, err := waus(weekPeriods)
	if err != nil {
		return nil, err
	}
	maus, err := maus(monthPeriods)
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

// uniques calculates the list of unique users starting at 00:00:00 on a given UTC date over a
// period of time.
func uniques(dayStart time.Time, period *UsageDuration) (*ActiveUsers, error) {
	c := pool.Get()
	defer c.Close()

	var (
		allUniqueUserIDs   = map[string]bool{}
		registeredUserIDs  = map[string]bool{}
		anonymousUserIDs   = map[string]bool{}
		integrationUserIDs = map[string]bool{}
	)

	dayStart = time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, period.Months, period.Days)

	// Start at 00:00:00 UTC of the last day, and loop backwards until reaching the start
	d := dayEnd.AddDate(0, 0, -1)
	for d.After(dayStart) || d.Equal(dayStart) {
		values, err := redis.Values(c.Do("SMEMBERS", usersActiveKeyFromDate(d)))
		if err != nil && err != redis.ErrNil {
			return nil, err
		}
		for _, id := range values {
			bid := id.([]byte)
			sid := string(bid)
			allUniqueUserIDs[sid] = true
			if len(bid) != 36 { // id is a numerical Sourcegraph user id, not an anonymous user's UUID
				registeredUserIDs[sid] = true
			} else {
				anonymousUserIDs[sid] = true
			}
		}

		d = d.AddDate(0, 0, -1)
	}

	// Loop through all id'd unique users, determine if they were active in a code host integration in the time period.
	//
	// Despite O(n) Redis requests (n == # of users in a given period), this performs acceptably well at large instances.
	// On an instance with 25K users, average GraphQL requests for default site activity data (14 days of DAUs, 10 weeks
	// of WAUs, and 3 months of MAUs) takes ~1.2s.
	for uid := range allUniqueUserIDs {
		userKey := keyPrefix + uid
		err := c.Send("HGET", userKey, fLastActiveCodeHostIntegration)
		if err != nil && err != redis.ErrNil {
			return nil, err
		}
	}
	c.Flush()
	for uid := range allUniqueUserIDs {
		lastActiveCodeHostStr, err := redis.String(c.Receive())
		if err != nil && err != redis.ErrNil {
			return nil, err
		}
		if lastActiveCodeHostStr != "" {
			t, err := time.Parse(time.RFC3339, lastActiveCodeHostStr)
			if err != nil {
				return nil, err
			}
			if (t.After(dayStart) || t.Equal(dayStart)) && t.Before(dayEnd) {
				integrationUserIDs[uid] = true
			}
		}
	}

	return &ActiveUsers{
		All:              keys(allUniqueUserIDs),
		Registered:       keys(registeredUserIDs),
		Anonymous:        keys(anonymousUserIDs),
		UsedIntegrations: keys(integrationUserIDs),
	}, nil
}

// uniquesCount calculates the number of unique users starting at 00:00:00 on a given UTC date over a
// period of time (years, months, and days).
func uniquesCount(dayStart time.Time, period *UsageDuration) (*types.SiteActivityPeriod, error) {
	userIDs, err := uniques(dayStart, period)
	if err != nil {
		return nil, err
	}

	return &types.SiteActivityPeriod{
		StartTime:            time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC),
		UserCount:            int32(len(userIDs.All)),
		RegisteredUserCount:  int32(len(userIDs.Registered)),
		AnonymousUserCount:   int32(len(userIDs.Anonymous)),
		IntegrationUserCount: int32(len(userIDs.UsedIntegrations)),
	}, nil
}

// daus returns a count of daily active users for the last daysCount days (including the current, partial day).
func daus(dayPeriods int) ([]*types.SiteActivityPeriod, error) {
	var daus []*types.SiteActivityPeriod
	now := timeNow().UTC()
	for daysAgo := 0; daysAgo < dayPeriods; daysAgo++ {
		uniques, err := uniquesCount(now.AddDate(0, 0, -daysAgo), &UsageDuration{Days: 1})
		if err != nil {
			return nil, err
		}
		daus = append(daus, uniques)
	}
	return daus, nil
}

// ListUsersToday returns a list of users active since today at 00:00 UTC.
func ListUsersToday() (*ActiveUsers, error) {
	return uniques(timeNow().UTC(), &UsageDuration{Days: 1})
}

// waus returns a count of daily active users for the last weeksCount calendar weeks (including the current, partial week).
func waus(weekPeriods int) ([]*types.SiteActivityPeriod, error) {
	var waus []*types.SiteActivityPeriod

	for w := 0; w < weekPeriods; w++ {
		weekStartDate := startOfWeek(w)
		uniques, err := uniquesCount(weekStartDate, &UsageDuration{Days: 7})
		if err != nil {
			return nil, err
		}
		waus = append(waus, uniques)
	}
	return waus, nil
}

// ListUsersThisWeek returns a list of users active since the latest Monday at 00:00 UTC.
func ListUsersThisWeek() (*ActiveUsers, error) {
	weekStartDate := startOfWeek(0)
	return uniques(weekStartDate, &UsageDuration{Days: 7})
}

// maus returns a count of daily active users for the last monthsCount calendar months (including the current, partial month).
func maus(monthPeriods int) ([]*types.SiteActivityPeriod, error) {
	var maus []*types.SiteActivityPeriod

	for m := 0; m < monthPeriods; m++ {
		monthStartDate := startOfMonth(m)
		uniques, err := uniquesCount(monthStartDate, &UsageDuration{Months: 1})
		if err != nil {
			return nil, err
		}
		maus = append(maus, uniques)
	}
	return maus, nil
}

// ListUsersThisMonth returns a list of users active since the first day of the month at 00:00 UTC.
func ListUsersThisMonth() (*ActiveUsers, error) {
	monthStartDate := startOfMonth(0)
	return uniques(monthStartDate, &UsageDuration{Months: 1})
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
	// Setup our GC of active key goroutine
	gcOnce.Do(func() {
		go gc()
	})

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

func keys(m map[string]bool) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

// gc expires active user sets after the max of daysOfHistory, weeksOfHistory, and monthsOfHistory have passed.
func gc() {
	for {
		key := usersActiveKeyFromDaysAgo(0)

		c := pool.Get()
		err := c.Send("EXPIRE", key, 60*60*24*int(maxStorageDays))
		c.Close()

		if err != nil {
			log15.Warn("EXPIRE failed", "key", key, "error", err)
		}

		jitter := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(time.Hour + jitter)
	}
}
