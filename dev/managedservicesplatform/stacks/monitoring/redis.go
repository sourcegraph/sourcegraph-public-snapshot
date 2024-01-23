package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringnotificationchannel"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
)

func createRedisAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	// Iterate over a list of Redis alert configurations. Custom struct defines
	// the field we expect to vary between each.
	for _, config := range []struct {
		ID                   string
		Name                 string
		Description          string
		ThresholdAggregation *alertpolicy.ThresholdAggregation
	}{
		{
			ID:          "memory",
			Name:        "Cloud Redis - System Memory Utilization",
			Description: "Redis System memory utilization is above the set threshold. The utilization is measured on a scale of 0 to 1.",
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "redis.googleapis.com/stats/memory/system_memory_usage_ratio"},
				Aligner:   alertpolicy.MonitoringAlignMean,
				Reducer:   alertpolicy.MonitoringReduceNone,
				Period:    "300s",
				Threshold: 0.8,
			},
		},
		{
			ID:          "cpu",
			Name:        "Cloud Redis - System CPU Utilization",
			Description: "Redis Engine CPU Utilization goes above the set threshold. The utilization is measured on a scale of 0 to 1.",
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters:       map[string]string{"metric.type": "redis.googleapis.com/stats/cpu_utilization_main_thread"},
				GroupByFields: []string{"resource.label.instance_id", "resource.label.node_id"},
				Aligner:       alertpolicy.MonitoringAlignRate,
				Reducer:       alertpolicy.MonitoringReduceSum,
				Period:        "300s",
				Threshold:     0.9,
			},
		},
		{
			ID:          "failover",
			Name:        "Cloud Redis - Standard Instance Failover",
			Description: "Instance failover occured for a standard tier Redis instance.",
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "redis.googleapis.com/replication/role"},
				Aligner:   alertpolicy.MonitoringAlignStddev,
				Period:    "300s",
				Threshold: 0,
			},
		},
	} {
		if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			// Resource we are targetting in this helper
			ResourceKind: alertpolicy.CloudRedis,
			ResourceName: *vars.RedisInstanceID,

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
