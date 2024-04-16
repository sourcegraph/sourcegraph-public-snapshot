package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func OtelCollector() *monitoring.Dashboard {
	containerName := "otel-collector"

	return &monitoring.Dashboard{
		Name:        containerName,
		Title:       "OpenTelemetry Collector",
		Description: "The OpenTelemetry collector ingests OpenTelemetry data from Sourcegraph and exports it to the configured backends.",
		Groups: []monitoring.Group{
			{
				Title:  "Receivers",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:        "otel_span_receive_rate",
							Description: "spans received per receiver per minute",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("receiver: {{receiver}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							Query:       "sum by (receiver) (rate(otelcol_receiver_accepted_spans[1m]))",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver

								A Trace is a collection of spans and a span represents a unit of work or operation. Spans are the building blocks of Traces.
								The spans have only been accepted by the receiver, which means they still have to move through the configured pipeline to be exported.
								For more information on tracing and configuration of a OpenTelemetry receiver see https://opentelemetry.io/docs/collector/configuration/#receivers.

								See the Exporters section see spans that have made it through the pipeline and are exported.

								Depending the configured processors, received spans might be dropped and not exported. For more information on configuring processors see
								https://opentelemetry.io/docs/collector/configuration/#processors.
							`,
						},
						{
							Name:        "otel_span_refused",
							Description: "spans refused per receiver",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("receiver: {{receiver}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							Query:       "sum by (receiver) (rate(otelcol_receiver_refused_spans[1m]))",
							Warning:     monitoring.Alert().Greater(1).For(5 * time.Minute),
							NextSteps:   "Check logs of the collector and configuration of the receiver",
							Interpretation: `
								Shows the amount of spans that have been refused by a receiver.

								A Trace is a collection of spans. A Span represents a unit of work or operation. Spans are the building blocks of Traces.

 								Spans can be rejected either due to a misconfigured receiver or receiving spans in the wrong format. The log of the collector will have more information on why a span was rejected.
								For more information on tracing and configuration of a OpenTelemetry receiver see https://opentelemetry.io/docs/collector/configuration/#receivers.
							`,
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
							Name:        "otel_span_export_rate",
							Description: "spans exported per exporter per minute",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							Query:       "sum by (exporter) (rate(otelcol_exporter_sent_spans[1m]))",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans being sent by the exporter

								A Trace is a collection of spans. A Span represents a unit of work or operation. Spans are the building blocks of Traces.
								The rate of spans here indicates spans that have made it through the configured pipeline and have been sent to the configured export destination.

								For more information on configuring a exporter for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configuration/#exporters.
							`,
						},
						{
							Name:        "otel_span_export_failures",
							Description: "span export failures by exporter",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							Query:       "sum by (exporter) (rate(otelcol_exporter_send_failed_spans[1m]))",
							Warning:     monitoring.Alert().Greater(1).For(5 * time.Minute),
							NextSteps:   "Check the configuration of the exporter and if the service being exported is up",
							Interpretation: `
								Shows the rate of spans failed to be sent by the configured reveiver. A number higher than 0 for a long period can indicate a problem with the exporter configuration or with the service that is being exported too

								For more information on configuring a exporter for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configuration/#exporters.
							`,
						},
					},
				},
			},
			{
				Title:  "Queue Length",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:           "otelcol_exporter_queue_capacity",
							Description:    "exporter queue capacity",
							Panel:          monitoring.Panel().LegendFormat("exporter: {{exporter}}"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Query:          "sum by (exporter) (rate(otelcol_exporter_queue_capacity{job=~\"^.*\"}[1m]))",
							NoAlert:        true,
							Interpretation: `Shows the the capacity of the retry queue (in batches).`,
						},
						{
							Name:           "otelcol_exporter_queue_size",
							Description:    "exporter queue size",
							Panel:          monitoring.Panel().LegendFormat("exporter: {{exporter}}"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Query:          "sum by (exporter) (rate(otelcol_exporter_queue_size{job=~\"^.*\"}[1m]))",
							NoAlert:        true,
							Interpretation: `Shows the current size of retry queue`,
						},
						{
							Name:           "otelcol_exporter_enqueue_failed_spans",
							Description:    "exporter enqueue failed spans",
							Panel:          monitoring.Panel().LegendFormat("exporter: {{exporter}}"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Query:          "sum by (exporter) (rate(otelcol_exporter_enqueue_failed_spans{job=~\"^.*\"}[1m]))",
							Warning:        monitoring.Alert().Greater(0).For(5 * time.Minute),
							NextSteps:      "Check the configuration of the exporter and if the service being exported is up. This may be caused by a queue full of unsettled elements, so you may need to decrease your sending rate or horizontally scale collectors.",
							Interpretation: `Shows the rate of spans failed to be enqueued by the configured exporter. A number higher than 0 for a long period can indicate a problem with the exporter configuration`,
						},
					},
				},
			},
			{
				Title:  "Processors",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:           "otelcol_processor_dropped_spans",
							Description:    "spans dropped per processor per minute",
							Panel:          monitoring.Panel().Unit(monitoring.Number).LegendFormat("processor: {{processor}}"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Query:          "sum by (processor) (rate(otelcol_processor_dropped_spans[1m]))",
							Warning:        monitoring.Alert().Greater(0).For(5 * time.Minute),
							NextSteps:      "Check the configuration of the processor",
							Interpretation: `Shows the rate of spans dropped by the configured processor`,
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
							Name:           "otel_cpu_usage",
							Description:    "cpu usage of the collector",
							Panel:          monitoring.Panel().Unit(monitoring.Seconds).LegendFormat("{{job}}"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Query:          "sum by (job) (rate(otelcol_process_cpu_seconds{job=~\"^.*\"}[1m]))",
							NoAlert:        true,
							Interpretation: `Shows CPU usage as reported by the OpenTelemetry collector.`,
						},
						{
							Name:           "otel_memory_resident_set_size",
							Description:    "memory allocated to the otel collector",
							Panel:          monitoring.Panel().Unit(monitoring.Bytes).LegendFormat("{{job}}"),
							Owner:          monitoring.ObservableOwnerInfraOrg,
							Query:          "sum by (job) (rate(otelcol_process_memory_rss{job=~\"^.*\"}[1m]))",
							NoAlert:        true,
							Interpretation: `Shows the allocated memory Resident Set Size (RSS) as reported by the OpenTelemetry collector.`,
						},
						{
							Name:        "otel_memory_usage",
							Description: "memory used by the collector",
							Panel:       monitoring.Panel().Unit(monitoring.Bytes).LegendFormat("{{job}}"),
							Owner:       monitoring.ObservableOwnerInfraOrg,
							Query:       "sum by (job) (rate(otelcol_process_runtime_total_alloc_bytes{job=~\"^.*\"}[1m]))",
							NoAlert:     true,
							Interpretation: `
								Shows how much memory is being used by the otel collector.

								* High memory usage might indicate thad the configured pipeline is keeping a lot of spans in memory for processing
								* Spans failing to be sent and the exporter is configured to retry
								* A high batch count by using a batch processor

								For more information on configuring processors for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configuration/#processors.
							`,
						},
					},
				},
			},
			shared.NewContainerMonitoringGroup("otel-collector", monitoring.ObservableOwnerInfraOrg, nil),
			shared.NewKubernetesMonitoringGroup("otel-collector", monitoring.ObservableOwnerInfraOrg, nil),
		},
	}
}
