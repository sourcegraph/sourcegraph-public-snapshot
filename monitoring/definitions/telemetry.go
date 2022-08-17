package definitions

import (
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
			shared.DataAnalytics.NewTelemetryJobOperationsGroup(containerName),
			shared.DataAnalytics.NewTelemetryJobQueueGroup(containerName),
			shared.DataAnalytics.TelemetryJobThroughputGroup(containerName),
		},
	}
}
