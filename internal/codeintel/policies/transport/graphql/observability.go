package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	codeIntelligenceConfiogurationPolicies    *observation.Operation
	configurationPolicyByID                   *observation.Operation
	createCodeIntelligenceConfigurationPolicy *observation.Operation
	deleteCodeIntelligenceConfigurationPolicy *observation.Operation
	previewGitObjectFilter                    *observation.Operation
	previewRepositoryFilter                   *observation.Operation
	updateCodeIntelligenceConfigurationPolicy *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_policies_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.policies.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		codeIntelligenceConfiogurationPolicies:    op("CodeIntelligenceConfiogurationPolicies"),
		configurationPolicyByID:                   op("ConfigurationPolicyByID"),
		createCodeIntelligenceConfigurationPolicy: op("CreateCodeIntelligenceConfigurationPolicy"),
		deleteCodeIntelligenceConfigurationPolicy: op("DeleteCodeIntelligenceConfigurationPolicy"),
		previewGitObjectFilter:                    op("PreviewGitObjectFilter"),
		previewRepositoryFilter:                   op("PreviewRepositoryFilter"),
		updateCodeIntelligenceConfigurationPolicy: op("UpdateCodeIntelligenceConfigurationPolicy"),
	}
}
