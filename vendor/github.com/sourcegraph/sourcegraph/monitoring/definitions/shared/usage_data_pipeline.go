package shared

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

var DataAnalytics dataAnalytics

// codeInsights provides `CodeInsights` implementations.
type dataAnalytics struct{}

var usageDataExporterNamespace = "Usage data exporter (legacy)"

// src_telemetry_job_queue_size
func (dataAnalytics) NewTelemetryJobQueueGroup(containerName string) monitoring.Group {
	return Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeInsights, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       usageDataExporterNamespace,
			DescriptionRoot: "Queue size",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "telemetry_job_queue_size",
				MetricDescriptionRoot: "event level usage data",
			},
		},

		QueueSize: NoAlertsOption("none"),
		QueueGrowthRate: NoAlertsOption(`
			This value compares the rate of enqueues against the rate of finished jobs.

				- A value < than 1 indicates that process rate > enqueue rate
				- A value = than 1 indicates that process rate = enqueue rate
				- A value > than 1 indicates that process rate < enqueue rate
		`),
	})
}

func (dataAnalytics) NewTelemetryJobOperationsGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerDataAnalytics, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "telemetry_job",
				MetricDescriptionRoot: "usage data exporter",
				By:                    []string{"op"},
			},
			Namespace:       usageDataExporterNamespace,
			DescriptionRoot: "Job operations",
			Hidden:          true,
		},
		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: WarningOption(monitoring.Alert().Greater(0).For(time.Minute*30), "Involved cloud team to inspect logs of the managed instance to determine error sources."),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

func (dataAnalytics) TelemetryJobThroughputGroup(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  "Usage data exporter (legacy): Utilization",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Name:          "telemetry_job_utilized_throughput",
					Description:   "utilized percentage of maximum throughput",
					Owner:         monitoring.ObservableOwnerDataAnalytics,
					Query:         `rate(src_telemetry_job_total{op="SendEvents"}[1h]) / on() group_right() src_telemetry_job_max_throughput * 100`,
					DataMustExist: false,
					Warning:       monitoring.Alert().Greater(90).For(time.Minute * 30),
					NextSteps:     "Throughput utilization is high. This could be a signal that this instance is producing too many events for the export job to keep up. Configure more throughput using the maxBatchSize option.",
					Panel:         monitoring.Panel().LegendFormat("percent utilized").Unit(monitoring.Percentage),
				},
			},
		},
	}
}
