package monitoring

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringnotificationchannel"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringuptimecheckconfig"
	opsgenieintegration "github.com/sourcegraph/managed-services-platform-cdktf/gen/opsgenie/apiintegration"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/opsgenie/dataopsgenieteam"
	opsgenieservice "github.com/sourcegraph/managed-services-platform-cdktf/gen/opsgenie/service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/sentry/datasentryorganizationintegration"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/sentry/notificationaction"
	sentryproject "github.com/sourcegraph/managed-services-platform-cdktf/gen/sentry/project"
	slackconversation "github.com/sourcegraph/managed-services-platform-cdktf/gen/slack/conversation"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/random"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/sentryalert"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/opsgenieprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/sentryprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/slackprovider"
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
	ProjectID string
	Service   spec.ServiceSpec
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
	Monitoring    spec.MonitoringSpec
	// MaxInstanceCount informs service scaling alerts.
	MaxInstanceCount *int
	// ExternalDomain informs external health checks on the service domain.
	ExternalDomain *spec.EnvironmentServiceDomainSpec
	// ServiceAuthentication informs external health checks on the service
	// domain. Currently, any configuration will disable external health checks.
	ServiceAuthentication *spec.EnvironmentServiceAuthenticationSpec
	// DiagnosticsSecret is used to configure external health checks.
	DiagnosticsSecret *random.Output
	// If Redis is enabled we configure alerts for it
	RedisInstanceID *string
	// ServiceHealthProbes is used to determine the threshold for service
	// startup latency alerts.
	ServiceHealthProbes *spec.EnvironmentServiceHealthProbesSpec
	// SentryProject is the project in Sentry for the service environment
	SentryProject sentryproject.Project
}

const StackName = "monitoring"

const sharedAlertsSlackChannel = "#alerts-msp"

func NewStack(stacks *stack.Set, vars Variables) (*CrossStackOutput, error) {
	stack, _, err := stacks.New(StackName,
		googleprovider.With(vars.ProjectID),
		opsgenieprovider.With(gsmsecret.DataConfig{
			Secret:    googlesecretsmanager.SecretOpsgenieAPIToken,
			ProjectID: googlesecretsmanager.SharedSecretsProjectID,
		}),
		slackprovider.With(gsmsecret.DataConfig{
			Secret:    googlesecretsmanager.SecretSlackOperatorOAuthToken,
			ProjectID: googlesecretsmanager.SharedSecretsProjectID,
		}),
		sentryprovider.With(gsmsecret.DataConfig{
			Secret:    googlesecretsmanager.SecretSentryAuthToken,
			ProjectID: googlesecretsmanager.SharedSecretsProjectID,
		}))
	if err != nil {
		return nil, err
	}

	id := resourceid.New("monitoring")

	// Prepare GCP monitoring channels on which to notify on when an alert goes
	// off.
	var channels []monitoringnotificationchannel.MonitoringNotificationChannel

	// Configure opsgenie channels
	// TODO: Enable after we dogfood the alerts for a while.
	// var opsgenieAlerts bool
	// switch vars.EnvironmentCategory {
	// case spec.EnvironmentCategoryInternal, spec.EnvironmentCategoryExternal:
	// 	opsgenieAlerts = true
	// }
	for i, owner := range vars.Service.Owners {
		// Use index because Opsgenie team names has lax character requirements
		id := id.Group("opsgenie_owner_%d", i)
		// Opsgenie team corresponding to owner must exist
		team := dataopsgenieteam.NewDataOpsgenieTeam(stack,
			id.TerraformID("opsgenie_team"),
			&dataopsgenieteam.DataOpsgenieTeamConfig{
				Name: &owner,
			})
		// Create a "Opsgenie service" representing this service. We can't
		// attach alerts to this so opsgenie-wise it's not very useful, but
		// it syncs to Incident.io, which could be useful. Either way, it seems
		// harmless to add, so let's add it and see what comes of it.
		_ = opsgenieservice.NewService(stack,
			id.TerraformID("opsgenie_service"),
			&opsgenieservice.ServiceConfig{
				Name:        pointers.Stringf("%s - %s", vars.Service.GetName(), vars.EnvironmentID),
				TeamId:      team.Id(),
				Description: &vars.Service.Description,
				Tags: &[]*string{
					pointers.Ptr("msp"),
					pointers.Ptr(vars.Service.ID),
					pointers.Ptr(string(vars.EnvironmentCategory)),
				},
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
				// TODO: Enable after we dogfood the alerts for a while.
				SuppressNotifications: pointers.Ptr(true),

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
		ProjectID: googlesecretsmanager.SharedSecretsProjectID,
	})
	for _, channel := range []struct {
		Name             string
		ProvisionChannel bool
	}{
		{
			Name: sharedAlertsSlackChannel,
			// Do not try to provision preexisting shared channel
			ProvisionChannel: false,
		},
		{
			// service-env-specific channel
			Name: fmt.Sprintf("#alerts-%s-%s",
				vars.Service.ID, vars.EnvironmentID),
			ProvisionChannel: true,
		},
	} {
		id := id.Group("slack_%s", strings.TrimPrefix(channel.Name, "#"))

		var slackChannel slackconversation.Conversation
		if channel.ProvisionChannel {
			description := pointers.Stringf(
				"Alerts from %s (%s) deployed on Managed Services Platform",
				vars.Service.GetName(), vars.EnvironmentID)
			// https://registry.terraform.io/providers/pablovarela/slack/latest/docs/resources/conversation#argument-reference
			slackChannel = slackconversation.NewConversation(stack, id.TerraformID("channel"), &slackconversation.ConversationConfig{
				Name:      pointers.Ptr(strings.TrimPrefix(channel.Name, "#")),
				Topic:     description,
				Purpose:   description,
				IsPrivate: pointers.Ptr(false),

				// In case it already exists
				AdoptExistingChannel: pointers.Ptr(true),
			})

			// Sentry Slack integration
			dataSentryOrganizationIntegration := datasentryorganizationintegration.NewDataSentryOrganizationIntegration(stack, id.TerraformID("sentry_integration"), &datasentryorganizationintegration.DataSentryOrganizationIntegrationConfig{
				Organization: vars.SentryProject.Organization(),
				ProviderKey:  pointers.Ptr("slack"),
				Name:         pointers.Ptr("Sourcegraph"),
			})

			// Provision Sentry Slack notification
			_ = notificationaction.NewNotificationAction(stack, id.TerraformID("sentry_notification_channel"), &notificationaction.NotificationActionConfig{
				Organization:     vars.SentryProject.Organization(),
				Projects:         &[]*string{vars.SentryProject.Slug()},
				ServiceType:      pointers.Ptr("slack"),
				IntegrationId:    dataSentryOrganizationIntegration.Id(),
				TargetDisplay:    slackChannel.Name(),
				TargetIdentifier: slackChannel.Id(),
				TriggerType:      pointers.Ptr("spike-protection"),
			})

			if err = createSentryAlerts(stack, id.Group("sentry_alerts"), vars, slackChannel, dataSentryOrganizationIntegration); err != nil {
				return nil, errors.Wrap(err, "failed to create Sentry alerts")
			}
		}

		channels = append(channels,
			monitoringnotificationchannel.NewMonitoringNotificationChannel(stack,
				id.TerraformID("notification_channel"),
				&monitoringnotificationchannel.MonitoringNotificationChannelConfig{
					Project:     &vars.ProjectID,
					DisplayName: pointers.Stringf("Slack - %s", channel.Name),
					Type:        pointers.Ptr("slack"),
					Labels: &map[string]*string{
						"channel_name": &channel.Name,
					},
					SensitiveLabels: &monitoringnotificationchannel.MonitoringNotificationChannelSensitiveLabels{
						AuthToken: &slackToken.Value,
					},
					DependsOn: func() *[]cdktf.ITerraformDependable {
						if slackChannel != nil {
							return pointers.Ptr([]cdktf.ITerraformDependable{slackChannel})
						}
						return nil
					}(),
				}))
	}

	// Set up alerts, configuring each with all our notification channels
	err = createCommonAlerts(stack, id.Group("common"), vars, channels)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create common alerts")
	}

	switch vars.Service.GetKind() {
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
				Period:    "300s",
				Threshold: 0.8,
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

func createServiceAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	// Only provision if MaxCount is specified above 5
	if pointers.Deref(vars.MaxInstanceCount, 0) > 5 {
		if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			Service:       vars.Service,
			EnvironmentID: vars.EnvironmentID,

			ID:           "instance_count",
			Name:         "Container Instance Count",
			Description:  "There are a lot of Cloud Run instances running - we may need to increase per-instance requests make make sure we won't hit the configured max instance count",
			ProjectID:    vars.ProjectID,
			ResourceName: vars.Service.ID,
			ResourceKind: alertpolicy.CloudRunService,
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				Filters: map[string]string{"metric.type": "run.googleapis.com/container/instance_count"},
				Aligner: alertpolicy.MonitoringAlignMax,
				Reducer: alertpolicy.MonitoringReduceMax,
				Period:  "60s",
			},
			NotificationChannels: channels,
		}); err != nil {
			return errors.Wrap(err, "instance_count")
		}
	}

	// If an external DNS name is provisioned, use it to check service availability
	// from outside Cloud Run. The service must not use IAM auth.
	if vars.ServiceAuthentication == nil && vars.ExternalDomain.GetDNSName() != "" {
		if err := createExternalHealthcheckAlert(stack, id, vars, channels); err != nil {
			return errors.Wrap(err, "external_healthcheck")
		}
	}
	return nil
}

func createExternalHealthcheckAlert(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	var (
		healthcheckPath    = "/"
		healthcheckHeaders = map[string]*string{}
	)
	// Only use MSP runtime standards if we know the service supports it.
	if vars.ServiceHealthProbes.UseHealthzProbes() {
		healthcheckPath = "/-/healthz"
		healthcheckHeaders = map[string]*string{
			"Authorization": pointers.Stringf("Bearer %s", vars.DiagnosticsSecret.HexValue),
		}
	}

	externalDNS := vars.ExternalDomain.GetDNSName()
	uptimeCheck := monitoringuptimecheckconfig.NewMonitoringUptimeCheckConfig(stack, id.TerraformID("external_uptime_check"), &monitoringuptimecheckconfig.MonitoringUptimeCheckConfigConfig{
		Project:     &vars.ProjectID,
		DisplayName: pointers.Stringf("External Uptime Check for %s", externalDNS),

		// https://cloud.google.com/monitoring/api/resources#tag_uptime_url
		MonitoredResource: &monitoringuptimecheckconfig.MonitoringUptimeCheckConfigMonitoredResource{
			Type: pointers.Ptr("uptime_url"),
			Labels: &map[string]*string{
				"project_id": &vars.ProjectID,
				"host":       &externalDNS,
			},
		},

		// 1 to 60 seconds.
		Timeout: pointers.Stringf("%ds", vars.ServiceHealthProbes.GetTimeoutSeconds()),
		// Only supported values are 60s (1 minute), 300s (5 minutes),
		// 600s (10 minutes), and 900s (15 minutes)
		Period: pointers.Ptr("60s"),
		HttpCheck: &monitoringuptimecheckconfig.MonitoringUptimeCheckConfigHttpCheck{
			Port:        pointers.Float64(443),
			UseSsl:      pointers.Ptr(true),
			ValidateSsl: pointers.Ptr(true),
			Path:        &healthcheckPath,
			Headers:     &healthcheckHeaders,
			AcceptedResponseStatusCodes: &[]*monitoringuptimecheckconfig.MonitoringUptimeCheckConfigHttpCheckAcceptedResponseStatusCodes{
				{
					StatusClass: pointers.Ptr("STATUS_CLASS_2XX"),
				},
			},
		},
	})

	if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
		Service:       vars.Service,
		EnvironmentID: vars.EnvironmentID,

		ID:          "external_health_check",
		Name:        "External Uptime Check",
		Description: fmt.Sprintf("Service is failing to repond on https://%s - this may be expected if the service was recently provisioned or if its external domain has changed.", externalDNS),
		ProjectID:   vars.ProjectID,

		ResourceKind: alertpolicy.URLUptime,
		ResourceName: *uptimeCheck.UptimeCheckId(),

		ThresholdAggregation: &alertpolicy.ThresholdAggregation{
			Filters: map[string]string{
				"metric.type": "monitoring.googleapis.com/uptime_check/check_passed",
			},
			Aligner: alertpolicy.MonitoringAlignFractionTrue,
			// Checks occur every 60s, in a 300s window if 2/5 fail we are in trouble
			Period:     "300s",
			Duration:   "0s",
			Comparison: alertpolicy.ComparisonLT,
			Threshold:  0.4,
			// Alert when all locations go down
			Trigger: alertpolicy.TriggerKindAllInViolation,
		},
		NotificationChannels: channels,
	}); err != nil {
		return err
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

func createResponseCodeMetrics(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels []monitoringnotificationchannel.MonitoringNotificationChannel,
) error {
	for _, config := range vars.Monitoring.Alerts.ResponseCodeRatios {
		if _, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			Service:       vars.Service,
			EnvironmentID: vars.EnvironmentID,

			ID:           config.ID,
			ProjectID:    vars.ProjectID,
			Name:         config.Name,
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

func createSentryAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channel slackconversation.Conversation,
	slackIntegration datasentryorganizationintegration.DataSentryOrganizationIntegration,
) error {
	for _, config := range []sentryalert.Config{
		{
			ID:            "all-issues",
			SentryProject: vars.SentryProject,
			AlertConfig: sentryalert.AlertConfig{
				Name:      "Notify in Slack",
				Frequency: 15, // Notify for an issue at most once every 15 minutes
				Conditions: []sentryalert.Condition{
					{
						ID:       sentryalert.EventFrequencyCondition,
						Value:    pointers.Ptr(0), // Always (seen more than 0 times) during interval
						Interval: pointers.Ptr("15m"),
					},
				},
				ActionMatch: sentryalert.ActionMatchAny,
				Actions: []sentryalert.Action{
					{
						ID: sentryalert.SlackNotifyServiceAction,
						ActionParameters: map[string]any{
							"workspace":  slackIntegration.Id(),
							"channel":    channel.Name(),
							"channel_id": channel.Id(),
							"tags": pointers.Stringf("msp-%s-%s",
								vars.Service.ID, vars.EnvironmentID),
						},
					},
				},
			},
		},
	} {
		if _, err := sentryalert.New(stack, id.Group(config.ID), config); err != nil {
			return err
		}
	}
	return nil
}
