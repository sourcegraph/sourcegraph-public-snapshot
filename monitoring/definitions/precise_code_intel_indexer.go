package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func PreciseCodeIntelIndexer() *monitoring.Container {
	return &monitoring.Container{
		Name:        "precise-code-intel-indexer",
		Title:       "Precise Code Intel Indexer",
		Description: `Executes jobs from the "codeintel" work queue.`,
		Groups: []monitoring.Group{
			{
				Title: "Executor",
				Rows: []monitoring.Row{
					{
						{
							Name:           "codeintel_job_99th_percentile_duration",
							Description:    "99th percentile successful job duration over 5m",
							Query:          `histogram_quantile(0.99, sum by (le)(rate(src_executor_queue_processor_duration_seconds_bucket{queue="codeintel"}[5m])))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("jobs").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
						{
							Name:           "codeintel_active_handlers",
							Description:    "active handlers processing jobs",
							Query:          `sum(src_executor_queue_processor_handlers{queue="codeintel"})`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("handlers"),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
						{
							Name:              "codeintel_job_errors",
							Description:       "job errors every 5m",
							Query:             `sum(increase(src_executor_queue_processor_errors_total{queue="codeintel"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title: "Stores and clients",
				Rows: []monitoring.Row{
					{
						{
							Name:              "executor_apiclient_99th_percentile_duration",
							Description:       "99th percentile successful API request duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_apiworker_apiclient_duration_seconds_bucket{job="sourcegraph-code-intel-indexers"}[5m])))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("requests").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "executor_apiclient_errors",
							Description:       "aPI errors every 5m", // DUMB
							Query:             `sum(increase(src_apiworker_apiclient_errors_total{job="sourcegraph-code-intel-indexers"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title: "Commands",
				Rows: []monitoring.Row{
					{
						{
							Name:           "executor_setup_command_99th_percentile_duration",
							Description:    "99th percentile successful setup command duration over 5m",
							Query:          `histogram_quantile(0.99, sum by (le)(rate(src_apiworker_command_duration_seconds_bucket{job="sourcegraph-code-intel-indexers", op=~"setup.*"}[5m])))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("commands").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
						{
							Name:              "executor_setup_command_errors",
							Description:       "setup command errors every 5m",
							Query:             `sum(increase(src_apiworker_command_errors_total{job="sourcegraph-code-intel-indexers", op=~"setup.*"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:           "executor_exec_command_99th_percentile_duration",
							Description:    "99th percentile successful exec command duration over 5m",
							Query:          `histogram_quantile(0.99, sum by (le)(rate(src_apiworker_command_duration_seconds_bucket{job="sourcegraph-code-intel-indexers", op=~"exec.*"}[5m])))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("commands").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
						{
							Name:              "executor_exec_command_errors",
							Description:       "exec command errors every 5m",
							Query:             `sum(increase(src_apiworker_command_errors_total{job="sourcegraph-code-intel-indexers", op=~"exec.*"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:           "executor_teardown_command_99th_percentile_duration",
							Description:    "99th percentile successful teardown command duration over 5m",
							Query:          `histogram_quantile(0.99, sum by (le)(rate(src_apiworker_teardown_command_duration_seconds_bucket{job="sourcegraph-code-intel-indexers", op=~"teardown.*"}[5m])))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("commands").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
						{
							Name:              "executor_teardown_command_errors",
							Description:       "teardown command errors every 5m",
							Query:             `sum(increase(src_apiworker_teardown_command_errors_total{job="sourcegraph-code-intel-indexers", op=~"teardown.*"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ContainerMemoryUsage("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ContainerMissing("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleGolangMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.GoGcDuration("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
		},
	}
}
