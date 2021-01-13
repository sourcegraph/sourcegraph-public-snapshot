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
							Name:              "prometheus_metrics_bloat",
							Description:       "prometheus metrics payload size",
							Query:             `http_response_size_bytes{handler="prometheus",job!="kubernetes-apiservers",job!="kubernetes-nodes",quantile="0.5"}`,
							Warning:           monitoring.Alert().GreaterOrEqual(20000),
							Panel:             monitoring.Panel().Unit(monitoring.Bytes).LegendFormat("{{instance}}"),
							Owner:             monitoring.ObservableOwnerDistribution,
							PossibleSolutions: "none",
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
	}
}
