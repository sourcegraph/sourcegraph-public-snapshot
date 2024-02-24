package monitoring

import (
	"fmt"
	"math"

	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func createJobAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) error {
	// Alert whenever a Cloud Run Job fails
	if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
		Service:       vars.Service,
		EnvironmentID: vars.EnvironmentID,

		ID:          "job_failures",
		Name:        "Cloud Run Job Failures",
		Description: "Cloud Run Job executions failed",
		ProjectID:   vars.ProjectID,
		ThresholdAggregation: &alertpolicy.ThresholdAggregation{
			ConditionBuilder: alertpolicy.ConditionBuilder{
				ResourceName: vars.Service.ID,
				ResourceKind: alertpolicy.CloudRunJob,
				Filters: map[string]string{
					"metric.type":          "run.googleapis.com/job/completed_task_attempt_count",
					"metric.labels.result": "failed",
				},
				GroupByFields: []string{"metric.label.result"},
				Aligner:       alertpolicy.MonitoringAlignCount,
				Reducer:       alertpolicy.MonitoringReduceSum,
				Period:        "60s",
			},
			Threshold:  0,
			Comparison: alertpolicy.ComparisonGT,
		},
		NotificationChannels: channels,
	}); err != nil {
		return errors.Wrap(err, "job_failures")
	}

	interval, err := vars.JobSchedule.FindMaxCronInterval()
	if err != nil {
		return errors.Wrap(err, "JobSchedule.FindMaxCronInterval")
	}
	if interval != nil {
		// Use the duration calculated from the cron, with some leeway. Use
		// math.Ceil just in case, as Minutes() may give us floats
		absentMinutes := int(math.Ceil(interval.Minutes())) + 10

		if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			Service:       vars.Service,
			EnvironmentID: vars.EnvironmentID,

			ID:          "job_execution_absence",
			Name:        "Cloud Run Job Execution Absence",
			Description: fmt.Sprintf("No Cloud Run Job executions were detected in expected window (%dm)", absentMinutes),
			ProjectID:   vars.ProjectID,
			MetricAbsence: &alertpolicy.MetricAbsence{
				ConditionBuilder: alertpolicy.ConditionBuilder{
					ResourceName: vars.Service.ID,
					ResourceKind: alertpolicy.CloudRunJob,
					Filters: map[string]string{
						"metric.type": "run.googleapis.com/job/completed_task_attempt_count",
					},
					Aligner: alertpolicy.MonitoringAlignCount,
					Reducer: alertpolicy.MonitoringReduceSum,
					Period:  "60s",
				},
				// Must be in seconds
				Duration: fmt.Sprintf("%ds", absentMinutes*60),
			},
			NotificationChannels: channels,
		}); err != nil {
			return errors.Wrap(err, "job_execution_absence")
		}
	}

	return nil
}
