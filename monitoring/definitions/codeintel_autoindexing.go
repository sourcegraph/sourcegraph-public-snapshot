package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelAutoIndexing() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-autoindexing",
		Title:       "Code Intelligence > Autoindexing",
		Description: "The service at `enterprise/internal/codeintel/autoindexing`.",
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Source",
				Name:  "source",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_codeintel_autoindexing_total{}",
					LabelName:     "app",
					ExampleOption: "frontend",
				},
				WildcardAllValue: true,
				Multi:            false,
			},
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewAutoindexingSummaryGroup("${source:regex}"),
			shared.CodeIntelligence.NewAutoindexingServiceGroup("${source:regex}"),
			shared.CodeIntelligence.NewAutoindexingGraphQLTransportGroup("${source:regex}"),
			shared.CodeIntelligence.NewAutoindexingStoreGroup("${source:regex}"),
			shared.CodeIntelligence.NewAutoindexingBackgroundJobGroup("${source:regex}"),
			shared.CodeIntelligence.NewAutoindexingInferenceServiceGroup("${source:regex}"),
			shared.CodeIntelligence.NewLuasandboxServiceGroup("${source:regex}"),
		},
	}
}
