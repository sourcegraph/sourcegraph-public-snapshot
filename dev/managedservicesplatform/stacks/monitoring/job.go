package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringnotificationchannel"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
)

func createJobAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	// Alert whenever a Cloud Run Job fails
	if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
		Service:       vars.Service,
		EnvironmentID: vars.EnvironmentID,

		ID:           "job_failures",
		Name:         "Cloud Run Job Failures",
		Description:  "Cloud Run Job executions failed",
		ProjectID:    vars.ProjectID,
		ResourceName: vars.Service.ID,
		ResourceKind: alertpolicy.CloudRunJob,
		ThresholdAggregation: &alertpolicy.ThresholdAggregation{
			Filters: map[string]string{
				"metric.type":          "run.googleapis.com/job/completed_task_attempt_count",
				"metric.labels.result": "failed",
			},
			GroupByFields: []string{"metric.label.result"},
			Aligner:       alertpolicy.MonitoringAlignCount,
			Reducer:       alertpolicy.MonitoringReduceSum,
			Period:        "60s",
			Threshold:     0,
		},
		NotificationChannels: channels,
	}); err != nil {
		return err
	}

	return nil
}
