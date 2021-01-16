package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Prometheus() *monitoring.Container {
	return &monitoring.Container{
		Name:        "prometheus",
		Title:       "Prometheus",
		Description: "Sourcegraph's all-in-one Prometheus and Alertmanager service.",
		Groups: []monitoring.Group{
			{
				Title: "Metrics",
				Rows: []monitoring.Row{
					{
						{
							Name:        "prometheus_rule_group_evaluation",
							Description: "average prometheus rule group evaluation duration over 10m",
							Query:       `sum by(rule_group) (avg_over_time(prometheus_rule_group_last_duration_seconds[10m]))`,
							Warning:     monitoring.Alert().GreaterOrEqual(30), // standard prometheus_rule_group_interval_seconds
							Panel:       monitoring.Panel().Unit(monitoring.Seconds).MinAuto().LegendFormat("{{rule_group}}"),
							Owner:       monitoring.ObservableOwnerDistribution,
							Interpretation: `
								A high value here indicates Prometheus rule evaluation is taking longer than expected.
								It might indicate that certain rule groups are taking too long to evaluate, or Prometheus is underprovisioned.
							`,
							PossibleSolutions: `
								- Try increasing resources for Prometheus.
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
							Name:              "alertmanager_notifications_failed_total",
							Description:       "failed alertmanager notifications over 1m",
							Query:             `sum by(integration) (rate(alertmanager_notifications_failed_total[1m]))`,
							Warning:           monitoring.Alert().Greater(0),
							Panel:             monitoring.Panel().LegendFormat("{{integration}}"),
							Owner:             monitoring.ObservableOwnerDistribution,
							PossibleSolutions: "Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.",
						},
						{
							Name:              "alertmanager_config_status",
							Description:       "alertmanager configuration reload status",
							Query:             `alertmanager_config_last_reload_successful`,
							Warning:           monitoring.Alert().Less(1),
							Panel:             monitoring.Panel().LegendFormat("reload success").Max(1),
							Owner:             monitoring.ObservableOwnerDistribution,
							Interpretation:    "A '1' indicates Alertmanager reloaded its configuration successfully.",
							PossibleSolutions: "Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.",
						},
					},
				},
			},
			{
				Title:  "Prometheus internals",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:              "prometheus_tsdb_op_failure",
							Description:       "prometheus tsdb failures by operation over 1m",
							Query:             `increase(label_replace({__name__=~"prometheus_tsdb_(.*)_failed_total"}, "operation", "$1", "__name__", "(.+)s_failed_total")[5m:1m])`,
							Warning:           monitoring.Alert().Greater(0),
							Panel:             monitoring.Panel().LegendFormat("{{operation}}"),
							Owner:             monitoring.ObservableOwnerDistribution,
							PossibleSolutions: "Check Prometheus logs for messages related to the failing operation.",
						},
						{
							Name:           "prometheus_config_status",
							Description:    "prometheus configuration reload status",
							Query:          `prometheus_config_last_reload_successful`,
							Warning:        monitoring.Alert().Less(1),
							Panel:          monitoring.Panel().LegendFormat("reload success").Max(1),
							Owner:          monitoring.ObservableOwnerDistribution,
							Interpretation: "A '1' indicates Prometheus reloaded its configuration successfully.",
							PossibleSolutions: `
								- Check Prometheus logs for messages related to configuration loading.
								- Ensure any custom configuration you have provided Prometheus is valid.
							`,
						},
					},
					{
						{
							Name:              "prometheus_target_sample_exceeded",
							Description:       "prometheus scrapes that exceed the sample limit over 10m",
							Query:             "increase(prometheus_target_scrapes_exceeded_sample_limit_total[10m])",
							Warning:           monitoring.Alert().Greater(0),
							Panel:             monitoring.Panel().LegendFormat("rejected scrapes"),
							Owner:             monitoring.ObservableOwnerDistribution,
							PossibleSolutions: "Check Prometheus logs for messages related to target scrape failures.",
						},
						{
							Name:              "prometheus_target_sample_duplicate",
							Description:       "prometheus scrapes rejected due to duplicate timestamps over 10m",
							Query:             "increase(prometheus_target_scrapes_sample_duplicate_timestamp_total[10m])",
							Warning:           monitoring.Alert().Greater(0),
							Panel:             monitoring.Panel().LegendFormat("rejected scrapes"),
							Owner:             monitoring.ObservableOwnerDistribution,
							PossibleSolutions: "Check Prometheus logs for messages related to target scrape failures.",
						},
					},
				},
			},
			{
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("prometheus", monitoring.ObservableOwnerDistribution).Observable(),
						shared.ContainerMemoryUsage("prometheus", monitoring.ObservableOwnerDistribution).Observable(),
					},
					{
						shared.ContainerRestarts("prometheus", monitoring.ObservableOwnerDistribution).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("prometheus", monitoring.ObservableOwnerDistribution).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("prometheus", monitoring.ObservableOwnerDistribution).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("prometheus", monitoring.ObservableOwnerDistribution).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("prometheus", monitoring.ObservableOwnerDistribution).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("prometheus", monitoring.ObservableOwnerDistribution).Observable(),
					},
				},
			},
		},

		// This is third-party service
		NoSourcegraphDebugServer: true,
	}
}
