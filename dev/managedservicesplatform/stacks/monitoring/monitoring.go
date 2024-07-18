package monitoring

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringnotificationchannel"
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
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/nobl9provider"
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
	// Alerting configuration for the environment.
	Alerting spec.EnvironmentAlertingSpec
	// Monitoring spec.
	Monitoring spec.MonitoringSpec
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
	// If CloudSQL is enabled we configure alerts for it
	CloudSQLInstanceID    *string
	CloudSQLMaxConections *int
	// ServiceHealthProbes is used to determine the threshold for service
	// startup latency alerts.
	ServiceHealthProbes *spec.EnvironmentServiceHealthProbesSpec
	// SentryProject is the project in Sentry for the service environment
	SentryProject sentryproject.Project
	// JobSchedule is used to determine if an alert on job absence should be
	// provisioned and the appropriate thresholds.
	JobSchedule *spec.EnvironmentJobScheduleSpec
}

const StackName = "monitoring"

const sharedAlertsSlackChannel = "#alerts-msp"

// mspRolloutsBotSlackUserID  is the user ID of MSP Rollouts Slack bot.
const mspRolloutsBotSlackUserID string = "U072KKSCLSJ"

// nobl9ClientID user account (@jac) for trial
const nobl9ClientID = "0oab428uphKZbY1jy417"
const nobl9OrganizationID = "sourcegraph-n8JWJzlFjsCw"

func NewStack(stacks *stack.Set, vars Variables) (*CrossStackOutput, error) {
	stack, locals, err := stacks.New(StackName,
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
		}),
		nobl9provider.With(nobl9provider.Config{
			ClientID:     nobl9ClientID,
			Organization: nobl9OrganizationID,
			Nobl9Token: gsmsecret.DataConfig{
				Secret:    googlesecretsmanager.SecretNobl9ClientSecret,
				ProjectID: googlesecretsmanager.SharedSecretsProjectID,
			},
		}))
	if err != nil {
		return nil, err
	}

	id := resourceid.New("monitoring")

	// Prepare GCP monitoring channels on which to notify on when an alert goes
	// off.
	channels := make(map[spec.AlertSeverityLevel][]monitoringnotificationchannel.MonitoringNotificationChannel)
	addChannel := func(level spec.AlertSeverityLevel, c monitoringnotificationchannel.MonitoringNotificationChannel) {
		channels[level] = append(channels[level], c)
	}

	// Configure opsgenie channels
	var opsgenieChannels []monitoringnotificationchannel.MonitoringNotificationChannel
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

				// Supress all notifications if Alerting.Opsgenie is explcitly set to false or
				// if Alerting.Opsgenie is nil suppress notifications if this is not a production environment.
				// This allows us to see the alerts, but not necessarily get paged by it.
				SuppressNotifications: !vars.Alerting.ShouldEnableOpsgenie(vars.EnvironmentCategory.IsProduction()),

				// Point alerts sent through this integration at the Opsgenie team.
				Responders: []*opsgenieintegration.ApiIntegrationResponders{{
					// Possible values for Type are:
					// team, user, escalation and schedule
					Type: pointers.Ptr("team"),
					Id:   team.Id(),
				}},
			})

		channel := monitoringnotificationchannel.NewMonitoringNotificationChannel(stack,
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
			})
		addChannel(spec.AlertSeverityLevelCritical, channel)
		opsgenieChannels = append(opsgenieChannels, channel)
	}

	// Configure Slack channels
	var slackChannels []monitoringnotificationchannel.MonitoringNotificationChannel
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
			description := fmt.Sprintf("Alerts from %s (%s) deployed on Managed Services Platform.",
				vars.Service.GetName(), vars.EnvironmentID)
			if vars.Service.NotionPageID != nil {
				description += fmt.Sprintf(" Operational handbook: %s", vars.Service.GetHandbookPageURL())
			}
			// https://registry.terraform.io/providers/pablovarela/slack/latest/docs/resources/conversation#argument-reference
			slackChannel = slackconversation.NewConversation(stack, id.TerraformID("channel"), &slackconversation.ConversationConfig{
				Name:             pointers.Ptr(strings.TrimPrefix(channel.Name, "#")),
				Topic:            &description,
				Purpose:          &description,
				IsPrivate:        pointers.Ptr(false),
				PermanentMembers: pointers.Ptr(pointers.Slice([]string{mspRolloutsBotSlackUserID})),
				// Do not kick out other users in the channel
				ActionOnUpdatePermanentMembers: pointers.Ptr("none"),

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

		notificationChannel := monitoringnotificationchannel.NewMonitoringNotificationChannel(stack,
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
			})

		addChannel(spec.AlertSeverityLevelWarning, notificationChannel)
		addChannel(spec.AlertSeverityLevelCritical, notificationChannel)
		slackChannels = append(slackChannels, notificationChannel)
	}

	// Now that we're done with configure channels, make the generalized
	// level-based channels available as vars, so that custom alerts can mimic
	// the behaviour of our built-ins.
	levels := maps.Keys(channels)
	slices.Sort(levels)
	for _, l := range levels {
		levelString := strings.ToLower(string(l))
		locals.AddSlice(
			fmt.Sprintf("notifications_%s", levelString),
			notificationChannelIDs(channels[l]),
			fmt.Sprintf("Google Monitoring notification channels for %s-level alerts", levelString),
		)
	}
	// Also add per-owner notification channels, such that using one of these
	// routes alerts to an Opsgenie owner and the usual Slack channels.
	for i, opsgenieChannel := range opsgenieChannels {
		locals.AddSlice(
			// Use index because Opsgenie team names has lax character
			// requirements that are hard to adapt to valid Terraform IDs - this
			// is just the order of the spec owners list so should be fairly
			// easy to reason with.
			fmt.Sprintf("notifications_owner_%d", i),
			notificationChannelIDs(append(slackChannels, opsgenieChannel)),
			fmt.Sprintf("Google Monitoring notification channel IDs for %s", *opsgenieChannel.DisplayNameInput()),
		)
	}

	// Add some variables that might help with writing alerts.
	locals.Add("service_id", vars.Service.ID, "Service ID")
	locals.Add("environment_id", vars.EnvironmentID, "Environment ID")
	locals.Add("service_name", vars.Service.GetName(), "Human-readable service name")
	locals.Add("alert_description_suffix", alertpolicy.DescriptionSuffix(vars.Service, vars.EnvironmentID),
		"Supplemental MSP help text intended to be added to alert descriptions")

	// Group alerts by type for dashboard
	alertGroups := make(map[string][]monitoringalertpolicy.MonitoringAlertPolicy)

	// Set up alerts, configuring each with all our notification channels
	commonAlerts, err := createCommonAlerts(stack, id.Group("common"), vars, channels)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create common alerts")
	}

	alertGroups["Container Alerts"] = commonAlerts

	switch vars.Service.GetKind() {
	case spec.ServiceKindService:
		serviceAlerts, err := createServiceAlerts(stack, id.Group("service"), vars, channels)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create service alerts")
		}
		alertGroups["Cloud Run Service Alerts"] = serviceAlerts

		if vars.Monitoring.Alerts.ResponseCodeRatios != nil {
			responseCodeRatioAlerts, err := createResponseCodeAlerts(stack, id.Group("response-code"), vars, channels)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create response code metrics")
			}
			alertGroups[responseCodeRatioAlertsGroupName] = responseCodeRatioAlerts
		}
	case spec.ServiceKindJob:
		jobAlerts, err := createJobAlerts(stack, id.Group("job"), vars, channels)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create job alerts")
		}
		alertGroups["Cloud Run Job Alerts"] = jobAlerts
	default:
		return nil, errors.New("unknown service kind")
	}

	if vars.Monitoring.Alerts.CustomAlerts != nil {
		customAlerts, err := createCustomAlerts(stack, id.Group("custom"), vars, channels)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create custom alerts")
		}
		alertGroups[customAlertsGroupName] = customAlerts
	}

	if vars.RedisInstanceID != nil {
		redisAlerts, err := createRedisAlerts(stack, id.Group("redis"), vars, channels)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create redis alerts")
		}
		alertGroups["Redis Alerts"] = redisAlerts
	}

	if vars.CloudSQLInstanceID != nil {
		cloudSQLAlerts, err := createCloudSQLAlerts(stack, id.Group("cloudsql"), vars, channels)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create CloudSQL alerts")
		}
		alertGroups["Cloud SQL Alerts"] = cloudSQLAlerts
	}

	if pointers.DerefZero(vars.Monitoring.Nobl9) {
		createNobl9Project(stack, id.Group("nobl9"), vars)
	}

	// Create a dashboard containing all MSP alerts
	err = createMonitoringDashboard(stack, id.Group("dashboard"), vars, alertGroups)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dashboard")
	}

	return &CrossStackOutput{}, nil
}

func notificationChannelIDs(channels []monitoringnotificationchannel.MonitoringNotificationChannel) []string {
	if len(channels) == 0 {
		return nil
	}
	var ids []string
	for _, c := range channels {
		ids = append(ids, *c.Id())
	}
	return ids
}
