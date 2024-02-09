package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
)

func createResponseCodeAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) error {
	for _, config := range vars.Monitoring.Alerts.ResponseCodeRatios {
		if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			Service:       vars.Service,
			EnvironmentID: vars.EnvironmentID,

			ID:           config.ID,
			ProjectID:    vars.ProjectID,
			Name:         config.Name,
			Description:  config.Description,
			ResourceName: vars.Service.ID,
			ResourceKind: alertpolicy.CloudRunService,
			ResponseCodeMetric: &alertpolicy.ResponseCodeMetric{
				Code:         config.Code,
				CodeClass:    config.CodeClass,
				ExcludeCodes: config.ExcludeCodes,
				Ratio:        config.Ratio,
				Duration:     config.Duration,
			},
			NotificationChannels: channels,
		}); err != nil {
			return err
		}
	}

	return nil
}
