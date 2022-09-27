package usagestats

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGroupSiteUsageStats(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := t1.Add(time.Hour)
	t3 := t2.Add(time.Hour)

	summary := types.SiteUsageSummary{
		Month:                   t1,
		Week:                    t2,
		Day:                     t3,
		UniquesMonth:            4,
		UniquesWeek:             5,
		UniquesDay:              6,
		RegisteredUniquesMonth:  1,
		RegisteredUniquesWeek:   2,
		RegisteredUniquesDay:    3,
		IntegrationUniquesMonth: 7,
		IntegrationUniquesWeek:  8,
		IntegrationUniquesDay:   9,
	}
	siteUsageStats := groupSiteUsageStats(summary, false)

	expectedSiteUsageStats := &types.SiteUsageStatistics{
		DAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            t3,
				UserCount:            6,
				RegisteredUserCount:  3,
				AnonymousUserCount:   3,
				IntegrationUserCount: 9,
			},
		},
		WAUs: []*types.SiteActivityPeriod{
			{
				StartTime:            t2,
				UserCount:            5,
				RegisteredUserCount:  2,
				AnonymousUserCount:   3,
				IntegrationUserCount: 8,
			},
		},
		MAUs: []*types.SiteActivityPeriod{
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
	t1 := time.Now().UTC()
	t2 := t1.Add(time.Hour)
	t3 := t2.Add(time.Hour)

	summary := types.SiteUsageSummary{
		Month:                   t1,
		Week:                    t2,
		Day:                     t3,
		UniquesMonth:            4,
		UniquesWeek:             5,
		UniquesDay:              6,
		RegisteredUniquesMonth:  1,
		RegisteredUniquesWeek:   2,
		RegisteredUniquesDay:    3,
		IntegrationUniquesMonth: 7,
		IntegrationUniquesWeek:  8,
		IntegrationUniquesDay:   9,
	}
	siteUsageStats := groupSiteUsageStats(summary, true)

	expectedSiteUsageStats := &types.SiteUsageStatistics{
		DAUs: []*types.SiteActivityPeriod{},
		WAUs: []*types.SiteActivityPeriod{},
		MAUs: []*types.SiteActivityPeriod{
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
