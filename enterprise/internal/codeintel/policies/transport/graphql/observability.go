package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Configurations
	createConfigurationPolicy *observation.Operation
	configurationPolicies     *observation.Operation
	configurationPolicyByID   *observation.Operation
	updateConfigurationPolicy *observation.Operation
	deleteConfigurationPolicy *observation.Operation

	// Retention
	previewGitObjectFilter *observation.Operation

	// Repository
	previewRepoFilter *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_policies_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.policies.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		// Configurations
		createConfigurationPolicy: op("CreateConfigurationPolicy"),
		configurationPolicies:     op("ConfigurationPolicies"),
		configurationPolicyByID:   op("ConfigurationPolicyByID"),
		updateConfigurationPolicy: op("UpdateConfigurationPolicy"),
		deleteConfigurationPolicy: op("DeleteConfigurationPolicy"),

		// Retention
		previewGitObjectFilter: op("PreviewGitObjectFilter"),

		// Repository
		previewRepoFilter: op("PreviewRepoFilter"),
	}
}
