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
		ManageUniquesMonth:      10,
		CodeUniquesMonth:        11,
		VerifyUniquesMonth:      12,
		MonitorUniquesMonth:     13,
		ManageUniquesWeek:       14,
		CodeUniquesWeek:         15,
		VerifyUniquesWeek:       16,
		MonitorUniquesWeek:      17,
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
		ManageUniquesMonth:      10,
		CodeUniquesMonth:        11,
		VerifyUniquesMonth:      12,
		MonitorUniquesMonth:     13,
		ManageUniquesWeek:       14,
		CodeUniquesWeek:         15,
		VerifyUniquesWeek:       16,
		MonitorUniquesWeek:      17,
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

func TestGroupAggregateSearchStats(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := t1.Add(time.Hour)
	t3 := t2.Add(time.Hour)

	searchStats := groupAggregatedSearchStats([]types.AggregatedEvent{
		{
			Name:           "search.latencies.structural",
			Month:          t1,
			Week:           t2,
			Day:            t3,
			TotalMonth:     31,
			TotalWeek:      32,
			TotalDay:       33,
			UniquesMonth:   34,
			UniquesWeek:    35,
			UniquesDay:     36,
			LatenciesMonth: []float64{31, 32, 33},
			LatenciesWeek:  []float64{34, 35, 36},
			LatenciesDay:   []float64{37, 38, 39},
		},
		{
			Name:           "search.latencies.commit",
			Month:          t1,
			Week:           t2,
			Day:            t3,
			TotalMonth:     41,
			TotalWeek:      42,
			TotalDay:       43,
			UniquesMonth:   44,
			UniquesWeek:    45,
			UniquesDay:     46,
			LatenciesMonth: []float64{41, 42, 43},
			LatenciesWeek:  []float64{44, 45, 46},
			LatenciesDay:   []float64{47, 48, 49},
		},
	})

	intptr := func(i int32) *int32 {
		return &i
	}

	expectedSearchStats := &types.SearchUsageStatistics{
		Daily: []*types.SearchUsagePeriod{
			{
				StartTime: t3,
				Literal:   newSearchEventStatistics(),
				Regexp:    newSearchEventStatistics(),
				Structural: &types.SearchEventStatistics{
					EventsCount:    intptr(33),
					UserCount:      intptr(36),
					EventLatencies: &types.SearchEventLatencies{P50: 37, P90: 38, P99: 39},
				},
				File: newSearchEventStatistics(),
				Repo: newSearchEventStatistics(),
				Diff: newSearchEventStatistics(),
				Commit: &types.SearchEventStatistics{
					EventsCount:    intptr(43),
					UserCount:      intptr(46),
					EventLatencies: &types.SearchEventLatencies{P50: 47, P90: 48, P99: 49},
				},
				Symbol:             newSearchEventStatistics(),
				Case:               newSearchCountStatistics(),
				Committer:          newSearchCountStatistics(),
				Lang:               newSearchCountStatistics(),
				Fork:               newSearchCountStatistics(),
				Archived:           newSearchCountStatistics(),
				Count:              newSearchCountStatistics(),
				Timeout:            newSearchCountStatistics(),
				Content:            newSearchCountStatistics(),
				Before:             newSearchCountStatistics(),
				After:              newSearchCountStatistics(),
				Author:             newSearchCountStatistics(),
				Message:            newSearchCountStatistics(),
				Index:              newSearchCountStatistics(),
				Repogroup:          newSearchCountStatistics(),
				Repohasfile:        newSearchCountStatistics(),
				Repohascommitafter: newSearchCountStatistics(),
				PatternType:        newSearchCountStatistics(),
				Type:               newSearchCountStatistics(),
				SearchModes:        newSearchModeUsageStatistics(),
			},
		},
		Weekly: []*types.SearchUsagePeriod{
			{
				StartTime: t2,
				Literal:   newSearchEventStatistics(),
				Regexp:    newSearchEventStatistics(),
				Structural: &types.SearchEventStatistics{
					EventsCount:    intptr(32),
					UserCount:      intptr(35),
					EventLatencies: &types.SearchEventLatencies{P50: 34, P90: 35, P99: 36},
				},
				File: newSearchEventStatistics(),
				Repo: newSearchEventStatistics(),
				Diff: newSearchEventStatistics(),
				Commit: &types.SearchEventStatistics{
					EventsCount:    intptr(42),
					UserCount:      intptr(45),
					EventLatencies: &types.SearchEventLatencies{P50: 44, P90: 45, P99: 46},
				},
				Symbol:             newSearchEventStatistics(),
				Case:               newSearchCountStatistics(),
				Committer:          newSearchCountStatistics(),
				Lang:               newSearchCountStatistics(),
				Fork:               newSearchCountStatistics(),
				Archived:           newSearchCountStatistics(),
				Count:              newSearchCountStatistics(),
				Timeout:            newSearchCountStatistics(),
				Content:            newSearchCountStatistics(),
				Before:             newSearchCountStatistics(),
				After:              newSearchCountStatistics(),
				Author:             newSearchCountStatistics(),
				Message:            newSearchCountStatistics(),
				Index:              newSearchCountStatistics(),
				Repogroup:          newSearchCountStatistics(),
				Repohasfile:        newSearchCountStatistics(),
				Repohascommitafter: newSearchCountStatistics(),
				PatternType:        newSearchCountStatistics(),
				Type:               newSearchCountStatistics(),
				SearchModes:        newSearchModeUsageStatistics(),
			},
		},
		Monthly: []*types.SearchUsagePeriod{
			{
				StartTime: t1,
				Literal:   newSearchEventStatistics(),
				Regexp:    newSearchEventStatistics(),
				Structural: &types.SearchEventStatistics{
					EventsCount:    intptr(31),
					UserCount:      intptr(34),
					EventLatencies: &types.SearchEventLatencies{P50: 31, P90: 32, P99: 33},
				},
				File: newSearchEventStatistics(),
				Repo: newSearchEventStatistics(),
				Diff: newSearchEventStatistics(),
				Commit: &types.SearchEventStatistics{
					EventsCount:    intptr(41),
					UserCount:      intptr(44),
					EventLatencies: &types.SearchEventLatencies{P50: 41, P90: 42, P99: 43},
				},
				Symbol:             newSearchEventStatistics(),
				Case:               newSearchCountStatistics(),
				Committer:          newSearchCountStatistics(),
				Lang:               newSearchCountStatistics(),
				Fork:               newSearchCountStatistics(),
				Archived:           newSearchCountStatistics(),
				Count:              newSearchCountStatistics(),
				Timeout:            newSearchCountStatistics(),
				Content:            newSearchCountStatistics(),
				Before:             newSearchCountStatistics(),
				After:              newSearchCountStatistics(),
				Author:             newSearchCountStatistics(),
				Message:            newSearchCountStatistics(),
				Index:              newSearchCountStatistics(),
				Repogroup:          newSearchCountStatistics(),
				Repohasfile:        newSearchCountStatistics(),
				Repohascommitafter: newSearchCountStatistics(),
				PatternType:        newSearchCountStatistics(),
				Type:               newSearchCountStatistics(),
				SearchModes:        newSearchModeUsageStatistics(),
			},
		},
	}
	if diff := cmp.Diff(expectedSearchStats, searchStats); diff != "" {
		t.Fatal(diff)
	}
}

func TestGroupAggregatedCodeIntelStats(t *testing.T) {
	lang1 := "go"
	lang2 := "typescript"
	t1 := time.Now().UTC().Add(time.Hour)

	codeIntelStats := groupAggregatedCodeIntelStats([]types.CodeIntelAggregatedEvent{
		{Name: "codeintel.lsifHover", Week: t1, TotalWeek: 10, UniquesWeek: 1},
		{Name: "codeintel.searchDefinitions", Week: t1, TotalWeek: 20, UniquesWeek: 2, LanguageID: &lang1},
		{Name: "codeintel.lsifDefinitions", Week: t1, TotalWeek: 30, UniquesWeek: 3},
		{Name: "codeintel.searchReferences.xrepo", Week: t1, TotalWeek: 40, UniquesWeek: 4, LanguageID: &lang2},
	})

	expectedCodeIntelStats := &types.NewCodeIntelUsageStatistics{
		StartOfWeek: t1,
		EventSummaries: []types.CodeIntelEventSummary{
			{
				Action:          types.HoverAction,
				Source:          types.PreciseSource,
				LanguageID:      "",
				CrossRepository: false,
				WAUs:            1,
				TotalActions:    10,
			},
			{
				Action:          types.DefinitionsAction,
				Source:          types.SearchSource,
				LanguageID:      "go",
				CrossRepository: false,
				WAUs:            2,
				TotalActions:    20,
			},
			{
				Action:          types.DefinitionsAction,
				Source:          types.PreciseSource,
				LanguageID:      "",
				CrossRepository: false,
				WAUs:            3,
				TotalActions:    30,
			},
			{
				Action:          types.ReferencesAction,
				Source:          types.SearchSource,
				LanguageID:      "typescript",
				CrossRepository: true,
				WAUs:            4,
				TotalActions:    40,
			},
		},
	}
	if diff := cmp.Diff(expectedCodeIntelStats, codeIntelStats); diff != "" {
		t.Fatal(diff)
	}
}
