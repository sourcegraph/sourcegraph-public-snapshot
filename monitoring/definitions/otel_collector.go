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
						{
							Name:        "otel-span-receive-rate",
							Description: "Spans received per second",
							Panel:       monitoring.Panel().Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_receiver_accepted_spans[1m])) by (receiver)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver`,
						},
						{
							Name:        "otel-span-export-rate",
							Description: "Spans exported per second",
							Panel:       monitoring.Panel().Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_accepted_spans[1m])) by (receiver)",
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
