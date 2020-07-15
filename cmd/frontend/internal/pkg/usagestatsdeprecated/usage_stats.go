// Package usagestatsdeprecated is deprecated in favor of package usagestats.
//
// Package usagestatsdeprecated provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
//
// Note that this package should not be used on sourcegraph.com, only on self-hosted
// deployments.
package usagestatsdeprecated

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	gcOnce sync.Once // ensures we only have 1 redis gc goroutine

	pool = redispool.Store

	timeNow = time.Now
)

const (
	defaultDays   = 14
	defaultWeeks  = 10
	defaultMonths = 3

	maxStorageDays = 93
)

var MockGetByUserID func(userID int32) (*types.UserUsageStatistics, error)

// GetByUserID returns a single user's UserUsageStatistics.
func GetByUserID(userID int32) (*types.UserUsageStatistics, error) {
	if MockGetByUserID != nil {
		return MockGetByUserID(userID)
	}

	userIDStr := strconv.Itoa(int(userID))
	key := keyPrefix + userIDStr

	c := pool.Get()
	values, err := redis.Values(c.Do("HMGET", key, fPageViews, fSearchQueries, fLastActive, fCodeIntelActions, fFindRefsActions, fLastActiveCodeHostIntegration))
	c.Close()
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	var lastActiveStr, lastActiveCodeHostStr string
	a := &types.UserUsageStatistics{
		UserID: userID,
	}
	_, err = redis.Scan(values, &a.PageViews, &a.SearchQueries, &lastActiveStr, &a.CodeIntelligenceActions, &a.FindReferencesActions, &lastActiveCodeHostStr)
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

// SiteUsageStatisticsOptions contains options for the number of daily, weekly, and monthly periods in
// which to calculate the number of unique users (i.e., how many days of Daily Active Users, or DAUs,
// how many weeks of Weekly Active Users, or WAUs, and how many months of Monthly Active Users, or MAUs).
type SiteUsageStatisticsOptions struct {
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

// GetSiteUsageStatistics returns the current site's SiteActivity.
func GetSiteUsageStatistics(opt *SiteUsageStatisticsOptions) (*types.SiteUsageStatistics, error) {
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
	return &types.SiteUsageStatistics{
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

func HasSearchOccurred() (bool, error) {
	c := pool.Get()
	defer c.Close()
	s, err := redis.Bool(c.Do("GET", keyPrefix+fSearchOccurred))
	if err != nil && err != redis.ErrNil {
		return s, err
	}
	return s, nil
}

func HasFindRefsOccurred() (bool, error) {
	c := pool.Get()
	defer c.Close()
	r, err := redis.Bool(c.Do("GET", keyPrefix+fFindRefsOccurred))
	if err != nil && err != redis.ErrNil {
		return r, err
	}
	return r, nil
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
	nextValue:
		for _, id := range values {
			sid := string(id.([]byte))
			allUniqueUserIDs[sid] = true

			// If any character is not a digit, then treat it as an anonymous user ID.
			for _, s := range sid {
				if s < '0' || s > '9' {
					anonymousUserIDs[sid] = true
					continue nextValue
				}
			}
			registeredUserIDs[sid] = true
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
func uniquesCount(dayStart time.Time, period *UsageDuration, calcStages bool) (*types.SiteActivityPeriod, error) {
	userIDs, err := uniques(dayStart, period)
	if err != nil {
		return nil, err
	}

	var stages *types.Stages
	if calcStages {
		stages, err = stageUniques(dayStart, period, userIDs.Registered)
		if err != nil {
			return nil, err
		}
	}

	return &types.SiteActivityPeriod{
		StartTime:            time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC),
		UserCount:            int32(len(userIDs.All)),
		RegisteredUserCount:  int32(len(userIDs.Registered)),
		AnonymousUserCount:   int32(len(userIDs.Anonymous)),
		IntegrationUserCount: int32(len(userIDs.UsedIntegrations)),
		Stages:               stages,
	}, nil
}

// daus returns a count of daily active users for the last daysCount days (including the current, partial day).
func daus(dayPeriods int) ([]*types.SiteActivityPeriod, error) {
	var daus []*types.SiteActivityPeriod
	now := timeNow().UTC()
	for daysAgo := 0; daysAgo < dayPeriods; daysAgo++ {
		uniques, err := uniquesCount(now.AddDate(0, 0, -daysAgo), &UsageDuration{Days: 1}, false)
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
		uniques, err := uniquesCount(weekStartDate, &UsageDuration{Days: 7}, w == 0)
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
		uniques, err := uniquesCount(monthStartDate, &UsageDuration{Months: 1}, m == 0)
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

var MockStageUniques func(dayStart time.Time, period *UsageDuration, registeredActives []string) (*types.Stages, error)

func stageUniques(dayStart time.Time, period *UsageDuration, registeredActives []string) (*types.Stages, error) {
	if MockStageUniques != nil {
		return MockStageUniques(dayStart, period, registeredActives)
	}

	ctx := context.Background()

	var (
		manageUserIDs    = map[string]bool{}
		planUserIDs      map[string]bool // none currently
		codeUserIDs      map[string]bool
		reviewUserIDs    map[string]bool
		verifyUserIDs    map[string]bool
		packageUserIDs   map[string]bool // none currently
		deployUserIDs    map[string]bool // none currently
		configureUserIDs map[string]bool // none currently
		monitorUserIDs   map[string]bool
		secureUserIDs    map[string]bool // none currently
		automateUserIDs  map[string]bool // none currently
	)

	dayStart = time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, period.Months, period.Days)

	var activeUserIDs []int32
	for _, userID := range registeredActives {
		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			return nil, err
		}
		activeUserIDs = append(activeUserIDs, int32(userIDInt))
	}

	//// MANAGE ////
	// 1) any activity from a site admin
	// 2) any usage of an API access token

	// Loop through all active registered users, see if any are admins
	users, err := db.Users.List(ctx, &db.UsersListOptions{UserIDs: activeUserIDs})
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.SiteAdmin {
			manageUserIDs[strconv.Itoa(int(user.ID))] = true
		}
	}
	// Loop through all access tokens used in the past week
	tokens, err := db.AccessTokens.List(ctx, db.AccessTokensListOptions{LastUsedAfter: &dayStart, LastUsedBefore: &dayEnd})
	if err != nil {
		return nil, err
	}
	for _, token := range tokens {
		manageUserIDs[strconv.Itoa(int(token.CreatorUserID))] = true
	}

	//// PLAN ////
	// none currently

	//// CODE ////
	// 1) any searches
	// 2) any file, repo, tree views
	// 3) TODO(Dan): any code host integration usage (other than for code review)

	// Loop through all active user IDs, see if any executed searches in the window
	codeUserIDs, err = usersSince(registeredActives, dayStart, keyFromStage("STAGECODE"))
	if err != nil {
		return nil, err
	}

	//// REVIEW ////
	// 1) TODO(Dan): code host integration usage for code review

	reviewUserIDs, err = usersSince(registeredActives, dayStart, keyFromStage("STAGEREVIEW"))
	if err != nil {
		return nil, err
	}

	//// VERIFY ////
	// 1) receiving a saved search notification (email)
	// 2) TODO(Dan): receiving a saved search notification (slack)
	// 3) clicking a saved search notification (email or slack)
	// 4) TODO(Dan): having a saved search defined in your user or org settings
	verifyUserIDs, err = usersSince(registeredActives, dayStart, keyFromStage("STAGEVERIFY"))
	if err != nil {
		return nil, err
	}

	//// PACKAGE ////
	// none currently

	//// DEPLOY ////
	// none currently

	//// CONFIGURE ////
	// none currently

	//// MONITOR ////
	// 1) running a diff search
	// 2) TODO(Dan): monitoring extension enabled (e.g. LightStep, Sentry, Datadog)
	monitorUserIDs, err = usersSince(registeredActives, dayStart, keyFromStage("STAGEMONITOR"))
	if err != nil {
		return nil, err
	}

	//// SECURE ////
	// none currently

	//// AUTOMATE ////
	// none currently

	return &types.Stages{
		Manage:    int32(len(keys(manageUserIDs))),
		Plan:      int32(len(keys(planUserIDs))),
		Code:      int32(len(keys(codeUserIDs))),
		Review:    int32(len(keys(reviewUserIDs))),
		Verify:    int32(len(keys(verifyUserIDs))),
		Package:   int32(len(keys(packageUserIDs))),
		Deploy:    int32(len(keys(deployUserIDs))),
		Configure: int32(len(keys(configureUserIDs))),
		Monitor:   int32(len(keys(monitorUserIDs))),
		Secure:    int32(len(keys(secureUserIDs))),
		Automate:  int32(len(keys(automateUserIDs))),
	}, nil
}

func usersSince(userList []string, dayStart time.Time, userField string) (map[string]bool, error) {
	c := pool.Get()
	defer c.Close()

	userIDs := map[string]bool{}
	for _, uid := range userList {
		userKey := keyPrefix + uid
		err := c.Send("HGET", userKey, userField)
		if err != nil && err != redis.ErrNil {
			return nil, err
		}
	}
	c.Flush()
	for _, uid := range userList {
		lastEvent, err := redis.String(c.Receive())
		if err != nil && err != redis.ErrNil {
			return nil, err
		}
		if lastEvent != "" {
			t, err := time.Parse(time.RFC3339, lastEvent)
			if err != nil {
				return nil, err
			}
			if t.After(dayStart) || t.Equal(dayStart) {
				userIDs[uid] = true
			}
		}
	}
	return userIDs, nil
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
