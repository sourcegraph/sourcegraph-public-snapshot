// Package usagestats2 provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
package usagestats2

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

const (
	defaultDays   = 14
	defaultWeeks  = 10
	defaultMonths = 3

	maxStorageDays = 93
)

var (
	timeNow = time.Now
)

var MockGetByUserID func(userID int32) (*types.UserUsageStatistics, error)

// GetByUserID returns a single user's UserUsageStatistics.
func GetByUserID(ctx context.Context, userID int32) (*types.UserUsageStatistics, error) {
	if MockGetByUserID != nil {
		return MockGetByUserID(userID)
	}

	pageViews, err := db.EventLogs.CountByUserIDAndEventNamePrefix(ctx, userID, "View")
	if err != nil {
		return nil, err
	}
	searchQueries, err := db.EventLogs.CountByUserIDAndEventName(ctx, userID, "SearchSubmitted")
	if err != nil {
		return nil, err
	}
	codeIntelligenceActions, err := db.EventLogs.CountByUserIDAndEventNames(ctx, userID, []string{"hover", "findReferences", "goToDefinition.preloaded", "goToDefinition"})
	if err != nil {
		return nil, err
	}
	findReferencesActions, err := db.EventLogs.CountByUserIDAndEventName(ctx, userID, "findReferences")
	if err != nil {
		return nil, err
	}
	lastActiveTime, err := db.EventLogs.MaxTimestampByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	lastCodeHostIntegrationTime, err := db.EventLogs.MaxTimestampByUserIDAndSource(ctx, userID, "CODEHOSTINTEGRATION")
	if err != nil {
		return nil, err
	}
	return &types.UserUsageStatistics{
		UserID:                      userID,
		PageViews:                   int32(pageViews),
		SearchQueries:               int32(searchQueries),
		CodeIntelligenceActions:     int32(codeIntelligenceActions),
		FindReferencesActions:       int32(findReferencesActions),
		LastActiveTime:              lastActiveTime,
		LastCodeHostIntegrationTime: lastCodeHostIntegrationTime,
	}, nil
}

// GetUsersActiveTodayCount returns a count of users that have been active today.
func GetUsersActiveTodayCount(ctx context.Context) (int, error) {
	now := timeNow().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return db.EventLogs.CountUniquesAll(ctx, today, today.AddDate(0, 0, 1))
}

// ListRegisteredUsersToday returns a list of the registered users that were active today.
func ListRegisteredUsersToday(ctx context.Context) ([]int32, error) {
	now := timeNow().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return db.EventLogs.ListUniquesAll(ctx, start, start.AddDate(0, 0, 1))
}

// ListRegisteredUsersThisWeek returns a list of the registered users that were active this week.
func ListRegisteredUsersThisWeek(ctx context.Context) ([]int32, error) {
	start := startOfWeek(0)
	return db.EventLogs.ListUniquesAll(ctx, start, start.AddDate(0, 0, 7))
}

// ListRegisteredUsersThisMonth returns a list of the registered users that were active this month.
func ListRegisteredUsersThisMonth(ctx context.Context) ([]int32, error) {
	now := timeNow().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return db.EventLogs.ListUniquesAll(ctx, start, start.AddDate(0, 1, 0))
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

// GetSiteUsageStatistics returns the current site's SiteActivity.
func GetSiteUsageStatistics(ctx context.Context, opt *SiteUsageStatisticsOptions) (*types.SiteUsageStatistics, error) {
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

	daus, err := daus(ctx, dayPeriods)
	if err != nil {
		return nil, err
	}
	waus, err := waus(ctx, weekPeriods)
	if err != nil {
		return nil, err
	}
	maus, err := maus(ctx, monthPeriods)
	if err != nil {
		return nil, err
	}
	return &types.SiteUsageStatistics{
		DAUs: daus,
		WAUs: waus,
		MAUs: maus,
	}, nil
}

// daus returns a count of daily active users for the last daysCount days (including the current, partial day).
func daus(ctx context.Context, periods int) ([]*types.SiteActivityPeriod, error) {
	if periods == 0 {
		return []*types.SiteActivityPeriod{}, nil
	}
	var daus []*types.SiteActivityPeriod
	now := timeNow().UTC()
	current := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	p := (periods - 1)
	startDate := current.AddDate(0, 0, -p)
	uniques, err := db.EventLogs.CountDAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	registeredUniques, err := db.EventLogs.CountRegisteredDAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	integrationUniques, err := db.EventLogs.CountIntegrationDAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	for i, u := range uniques {
		actPer := &types.SiteActivityPeriod{
			StartTime:            u.Start,
			UserCount:            int32(u.Count),
			RegisteredUserCount:  int32(registeredUniques[i].Count),
			AnonymousUserCount:   int32(u.Count - registeredUniques[i].Count),
			IntegrationUserCount: int32(integrationUniques[i].Count),
			Stages:               nil,
		}
		daus = append(daus, actPer)
	}
	return daus, nil
}

// waus returns a count of weekly active users for the last weekssCount weeks (including the current, partial week).
func waus(ctx context.Context, periods int) ([]*types.SiteActivityPeriod, error) {
	if periods == 0 {
		return []*types.SiteActivityPeriod{}, nil
	}
	var waus []*types.SiteActivityPeriod
	current := startOfWeek(0)
	p := (periods - 1)
	startDate := current.AddDate(0, 0, -p*7)
	uniques, err := db.EventLogs.CountWAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	registeredUniques, err := db.EventLogs.CountRegisteredWAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	integrationUniques, err := db.EventLogs.CountIntegrationWAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	for i, u := range uniques {
		actPer := &types.SiteActivityPeriod{
			StartTime:            u.Start,
			UserCount:            int32(u.Count),
			RegisteredUserCount:  int32(registeredUniques[i].Count),
			AnonymousUserCount:   int32(u.Count - registeredUniques[i].Count),
			IntegrationUserCount: int32(integrationUniques[i].Count),
			Stages:               nil,
		}
		waus = append(waus, actPer)
	}

	// Add stage uniques For the latest week only.
	waus[0].Stages, err = stageUniques(current, &UsageDuration{Months: 0, Days: 7})
	if err != nil {
		return nil, err
	}

	return waus, nil
}

// maus returns a count of monthly active users for the last monthsCount months (including the current, partial month).
func maus(ctx context.Context, periods int) ([]*types.SiteActivityPeriod, error) {
	if periods == 0 {
		return []*types.SiteActivityPeriod{}, nil
	}
	var maus []*types.SiteActivityPeriod
	now := timeNow().UTC()
	current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	p := periods - 1
	startDate := current.AddDate(0, -p, 0)
	uniques, err := db.EventLogs.CountMAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	registeredUniques, err := db.EventLogs.CountRegisteredMAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	integrationUniques, err := db.EventLogs.CountIntegrationMAUs(ctx, startDate, p)
	if err != nil {
		return nil, err
	}
	for i, u := range uniques {
		actPer := &types.SiteActivityPeriod{
			StartTime:            u.Start,
			UserCount:            int32(u.Count),
			RegisteredUserCount:  int32(registeredUniques[i].Count),
			AnonymousUserCount:   int32(u.Count - registeredUniques[i].Count),
			IntegrationUserCount: int32(integrationUniques[i].Count),
			Stages:               nil,
		}
		maus = append(maus, actPer)
	}

	// Add stage uniques For the latest month only.
	maus[0].Stages, err = stageUniques(current, &UsageDuration{Months: 1, Days: 0})
	if err != nil {
		return nil, err
	}

	return maus, nil
}

var MockStageUniques func(dayStart time.Time, period *UsageDuration) (*types.Stages, error)

func stageUniques(dayStart time.Time, period *UsageDuration) (*types.Stages, error) {
	if MockStageUniques != nil {
		return MockStageUniques(dayStart, period)
	}

	ctx := context.Background()
	dayStart = time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, period.Months, period.Days)

	//// MANAGE ////
	// 1) any activity from a site admin
	// 2) any usage of an API access token

	manageUniques, err := db.EventLogs.CountUniquesByEventNamePrefix(ctx, dayStart, dayEnd, "ViewSiteAdmin")
	if err != nil {
		return nil, err
	}

	//// PLAN ////
	// none currently

	//// CODE ////
	// 1) any searches
	// 2) any file, repo, tree views
	// 3) TODO(Dan): any code host integration usage (other than for code review)

	codeUniques, err := db.EventLogs.CountUniquesByEventNames(ctx, dayStart, dayEnd, []string{"ViewRepository", "ViewBlob", "ViewTree", "SearchResultsQueried"})
	if err != nil {
		return nil, err
	}

	//// REVIEW ////
	// 1) TODO(Dan): code host integration usage for code review

	//// VERIFY ////
	// 1) receiving a saved search notification (email)
	// 2) TODO(Dan): receiving a saved search notification (slack)
	// 3) clicking a saved search notification (email or slack)
	// 4) TODO(Dan): having a saved search defined in your user or org settings
	verifyUniques, err := db.EventLogs.CountUniquesByEventNames(ctx, dayStart, dayEnd, []string{"SavedSearchEmailClicked", "SavedSearchSlackClicked", "SavedSearchEmailNotificationSent"})
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
	monitorUniques, err := db.EventLogs.CountUniquesByEventName(ctx, dayStart, dayEnd, "DiffSearchResultsQueried")
	if err != nil {
		return nil, err
	}

	//// SECURE ////
	// none currently

	//// AUTOMATE ////
	// none currently

	return &types.Stages{
		Manage:    int32(manageUniques),
		Plan:      0,
		Code:      int32(codeUniques),
		Review:    0,
		Verify:    int32(verifyUniques),
		Package:   0,
		Deploy:    0,
		Configure: 0,
		Monitor:   int32(monitorUniques),
		Secure:    0,
		Automate:  0,
	}, nil
}
