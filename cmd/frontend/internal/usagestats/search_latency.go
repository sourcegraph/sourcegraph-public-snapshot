package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// GetSearchLatencyStatistics returns search latency statistics for the
// specified span of days, weeks, and months respectively.
func GetSearchLatencyStatistics(ctx context.Context, days, weeks, months int) (*types.SearchLatencyStatistics, error) {
	var err error
	var stats types.SearchLatencyStatistics
	if stats.Daily, err = searchQueryLatency(ctx, db.Daily, minIntOrZero(maxStorageDays, days)); err != nil {
		return nil, err
	}
	if stats.Weekly, err = searchQueryLatency(ctx, db.Weekly, minIntOrZero(maxStorageDays/7, weeks)); err != nil {
		return nil, err
	}
	if stats.Monthly, err = searchQueryLatency(ctx, db.Monthly, minIntOrZero(maxStorageDays/31, months)); err != nil {
		return nil, err
	}
	return &stats, nil
}

func searchQueryLatency(ctx context.Context, periodType db.PeriodType, periods int) ([]*types.SearchLatencyPeriod, error) {
	if periods == 0 {
		return []*types.SearchLatencyPeriod{}, nil
	}

	activityPeriods := makeActivityPeriods(periods)
	latenciesByName := map[string]func(p *types.SearchLatencyPeriod) *types.SearchLatency{
		"search.latencies.literal":    func(p *types.SearchLatencyPeriod) *types.SearchLatency { return p.Latencies.Literal },
		"search.latencies.regexp":     func(p *types.SearchLatencyPeriod) *types.SearchLatency { return p.Latencies.Regexp },
		"search.latencies.structural": func(p *types.SearchLatencyPeriod) *types.SearchLatency { return p.Latencies.Structural },
		"search.latencies.file":       func(p *types.SearchLatencyPeriod) *types.SearchLatency { return p.Latencies.File },
		"search.latencies.repo":       func(p *types.SearchLatencyPeriod) *types.SearchLatency { return p.Latencies.Repo },
		"search.latencies.diff":       func(p *types.SearchLatencyPeriod) *types.SearchLatency { return p.Latencies.Diff },
		"search.latencies.commit":     func(p *types.SearchLatencyPeriod) *types.SearchLatency { return p.Latencies.Commit },
	}

	const durationField = "durationMs"
	durationPercentiles := []float64{0.5, 0.9, 0.99}

	for name, getLatencies := range latenciesByName {
		date := timeNow().UTC()
		// date = date.AddDate(0, 0, 1)
		percentiles, err := db.EventLogs.PercentilesPerPeriod(ctx, periodType, date, periods, durationField, durationPercentiles, &db.EventFilterOptions{
			ByEventName: name,
		})
		if err != nil {
			return nil, err
		}
		for i, p := range percentiles {
			getLatencies(activityPeriods[i]).P50 = p.Values[0]
			getLatencies(activityPeriods[i]).P90 = p.Values[1]
			getLatencies(activityPeriods[i]).P99 = p.Values[2]
		}
	}

	return activityPeriods, nil
}

func makeActivityPeriods(periods int) []*types.SearchLatencyPeriod {
	activityPeriods := []*types.SearchLatencyPeriod{}
	for i := 0; i <= periods; i++ {
		activityPeriods[i] = &types.SearchLatencyPeriod{
			Latencies: &types.SearchTypeLatency{
				Literal:    &types.SearchLatency{},
				Regexp:     &types.SearchLatency{},
				Structural: &types.SearchLatency{},
				File:       &types.SearchLatency{},
				Repo:       &types.SearchLatency{},
				Diff:       &types.SearchLatency{},
				Commit:     &types.SearchLatency{},
			},
		}
	}
	return activityPeriods
}
