package alertpolicy

import (
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newCustomAlertCondition(config *Config) *monitoringalertpolicy.MonitoringAlertPolicyConditions {
	if config.CustomAlert.DurationMinutes == nil {
		// default to 1 minute
		config.CustomAlert.DurationMinutes = pointers.Ptr(uint(1))
	}

	switch config.CustomAlert.Type {
	case spec.CustomAlertQueryTypeMQL:
		return &monitoringalertpolicy.MonitoringAlertPolicyConditions{
			DisplayName: pointers.Ptr(config.Name),
			ConditionMonitoringQueryLanguage: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage{
				Query:    pointers.Ptr(config.CustomAlert.Query),
				Duration: pointers.Stringf("%ds", *config.CustomAlert.DurationMinutes*60),
			},
		}
	case spec.CustomAlertQueryTypePromQL:
		return &monitoringalertpolicy.MonitoringAlertPolicyConditions{
			DisplayName: pointers.Ptr(config.Name),
			ConditionPrometheusQueryLanguage: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage{
				Query:    pointers.Ptr(config.CustomAlert.Query),
				Duration: pointers.Stringf("%ds", *config.CustomAlert.DurationMinutes*60),
			},
		}
	}

	return nil
}
