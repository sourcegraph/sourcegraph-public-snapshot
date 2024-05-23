package alertpolicy

import (
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newMQLCondition(config *Config) *monitoringalertpolicy.MonitoringAlertPolicyConditions {
	if config.MQL.Duration == nil {
		config.MQL.Duration = pointers.Ptr("60s")
	}

	return &monitoringalertpolicy.MonitoringAlertPolicyConditions{
		DisplayName: pointers.Ptr(config.Name),
		ConditionMonitoringQueryLanguage: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage{
			Query:    pointers.Ptr(config.MQL.Query),
			Duration: config.MQL.Duration,
		},
	}
}
