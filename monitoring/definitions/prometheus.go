package definitions

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Prometheus() *monitoring.Dashboard {
	const (
		containerName = "prometheus"

		// ruleGroupInterpretation provides interpretation documentation for observables that are per prometheus rule_group.
		ruleGroupInterpretation = `Rules that Sourcegraph ships with are grouped under '/sg_config_prometheus'. [Custom rules are grouped under '/sg_prometheus_addons'](https://docs.sourcegraph.com/admin/observability/metrics#prometheus-configuration).`
	)

	return &monitoring.Dashboard{
		Name:                     "prometheus",
		Title:                    "Prometheus",
		Description:              "Sourcegraph's all-in-one Prometheus and Alertmanager service.",
		NoSourcegraphDebugServer: true, // This is third-party service
		Groups: []monitoring.Group{
			{
				Title: "Metrics",
				Rows: []monitoring.Row{
					{
						{
							Name:           "metrics_cardinality",
							Description:    "metrics with highest cardinalities",
							Query:          `topk(10, count by (__name__, job)({__name__!=""}))`,
							Panel:          monitoring.Panel().LegendFormat("{{__name__}} ({{job}})"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							NoAlert:        true,
							Interpretation: "The 10 highest-cardinality metrics collected by this Prometheus instance.",
						},
						{
							Name:           "samples_scraped",
							Description:    "samples scraped by job",
							Query:          `sum by(job) (scrape_samples_post_metric_relabeling{job!=""})`,
							Panel:          monitoring.Panel().LegendFormat("{{job}}"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							NoAlert:        true,
							Interpretation: "The number of samples scraped after metric relabeling was applied by this Prometheus instance.",
						},
					},
					{
						{
							Name:        "prometheus_rule_eval_duration",
							Description: "average prometheus rule group evaluation duration over 10m by rule group",
							Query:       `sum by(rule_group) (avg_over_time(prometheus_rule_group_last_duration_seconds[10m]))`,
							Warning:     monitoring.Alert().GreaterOrEqual(30), // standard prometheus_rule_group_interval_seconds
							Panel:       monitoring.Panel().Unit(monitoring.Seconds).MinAuto().LegendFormat("{{rule_group}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							Interpretation: fmt.Sprintf(`
								A high value here indicates Prometheus rule evaluation is taking longer than expected.
								It might indicate that certain rule groups are taking too long to evaluate, or Prometheus is underprovisioned.

								%s
							`, ruleGroupInterpretation),
							NextSteps: fmt.Sprintf(`
								- Check the %s panels and try increasing resources for Prometheus if necessary.
								- If the rule group taking a long time to evaluate belongs to '/sg_prometheus_addons', try reducing the complexity of any custom Prometheus rules provided.
								- If the rule group taking a long time to evaluate belongs to '/sg_config_prometheus', please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title=).
							`, shared.TitleContainerMonitoring),
						},
						{
							Name:           "prometheus_rule_eval_failures",
							Description:    "failed prometheus rule evaluations over 5m by rule group",
							Query:          `sum by(rule_group) (rate(prometheus_rule_evaluation_failures_total[5m]))`,
							Warning:        monitoring.Alert().Greater(0),
							Panel:          monitoring.Panel().LegendFormat("{{rule_group}}"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Interpretation: ruleGroupInterpretation,
							NextSteps: `
								- Check Prometheus logs for messages related to rule group evaluation (generally with log field 'component="rule manager"').
								- If the rule group failing to evaluate belongs to '/sg_prometheus_addons', ensure any custom Prometheus configuration provided is valid.
								- If the rule group taking a long time to evaluate belongs to '/sg_config_prometheus', please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title=).
							`,
						},
					},
				},
			},
			{
				Title: "Alerts",
				Rows: []monitoring.Row{
					{
						{
							Name:        "alertmanager_notification_latency",
							Description: "alertmanager notification latency over 1m by integration",
							Query:       `sum by(integration) (rate(alertmanager_notification_latency_seconds_sum[1m]))`,
							Warning:     monitoring.Alert().GreaterOrEqual(1),
							Panel:       monitoring.Panel().Unit(monitoring.Seconds).LegendFormat("{{integration}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							NextSteps: fmt.Sprintf(`
								- Check the %s panels and try increasing resources for Prometheus if necessary.
								- Ensure that your ['observability.alerts' configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.
								- Check if the relevant alert integration service is experiencing downtime or issues.
							`, shared.TitleContainerMonitoring),
						},
						{
							Name:        "alertmanager_notification_failures",
							Description: "failed alertmanager notifications over 1m by integration",
							Query:       `sum by(integration) (rate(alertmanager_notifications_failed_total[1m]))`,
							Warning:     monitoring.Alert().Greater(0),
							Panel:       monitoring.Panel().LegendFormat("{{integration}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							NextSteps: `
								- Ensure that your ['observability.alerts' configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.
								- Check if the relevant alert integration service is experiencing downtime or issues.
							`,
						},
					},
				},
			},
			{
				Title:  "Internals",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:           "prometheus_config_status",
							Description:    "prometheus configuration reload status",
							Query:          `prometheus_config_last_reload_successful`,
							Warning:        monitoring.Alert().Less(1),
							Panel:          monitoring.Panel().LegendFormat("reload success").Max(1),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Interpretation: "A '1' indicates Prometheus reloaded its configuration successfully.",
							NextSteps: `
								- Check Prometheus logs for messages related to configuration loading.
								- Ensure any [custom configuration you have provided Prometheus](https://docs.sourcegraph.com/admin/observability/metrics#prometheus-configuration) is valid.
							`,
						},
						{
							Name:           "alertmanager_config_status",
							Description:    "alertmanager configuration reload status",
							Query:          `alertmanager_config_last_reload_successful`,
							Warning:        monitoring.Alert().Less(1),
							Panel:          monitoring.Panel().LegendFormat("reload success").Max(1),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Interpretation: "A '1' indicates Alertmanager reloaded its configuration successfully.",
							NextSteps:      "Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.",
						},
					},
					{
						{
							Name:        "prometheus_tsdb_op_failure",
							Description: "prometheus tsdb failures by operation over 1m by operation",
							Query:       `increase(label_replace({__name__=~"prometheus_tsdb_(.*)_failed_total"}, "operation", "$1", "__name__", "(.+)s_failed_total")[5m:1m])`,
							Warning:     monitoring.Alert().Greater(0),
							Panel:       monitoring.Panel().LegendFormat("{{operation}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							NextSteps:   "Check Prometheus logs for messages related to the failing operation.",
						},
						{
							Name:        "prometheus_target_sample_exceeded",
							Description: "prometheus scrapes that exceed the sample limit over 10m",
							Query:       "increase(prometheus_target_scrapes_exceeded_sample_limit_total[10m])",
							Warning:     monitoring.Alert().Greater(0),
							Panel:       monitoring.Panel().LegendFormat("rejected scrapes"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							NextSteps:   "Check Prometheus logs for messages related to target scrape failures.",
						},
						{
							Name:        "prometheus_target_sample_duplicate",
							Description: "prometheus scrapes rejected due to duplicate timestamps over 10m",
							Query:       "increase(prometheus_target_scrapes_sample_duplicate_timestamp_total[10m])",
							Warning:     monitoring.Alert().Greater(0),
							Panel:       monitoring.Panel().LegendFormat("rejected scrapes"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							NextSteps:   "Check Prometheus logs for messages related to target scrape failures.",
						},
					},
				},
			},

			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerInfraOrg, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerInfraOrg, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerInfraOrg, nil),
		},
	}
}
