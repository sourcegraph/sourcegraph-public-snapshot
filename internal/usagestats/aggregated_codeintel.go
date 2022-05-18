package usagestats

import (
	"context"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// GetAggregatedCodeIntelStats returns aggregated statistics for code intelligence usage.
func GetAggregatedCodeIntelStats(ctx context.Context, db database.DB) (*types.NewCodeIntelUsageStatistics, error) {
	eventLogs := db.EventLogs()

	codeIntelEvents, err := eventLogs.AggregatedCodeIntelEvents(ctx)
	if err != nil {
		return nil, err
	}
	codeIntelInvestigationEvents, err := eventLogs.AggregatedCodeIntelInvestigationEvents(ctx)
	if err != nil {
		return nil, err
	}
	if len(codeIntelEvents) == 0 && len(codeIntelInvestigationEvents) == 0 {
		return nil, nil
	}

	// NOTE: this requires at least one of these to be non-empty (to get an initial date)
	stats := groupAggregatedCodeIntelStats(codeIntelEvents, codeIntelInvestigationEvents)

	pairs := []struct {
		fetch  func(ctx context.Context) (int, error)
		target **int32
	}{
		{eventLogs.CodeIntelligenceWAUs, &stats.WAUs},
		{eventLogs.CodeIntelligencePreciseWAUs, &stats.PreciseWAUs},
		{eventLogs.CodeIntelligenceSearchBasedWAUs, &stats.SearchBasedWAUs},
		{eventLogs.CodeIntelligenceCrossRepositoryWAUs, &stats.CrossRepositoryWAUs},
		{eventLogs.CodeIntelligencePreciseCrossRepositoryWAUs, &stats.PreciseCrossRepositoryWAUs},
		{eventLogs.CodeIntelligenceSearchBasedCrossRepositoryWAUs, &stats.SearchBasedCrossRepositoryWAUs},
	}

	for _, pair := range pairs {
		count, err := pair.fetch(ctx)
		if err != nil {
			return nil, err
		}

		v := int32(count)
		*pair.target = &v
	}

	counts, err := eventLogs.CodeIntelligenceRepositoryCounts(ctx)
	if err != nil {
		return nil, err
	}

	countsByLanguage, err := eventLogs.CodeIntelligenceRepositoryCountsByLanguage(ctx)
	if err != nil {
		return nil, err
	}

	settingsPageViewCount, err := eventLogs.CodeIntelligenceSettingsPageViewCount(ctx)
	if err != nil {
		return nil, err
	}

	stats.NumRepositories = int32Ptr(counts.NumRepositories)
	stats.NumRepositoriesWithUploadRecords = int32Ptr(counts.NumRepositoriesWithUploadRecords)
	stats.NumRepositoriesWithFreshUploadRecords = int32Ptr(counts.NumRepositoriesWithFreshUploadRecords)
	stats.NumRepositoriesWithIndexRecords = int32Ptr(counts.NumRepositoriesWithIndexRecords)
	stats.NumRepositoriesWithFreshIndexRecords = int32Ptr(counts.NumRepositoriesWithFreshIndexRecords)
	stats.NumRepositoriesWithAutoIndexConfigurationRecords = int32Ptr(counts.NumRepositoriesWithAutoIndexConfigurationRecords)
	stats.SettingsPageViewCount = int32Ptr(settingsPageViewCount)

	stats.CountsByLanguage = make(map[string]types.CodeIntelRepositoryCountsByLanguage, len(countsByLanguage))
	for language, counts := range countsByLanguage {
		stats.CountsByLanguage[language] = types.CodeIntelRepositoryCountsByLanguage{
			NumRepositoriesWithUploadRecords:      int32Ptr(counts.NumRepositoriesWithUploadRecords),
			NumRepositoriesWithFreshUploadRecords: int32Ptr(counts.NumRepositoriesWithFreshUploadRecords),
			NumRepositoriesWithIndexRecords:       int32Ptr(counts.NumRepositoriesWithIndexRecords),
			NumRepositoriesWithFreshIndexRecords:  int32Ptr(counts.NumRepositoriesWithFreshIndexRecords),
		}
	}

	requestCountsByLanguage, err := eventLogs.RequestsByLanguage(ctx)
	if err != nil {
		return nil, err
	}

	languageRequests := make([]types.LanguageRequest, 0, len(countsByLanguage))
	for languageID, value := range requestCountsByLanguage {
		languageRequests = append(languageRequests, types.LanguageRequest{
			LanguageID:  languageID,
			NumRequests: int32(value),
		})
	}
	stats.LanguageRequests = languageRequests

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

var investigationTypeMap = map[string]types.CodeIntelInvestigationType{
	"CodeIntelligenceIndexerSetupInvestigated": types.CodeIntelIndexerSetupInvestigationType,
	"CodeIntelligenceUploadErrorInvestigated":  types.CodeIntelUploadErrorInvestigationType,
	"CodeIntelligenceIndexErrorInvestigated":   types.CodeIntelIndexErrorInvestigationType,
}

func groupAggregatedCodeIntelStats(
	rawEvents []types.CodeIntelAggregatedEvent,
	rawInvestigationEvents []types.CodeIntelAggregatedInvestigationEvent,
) *types.NewCodeIntelUsageStatistics {
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

	var investigationEvents []types.CodeIntelInvestigationEvent
	for _, event := range rawInvestigationEvents {
		investigationEvents = append(investigationEvents, types.CodeIntelInvestigationEvent{
			Type:  investigationTypeMap[event.Name],
			WAUs:  event.UniquesWeek,
			Total: event.TotalWeek,
		})
	}

	var startOfWeek time.Time
	if len(rawEvents) > 0 {
		startOfWeek = rawEvents[0].Week
	} else if len(rawInvestigationEvents) > 0 {
		startOfWeek = rawInvestigationEvents[0].Week
	}

	return &types.NewCodeIntelUsageStatistics{
		StartOfWeek:         startOfWeek,
		EventSummaries:      eventSummaries,
		InvestigationEvents: investigationEvents,
	}
}
