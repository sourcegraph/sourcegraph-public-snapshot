package policies

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Configurations
	getConfigurationPolicies      *observation.Operation
	getConfigurationPoliciesByID  *observation.Operation
	createConfigurationPolicy     *observation.Operation
	updateConfigurationPolicy     *observation.Operation
	deleteConfigurationPolicyByID *observation.Operation

	// Retention Policy
	getRetentionPolicyOverview *observation.Operation

	// Repository
	getPreviewRepositoryFilter                  *observation.Operation
	getPreviewGitObjectFilter                   *observation.Operation
	selectPoliciesForRepositoryMembershipUpdate *observation.Operation
	updateReposMatchingPatterns                 *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_policies",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.policies.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	return &operations{
		// Configurations
		getConfigurationPolicies:      op("GetConfigurationPolicies"),
		getConfigurationPoliciesByID:  op("GetConfigurationPoliciesByID"),
		createConfigurationPolicy:     op("CreateConfigurationPolicy"),
		updateConfigurationPolicy:     op("UpdateConfigurationPolicy"),
		deleteConfigurationPolicyByID: op("DeleteConfigurationPolicyByID"),

		// Retention
		getRetentionPolicyOverview: op("GetRetentionPolicyOverview"),

		// Repository
		getPreviewRepositoryFilter:                  op("GetPreviewRepositoryFilter"),
		getPreviewGitObjectFilter:                   op("GetPreviewGitObjectFilter"),
		selectPoliciesForRepositoryMembershipUpdate: op("SelectPoliciesForRepositoryMembershipUpdate"),
		updateReposMatchingPatterns:                 op("UpdateReposMatchingPatterns"),
	}
}
