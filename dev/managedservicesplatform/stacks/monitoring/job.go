package monitoring

import (
	"fmt"
	"math"
	"time"

	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func createJobAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) ([]monitoringalertpolicy.MonitoringAlertPolicy, error) {
	// Collect all alerts to aggregate in a dashboard
	var alerts []monitoringalertpolicy.MonitoringAlertPolicy

	// Alert whenever a Cloud Run Job fails
	alert, err := alertpolicy.New(stack, id, &alertpolicy.Config{
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
	})
	if err != nil {
		return nil, errors.Wrap(err, "job_failures")
	}
	alerts = append(alerts, alert.AlertPolicy)

	interval, err := vars.JobSchedule.FindMaxCronInterval(time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "JobSchedule.FindMaxCronInterval")
	}
	if interval != nil {
		// Use the duration calculated from the cron, with some leeway. Use
		// math.Ceil just in case, as Minutes() may give us floats
		absentMinutes := int(math.Ceil(interval.Minutes())) + 10
		// GCP does not allow absence alerts above 23h30m - for jobs that run
		// at a longer interval, depend on the implementor using the MSP job
		// runtime to report runs to Sentry, and have Sentry handle alerting
		// on missing executions instead.
		if absentMinutes < int((23*time.Hour + 30*time.Minute).Minutes()) {
			alert, err := alertpolicy.New(stack, id, &alertpolicy.Config{
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
			})
			if err != nil {
				return nil, errors.Wrap(err, "job_execution_absence")
			}
			alerts = append(alerts, alert.AlertPolicy)
		}

	}

	return alerts, nil
}
