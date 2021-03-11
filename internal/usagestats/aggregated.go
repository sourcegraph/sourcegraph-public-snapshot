package usagestats

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetSiteUsageStats(ctx context.Context, db dbutil.DB, monthsOnly bool) (*types.SiteUsageStatistics, error) {
	summary, err := database.EventLogs(db).SiteUsage(ctx)
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

// GetAggregatedCodeIntelStats returns aggregated statistics for code intelligence usage.
func GetAggregatedCodeIntelStats(ctx context.Context, db dbutil.DB) (*types.NewCodeIntelUsageStatistics, error) {
	codeIntelEvents, err := database.EventLogs(db).AggregatedCodeIntelEvents(ctx)
	if err != nil {
		return nil, err
	} else if len(codeIntelEvents) == 0 {
		return nil, nil
	}
	stats := groupAggregatedCodeIntelStats(codeIntelEvents)

	pairs := []struct {
		fetch  func(ctx context.Context) (int, error)
		target **int32
	}{
		{database.EventLogs(db).CodeIntelligenceWAUs, &stats.WAUs},
		{database.EventLogs(db).CodeIntelligencePreciseWAUs, &stats.PreciseWAUs},
		{database.EventLogs(db).CodeIntelligenceSearchBasedWAUs, &stats.SearchBasedWAUs},
		{database.EventLogs(db).CodeIntelligenceCrossRepositoryWAUs, &stats.CrossRepositoryWAUs},
		{database.EventLogs(db).CodeIntelligencePreciseCrossRepositoryWAUs, &stats.PreciseCrossRepositoryWAUs},
		{database.EventLogs(db).CodeIntelligenceSearchBasedCrossRepositoryWAUs, &stats.SearchBasedCrossRepositoryWAUs},
	}

	for _, pair := range pairs {
		count, err := pair.fetch(ctx)
		if err != nil {
			return nil, err
		}

		v := int32(count)
		*pair.target = &v
	}

	return stats, nil
}

var actionMap = map[string]types.CodeIntelAction{
	"codeintel.lsifHover":               types.HoverAction,
	"codeintel.searchHover":             types.HoverAction,
	"codeintel.lsifDefinitions":         types.DefinitionsAction,
	"codeintel.lsifDefinitions.xrepo":   types.DefinitionsAction,
	"codeintel.searchDefinitions":       types.DefinitionsAction,
	"codeintel.searchDefinitions.xrepo": types.DefinitionsAction,
	"codeintel.lsifReferences":          types.ReferencesAction,
	"codeintel.lsifReferences.xrepo":    types.ReferencesAction,
	"codeintel.searchReferences":        types.ReferencesAction,
	"codeintel.searchReferences.xrepo":  types.ReferencesAction,
}

var sourceMap = map[string]types.CodeIntelSource{
	"codeintel.lsifHover":               types.PreciseSource,
	"codeintel.lsifDefinitions":         types.PreciseSource,
	"codeintel.lsifDefinitions.xrepo":   types.PreciseSource,
	"codeintel.lsifReferences":          types.PreciseSource,
	"codeintel.lsifReferences.xrepo":    types.PreciseSource,
	"codeintel.searchHover":             types.SearchSource,
	"codeintel.searchDefinitions":       types.SearchSource,
	"codeintel.searchDefinitions.xrepo": types.SearchSource,
	"codeintel.searchReferences":        types.SearchSource,
	"codeintel.searchReferences.xrepo":  types.SearchSource,
}

func groupAggregatedCodeIntelStats(rawEvents []types.CodeIntelAggregatedEvent) *types.NewCodeIntelUsageStatistics {
	var eventSummaries []types.CodeIntelEventSummary
	for _, event := range rawEvents {
		languageID := ""
		if event.LanguageID != nil {
			languageID = *event.LanguageID
		}

		eventSummaries = append(eventSummaries, types.CodeIntelEventSummary{
			Action:          actionMap[event.Name],
			Source:          sourceMap[event.Name],
			LanguageID:      languageID,
			CrossRepository: strings.HasSuffix(event.Name, ".xrepo"),
			WAUs:            event.UniquesWeek,
			TotalActions:    event.TotalWeek,
		})
	}

	return &types.NewCodeIntelUsageStatistics{
		StartOfWeek:    rawEvents[0].Week,
		EventSummaries: eventSummaries,
	}
}

// GetAggregatedSearchStats returns aggregates statistics for search usage.
func GetAggregatedSearchStats(ctx context.Context, db dbutil.DB) (*types.SearchUsageStatistics, error) {
	events, err := database.EventLogs(db).AggregatedSearchEvents(ctx)
	if err != nil {
		return nil, err
	}

	return groupAggregatedSearchStats(events), nil
}

func groupAggregatedSearchStats(events []types.AggregatedEvent) *types.SearchUsageStatistics {
	searchUsageStats := &types.SearchUsageStatistics{
		Daily:   []*types.SearchUsagePeriod{newSearchEventPeriod()},
		Weekly:  []*types.SearchUsagePeriod{newSearchEventPeriod()},
		Monthly: []*types.SearchUsagePeriod{newSearchEventPeriod()},
	}

	for _, event := range events {
		insertSearchEventStatistics(event, searchUsageStats)
	}

	return searchUsageStats
}

var searchExtractors = map[string]func(p *types.SearchUsagePeriod) *types.SearchEventStatistics{
	"search.latencies.literal":    func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Literal },
	"search.latencies.regexp":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Regexp },
	"search.latencies.structural": func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Structural },
	"search.latencies.file":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.File },
	"search.latencies.repo":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Repo },
	"search.latencies.diff":       func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Diff },
	"search.latencies.commit":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Commit },
	"search.latencies.symbol":     func(p *types.SearchUsagePeriod) *types.SearchEventStatistics { return p.Symbol },
}

func insertSearchEventStatistics(event types.AggregatedEvent, statistics *types.SearchUsageStatistics) {
	extractor, ok := searchExtractors[event.Name]
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
