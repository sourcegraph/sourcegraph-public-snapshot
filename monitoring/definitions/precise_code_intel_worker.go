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
							Warning:           monitoring.Alert().GreaterOrEqual(100, nil),
							Panel:             monitoring.Panel().LegendFormat("unprocessed uploads"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "upload_queue_growth_rate",
							Description:       "queue growth rate over 30m",
							Query:             `sum(increase(src_upload_queue_uploads_total[30m])) / sum(increase(src_codeintel_upload_queue_processor_total[30m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(5, nil),
							Panel:             monitoring.Panel().LegendFormat("rate of (enqueued / processed)"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "job_errors",
							Description:       "job errors errors every 5m",
							Query:             `sum(increase(src_codeintel_upload_queue_processor_errors_total[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:           "active_workers",
							Description:    "active workers processing uploads",
							Query:          `max(up{job="precise-code-intel-worker"})`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("workers"),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
						{
							Name:           "active_jobs",
							Description:    "active jobs",
							Query:          `sum(src_codeintel_upload_queue_processor_handlers)`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("jobs"),
							Owner:          monitoring.ObservableOwnerCodeIntel,
							Interpretation: "none",
						},
					},
				},
			},
			{
				Title: "Workers",
				Rows: []monitoring.Row{
					{
						{
							Name:           "job_99th_percentile_duration",
							Description:    "99th percentile successful job duration over 5m",
							Query:          `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_upload_queue_processor_duration_seconds_bucket[5m])))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("jobs").Unit(monitoring.Seconds),
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
							Name:              "codeintel_dbstore_99th_percentile_duration",
							Description:       "99th percentile successful database store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_dbstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_dbstore_errors",
							Description:       "database store errors every 5m",
							Query:             `sum(increase(src_codeintel_dbstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_workerstore_99th_percentile_duration",
							Description:       "99th percentile successful worker store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_workerutil_dbworker_store_precise_code_intel_upload_worker_store_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_workerstore_errors",
							Description:       "worker store errors every 5m",
							Query:             `sum(increase(src_workerutil_dbworker_store_precise_code_intel_upload_worker_store_errors_total{job="precise-code-intel-worker"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_lsifstore_99th_percentile_duration",
							Description:       "99th percentile successful LSIF store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_lsifstore_errors",
							Description:       "lSIF store errors every 5m", // DUMB
							Query:             `sum(increase(src_codeintel_lsifstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_uploadstore_99th_percentile_duration",
							Description:       "99th percentile successful upload store operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_uploadstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_uploadstore_errors",
							Description:       "upload store errors every 5m",
							Query:             `sum(increase(src_codeintel_uploadstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_gitserverclient_99th_percentile_duration",
							Description:       "99th percentile successful gitserver client operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_gitserver_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							Warning:           monitoring.Alert().GreaterOrEqual(20, nil),
							Panel:             monitoring.Panel().LegendFormat("operations").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_gitserverclient_errors",
							Description:       "gitserver client errors every 5m",
							Query:             `sum(increase(src_codeintel_gitserver_errors_total{job="precise-code-intel-worker"}[5m]))`,
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
						shared.FrontendInternalAPIErrorResponses("precise-code-intel-worker", monitoring.ObservableOwnerCodeIntel).Observable(),
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
