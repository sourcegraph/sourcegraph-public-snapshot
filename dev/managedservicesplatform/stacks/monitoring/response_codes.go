package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
)

func createResponseCodeAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) ([]monitoringalertpolicy.MonitoringAlertPolicy, error) {
	// Collect all alerts to aggregate in a dashboard
	var alerts []monitoringalertpolicy.MonitoringAlertPolicy

	for _, config := range vars.Monitoring.Alerts.ResponseCodeRatios {
		alert, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			Service:       vars.Service,
			EnvironmentID: vars.EnvironmentID,

			ID:          config.ID,
			ProjectID:   vars.ProjectID,
			Name:        config.Name,
			Description: config.Description,

			ResponseCodeMetric: &alertpolicy.ResponseCodeMetric{
				Code:            config.Code,
				CodeClass:       config.CodeClass,
				ExcludeCodes:    config.ExcludeCodes,
				Ratio:           config.Ratio,
				DurationMinutes: config.DurationMinutes,
			},
			NotificationChannels: channels,
		})
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert.AlertPolicy)
	}

	return alerts, nil
}
