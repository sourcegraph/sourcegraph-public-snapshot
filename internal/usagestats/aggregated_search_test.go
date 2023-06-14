package usagestats

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestGroupAggregateSearchStats(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := t1.Add(time.Hour)
	t3 := t2.Add(time.Hour)

	searchStats := groupAggregatedSearchStats([]types.SearchAggregatedEvent{
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

	expectDailyStructural := newSearchTestEvent(33, 36, 37, 38, 39)
	expectDailyCommit := newSearchTestEvent(43, 46, 47, 48, 49)
	expectWeeklyStructural := newSearchTestEvent(32, 35, 34, 35, 36)
	expectWeeklyCommit := newSearchTestEvent(42, 45, 44, 45, 46)
	expectMonthlyStructural := newSearchTestEvent(31, 34, 31, 32, 33)
	expectMonthlyCommit := newSearchTestEvent(41, 44, 41, 42, 43)

	expectedSearchStats := &types.SearchUsageStatistics{
		Daily:   newSearchUsagePeriod(t3, expectDailyStructural, expectDailyCommit),
		Weekly:  newSearchUsagePeriod(t2, expectWeeklyStructural, expectWeeklyCommit),
		Monthly: newSearchUsagePeriod(t1, expectMonthlyStructural, expectMonthlyCommit),
	}
	if diff := cmp.Diff(expectedSearchStats, searchStats); diff != "" {
		t.Fatal(diff)
	}
}

func newSearchTestEvent(eventCount, userCount int32, p50, p90, p99 float64) *types.SearchEventStatistics {
	return &types.SearchEventStatistics{
		EventsCount:    pointers.Ptr(eventCount),
		UserCount:      pointers.Ptr(userCount),
		EventLatencies: &types.SearchEventLatencies{P50: p50, P90: p90, P99: p99},
	}
}

func newSearchUsagePeriod(t time.Time, structuralEvent, commitEvent *types.SearchEventStatistics) []*types.SearchUsagePeriod {
	return []*types.SearchUsagePeriod{
		{
			StartTime:  t,
			Literal:    newSearchEventStatistics(),
			Regexp:     newSearchEventStatistics(),
			Structural: structuralEvent,
			File:       newSearchEventStatistics(),
			Repo:       newSearchEventStatistics(),
			Diff:       newSearchEventStatistics(),
			Commit:     commitEvent,
			Symbol:     newSearchEventStatistics(),

			// Counts of search query attributes. Ref: RFC 384.
			OperatorOr:              newSearchCountStatistics(),
			OperatorAnd:             newSearchCountStatistics(),
			OperatorNot:             newSearchCountStatistics(),
			SelectRepo:              newSearchCountStatistics(),
			SelectFile:              newSearchCountStatistics(),
			SelectContent:           newSearchCountStatistics(),
			SelectSymbol:            newSearchCountStatistics(),
			SelectCommitDiffAdded:   newSearchCountStatistics(),
			SelectCommitDiffRemoved: newSearchCountStatistics(),
			RepoContains:            newSearchCountStatistics(),
			RepoContainsFile:        newSearchCountStatistics(),
			RepoContainsContent:     newSearchCountStatistics(),
			RepoContainsCommitAfter: newSearchCountStatistics(),
			RepoDependencies:        newSearchCountStatistics(),
			CountAll:                newSearchCountStatistics(),
			NonGlobalContext:        newSearchCountStatistics(),
			OnlyPatterns:            newSearchCountStatistics(),
			OnlyPatternsThreeOrMore: newSearchCountStatistics(),

			// DEPRECATED.
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
	}
}
