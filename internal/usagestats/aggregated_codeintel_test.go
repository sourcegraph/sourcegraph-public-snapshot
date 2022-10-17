package usagestats

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGroupAggregatedCodeIntelStats(t *testing.T) {
	lang1 := "go"
	lang2 := "typescript"
	t1 := time.Now().UTC().Add(time.Hour)

	codeIntelStats := groupAggregatedCodeIntelStats([]types.CodeIntelAggregatedEvent{
		{Name: "codeintel.lsifHover", Week: t1, TotalWeek: 10, UniquesWeek: 1},
		{Name: "codeintel.searchDefinitions", Week: t1, TotalWeek: 20, UniquesWeek: 2, LanguageID: &lang1},
		{Name: "codeintel.lsifDefinitions", Week: t1, TotalWeek: 30, UniquesWeek: 3},
		{Name: "codeintel.searchReferences.xrepo", Week: t1, TotalWeek: 40, UniquesWeek: 4, LanguageID: &lang2},
	}, nil)

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
