package alertpolicy

import (
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newPromQLCondition(config *Config) *monitoringalertpolicy.MonitoringAlertPolicyConditions {
	if config.PromQL.Duration == nil {
		config.PromQL.Duration = pointers.Ptr("60s")
	}

	return &monitoringalertpolicy.MonitoringAlertPolicyConditions{
		DisplayName: pointers.Ptr(config.Name),
		ConditionPrometheusQueryLanguage: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage{
			Query:    pointers.Ptr(config.PromQL.Query),
			Duration: config.PromQL.Duration,
		},
	}
}
