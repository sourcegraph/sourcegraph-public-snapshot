package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelPolicies() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-policies",
		Title:       "Code Intelligence > Policies",
		Description: "The service at `internal/codeintel/policies`.",
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Source",
				Name:  "source",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_codeintel_policies_total{}",
					LabelName:     "app",
					ExampleOption: "frontend",
				},
				WildcardAllValue: true,
				Multi:            false,
			},
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewPoliciesServiceGroup("${source:regex}"),
			shared.CodeIntelligence.NewPoliciesStoreGroup("${source:regex}"),
			shared.CodeIntelligence.NewPoliciesGraphQLTransportGroup("${source:regex}"),
			shared.CodeIntelligence.NewRepoMatcherTaskGroup("${source:regex}"),
		},
	}
}
