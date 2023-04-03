package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	configurationPolicies     *observation.Operation
	configurationPolicyByID   *observation.Operation
	createConfigurationPolicy *observation.Operation
	deleteConfigurationPolicy *observation.Operation
	previewGitObjectFilter    *observation.Operation
	previewRepoFilter         *observation.Operation
	updateConfigurationPolicy *observation.Operation
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
		configurationPolicies:     op("ConfigurationPolicies"),
		configurationPolicyByID:   op("ConfigurationPolicyByID"),
		createConfigurationPolicy: op("CreateConfigurationPolicy"),
		deleteConfigurationPolicy: op("DeleteConfigurationPolicy"),
		previewGitObjectFilter:    op("PreviewGitObjectFilter"),
		previewRepoFilter:         op("PreviewRepoFilter"),
		updateConfigurationPolicy: op("UpdateConfigurationPolicy"),
	}
}
