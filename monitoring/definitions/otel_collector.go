package definitions

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

func OtelCollector() *monitoring.Dashboard {
	containerName := "otel-collector"

	return &monitoring.Dashboard{
		Name:        containerName,
		Title:       "Open Telemetry Collector",
		Description: "Metrics about the operation of the open telemetry collector.",
		Groups: []monitoring.Group{
			{
				Title:  "Receivers",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						// TODO(burmudar): look into adding a Guage as a Panel type
						{
							Name:        "otel-span-receive-rate",
							Description: "spans received per second",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("receiver: {{receiver}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_receiver_accepted_spans{receiver=~\"^.*\"}[1m])) by (receiver)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver`,
						},
						{
							Name:        "otel-span-refused",
							Description: "spans that the receiver refused",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("receiver: {{receiver}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_receiver_refused_spans{receiver=~\"^.*.*\"}[1m])) by (receiver)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver`,
						},
					},
				},
			},
			{
				Title:  "Exporters",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:        "otel-span-export-rate",
							Description: "spans exported per second",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_sent_spans{exporter=~\"^.*\"}[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans being sent by the exporter`,
						},
						{
							Name:        "otel-span-failed-send-size",
							Description: "spans that the exporter failed to send",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_send_failed_spans{exporter=~\"^.*\"}[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans failed to be sent by the configured reveiver. A number higher than 0 for a long period can indicate a problem with the exporter configuration or with the service that is being exported too`,
						},
						{
							Name:        "otel-span-queue-size",
							Description: "spans pending to be sent",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_queue_size{exporter=~\"^.*\"}[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Indicates the amount of spans that are in the queue to be sent (exported). A high queue count might indicate a high volume of spans or a problem with the receiving service`,
						},
						{
							Name:        "otel-span-queue-capacity",
							Description: "spans max items that can be pending to be sent",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_queue_capacity{exporter=~\"^.*\"}[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Indicates the amount of spans that are in the queue to be sent (exported). A high queue count might indicate a high volume of spans or a problem with the receiving service`,
						},
					},
				},
			},
			{
				Title:  "Collector resource usage",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:        "otel-cpu-usage",
							Description: "cpu usuge of the collector",
							Panel:       monitoring.Panel().Unit(monitoring.Seconds).LegendFormat("{{job}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_process_cpu_seconds{job=~\"^.*\"}[1m])) by (job)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver`,
						},
						{
							Name:        "otel-memory-rss",
							Description: "memory allocated to the otel collector",
							Panel:       monitoring.Panel().Unit(monitoring.Bytes).LegendFormat("{{job}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_process_memory_rss{job=~\"^.*\"}[1m])) by (job)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver`,
						},
						{
							Name:        "otel-memory-usage",
							Description: "total memory usage",
							Panel:       monitoring.Panel().Unit(monitoring.Bytes).LegendFormat("{{job}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_process_runtime_total_alloc_bytes{job=~\"^.*\"}[1m])) by (job)",
							NoAlert:     true,
							Interpretation: `
								Shows how much memory is being used by the otel collector.

								* High memory usage might indicate thad the configured pipeline is keeping a lot of spans in memory for processing
								* Spans failing to be sent and the exporter is configured to retry
								* A high bacth count by using a batch processor`,
						},
					},
				},
			},
		},
	}
}
