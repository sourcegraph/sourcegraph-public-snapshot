package main

func Prometheus() *Container {
	return &Container{
		Name:        "prometheus",
		Title:       "Prometheus",
		Description: "Sourcegraph's all-in-one Prometheus and Alertmanager service.",
		Groups: []Group{
			{
				Title: "Metrics",
				Rows: []Row{
					{
						{
							Name:              "prometheus_metrics_bloat",
							Description:       "prometheus metrics payload size",
							Query:             `http_response_size_bytes{handler="prometheus",job!="kubernetes-apiservers",job!="kubernetes-nodes",quantile="0.5"}`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20000},
							Owner:             ObservableOwnerDistribution,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title: "Alerts",
				Rows: []Row{
					{
						{
							Name:              "alertmanager_notifications_failed_total",
							Description:       "failed alertmanager notifications rate over 1m",
							Query:             `rate(alertmanager_notifications_failed_total[1m])`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 1},
							Owner:             ObservableOwnerDistribution,
							PossibleSolutions: "Ensure that your `observability.alerts` configuration (in site configuration) is valid.",
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("prometheus"),
						sharedContainerMemoryUsage("prometheus"),
					},
					{
						sharedContainerRestarts("prometheus"),
						sharedContainerFsInodes("prometheus"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsage7d("prometheus"),
						sharedProvisioningMemoryUsage7d("prometheus"),
					},
					{
						sharedProvisioningCPUUsage5m("prometheus"),
						sharedProvisioningMemoryUsage5m("prometheus"),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("prometheus"),
					},
				},
			},
			{
				Title:  "Test alerts",
				Hidden: true,
				Rows: []Row{
					{
						{
							Name:              "observability_sample_alert_warning",
							Description:       "sample warning alert metric",
							Query:             `max(observability_sample_metric_warning)`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 1},
							Owner:             ObservableOwnerDistribution,
							PossibleSolutions: "Disable this alert via the `setObservabilityTestAlertState` GraphQL endpoint.",
						},
						{
							Name:              "observability_sample_alert_critical",
							Description:       "sample critical alert metric",
							Query:             `max(observability_sample_metric_critical)`,
							DataMayNotExist:   true,
							Critical:          Alert{GreaterOrEqual: 1},
							Owner:             ObservableOwnerDistribution,
							PossibleSolutions: "Disable this alert via the `setObservabilityTestAlertState` GraphQL endpoint.",
						},
					},
				},
			},
		},
	}
}
