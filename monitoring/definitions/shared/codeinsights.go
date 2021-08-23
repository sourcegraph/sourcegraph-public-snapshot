package shared

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

var CodeInsights codeInsights

var namespace string = "codeinsights"

// codeInsights provides `CodeInsights` implementations.
type codeInsights struct{}

// src_insights_search_queue_total
// src_insights_search_queue_processor_total
func (codeInsights) NewInsightsQueryRunnerQueueGroup(containerName string) monitoring.Group {
	return Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeInsights, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       namespace,
			DescriptionRoot: "Query Runner Queue",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "insights_search_queue",
				MetricDescriptionRoot: "code insights search queue",
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

// src_insights_search_queue_processor_total
// src_insights_search_queue_processor_duration_seconds_bucket
// src_insights_search_queue_processor_errors_total
// src_insights_search_queue_processor_handlers
func (codeInsights) NewInsightsQueryRunnerWorkerGroup(containerName string) monitoring.Group {
	return Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeInsights, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       namespace,
			DescriptionRoot: "insights queue processor",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "insights_search_queue",
				MetricDescriptionRoot: "handler",
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Handlers: NoAlertsOption("none"),
	})
}
