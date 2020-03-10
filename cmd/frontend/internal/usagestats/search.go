package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// SearchUsageStatisticsOptions contains options for the number of daily, weekly, and monthly
// periods in which to calculate the latency percentiles.
type SearchUsageStatisticsOptions struct {
	DayPeriods   *int
	WeekPeriods  *int
	MonthPeriods *int
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

	daily, err := searchActivity(ctx, db.Daily, dayPeriods)
	if err != nil {
		return nil, err
	}
	weekly, err := searchActivity(ctx, db.Weekly, weekPeriods)
	if err != nil {
		return nil, err
	}
	monthly, err := searchActivity(ctx, db.Monthly, monthPeriods)
	if err != nil {
		return nil, err
	}
	return &types.SearchUsageStatistics{
		Daily:   daily,
		Weekly:  weekly,
		Monthly: monthly,
	}, nil
}

func searchActivity(ctx context.Context, periodType db.PeriodType, periods int) ([]*types.SearchUsagePeriod, error) {
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

	for eventName, getEventStatistics := range latenciesByName {
		percentiles, err := db.EventLogs.PercentilesPerPeriod(ctx, periodType, timeNow().UTC(), periods, sDurationField, sDurationPercentiles, &db.EventFilterOptions{
			ByEventName: eventName,
		})
		if err != nil {
			return nil, err
		}
		for i, p := range percentiles {
			getEventStatistics(activityPeriods[i]).EventLatencies.P50 = p.Values[0]
			getEventStatistics(activityPeriods[i]).EventLatencies.P90 = p.Values[1]
			getEventStatistics(activityPeriods[i]).EventLatencies.P99 = p.Values[2]
		}
	}

	return activityPeriods, nil
}

func newSearchEventPeriod() *types.SearchUsagePeriod {
	return &types.SearchUsagePeriod{
		Literal:    &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		Regexp:     &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		Structural: &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		File:       &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		Repo:       &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		Diff:       &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		Commit:     &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
		Symbol:     &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}},
	}
}
