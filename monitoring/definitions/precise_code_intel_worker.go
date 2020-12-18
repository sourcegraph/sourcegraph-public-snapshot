package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func PreciseCodeIntelWorker() *monitoring.Container {
	return &monitoring.Container{
		Name:        "precise-code-intel-worker",
		Title:       "Precise Code Intel Worker",
		Description: "Handles conversion of uploaded precise code intelligence bundles.",
		Groups: []monitoring.Group{
			{
				Title: "Upload queue",
				Rows: []monitoring.Row{
					{
						{
							Name:              "upload_queue_size",
							Description:       "queue size",
							Query:             `max(src_upload_queue_uploads_total)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(100),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("unprocessed uploads"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "upload_queue_growth_rate",
							Description:       "queue growth rate over 30m",
							Query:             `sum(increase(src_upload_queue_uploads_total[30m])) / sum(increase(src_codeintel_upload_queue_processor_total[30m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("rate of (enqueued / processed)"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "job_errors",
							Description:       "job errors errors every 5m",
							Query:             `sum(increase(src_codeintel_upload_queue_processor_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "active_workers",
							Description:       "active workers processing uploads",
							Query:             `max(up{job="precise-code-intel-worker"})`,
							DataMayNotExist:   true,
							NoAlert:           true,
							PanelOptions:      monitoring.PanelOptions().LegendFormat("workers"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "active_jobs",
							Description:       "active jobs",
							Query:             `sum(src_codeintel_upload_queue_processor_handlers)`,
							DataMayNotExist:   true,
							NoAlert:           true,
							PanelOptions:      monitoring.PanelOptions().LegendFormat("jobs"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title: "Workers",
				Rows: []monitoring.Row{
					{
						{
							Name:              "job_99th_percentile_duration",
							Description:       "99th percentile successful job duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_upload_queue_processor_duration_seconds_bucket[5m])))`,
							DataMayNotExist:   true,
							NoAlert:           true,
							PanelOptions:      monitoring.PanelOptions().LegendFormat("jobs").Unit(monitoring.Seconds),
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
							Name:              "codeintel_dbstore_99th_percentile_duration",
							Description:       "99th percentile successful database store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_dbstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_dbstore_errors",
							Description:       "database store errors every 5m",
							Query:             `sum(increase(src_codeintel_dbstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_workerstore_99th_percentile_duration",
							Description:       "99th percentile successful worker store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_workerutil_dbworker_store_precise_code_intel_upload_worker_store_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_workerstore_errors",
							Description:       "worker store errors every 5m",
							Query:             `sum(increase(src_workerutil_dbworker_store_precise_code_intel_upload_worker_store_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_lsifstore_99th_percentile_duration",
							Description:       "99th percentile successful LSIF store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_lsifstore_errors",
							Description:       "lSIF store errors every 5m", // DUMB
							Query:             `sum(increase(src_codeintel_lsifstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_uploadstore_99th_percentile_duration",
							Description:       "99th percentile successful upload store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_uploadstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_uploadstore_errors",
							Description:       "upload store errors every 5m",
							Query:             `sum(increase(src_codeintel_uploadstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_gitserverclient_99th_percentile_duration",
							Description:       "99th percentile successful gitserver client operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_gitserver_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_gitserverclient_errors",
							Description:       "gitserver client errors every 5m",
							Query:             `sum(increase(src_codeintel_gitserver_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("errors"),
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
						shared.FrontendInternalAPIErrorResponses("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
						shared.ContainerMemoryUsage("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
					},
					{
						shared.ContainerRestarts("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
						shared.ContainerFsInodes("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
						shared.ProvisioningMemoryUsageLongTerm("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
						shared.ProvisioningMemoryUsageShortTerm("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
						shared.GoGcDuration("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
