package usagestats

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// SearchUsageStatisticsOptions contains options for the number of daily, weekly, and monthly
// periods in which to calculate the latency percentiles.
type SearchUsageStatisticsOptions struct {
	DayPeriods         *int
	WeekPeriods        *int
	MonthPeriods       *int
	IncludeEventCounts bool
}

var (
	sDurationField       = "durationMs"
	sDurationPercentiles = []float64{0.5, 0.9, 0.99}
)

// GetSearchUsageStatistics returns the current site's search activity.
func GetSearchUsageStatistics(ctx context.Context, opt *SearchUsageStatisticsOptions) (*types.SearchUsageStatistics, error) {
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

	daily, err := searchActivity(ctx, db.Daily, dayPeriods, opt.IncludeEventCounts)
	if err != nil {
		return nil, err
	}
	weekly, err := searchActivity(ctx, db.Weekly, weekPeriods, opt.IncludeEventCounts)
	if err != nil {
		return nil, err
	}
	monthly, err := searchActivity(ctx, db.Monthly, monthPeriods, opt.IncludeEventCounts)
	if err != nil {
		return nil, err
	}
	return &types.SearchUsageStatistics{
		Daily:   daily,
		Weekly:  weekly,
		Monthly: monthly,
	}, nil
}

func searchActivity(ctx context.Context, periodType db.PeriodType, periods int, includeEventCounts bool) ([]*types.SearchUsagePeriod, error) {
	if periods == 0 {
		return []*types.SearchUsagePeriod{}, nil
	}

	activityPeriods := make([]*types.SearchUsagePeriod, 0, periods)
	for i := 0; i < periods; i++ {
		activityPeriods = append(activityPeriods, newSearchEventPeriod())
	}

	latenciesByName := map[string]func(p *types.SearchUsagePeriod) *types.SearchEventStatistics{
		"search.latencies.literal":    func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Literal },
		"search.latencies.regexp":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Regexp },
		"search.latencies.structural": func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Structural },
		"search.latencies.file":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.File },
		"search.latencies.repo":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Repo },
		"search.latencies.diff":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Diff },
		"search.latencies.commit":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Commit },
		"search.latencies.symbol":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Symbol },
	}

	for eventName, getEventStatistic := range latenciesByName {
		userCounts, err := db.EventLogs.CountUniqueUsersPerPeriod(ctx, periodType, timeNow().UTC(), periods, &db.CountUniqueUsersOptions{
			EventFilters: &db.EventFilterOptions{
				ByEventName: eventName,
			},
		})
		if err != nil {
			return nil, err
		}

		for i, uc := range userCounts {
			activityPeriods[i].StartTime = uc.Start
		}

		percentiles, err := db.EventLogs.PercentilesPerPeriod(ctx, periodType, timeNow().UTC(), periods, sDurationField, sDurationPercentiles, &db.EventFilterOptions{
			ByEventName: eventName,
		})
		if err != nil {
			return nil, err
		}
		for i, p := range percentiles {
			getEventStatistic(activityPeriods[i]).EventLatencies.P50 = p.Values[0]
			getEventStatistic(activityPeriods[i]).EventLatencies.P90 = p.Values[1]
			getEventStatistic(activityPeriods[i]).EventLatencies.P99 = p.Values[2]
		}
	}

	// Count total unique search users per period
	totalUniqueUsers, err := db.EventLogs.CountUniqueUsersPerPeriod(ctx, periodType, timeNow().UTC(), periods, &db.CountUniqueUsersOptions{
		EventFilters: &db.EventFilterOptions{ByEventName: "SearchResultsQueried"},
	})
	if err != nil {
		return nil, err
	}
	for i, uniqueUserCounts := range totalUniqueUsers {
		activityPeriods[i].StartTime = uniqueUserCounts.Start
		activityPeriods[i].TotalUsers = int32(uniqueUserCounts.Count)
	}

	// Count total unique users and events of each search mode per period
	searchModeNameToArgumentMatches := map[string]struct {
		eventName          string
		argumentName       string
		getEventStatistics func(p *types.SearchUsagePeriod) *types.SearchCountStatistics
	}{
		"plain": {
			eventName:          "SearchResultsQueried",
			argumentName:       "mode",
			getEventStatistics: func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.SearchModes.PlainText },
		},
		"interactive": {
			eventName:          "SearchResultsQueried",
			argumentName:       "mode",
			getEventStatistics: func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.SearchModes.Interactive },
		},
	}

	for searchMode, match := range searchModeNameToArgumentMatches {
		userCounts, err := db.EventLogs.CountUniqueUsersPerPeriod(ctx, periodType, timeNow().UTC(), periods, &db.CountUniqueUsersOptions{
			EventFilters: &db.EventFilterOptions{
				ByEventName:              match.eventName,
				ByEventNameWithCondition: sqlf.Sprintf("argument->>%s=%s", match.argumentName, searchMode),
			},
		})
		if err != nil {
			return nil, err
		}
		for i, uc := range userCounts {
			count := int32(uc.Count)
			match.getEventStatistics(activityPeriods[i]).UserCount = &count
		}
		if includeEventCounts {
			eventCounts, err := db.EventLogs.CountEventsPerPeriod(ctx, periodType, timeNow().UTC(), periods, &db.EventFilterOptions{
				ByEventName:              match.eventName,
				ByEventNameWithCondition: sqlf.Sprintf("argument->>%s=%s", match.argumentName, searchMode),
			})
			if err != nil {
				return nil, err
			}
			for i, ec := range eventCounts {
				count := int32(ec.Count)
				match.getEventStatistics(activityPeriods[i]).EventsCount = &count
			}
		}
	}

	filterEventStatistics := map[string]func(*types.SearchUsagePeriod) *types.SearchEventStatistics{
		"field_file": func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.File },
		"field_repo": func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Repo },
	}

	filterCountStatistics := map[string]func(*types.SearchUsagePeriod) *types.SearchCountStatistics{
		"field_after":              func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.After },
		"field_archived":           func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Archived },
		"field_author":             func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Author },
		"field_before":             func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Before },
		"field_case":               func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Case },
		"field_committer":          func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Committer },
		"field_content":            func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Content },
		"field_count":              func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Count },
		"field_fork":               func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Fork },
		"field_index":              func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Index },
		"field_lang":               func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Lang },
		"field_message":            func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Message },
		"field_patterntype":        func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.PatternType },
		"field_repogroup":          func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Repogroup },
		"field_repohascommitafter": func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Repohascommitafter },
		"field_repohasfile":        func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Repohasfile },
		"field_timeout":            func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Timeout },
		"field_type":               func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.Type },
	}

	for filter, getCountStatistics := range filterEventStatistics {
		userCounts, err := countSearchFilterUsersPerPeriod(ctx, periodType, time.Now().UTC(), periods, filter)
		if err != nil {
			return nil, err
		}

		for i, uc := range userCounts {
			count := int32(uc.Count)
			getCountStatistics(activityPeriods[i]).UserCount = &count
		}

		if includeEventCounts {
			filterCounts, err := countSearchFilterEventsPerPeriod(ctx, periodType, time.Now().UTC(), periods, filter)
			if err != nil {
				return nil, err
			}

			for i, fc := range filterCounts {
				count := int32(fc.Count)
				getCountStatistics(activityPeriods[i]).EventsCount = &count
			}
		}
	}

	for filter, getCountStatistics := range filterCountStatistics {
		userCounts, err := countSearchFilterUsersPerPeriod(ctx, periodType, time.Now().UTC(), periods, filter)
		if err != nil {
			return nil, err
		}

		for i, uc := range userCounts {
			count := int32(uc.Count)
			getCountStatistics(activityPeriods[i]).UserCount = &count
		}

		if includeEventCounts {
			eventCounts, err := countSearchFilterEventsPerPeriod(ctx, periodType, time.Now().UTC(), periods, filter)
			if err != nil {
				return nil, err
			}

			for i, fc := range eventCounts {
				count := int32(fc.Count)
				getCountStatistics(activityPeriods[i]).EventsCount = &count
			}
		}
	}
	return activityPeriods, nil
}

func countSearchFilterEventsPerPeriod(ctx context.Context, periodType db.PeriodType, now time.Time, periods int, fieldName string) ([]db.UsageValue, error) {
	ec, err := db.EventLogs.CountEventsPerPeriod(ctx, periodType, now, periods, &db.EventFilterOptions{ByEventName: "SearchResultsQueried", ByEventNameWithCondition: sqlf.Sprintf("argument->'code_search'->'query_data'->'query'->%s IS NOT NULL", fieldName)})
	if err != nil {
		return nil, err
	}
	return ec, nil
}

func countSearchFilterUsersPerPeriod(ctx context.Context, periodType db.PeriodType, now time.Time, periods int, fieldName string) ([]db.UsageValue, error) {
	ec, err := db.EventLogs.CountUniqueUsersPerPeriod(ctx, periodType, now, periods, &db.CountUniqueUsersOptions{EventFilters: &db.EventFilterOptions{ByEventName: "SearchResultsQueried", ByEventNameWithCondition: sqlf.Sprintf("argument->'code_search'->'query_data'->'query'->%s IS NOT NULL", fieldName)}})
	if err != nil {
		return nil, err
	}
	return ec, nil
}

func newSearchEventPeriod() *types.SearchUsagePeriod {
	return &types.SearchUsagePeriod{
		TotalUsers:         0,
		Literal:            &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		Regexp:             &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		Structural:         &types.SearchEventStatistics{UserCount: nil, EventsCount: nil, EventLatencies: &types.SearchEventLatencies{}},
		File:               &types.SearchEventStatistics{UserCount: nil, EventsCount: nil, EventLatencies: &types.SearchEventLatencies{}},
		Repo:               &types.SearchEventStatistics{UserCount: nil, EventsCount: nil, EventLatencies: &types.SearchEventLatencies{}},
		Diff:               &types.SearchEventStatistics{UserCount: nil, EventsCount: nil, EventLatencies: &types.SearchEventLatencies{}},
		Commit:             &types.SearchEventStatistics{UserCount: nil, EventsCount: nil, EventLatencies: &types.SearchEventLatencies{}},
		Symbol:             &types.SearchEventStatistics{UserCount: nil, EventsCount: nil, EventLatencies: &types.SearchEventLatencies{}},
		Case:               &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Committer:          &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Lang:               &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Fork:               &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Archived:           &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Count:              &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Timeout:            &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Content:            &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Before:             &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		After:              &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Author:             &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Message:            &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Index:              &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Repogroup:          &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Repohasfile:        &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Repohascommitafter: &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		PatternType:        &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		Type:               &types.SearchCountStatistics{UserCount: nil, EventsCount: nil},
		SearchModes:        &types.SearchModeUsageStatistics{Interactive: &types.SearchCountStatistics{}, PlainText: &types.SearchCountStatistics{}},
	}
}
