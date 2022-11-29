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
				Title:  "Export failures",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						// TODO(burmudar): look into adding a Guage as a Panel type
						{
							Name:        "otel-span-receive-rate",
							Description: "spans received per second",
							Panel:       monitoring.Panel().Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_receiver_accepted_spans[1m])) by (receiver)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver`,
						},
						{
							Name:        "otel-span-export-rate",
							Description: "spans exported per second",
							Panel:       monitoring.Panel().Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_sent_spans[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans being sent by the exporter`,
						},
						{
							Name:        "otel-span-failed-send",
							Description: "spans that the exporter failed to send size",
							Panel:       monitoring.Panel().Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(otelcol_exporter_send_failed_spans[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver`,
						},
					},
				},
			},
		},
	}
}
