package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetSiteUsageStats(ctx context.Context, db database.DB, monthsOnly bool) (*types.SiteUsageStatistics, error) {
	summary, err := db.EventLogs().SiteUsageCurrentPeriods(ctx)
	if err != nil {
		return nil, err
	}

	stats := groupSiteUsageStats(summary, monthsOnly)
	return stats, nil
}

func groupSiteUsageStats(summary types.SiteUsageSummary, monthsOnly bool) *types.SiteUsageStatistics {
	stats := &types.SiteUsageStatistics{
		DAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            summary.Day,
				UserCount:            summary.UniquesDay,
				RegisteredUserCount:  summary.RegisteredUniquesDay,
				AnonymousUserCount:   summary.UniquesDay - summary.RegisteredUniquesDay,
				IntegrationUserCount: summary.IntegrationUniquesDay,
			},
		},
		WAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            summary.Week,
				UserCount:            summary.UniquesWeek,
				RegisteredUserCount:  summary.RegisteredUniquesWeek,
				AnonymousUserCount:   summary.UniquesWeek - summary.RegisteredUniquesWeek,
				IntegrationUserCount: summary.IntegrationUniquesWeek,
			},
		},
		MAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            summary.Month,
				UserCount:            summary.UniquesMonth,
				RegisteredUserCount:  summary.RegisteredUniquesMonth,
				AnonymousUserCount:   summary.UniquesMonth - summary.RegisteredUniquesMonth,
				IntegrationUserCount: summary.IntegrationUniquesMonth,
			},
		},
	}

	if monthsOnly {
		stats.DAUs = []*types.SiteActivityPeriod{}
		stats.WAUs = []*types.SiteActivityPeriod{}
	}

	return stats
}
