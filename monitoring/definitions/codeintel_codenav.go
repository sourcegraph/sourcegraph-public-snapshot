package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelCodeNav() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-codenav",
		Title:       "Code Intelligence > Code Nav",
		Description: "The service at `internal/codeintel/codenav`.",
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
			shared.CodeIntelligence.NewCodeNavServiceGroup("$source"),
			shared.CodeIntelligence.NewCodeNavLsifStoreGroup("$source"),
			shared.CodeIntelligence.NewCodeNavGraphQLTransportGroup("$source"),
			shared.CodeIntelligence.NewCodeNavStoreGroup("$source"),
		},
	}
}
