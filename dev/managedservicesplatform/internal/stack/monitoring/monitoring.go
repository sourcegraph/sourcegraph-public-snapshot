package monitoring

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringnotificationchannel"
	opsgenieintegration "github.com/sourcegraph/managed-services-platform-cdktf/gen/opsgenie/apiintegration"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/opsgenie/dataopsgenieteam"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/opsgenieprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Common
// - Container (8)
//    - run.googleapis.com/container/billable_instance_time
//    - run.googleapis.com/container/cpu/allocation_time
//    * run.googleapis.com/container/cpu/utilizations
//    - run.googleapis.com/container/memory/allocation_time
//    * run.googleapis.com/container/memory/utilizations
//    * run.googleapis.com/container/startup_latencies
//    - run.googleapis.com/container/network/received_bytes_count
//    - run.googleapis.com/container/network/sent_bytes_count
// - Log-based metrics (2)
//    - logging.googleapis.com/byte_count
//    - logging.googleapis.com/log_entry_count
// Cloud Run Job
// - Job (4)
//    - run.googleapis.com/job/completed_execution_count
//    * run.googleapis.com/job/completed_task_attempt_count
//    - run.googleapis.com/job/running_executions
//    - run.googleapis.com/job/running_task_attempts
// Cloud Run Service
// - Container (9)
//    - run.googleapis.com/container/completed_probe_attempt_count
//    - run.googleapis.com/container/completed_probe_count
//    - run.googleapis.com/container/probe_attempt_latencies
//    - run.googleapis.com/container/probe_latencies
//    * run.googleapis.com/container/instance_count
//    - run.googleapis.com/container/max_request_concurrencies
//    - run.googleapis.com/container/cpu/usage
//    - run.googleapis.com/container/containers
//    - run.googleapis.com/container/memory/usage
// - Request_count (1)
//    - run.googleapis.com/request_count
// - Request_latencies (1)
//    * run.googleapis.com/request_latencies
// - Pending_queue (1)
//    - run.googleapis.com/pending_queue/pending_requests

type CrossStackOutput struct{}

type Variables struct {
	ProjectID  string
	Service    spec.ServiceSpec
	Monitoring spec.MonitoringSpec

	// MaxInstanceCount informs service scaling alerts.
	MaxInstanceCount *int
	// If Redis is enabled we configure alerts for it
	RedisInstanceID *string

	// EnvironmentCategory dictates what kind of notifications are set up:
	//
	// 1. 'test' services only generate Slack notifications.
	// 2. 'internal' and 'external' services generate Slack and Opsgenie notifications.
	//
	// Slack channels are expected to be named '#alerts-<service>-<environmentName>'.
	// Opsgenie teams are expected to correspond to service owners.
	//
	// Both Slack channels and Opsgenie teams are currently expected to be manually
	// configured. In particular, it seems that there is not a well-maintained
	// Terraform provider for Slack.
	EnvironmentCategory spec.EnvironmentCategory
	// EnvironmentID is the name of the service environment.
	EnvironmentID string
	// Owners is a list of team names. Each owner MUST correspond to the name
	// of a team in Opsgenie.
	Owners []string
}

const StackName = "monitoring"

func NewStack(stacks *stack.Set, vars Variables) (*CrossStackOutput, error) {
	stack, _, err := stacks.New(StackName,
		googleprovider.With(vars.ProjectID),
		opsgenieprovider.With(gsmsecret.DataConfig{
			Secret:    googlesecretsmanager.SecretOpsgenieAPIToken,
			ProjectID: googlesecretsmanager.ProjectID,
		}))
	if err != nil {
		return nil, err
	}

	id := resourceid.New("monitoring")

	// Prepare GCP monitoring channels on which to notify on when an alert goes
	// off.
	var channels []monitoringnotificationchannel.MonitoringNotificationChannel

	// Configure opsgenie channels
	var opsgenieAlerts bool
	switch vars.EnvironmentCategory {
	case spec.EnvironmentCategoryInternal, spec.EnvironmentCategoryExternal:
		opsgenieAlerts = true
	}
	for i, owner := range vars.Owners {
		// Use index because Opsgenie team names has lax character requirements
		id := id.Group("opsgenie_owner_%d", i)
		// Opsgenie team corresponding to owner must exist
		team := dataopsgenieteam.NewDataOpsgenieTeam(stack,
			id.TerraformID("opsgenie_team"),
			&dataopsgenieteam.DataOpsgenieTeamConfig{
				Name: &owner,
			})
		// Set up integration for us to post to
		integration := opsgenieintegration.NewApiIntegration(stack,
			id.TerraformID("opsgenie_integration"),
			&opsgenieintegration.ApiIntegrationConfig{
				// Must be unique, so include the TF team ID in it. Each team
				// will only get one integration per service environment.
				Name: pointers.Stringf("msp-%s-%s-%s",
					vars.Service.ID, vars.EnvironmentID, *team.Id()),
				// https://support.atlassian.com/opsgenie/docs/integration-types-to-be-used-with-the-api/
				Type: pointers.Ptr("GoogleStackdriver"),

				// Let the team own the integration.
				OwnerTeamId: team.Id(),

				// Supress all notifications if opsgenieAlerts is disabled -
				// this allows us to see the alerts, but not necessarily get
				// paged by it.
				SuppressNotifications: pointers.Ptr(!opsgenieAlerts),

				// Point alerts sent through this integration at the Opsgenie team.
				Responders: []*opsgenieintegration.ApiIntegrationResponders{{
					// Possible values for Type are:
					// team, user, escalation and schedule
					Type: pointers.Ptr("team"),
					Id:   team.Id(),
				}},
			})

		channels = append(channels,
			monitoringnotificationchannel.NewMonitoringNotificationChannel(stack,
				id.TerraformID("notification_channel"),
				&monitoringnotificationchannel.MonitoringNotificationChannelConfig{
					Project:     &vars.ProjectID,
					DisplayName: pointers.Stringf("Opsgenie - %s", owner),
					Type:        pointers.Ptr("webhook_tokenauth"),
					Labels: &map[string]*string{
						// This is kind of a secret, but we can't put this in
						// sensitive_labels so here it is. It seems we do this
						// in Cloud as well.
						"url": pointers.Stringf(
							"https://api.opsgenie.com/v1/json/googlestackdriver?apiKey=%s",
							*integration.ApiKey()),
					},
				}))
	}

	// Configure Slack channels
	slackToken := gsmsecret.Get(stack, id.Group("slack_token"), gsmsecret.DataConfig{
		Secret:    googlesecretsmanager.SecretSlackOAuthToken,
		ProjectID: googlesecretsmanager.ProjectID,
	})
	for _, channelName := range []string{
		"#alerts-msp", // central channel
		fmt.Sprintf("#alerts-%s-%s", // service-env-specific channel
			vars.Service.ID, vars.EnvironmentID),
	} {
		id := id.Group("slack_%s", strings.TrimPrefix(channelName, "#"))
		channels = append(channels,
			monitoringnotificationchannel.NewMonitoringNotificationChannel(stack,
				id.TerraformID("notification_channel"),
				&monitoringnotificationchannel.MonitoringNotificationChannelConfig{
					Project:     &vars.ProjectID,
					DisplayName: pointers.Stringf("Slack - %s", channelName),
					Type:        pointers.Ptr("slack"),
					Labels: &map[string]*string{
						"channel_name": pointers.Ptr(channelName),
					},
					SensitiveLabels: &monitoringnotificationchannel.MonitoringNotificationChannelSensitiveLabels{
						AuthToken: &slackToken.Value,
					},
				}))
	}

	// Set up alerts, configuring each with all our notification channels
	err = createCommonAlerts(stack, id.Group("common"), vars, channels)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create common alerts")
	}

	switch pointers.Deref(vars.Service.Kind, spec.ServiceKindService) {
	case spec.ServiceKindService:
		if err = createServiceAlerts(stack, id.Group("service"), vars, channels); err != nil {
			return nil, errors.Wrap(err, "failed to create service alerts")
		}

		if vars.Monitoring.Alerts.ResponseCodeRatios != nil {
			if err = createResponseCodeMetrics(stack, id.Group("response-code"), vars, channels); err != nil {
				return nil, errors.Wrap(err, "failed to create response code metrics")
			}
		}
	case spec.ServiceKindJob:
		if err = createJobAlerts(stack, id.Group("job"), vars, channels); err != nil {
			return nil, errors.Wrap(err, "failed to create job alerts")
		}
	default:
		return nil, errors.New("unknown service kind")
	}

	if vars.RedisInstanceID != nil {
		if err = createRedisAlerts(stack, id.Group("redis"), vars, channels); err != nil {
			return nil, errors.Wrap(err, "failed to create redis alerts")
		}
	}

	return &CrossStackOutput{}, nil
}

func createCommonAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	// Convert a spec.ServiceKind into a alertpolicy.ServiceKind
	serviceKind := alertpolicy.CloudRunService
	kind := pointers.Deref(vars.Service.Kind, "service")
	if kind == spec.ServiceKindJob {
		serviceKind = alertpolicy.CloudRunJob
	}

	for _, config := range []alertpolicy.Config{
		{
			ID:          "cpu",
			Name:        "High Container CPU Utilization",
			Description: pointers.Ptr("High CPU Usage - it may be neccessaru to reduce load or increase CPU allocation"),
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "run.googleapis.com/container/cpu/utilizations"},
				Aligner:   alertpolicy.MonitoringAlignPercentile99,
				Reducer:   alertpolicy.MonitoringReduceMax,
				Period:    "300s",
				Threshold: 0.8,
			},
		},
		{
			ID:          "memory",
			Name:        "High Container Memory Utilization",
			Description: pointers.Ptr("High Memory Usage - it may be neccessary to reduce load or increase memory allocation"),
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
			Description: pointers.Ptr("Instance is taking a long time to start up - something may be blocking startup"),
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "run.googleapis.com/container/startup_latencies"},
				Aligner:   alertpolicy.MonitoringAlignPercentile99,
				Reducer:   alertpolicy.MonitoringReduceMax,
				Period:    "60s",
				Threshold: 10000,
			},
		},
	} {
		config.ProjectID = vars.ProjectID
		config.ServiceName = vars.Service.ID
		config.ServiceKind = serviceKind
		config.NotificationChannels = channels
		if _, err := alertpolicy.New(stack, id, &config); err != nil {
			return err
		}
	}

	return nil
}

func createServiceAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	// Only provision if MaxCount is specified above 5
	if pointers.Deref(vars.MaxInstanceCount, 0) > 5 {
		if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			ID:          "instance_count",
			Name:        "Container Instance Count",
			Description: pointers.Ptr("There are a lot of Cloud Run instances running - we may need to increase per-instance requests make make sure we won't hit the configured max instance count"),
			ProjectID:   vars.ProjectID,
			ServiceName: vars.Service.ID,
			ServiceKind: alertpolicy.CloudRunService,
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters: map[string]string{"metric.type": "run.googleapis.com/container/instance_count"},
				Aligner: alertpolicy.MonitoringAlignMax,
				Reducer: alertpolicy.MonitoringReduceMax,
				Period:  "60s",
			},
			NotificationChannels: channels,
		}); err != nil {
			return err
		}
	}
	return nil
}

func createJobAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	// Alert whenever a Cloud Run Job fails
	if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
		ID:          "job_failures",
		Name:        "Cloud Run Job Failures",
		Description: pointers.Ptr("Failed executions of Cloud Run Job"),
		ProjectID:   vars.ProjectID,
		ServiceName: vars.Service.ID,
		ServiceKind: alertpolicy.CloudRunJob,
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

func createResponseCodeMetrics(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	for _, config := range vars.Monitoring.Alerts.ResponseCodeRatios {
		if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			ID:          config.ID,
			ProjectID:   vars.ProjectID,
			Name:        config.Name,
			ServiceName: vars.Service.ID,
			ServiceKind: alertpolicy.CloudRunService,
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

func createRedisAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	for _, config := range []alertpolicy.Config{
		{
			ID:          "memory",
			Name:        "Cloud Redis - System Memory Utilization",
			Description: pointers.Ptr("This alert fires if the system memory utilization is above the set threshold. The utilization is measured on a scale of 0 to 1."),
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
			Description: pointers.Ptr("This alert fires if the Redis Engine CPU Utilization goes above the set threshold. The utilization is measured on a scale of 0 to 1."),
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
			Description: pointers.Ptr("This alert fires if failover occurs for a standard tier instance."),
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "redis.googleapis.com/replication/role"},
				Aligner:   alertpolicy.MonitoringAlignStddev,
				Period:    "300s",
				Threshold: 0,
			},
		},
	} {
		config.ProjectID = vars.ProjectID
		config.ServiceName = *vars.RedisInstanceID
		config.ServiceKind = alertpolicy.CloudRedis
		config.NotificationChannels = channels
		if _, err := alertpolicy.New(stack, id, &config); err != nil {
			return err
		}
	}

	return nil
}
