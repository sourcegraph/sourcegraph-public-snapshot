package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

func createCommonAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) error {
	// Convert a spec.ServiceKind into a alertpolicy.ServiceKind
	serviceKind := alertpolicy.CloudRunService
	kind := vars.Service.GetKind()
	if kind == spec.ServiceKindJob {
		serviceKind = alertpolicy.CloudRunJob
	}

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
				Filters:   map[string]string{"metric.type": "run.googleapis.com/container/cpu/utilizations"},
				Aligner:   alertpolicy.MonitoringAlignPercentile99,
				Reducer:   alertpolicy.MonitoringReduceMax,
				Period:    "60s",
				Duration:  "600s",
				Threshold: 0.9,
			},
		},
		{
			ID:          "memory",
			Name:        "High Container Memory Utilization",
			Description: "High Memory Usage - it may be neccessary to reduce load or increase memory allocation",
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "run.googleapis.com/container/memory/utilizations"},
				Aligner:   alertpolicy.MonitoringAlignPercentile99,
				Reducer:   alertpolicy.MonitoringReduceMax,
				Period:    "300s",
				Threshold: 0.8,
			},
		},
		{
			ID:          "startup",
			Name:        "Container Startup Latency",
			Description: "Service containers are taking longer than configured timeouts to start up.",
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters: map[string]string{"metric.type": "run.googleapis.com/container/startup_latencies"},
				Aligner: alertpolicy.MonitoringAlignPercentile99,
				Reducer: alertpolicy.MonitoringReduceMax,
				Period:  "60s",
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
		if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			// Resource we are targetting in this helper
			ResourceKind: serviceKind,
			ResourceName: vars.Service.ID,

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
		}); err != nil {
			return err
		}
	}

	return nil
}
