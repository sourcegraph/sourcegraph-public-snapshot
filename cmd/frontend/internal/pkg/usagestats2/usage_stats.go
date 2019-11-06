// Package usagestats2 provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
//
// Note that this package should not be used on sourcegraph.com, only on self-hosted
// deployments.
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
	lastCodeHostIntegrationTime, err := db.EventLogs.MaxTimestampByUserIDAndSource(ctx, userID, "INTEGRATION")
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
			AnonymousUserCount:   int32(registeredUniques[i].Count - u.Count),
			IntegrationUserCount: int32(integrationUniques[i].Count),
			Stages:               nil,
		}
		daus = append(daus, actPer)
	}
	return daus, nil
}

// waus returns a count of weekly active users for the last weekssCount weeks (including the current, partial week).
func waus(ctx context.Context, periods int) ([]*types.SiteActivityPeriod, error) {
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
			AnonymousUserCount:   int32(registeredUniques[i].Count - u.Count),
			IntegrationUserCount: int32(integrationUniques[i].Count),
			Stages:               nil,
		}
		waus = append(waus, actPer)
	}
	return waus, nil
}

// maus returns a count of monthly active users for the last monthsCount months (including the current, partial month).
func maus(ctx context.Context, periods int) ([]*types.SiteActivityPeriod, error) {
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
			AnonymousUserCount:   int32(registeredUniques[i].Count - u.Count),
			IntegrationUserCount: int32(integrationUniques[i].Count),
			Stages:               nil,
		}
		maus = append(maus, actPer)
	}
	return maus, nil
}
