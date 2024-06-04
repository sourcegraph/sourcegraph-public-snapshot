package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func createCustomAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) ([]monitoringalertpolicy.MonitoringAlertPolicy, error) {
	// Collect all alerts to aggregate in a dashboard
	var alerts []monitoringalertpolicy.MonitoringAlertPolicy

	// Iterate over a list of custom alert configurations.
	for _, config := range vars.Monitoring.Alerts.CustomAlerts {

		alert, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			// Alert policy
			ID:          config.ID,
			Name:        config.Name,
			Description: config.Description,
			Severity:    config.SeverityLevel,

			CustomAlert: &config.Condition,

			// Shared configuration
			Service:              vars.Service,
			EnvironmentID:        vars.EnvironmentID,
			ProjectID:            vars.ProjectID,
			NotificationChannels: channels,
		})
		if err != nil {
			return nil, errors.Wrap(err, config.ID)
		}
		alerts = append(alerts, alert.AlertPolicy)
	}

	return alerts, nil
}
