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
	events, err := db.EventLogs().AggregatedCodyEvents(ctx, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return groupAggregatedCodyStats(events), nil
}

// groupAggregatedCodyStats takes a set of input events (originating from
// Sourcegraph's Postgres table) and returns a CodyUsageStatistics data type
// that ends up being stored in BigQuery. CodyUsageStatistics corresponds to
// the target DB schema.
func groupAggregatedCodyStats(events []types.CodyAggregatedEvent) *types.CodyUsageStatistics {
	codyUsageStats := &types.CodyUsageStatistics{
		Daily:   []*types.CodyUsagePeriod{newCodyEventPeriod()},
		Weekly:  []*types.CodyUsagePeriod{newCodyEventPeriod()},
		Monthly: []*types.CodyUsagePeriod{newCodyEventPeriod()},
	}

	// Iterate over events, updating codyUsageStats for each event
	for _, event := range events {
		populateCodyCountStatistics(event, codyUsageStats)
	}

	return codyUsageStats
}

// utility functions that resolve a CodyCountStatistics value for a given event name for some CodyUsagePeriod.
var codyEventCountExtractors = map[string]func(p *types.CodyUsagePeriod) *types.CodyCountStatistics{
	"CodyVSCodeExtension:recipe:rewrite-to-functional:executed":   func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:improve-variable-names:executed":  func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:replace:executed":                 func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:generate-docstring:executed":      func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:generate-unit-test:executed":      func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:rewrite-functional:executed":      func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:code-refactor:executed":           func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:fixup:executed":                   func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:translate-to-language:executed":   func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.CodeGenerationRequests },
	"CodyVSCodeExtension:recipe:explain-code-high-level:executed": func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.ExplanationRequests },
	"CodyVSCodeExtension:recipe:explain-code-detailed:executed":   func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.ExplanationRequests },
	"CodyVSCodeExtension:recipe:find-code-smells:executed":        func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.ExplanationRequests },
	"CodyVSCodeExtension:recipe:git-history:executed":             func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.ExplanationRequests },
	"CodyVSCodeExtension:recipe:rate-code:executed":               func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.ExplanationRequests },
	"CodyVSCodeExtension:recipe:chat-question:executed":           func(p *types.CodyUsagePeriod) *types.CodyCountStatistics { return p.TotalRequests },
}

func populateCodyCountStatistics(event types.CodyAggregatedEvent, statistics *types.CodyUsageStatistics) {
	extractor, ok := codyEventCountExtractors[event.Name]
	if !ok {
		return
	}

	statistics.Monthly[0].StartTime = event.Month
	month := extractor(statistics.Monthly[0])
	month.EventsCount = &event.TotalMonth
	month.UserCount = &event.UniquesMonth

	statistics.Weekly[0].StartTime = event.Week
	week := extractor(statistics.Weekly[0])
	week.EventsCount = &event.TotalWeek
	week.UserCount = &event.UniquesWeek

	statistics.Daily[0].StartTime = event.Day
	day := extractor(statistics.Daily[0])
	day.EventsCount = &event.TotalDay
	day.UserCount = &event.UniquesDay
}

func newCodyEventPeriod() *types.CodyUsagePeriod {
	return &types.CodyUsagePeriod{
		StartTime:              time.Now().UTC(),
		TotalUsers:             newCodyCountStatistics(),
		TotalRequests:          newCodyCountStatistics(),
		CodeGenerationRequests: newCodyCountStatistics(),
		ExplanationRequests:    newCodyCountStatistics(),
		InvalidRequests:        newCodyCountStatistics(),
	}
}

func newCodyCountStatistics() *types.CodyCountStatistics {
	return &types.CodyCountStatistics{}
}
