package usagestats

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// GetAggregatedCodyStats queries the database for Cody usage and returns
// the aggregates statistics in the format of our BigQuery schema.
func GetAggregatedCodyStats(ctx context.Context, db database.DB) (*types.CodyUsageStatistics, error) {
	events, err := db.EventLogs().AggregatedCodyUsage(ctx, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	stats := &types.CodyUsageStatistics{
		Daily:   []*types.CodyUsagePeriod{newCodyEventPeriod()},
		Weekly:  []*types.CodyUsagePeriod{newCodyEventPeriod()},
		Monthly: []*types.CodyUsagePeriod{newCodyEventPeriod()},
	}

	stats.Daily[0].StartTime = events.Day
	stats.Daily[0].TotalCodyUsers.EventsCount = &events.TotalDay
	stats.Daily[0].TotalCodyUsers.UserCount = &events.UniquesDay
	stats.Daily[0].TotalProductUsers.UserCount = &events.ProductUsersDay

	stats.Weekly[0].StartTime = events.Week
	stats.Weekly[0].TotalCodyUsers.EventsCount = &events.TotalWeek
	stats.Weekly[0].TotalCodyUsers.UserCount = &events.UniquesWeek
	stats.Weekly[0].TotalProductUsers.UserCount = &events.ProductUsersWeek

	stats.Monthly[0].StartTime = events.Month
	stats.Monthly[0].TotalCodyUsers.EventsCount = &events.TotalMonth
	stats.Monthly[0].TotalCodyUsers.UserCount = &events.UniquesMonth
	stats.Monthly[0].TotalProductUsers.UserCount = &events.ProductUsersMonth
	stats.Monthly[0].TotalVSCodeProductUsers.UserCount = &events.VSCodeProductUsersMonth
	stats.Monthly[0].TotalJetBrainsProductUsers.UserCount = &events.JetBrainsProductUsersMonth
	stats.Monthly[0].TotalNeovimProductUsers.UserCount = &events.NeovimProductUsersMonth
	stats.Monthly[0].TotalEmacsProductUsers.UserCount = &events.EmacsProductUsersMonth
	stats.Monthly[0].TotalWebProductUsers.UserCount = &events.WebProductUsersMonth

	return stats, nil
}

func newCodyEventPeriod() *types.CodyUsagePeriod {
	return &types.CodyUsagePeriod{
		StartTime:                  time.Now().UTC(),
		TotalCodyUsers:             newCodyCountStatistics(),
		TotalProductUsers:          newCodyCountStatistics(),
		TotalVSCodeProductUsers:    newCodyCountStatistics(),
		TotalJetBrainsProductUsers: newCodyCountStatistics(),
		TotalNeovimProductUsers:    newCodyCountStatistics(),
		TotalEmacsProductUsers:     newCodyCountStatistics(),
		TotalWebProductUsers:       newCodyCountStatistics(),
	}
}

func newCodyCountStatistics() *types.CodyCountStatistics {
	return &types.CodyCountStatistics{}
}
