package alertpolicy

import (
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newMetricAbsenceCondition(config *Config) *monitoringalertpolicy.MonitoringAlertPolicyConditions {
	return &monitoringalertpolicy.MonitoringAlertPolicyConditions{
		DisplayName: pointers.Ptr(config.Name),
		ConditionAbsent: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionAbsent{
			Aggregations: config.MetricAbsence.buildAbsentAggregations(),
			Duration:     pointers.Ptr(config.MetricAbsence.Duration),
			Filter:       pointers.Ptr(config.MetricAbsence.buildFilter()),
			Trigger:      config.MetricAbsence.buildAbsentTrigger(),
		},
	}
}
