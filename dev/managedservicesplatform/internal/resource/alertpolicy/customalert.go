package alertpolicy

import (
	"fmt"
	"strings"

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
				Query:    pointers.Ptr(flattenPromQLQuery(config.CustomAlert.Query)),
				Duration: pointers.Stringf("%ds", *config.CustomAlert.DurationMinutes*60),
			},
		}
	default:
		panic(fmt.Sprintf("unknown custom alert type %q", config.CustomAlert.Type))
	}
}

// GCP monitoring expects PromQL queries to be on a single line, so we do that
// automatically and also strip extra spaces for readability of the condensed
// version.
func flattenPromQLQuery(query string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(query, "\n", " ")), " ")
}
