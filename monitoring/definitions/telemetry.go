package definitions

import (
	"time"

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
							Query:          `sum(src_telemetrygatewayexport_queue_size)`,
							Panel:          monitoring.Panel().Min(0).LegendFormat("events"),
							NoAlert:        true,
							Interpretation: "The number of events queued to be exported.",
						},
						{
							Name:           "telemetry_gateway_exporter_exported_events",
							Description:    "events exported from queue per hour",
							Owner:          monitoring.ObservableOwnerDataAnalytics,
							Query:          `max(increase(src_telemetrygatewayexport_exported_events[1h]))`,
							Panel:          monitoring.Panel().Min(0).LegendFormat("events"),
							NoAlert:        true,
							Interpretation: "The number of events being exported.",
						},
						{
							Name:           "telemetry_gateway_exporter_export_batch_sizes",
							Description:    "95th percentile number of events exported per batch",
							Owner:          monitoring.ObservableOwnerDataAnalytics,
							Query:          "histogram_quantile(0.95, sum(rate(src_telemetrygatewayexport_batch_size_bucket[5m])) by (le))",
							Panel:          monitoring.Panel().Min(0).LegendFormat("events"),
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
						MetricDescriptionRoot: "telemetry events exporter",
					},
					Namespace:       "Telemetry Gateway Exporter",
					DescriptionRoot: "Export job operations",
					Hidden:          true, // TODO(@bobheadxi): not yet enabled by default, un-hide in 5.2.1
				},
				SharedObservationGroupOptions: shared.SharedObservationGroupOptions{
					Total:    shared.NoAlertsOption("none"),
					Duration: shared.NoAlertsOption("none"),
					Errors:   shared.NoAlertsOption("none"),
					ErrorRate: shared.CriticalOption(monitoring.Alert().Greater(0).For(time.Minute*30), `
						See worker logs in the 'worker.telemetrygateway-exporter' log scope for more details.
						If logs only indicate that exports failed, reach out to Sourcegraph with relevant log entries, as this may be an issue in Sourcegraph's Telemetry Gateway service.
					`),
				},
			}),
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerDataAnalytics, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "telemetrygatewayexporter_queue_cleanup",
						MetricDescriptionRoot: "telemetry export queue cleanup",
					},
					Namespace:       "Telemetry Gateway Exporter",
					DescriptionRoot: "Export queue cleanup job operations",
					Hidden:          true, // TODO(@bobheadxi): not yet enabled by default, un-hide in 5.2.1
				},
				SharedObservationGroupOptions: shared.SharedObservationGroupOptions{
					Total:    shared.NoAlertsOption("none"),
					Duration: shared.NoAlertsOption("none"),
					Errors:   shared.NoAlertsOption("none"),
					ErrorRate: shared.WarningOption(monitoring.Alert().Greater(0).For(time.Minute*30),
						"See worker logs in the `worker.telemetrygateway-exporter` log scope for more details."),
				},
			}),
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerDataAnalytics, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "telemetrygatewayexporter_queue_metrics_reporter",
						MetricDescriptionRoot: "telemetry export backlog metrics reporting",
					},
					Namespace:       "Telemetry Gateway Exporter",
					DescriptionRoot: "Export queue metrics reporting job operations",
					Hidden:          true,
				},
				SharedObservationGroupOptions: shared.SharedObservationGroupOptions{
					Total:    shared.NoAlertsOption("none"),
					Duration: shared.NoAlertsOption("none"),
					Errors:   shared.NoAlertsOption("none"),
					ErrorRate: shared.WarningOption(monitoring.Alert().Greater(0).For(time.Minute*30),
						"See worker logs in the `worker.telemetrygateway-exporter` log scope for more details."),
				},
			}),
		},
	}
}
