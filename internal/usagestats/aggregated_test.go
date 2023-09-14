package usagestats

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGroupSiteUsageStats(t *testing.T) {
	t1 := time.Now().Add(-30 * 24 * time.Hour).UTC()
	t2 := time.Now().UTC()
	t3 := t2.Add(time.Hour)
	t4 := t3.Add(time.Hour)

	summary := types.SiteUsageSummary{
		RollingMonth:                   t1,
		Month:                          t2,
		Week:                           t3,
		Day:                            t4,
		UniquesRollingMonth:            4,
		UniquesMonth:                   4,
		UniquesWeek:                    5,
		UniquesDay:                     6,
		RegisteredUniquesRollingMonth:  1,
		RegisteredUniquesMonth:         1,
		RegisteredUniquesWeek:          2,
		RegisteredUniquesDay:           3,
		IntegrationUniquesRollingMonth: 7,
		IntegrationUniquesMonth:        7,
		IntegrationUniquesWeek:         8,
		IntegrationUniquesDay:          9,
	}
	siteUsageStats := groupSiteUsageStats(summary, false)

	expectedSiteUsageStats := &types.SiteUsageStatistics{
		DAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            t4,
				UserCount:            6,
				RegisteredUserCount:  3,
				AnonymousUserCount:   3,
				IntegrationUserCount: 9,
			},
		},
		WAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            t3,
				UserCount:            5,
				RegisteredUserCount:  2,
				AnonymousUserCount:   3,
				IntegrationUserCount: 8,
			},
		},
		MAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            t2,
				UserCount:            4,
				RegisteredUserCount:  1,
				AnonymousUserCount:   3,
				IntegrationUserCount: 7,
			},
		},
		RMAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            t1,
				UserCount:            4,
				RegisteredUserCount:  1,
				AnonymousUserCount:   3,
				IntegrationUserCount: 7,
			},
		},
	}
	if diff := cmp.Diff(expectedSiteUsageStats, siteUsageStats); diff != "" {
		t.Fatal(diff)
	}
}

func TestGroupSiteUsageStatsMonthsOnly(t *testing.T) {
	t1 := time.Now().Add(-30 * 24 * time.Hour).UTC()
	t2 := time.Now().UTC()
	t3 := t2.Add(time.Hour)
	t4 := t3.Add(time.Hour)

	summary := types.SiteUsageSummary{
		RollingMonth:                   t1,
		Month:                          t2,
		Week:                           t3,
		Day:                            t4,
		UniquesRollingMonth:            4,
		UniquesMonth:                   4,
		UniquesWeek:                    5,
		UniquesDay:                     6,
		RegisteredUniquesRollingMonth:  1,
		RegisteredUniquesMonth:         1,
		RegisteredUniquesWeek:          2,
		RegisteredUniquesDay:           3,
		IntegrationUniquesRollingMonth: 7,
		IntegrationUniquesMonth:        7,
		IntegrationUniquesWeek:         8,
		IntegrationUniquesDay:          9,
	}
	siteUsageStats := groupSiteUsageStats(summary, true)

	expectedSiteUsageStats := &types.SiteUsageStatistics{
		DAUs: []*types.SiteActivityPeriod{},
		WAUs: []*types.SiteActivityPeriod{},
		MAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            t2,
				UserCount:            4,
				RegisteredUserCount:  1,
				AnonymousUserCount:   3,
				IntegrationUserCount: 7,
			},
		},
		RMAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            t1,
				UserCount:            4,
				RegisteredUserCount:  1,
				AnonymousUserCount:   3,
				IntegrationUserCount: 7,
			},
		},
	}
	if diff := cmp.Diff(expectedSiteUsageStats, siteUsageStats); diff != "" {
		t.Fatal(diff)
	}
}
