package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func createCommonAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) ([]monitoringalertpolicy.MonitoringAlertPolicy, error) {
	// Convert a spec.ServiceKind into a alertpolicy.ServiceKind
	serviceKind := alertpolicy.CloudRunService
	kind := vars.Service.GetKind()
	if kind == spec.ServiceKindJob {
		serviceKind = alertpolicy.CloudRunJob
	}

	// Collect all alerts to aggregate in a dashboard
	var alerts []monitoringalertpolicy.MonitoringAlertPolicy

	// Iterate over a list of Redis alert configurations. Custom struct defines
	// the field we expect to vary between each.
	for _, config := range []struct {
		ID                   string
		Name                 string
		Description          string
		ThresholdAggregation *alertpolicy.ThresholdAggregation
	}{
		{
			ID:          "cpu",
			Name:        "High Container CPU Utilization",
			Description: "High CPU Usage - it may be neccessary to reduce load or increase CPU allocation",
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				ConditionBuilder: alertpolicy.ConditionBuilder{
					Filters: map[string]string{"metric.type": "run.googleapis.com/container/cpu/utilizations"},
					Aligner: alertpolicy.MonitoringAlignPercentile99,
					Reducer: alertpolicy.MonitoringReduceMax,
					Period:  "60s",
				},
				Duration:  "600s",
				Threshold: 0.9,
			},
		},
		{
			ID:          "memory",
			Name:        "High Container Memory Utilization",
			Description: "High Memory Usage - it may be neccessary to reduce load or increase memory allocation",
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				ConditionBuilder: alertpolicy.ConditionBuilder{
					Filters: map[string]string{"metric.type": "run.googleapis.com/container/memory/utilizations"},
					Aligner: alertpolicy.MonitoringAlignPercentile99,
					Reducer: alertpolicy.MonitoringReduceMax,
					Period:  "300s",
				},
				Threshold: 0.8,
			},
		},
		{
			ID:          "startup",
			Name:        "Container Startup Latency",
			Description: "Service containers are taking longer than configured timeouts to start up.",
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				ConditionBuilder: alertpolicy.ConditionBuilder{
					Filters: map[string]string{"metric.type": "run.googleapis.com/container/startup_latencies"},
					Aligner: alertpolicy.MonitoringAlignPercentile99,
					Reducer: alertpolicy.MonitoringReduceMax,
					Period:  "60s",
				},
				Threshold: func() float64 {
					if serviceKind == alertpolicy.CloudRunJob {
						// jobs measure container startup, not service startup,
						// this should never take very long
						return 10 * 1000 // ms
					}
					// otherwise, use the startup probe configuration to
					// determine the threshold for how long we should be waiting
					return float64(vars.ServiceHealthProbes.MaximumStartupLatencySeconds()) * 1000 // ms
				}(),
			},
		},
	} {
		// Resource we are targeting in this helper
		config.ThresholdAggregation.ResourceKind = serviceKind
		config.ThresholdAggregation.ResourceName = vars.Service.ID

		alert, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			// Alert policy
			ID:                   config.ID,
			Name:                 config.Name,
			Description:          config.Description,
			ThresholdAggregation: config.ThresholdAggregation,

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
