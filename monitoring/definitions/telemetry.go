package definitions

import (
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/prometheus/common/model"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Telemetry() *monitoring.Dashboard {
	containerName := "worker"
	return &monitoring.Dashboard{
		Name:        "telemetry",
		Title:       "Telemetry",
		Description: "Monitoring telemetry services in Sourcegraph.",
		Groups: []monitoring.Group{
			// Legacy dashboards - TODO(@bobheadxi): remove after 5.2.2
			shared.DataAnalytics.NewTelemetryJobOperationsGroup(containerName),
			shared.DataAnalytics.NewTelemetryJobQueueGroup(containerName),
			shared.DataAnalytics.TelemetryJobThroughputGroup(containerName),

			// The new stuff - https://docs.sourcegraph.com/dev/background-information/telemetry
			{
				Title:  "Telemetry Gateway Exporter: Export and queue metrics",
				Hidden: true, // TODO(@bobheadxi): not yet enabled by default, un-hide in 5.2.1
				Rows: []monitoring.Row{
					{
						{
							Name:           "telemetry_gateway_exporter_queue_size",
							Description:    "telemetry event payloads pending export",
							Owner:          monitoring.ObservableOwnerDataAnalytics,
							Query:          `sum(src_telemetrygatewayexporter_queue_size)`,
							Panel:          monitoring.Panel().Min(0).LegendFormat("events"),
							NoAlert:        true,
							Interpretation: "The number of events queued to be exported.",
						},
						{
							Name:           "telemetry_gateway_exporter_queue_growth",
							Description:    "rate of growth of export queue over 30m",
							Owner:          monitoring.ObservableOwnerDataAnalytics,
							Query:          `max(deriv(src_telemetrygatewayexporter_queue_size[30m]))`,
							Panel:          monitoring.Panel().LegendFormat("growth").MinAuto(),
							Interpretation: `A positive value indicates the queue is growing.`,
							// Warn when steadily growing
							Warning: monitoring.Alert().Greater(1).For(1 * time.Hour),
							// Critical when it grows without ever reducing
							Critical: monitoring.Alert().Greater(1).For(36 * time.Hour),
							NextSteps: `
								- Increase 'TELEMETRY_GATEWAY_EXPORTER_EXPORT_BATCH_SIZE' to export more events per batch.
								- Reduce 'TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL' to schedule more export jobs.
								- See worker logs in the 'worker.telemetrygateway-exporter' log scope for more details to see if any export errors are occuring - if logs only indicate that exports failed, reach out to Sourcegraph with relevant log entries, as this may be an issue in Sourcegraph's Telemetry Gateway service.
							`,
						},
					},
					{
						{
							Name:           "src_telemetrygatewayexporter_exported_events",
							Description:    "events exported from queue per hour",
							Owner:          monitoring.ObservableOwnerDataAnalytics,
							Query:          `max(increase(src_telemetrygatewayexporter_exported_events[1h]))`,
							Panel:          monitoring.Panel().Min(0).LegendFormat("events"),
							NoAlert:        true,
							Interpretation: "The number of events being exported.",
						},
						{
							Name:        "telemetry_gateway_exporter_batch_size",
							Description: "number of events exported per batch over 30m",
							Owner:       monitoring.ObservableOwnerDataAnalytics,
							Query:       "sum by (le) (rate(src_telemetrygatewayexporter_batch_size_bucket[30m]))",
							Panel: monitoring.PanelHeatmap().
								With(func(o monitoring.Observable, p *sdk.Panel) {
									p.HeatmapPanel.YAxis.Format = "short"
									p.HeatmapPanel.YAxis.Decimals = pointers.Ptr(0)
									p.HeatmapPanel.DataFormat = "tsbuckets"
									p.HeatmapPanel.Targets[0].Format = "heatmap"
									p.HeatmapPanel.Targets[0].LegendFormat = "{{le}}"
								}),
							NoAlert:        true,
							Interpretation: "The number of events exported in each batch.",
						},
					},
				},
			},
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerDataAnalytics, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "telemetrygatewayexporter_exporter",
						MetricDescriptionRoot: "events exporter",
						RangeWindow:           model.Duration(30 * time.Minute),
					},
					Namespace:       "Telemetry Gateway Exporter",
					DescriptionRoot: "Export job operations",
					Hidden:          true, // TODO(@bobheadxi): not yet enabled by default, un-hide in 5.2.1
				},
				SharedObservationGroupOptions: shared.SharedObservationGroupOptions{
					Total:     shared.NoAlertsOption("none"),
					Duration:  shared.NoAlertsOption("none"),
					ErrorRate: shared.NoAlertsOption("none"),
					Errors: shared.WarningOption(monitoring.Alert().Greater(0), `
						See worker logs in the 'worker.telemetrygateway-exporter' log scope for more details.
						If logs only indicate that exports failed, reach out to Sourcegraph with relevant log entries, as this may be an issue in Sourcegraph's Telemetry Gateway service.
					`),
				},
			}),
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerDataAnalytics, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "telemetrygatewayexporter_queue_cleanup",
						MetricDescriptionRoot: "export queue cleanup",
						RangeWindow:           model.Duration(30 * time.Minute),
					},
					Namespace:       "Telemetry Gateway Exporter",
					DescriptionRoot: "Export queue cleanup job operations",
					Hidden:          true, // TODO(@bobheadxi): not yet enabled by default, un-hide in 5.2.1
				},
				SharedObservationGroupOptions: shared.SharedObservationGroupOptions{
					Total:     shared.NoAlertsOption("none"),
					Duration:  shared.NoAlertsOption("none"),
					ErrorRate: shared.NoAlertsOption("none"),
					Errors: shared.WarningOption(monitoring.Alert().Greater(0),
						"See worker logs in the `worker.telemetrygateway-exporter` log scope for more details."),
				},
			}),
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerDataAnalytics, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "telemetrygatewayexporter_queue_metrics_reporter",
						MetricDescriptionRoot: "export backlog metrics reporting",
						RangeWindow:           model.Duration(30 * time.Minute),
					},
					Namespace:       "Telemetry Gateway Exporter",
					DescriptionRoot: "Export queue metrics reporting job operations",
					Hidden:          true,
				},
				SharedObservationGroupOptions: shared.SharedObservationGroupOptions{
					Total:     shared.NoAlertsOption("none"),
					Duration:  shared.NoAlertsOption("none"),
					ErrorRate: shared.NoAlertsOption("none"),
					Errors: shared.WarningOption(monitoring.Alert().Greater(0),
						"See worker logs in the `worker.telemetrygateway-exporter` log scope for more details."),
				},
			}),
		},
	}
}
