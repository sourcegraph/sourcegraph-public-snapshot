package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelRanking() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-ranking",
		Title:       "Code Intelligence > Ranking",
		Description: "The service at `enterprise/internal/codeintel/ranking`.",
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Source",
				Name:  "source",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_codeintel_ranking_total{}",
					LabelName:     "app",
					ExampleOption: "frontend",
				},
				WildcardAllValue: true,
				Multi:            false,
			},
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewRankingServiceGroup("${source:regex}"),
			shared.CodeIntelligence.NewRankingStoreGroup("${source:regex}"),
			shared.CodeIntelligence.NewRankingPageRankGroup("${source:regex}"),
		},
	}
}
