package monitoring

import (
	"fmt"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringnotificationchannel"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringuptimecheckconfig"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

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
