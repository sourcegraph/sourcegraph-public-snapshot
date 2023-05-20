package usagestats

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// GetAggregatedSearchStats queries the database for search usage and returns
// the aggregates statistics in the format of our BigQuery schema.
func GetAggregatedSearchStats(ctx context.Context, db database.DB) (*types.SearchUsageStatistics, error) {
	events, err := db.EventLogs().AggregatedSearchEvents(ctx, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return groupAggregatedSearchStats(events), nil
}

// groupAggregatedSearchStats takes a set of input events (originating from
// Sourcegraph's Postgres table) and returns a SearchUsageStatistics data type
// that ends up being stored in BigQuery. SearchUsageStatistics corresponds to
// the target DB schema.
func groupAggregatedSearchStats(events []types.SearchAggregatedEvent) *types.SearchUsageStatistics {
	searchUsageStats := &types.SearchUsageStatistics{
		Daily:   []*types.SearchUsagePeriod{newSearchEventPeriod()},
		Weekly:  []*types.SearchUsagePeriod{newSearchEventPeriod()},
		Monthly: []*types.SearchUsagePeriod{newSearchEventPeriod()},
	}

	// Iterate over events, updating searchUsageStats for each event
	for _, event := range events {
		populateSearchEventStatistics(event, searchUsageStats)
		populateSearchFilterCountStatistics(event, searchUsageStats)
	}

	return searchUsageStats
}

// utility functions that resolve a SearchEventStatistics value for a given event name for some SearchUsagePeriod.
var searchLatencyExtractors = map[string]func(p *types.SearchUsagePeriod) *types.SearchEventStatistics{
	"search.latencies.literal":    func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Literal },
	"search.latencies.regexp":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Regexp },
	"search.latencies.structural": func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Structural },
	"search.latencies.file":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.File },
	"search.latencies.repo":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Repo },
	"search.latencies.diff":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Diff },
	"search.latencies.commit":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Commit },
	"search.latencies.symbol":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Symbol },
}

var searchFilterCountExtractors = map[string]func(p *types.SearchUsagePeriod) *types.SearchCountStatistics{
	"count_or":                          func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.OperatorOr },
	"count_and":                         func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.OperatorAnd },
	"count_not":                         func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.OperatorNot },
	"count_select_repo":                 func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.SelectRepo },
	"count_select_file":                 func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.SelectFile },
	"count_select_content":              func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.SelectContent },
	"count_select_symbol":               func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.SelectSymbol },
	"count_select_commit_diff_added":    func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.SelectCommitDiffAdded },
	"count_select_commit_diff_removed":  func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.SelectCommitDiffRemoved },
	"count_repo_contains":               func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.RepoContains },
	"count_repo_contains_file":          func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.RepoContainsFile },
	"count_repo_contains_content":       func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.RepoContainsContent },
	"count_repo_contains_commit_after":  func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.RepoContainsCommitAfter },
	"count_repo_dependencies":           func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.RepoDependencies },
	"count_count_all":                   func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.CountAll },
	"count_non_global_context":          func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.NonGlobalContext },
	"count_only_patterns":               func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.OnlyPatterns },
	"count_only_patterns_three_or_more": func(p *types.SearchUsagePeriod) *types.SearchCountStatistics { return p.OnlyPatternsThreeOrMore },
}

// populateSearchEventStatistics is a side-effecting function that populates the
// `statistics` object. The `statistics` event value is our target output type.
//
// Overview how it works:
// (1) To populate the `statistics` object, we expect an event to have a supported event.Name.
//
// (2) Create a SearchUsagePeriod target object based on the event's period (i.e., Month, Week, Day).
//
// (3) Use the SearchUsagePeriod object as an argument for the utility functions
// above, to get a handle on the (currently zero-valued) SearchEventStatistics
// value that it contains that corresponds to that event type.
//
// (4) Populate that SearchEventStatistics object in the SearchUsagePeriod object with usage stats (latencies, etc).
func populateSearchEventStatistics(event types.SearchAggregatedEvent, statistics *types.SearchUsageStatistics) {
	extractor, ok := searchLatencyExtractors[event.Name]
	if !ok {
		return
	}

	makeLatencies := func(values []float64) *types.SearchEventLatencies {
		for len(values) < 3 {
			// If event logs didn't have samples, add zero values
			values = append(values, 0)
		}

		return &types.SearchEventLatencies{P50: values[0], P90: values[1], P99: values[2]}
	}

	statistics.Monthly[0].StartTime = event.Month
	month := extractor(statistics.Monthly[0])
	month.EventsCount = &event.TotalMonth
	month.UserCount = &event.UniquesMonth
	month.EventLatencies = makeLatencies(event.LatenciesMonth)

	statistics.Weekly[0].StartTime = event.Week
	week := extractor(statistics.Weekly[0])
	week.EventsCount = &event.TotalWeek
	week.UserCount = &event.UniquesWeek
	week.EventLatencies = makeLatencies(event.LatenciesWeek)

	statistics.Daily[0].StartTime = event.Day
	day := extractor(statistics.Daily[0])
	day.EventsCount = &event.TotalDay
	day.UserCount = &event.UniquesDay
	day.EventLatencies = makeLatencies(event.LatenciesDay)
}

func populateSearchFilterCountStatistics(event types.SearchAggregatedEvent, statistics *types.SearchUsageStatistics) {
	extractor, ok := searchFilterCountExtractors[event.Name]
	if !ok {
		return
	}

	statistics.Monthly[0].StartTime = event.Month
	month := extractor(statistics.Monthly[0])
	month.EventsCount = &event.TotalMonth
	month.UserCount = &event.UniquesMonth

	statistics.Weekly[0].StartTime = event.Week
	week := extractor(statistics.Weekly[0])
	week.EventsCount = &event.TotalMonth
	week.UserCount = &event.UniquesMonth

	statistics.Daily[0].StartTime = event.Day
	day := extractor(statistics.Daily[0])
	day.EventsCount = &event.TotalMonth
	day.UserCount = &event.UniquesMonth
}

func newSearchEventPeriod() *types.SearchUsagePeriod {
	return &types.SearchUsagePeriod{
		Literal:    newSearchEventStatistics(),
		Regexp:     newSearchEventStatistics(),
		Structural: newSearchEventStatistics(),
		File:       newSearchEventStatistics(),
		Repo:       newSearchEventStatistics(),
		Diff:       newSearchEventStatistics(),
		Commit:     newSearchEventStatistics(),
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
