package shared

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

// Batches exports available shared observable and group constructors related to
// the batch changes team. Some of these panels are useful from multiple container
// contexts, so we maintain this struct as a place of authority over team alert
// definitions.
var Batches batches

// batches provides `Batches` implementations.
type batches struct{}

// src_batches_dbstore_total
// src_batches_dbstore_duration_seconds_bucket
// src_batches_dbstore_errors_total
func (batches) NewDBStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerBatches, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "batches",
			DescriptionRoot: "dbstore stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "batches_dbstore",
				MetricDescriptionRoot: "store",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_batches_webhooks_total
// src_batches_webhooks_duration_seconds_bucket
// src_batches_webhooks_errors_total
func (batches) NewWebhooksGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerBatches, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "batches",
			DescriptionRoot: "webhooks stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "batches_webhooks",
				MetricDescriptionRoot: "webhook",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}
