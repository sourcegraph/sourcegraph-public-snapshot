package monitoring

import (
	"fmt"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringuptimecheckconfig"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/alertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/logcountmetric"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func createServiceAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) ([]monitoringalertpolicy.MonitoringAlertPolicy, error) {
	// Collect all alerts to aggregate in a dashboard
	var alerts []monitoringalertpolicy.MonitoringAlertPolicy

	// Only provision if MaxCount is specified greater or equal 5 (the default).
	// If nil, it doesn't matter
	if vars.MaxInstanceCount != nil && *vars.MaxInstanceCount >= 5 {
		instanceCountAlert, err := alertpolicy.New(stack, id, &alertpolicy.Config{
			Service:       vars.Service,
			EnvironmentID: vars.EnvironmentID,

			ID:          "instance_count",
			Name:        "Container Instance Count",
			Description: "There are a lot of Cloud Run instances running - we may need to increase per-instance requests make make sure we won't hit the configured max instance count",
			ProjectID:   vars.ProjectID,
			ThresholdAggregation: &alertpolicy.ThresholdAggregation{
				ConditionBuilder: alertpolicy.ConditionBuilder{
					ResourceName: vars.Service.ID,
					ResourceKind: alertpolicy.CloudRunService,

					Filters: map[string]string{"metric.type": "run.googleapis.com/container/instance_count"},
					Aligner: alertpolicy.MonitoringAlignMax,
					Reducer: alertpolicy.MonitoringReduceMax,
					Period:  "60s",
				},
				// Fire when we are 1 instance away from hitting the limit.
				Threshold:  float64(*vars.MaxInstanceCount - 1),
				Comparison: alertpolicy.ComparisonGT,
			},
			NotificationChannels: channels,
		})
		if err != nil {
			return nil, errors.Wrap(err, "instance_count")
		}
		alerts = append(alerts, instanceCountAlert.AlertPolicy)
	}

	pendingRequestsAlert, err := alertpolicy.New(stack, id, &alertpolicy.Config{
		Service:       vars.Service,
		EnvironmentID: vars.EnvironmentID,

		ID:          "cloud_run_pending_requests",
		Name:        "Cloud Run Pending Requests",
		Description: "There are requests pending - we may need to increase  Cloud Run instance count, request concurrency, or investigate further.",
		ProjectID:   vars.ProjectID,
		ThresholdAggregation: &alertpolicy.ThresholdAggregation{
			ConditionBuilder: alertpolicy.ConditionBuilder{
				ResourceName: vars.Service.ID,
				ResourceKind: alertpolicy.CloudRunService,

				Filters: map[string]string{"metric.type": "run.googleapis.com/pending_queue/pending_requests"},
				Aligner: alertpolicy.MonitoringAlignSum,
				Reducer: alertpolicy.MonitoringReduceSum,
				Period:  "60s",
			},
			Threshold:  5,
			Comparison: alertpolicy.ComparisonGT,
		},
		NotificationChannels: channels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "cloud_run_pending_requests")
	}
	alerts = append(alerts, pendingRequestsAlert.AlertPolicy)

	// If an external DNS name is provisioned, use it to check service availability
	// from outside Cloud Run. The service must not use IAM auth.
	if vars.ServiceAuthentication == nil && vars.ExternalDomain.GetDNSName() != "" {
		externalHealthCheckAlert, err := createExternalHealthcheckAlert(stack, id, vars, channels)
		if err != nil {
			return nil, errors.Wrap(err, "external_healthcheck")
		}
		alerts = append(alerts, externalHealthCheckAlert)
	}

	cloudRunPreconditionFailedAlert, err := createCloudRunPreconditionFailedAlert(stack, id, vars, channels)
	if err != nil {
		return nil, errors.Wrap(err, "CloudRunPreconditionFailedAlert")
	}
	alerts = append(alerts, cloudRunPreconditionFailedAlert)

	return alerts, nil
}

func createExternalHealthcheckAlert(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) (monitoringalertpolicy.MonitoringAlertPolicy, error) {
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

	alert, err := alertpolicy.New(stack, id, &alertpolicy.Config{
		Service:       vars.Service,
		EnvironmentID: vars.EnvironmentID,
		ProjectID:     vars.ProjectID,

		ID:          "external_health_check",
		Name:        "External Uptime Check",
		Description: fmt.Sprintf("Service is failing to respond on https://%s - this may be expected if the service was recently provisioned or if its external domain has changed.", externalDNS),

		// If a service is not reachable, it's definitely a problem.
		Severity: spec.AlertSeverityLevelCritical,

		ThresholdAggregation: &alertpolicy.ThresholdAggregation{
			ConditionBuilder: alertpolicy.ConditionBuilder{
				ResourceKind: alertpolicy.URLUptime,
				ResourceName: *uptimeCheck.UptimeCheckId(),

				Filters: map[string]string{
					"metric.type": "monitoring.googleapis.com/uptime_check/check_passed",
				},
				Aligner: alertpolicy.MonitoringAlignFractionTrue,
				// Checks run once every 60s, if 2/3 fail we are in trouble.
				Period: "180s",
				// We want to alert when all locations go down, but right now that
				// sends 6 notifications when the alert fires, which is annoying -
				// there seems to be no way to change this. So we group by the check
				// target anyway.
				Trigger:       alertpolicy.TriggerKindAllInViolation,
				GroupByFields: []string{"metric.labels.host"},
				Reducer:       alertpolicy.MonitoringReduceMean,
			},
			Threshold:  0.4,
			Duration:   "0s",
			Comparison: alertpolicy.ComparisonLT,
		},
		NotificationChannels: channels,
	})
	if err != nil {
		return nil, err
	}
	return alert.AlertPolicy, nil
}

func createCloudRunPreconditionFailedAlert(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channels alertpolicy.NotificationChannels,
) (monitoringalertpolicy.MonitoringAlertPolicy, error) {
	metric, err := logcountmetric.New(stack, id.Group("cloud_run_precondition_failed"), &logcountmetric.Config{
		Name: "msp.sourcegraph.com/cloud_run_precondition_failed",
		// Status 9 indicates 'precondition failed'
		// https://github.com/googleapis/googleapis/blob/e3802a1c97ee876e01247f9d22c15219ef4d9c19/google/rpc/code.proto#L110-L128
		LogFilters: fmt.Sprintf(`
			resource.type="cloud_run_revision"
			AND protoPayload.status.code="9"
			AND resource.labels.service_name =~ "^%s.*"
		`, vars.Service.ID),
		LabelExtractors: map[string]logcountmetric.LabelExtractor{
			"service_name": {
				Expression:  "EXTRACT(resource.labels.service_name)",
				Type:        "STRING",
				Description: "Name of the affected Cloud Run service",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	alert, err := alertpolicy.New(stack, id.Group("cloud_run_precondition_failed_alert"), &alertpolicy.Config{
		Service:       vars.Service,
		EnvironmentID: vars.EnvironmentID,
		ProjectID:     vars.ProjectID,

		ID:   "cloud_run_precondition_failed",
		Name: "Cloud Run Instance Precondition Failed",
		Description: `Cloud Run instance failed to start due to a precondition failure.
This is unlikely to cause immediate downtime, and may auto-resolve if no new instances are created and/or we return to a healthy state, but you should follow up to ensure the latest Cloud Run revision is healthy.`,
		ThresholdAggregation: &alertpolicy.ThresholdAggregation{
			ConditionBuilder: alertpolicy.ConditionBuilder{
				Filters: map[string]string{
					"metric.type": metric.Metric,
					// HACK: Strangely, this seems required on our log-based metric
					"resource.type": "cloud_run_revision",
				},
				Aligner: alertpolicy.MonitoringAlignMax,
				Reducer: alertpolicy.MonitoringReduceSum,
				Period:  "60s",
				Trigger: alertpolicy.TriggerKindAnyViolation,
			},
			Threshold:  0, // any occurence is bad
			Comparison: alertpolicy.ComparisonGT,
		},
		NotificationChannels: channels,
	})
	if err != nil {
		return nil, err
	}

	return alert.AlertPolicy, nil
}
