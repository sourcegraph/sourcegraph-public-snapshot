package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelRanking() *monitoring.Dashboard {
	groups := []monitoring.Group{
		shared.CodeIntelligence.NewRankingServiceGroup("${source:regex}"),
		shared.CodeIntelligence.NewRankingStoreGroup("${source:regex}"),
		shared.CodeIntelligence.NewRankingLSIFStoreGroup("${source:regex}"),
	}
	groups = append(groups, shared.CodeIntelligence.NewRankingPipelineTaskGroups("${source:regex}")...)
	groups = append(groups, shared.CodeIntelligence.NewRankingJanitorTaskGroups("${source:regex}")...)

	return &monitoring.Dashboard{
		Name:        "codeintel-ranking",
		Title:       "Code Intelligence > Ranking",
		Description: "The service at `internal/codeintel/ranking`.",
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
		Groups: groups,
	}
}
