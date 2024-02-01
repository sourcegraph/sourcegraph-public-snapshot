package alertpolicy

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newPromQLAlert(scope constructs.Construct, id resourceid.ID, config *Config) (*Output, error) {

	if config.ResponseCodeMetric.Duration == nil {
		config.ResponseCodeMetric.Duration = pointers.Ptr("60s")
	}

	_ = monitoringalertpolicy.NewMonitoringAlertPolicy(scope,
		id.TerraformID(config.ID), &monitoringalertpolicy.MonitoringAlertPolicyConfig{
			Project:     pointers.Ptr(config.ProjectID),
			DisplayName: &config.Name,
			Documentation: &monitoringalertpolicy.MonitoringAlertPolicyDocumentation{
				Content:  pointers.Ptr(config.Description),
				MimeType: pointers.Ptr("text/markdown"),
			},
			Combiner: pointers.Ptr("OR"),
			Conditions: []monitoringalertpolicy.MonitoringAlertPolicyConditions{
				{
					DisplayName: pointers.Ptr(config.Name),
					ConditionPrometheusQueryLanguage: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage{
						Query:              &config.PromQL.Query,
						Duration:           &config.PromQL.Duration,
						EvaluationInterval: &config.PromQL.EvaluationInterval,
					},
				},
			},
			AlertStrategy: &monitoringalertpolicy.MonitoringAlertPolicyAlertStrategy{
				AutoClose: pointers.Ptr("604800s"),
			},
			NotificationChannels: notificationChannelIDs(config.NotificationChannels[config.Severity]),
		})
	return &Output{}, nil
}
