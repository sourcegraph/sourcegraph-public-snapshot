package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func ExecutorQueue() *monitoring.Container {
	return &monitoring.Container{
		Name:        "executor-queue",
		Title:       "Executor Queue",
		Description: "Coordinates the executor work queues.",
		Groups: []monitoring.Group{
			{
				Title: "Code intelligence queue",
				Rows: []monitoring.Row{
					{
						{
							Name:              "codeintel_queue_size",
							Description:       "queue size",
							Query:             `max(src_executor_queue_total{queue="codeintel"})`,
							Warning:           monitoring.Alert().GreaterOrEqual(100, nil),
							Panel:             monitoring.Panel().LegendFormat("unprocessed jobs"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_queue_growth_rate",
							Description:       "queue growth rate over 30m",
							Query:             `sum(increase(src_executor_queue_total{queue="codeintel"}[30m])) / sum(increase(src_executor_queue_processor_total{queue="codeintel"}[30m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(5, nil),
							Panel:             monitoring.Panel().LegendFormat("rate of (enqueued / processed)"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
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
					{
						{
							Name:           "codeintel_active_executors",
							Description:    "active executors processing codeintel jobs",
							Query:          `max(src_apiworker_apiserver_executors_total{queue="codeintel"})`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("executors"),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
						{
							Name:           "codeintel_active_jobs",
							Description:    "active jobs",
							Query:          `sum(src_apiworker_apiserver_jobs_total{queue="codeintel"})`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("jobs"),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
					},
				},
			},
			{
				Title: "Stores and clients",
				Rows: []monitoring.Row{
					{
						{
							Name:              "codeintel_workerstore_99th_percentile_duration",
							Description:       "99th percentile successful worker store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_workerutil_dbworker_store_precise_code_intel_index_worker_store_duration_seconds_bucket{job="executor-queue"}[5m])))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_workerstore_errors",
							Description:       "worker store errors every 5m",
							Query:             `sum(increase(src_workerutil_dbworker_store_precise_code_intel_index_worker_store_errors_total{job="executor-queue"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.FrontendInternalAPIErrorResponses("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ContainerMemoryUsage("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ContainerMissing("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleGolangMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.GoGcDuration("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("executor-queue", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
		},
	}
}
