package usagestats

import (
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

func newSearchEventPeriod() *types.SearchUsagePeriod {
	return &types.SearchUsagePeriod{
		Literal:            newSearchEventStatistics(),
		Regexp:             newSearchEventStatistics(),
		Structural:         newSearchEventStatistics(),
		File:               newSearchEventStatistics(),
		Repo:               newSearchEventStatistics(),
		Diff:               newSearchEventStatistics(),
		Commit:             newSearchEventStatistics(),
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
	}
}

func newSearchEventStatistics() *types.SearchEventStatistics {
	return &types.SearchEventStatistics{EventLatencies: &types.SearchEventLatencies{}}
}

func newSearchCountStatistics() *types.SearchCountStatistics {
	return &types.SearchCountStatistics{}
}

func newSearchModeUsageStatistics() *types.SearchModeUsageStatistics {
	return &types.SearchModeUsageStatistics{Interactive: &types.SearchCountStatistics{}, PlainText: &types.SearchCountStatistics{}}
}
