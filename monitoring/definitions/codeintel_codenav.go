package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelCodeNav() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-codenav",
		Title:       "Code Intelligence > Code Nav",
		Description: "The service at internal/codeintel/codenav`.",
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Source",
				Name:  "source",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_codeintel_codenav_total{}",
					LabelName:     "app",
					ExampleOption: "frontend",
				},
				WildcardAllValue: true,
				Multi:            false,
			},
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewCodeNavServiceGroup("${source:regex}"),
			shared.CodeIntelligence.NewCodeNavLsifStoreGroup("${source:regex}"),
			shared.CodeIntelligence.NewCodeNavGraphQLTransportGroup("${source:regex}"),
			shared.CodeIntelligence.NewCodeNavStoreGroup("${source:regex}"),
		},
	}
}
