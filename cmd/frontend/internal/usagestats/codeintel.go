package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// CodeIntelUsageStatisticsOptions contains options for the number of daily, weekly, and monthly
// periods in which to calculate the number of events and latency percentiles.
type CodeIntelUsageStatisticsOptions struct {
	DayPeriods            *int
	WeekPeriods           *int
	MonthPeriods          *int
	IncludeEventCounts    bool
	IncludeEventLatencies bool
}

type (
	usagePeriod             = types.CodeIntelUsagePeriod
	eventCategoryStatistics = types.CodeIntelEventCategoryStatistics
	eventStatistics         = types.CodeIntelEventStatistics
	eventLatencies          = types.CodeIntelEventLatencies
)

var (
	DurationField       = "durationMs"
	DurationPercentiles = []float64{0.5, 0.9, 0.99}
)

// GetCodeIntelUsageStatistics returns the current site's code intel activity.
func GetCodeIntelUsageStatistics(ctx context.Context, opt *CodeIntelUsageStatisticsOptions) (*types.CodeIntelUsageStatistics, error) {
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

	daily, err := codeIntelActivity(ctx, db.Daily, dayPeriods, opt.IncludeEventCounts, opt.IncludeEventLatencies)
	if err != nil {
		return nil, err
	}
	weekly, err := codeIntelActivity(ctx, db.Weekly, weekPeriods, opt.IncludeEventCounts, opt.IncludeEventLatencies)
	if err != nil {
		return nil, err
	}
	monthly, err := codeIntelActivity(ctx, db.Monthly, monthPeriods, opt.IncludeEventCounts, opt.IncludeEventLatencies)
	if err != nil {
		return nil, err
	}
	return &types.CodeIntelUsageStatistics{
		Daily:   daily,
		Weekly:  weekly,
		Monthly: monthly,
	}, nil
}

func codeIntelActivity(ctx context.Context, periodType db.PeriodType, periods int, includeEventCounts, includeEventLatencies bool) ([]*types.CodeIntelUsagePeriod, error) {
	if periods == 0 {
		return []*types.CodeIntelUsagePeriod{}, nil
	}

	activityPeriods := []*types.CodeIntelUsagePeriod{}
	for i := 0; i < periods; i++ {
		activityPeriods = append(activityPeriods, &usagePeriod{
			Hover:       newEventCategory(),
			Definitions: newEventCategory(),
			References:  newEventCategory(),
		})
	}

	eventStatisticByName := map[string]func(p *usagePeriod) *eventStatistics{
		"codeintel.lsifHover":         func(p *usagePeriod) *eventStatistics { return p.Hover.LSIF },
		"codeintel.lspHover":          func(p *usagePeriod) *eventStatistics { return p.Hover.LSP },
		"codeintel.searchHover":       func(p *usagePeriod) *eventStatistics { return p.Hover.Search },
		"codeintel.lsifDefinitions":   func(p *usagePeriod) *eventStatistics { return p.Definitions.LSIF },
		"codeintel.lspDefinitions":    func(p *usagePeriod) *eventStatistics { return p.Definitions.LSP },
		"codeintel.searchDefinitions": func(p *usagePeriod) *eventStatistics { return p.Definitions.Search },
		"codeintel.lsifReferences":    func(p *usagePeriod) *eventStatistics { return p.References.LSIF },
		"codeintel.lspReferences":     func(p *usagePeriod) *eventStatistics { return p.References.LSP },
		"codeintel.searchReferences":  func(p *usagePeriod) *eventStatistics { return p.References.Search },
	}

	for eventName, getEventStatistic := range eventStatisticByName {
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
			getEventStatistic(activityPeriods[i]).UsersCount = int32(uc.Count)
		}

		if includeEventCounts {
			eventCounts, err := db.EventLogs.CountEventsPerPeriod(ctx, periodType, timeNow().UTC(), periods, &db.EventFilterOptions{
				ByEventName: eventName,
			})
			if err != nil {
				return nil, err
			}

			for i, uc := range eventCounts {
				count := int32(uc.Count)
				getEventStatistic(activityPeriods[i]).EventsCount = &count
			}
		}

		if includeEventLatencies {
			percentiles, err := db.EventLogs.PercentilesPerPeriod(ctx, periodType, timeNow().UTC(), periods, DurationField, DurationPercentiles, &db.EventFilterOptions{
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
	}

	return activityPeriods, nil
}

func newEventCategory() *eventCategoryStatistics {
	return &eventCategoryStatistics{
		LSIF:   &eventStatistics{EventLatencies: &eventLatencies{}},
		LSP:    &eventStatistics{EventLatencies: &eventLatencies{}},
		Search: &eventStatistics{EventLatencies: &eventLatencies{}},
	}
}
