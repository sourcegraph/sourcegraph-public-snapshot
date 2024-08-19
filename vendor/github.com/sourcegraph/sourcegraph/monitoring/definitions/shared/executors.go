package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Executors exports available shared observable and group constructors related to
// executors.
//
// TODO: Maybe move more shared.CodeIntelligence group builders here.
var Executors executors

type executors struct{}

// src_executor_total
// src_executor_processor_total
// src_executor_queued_duration_seconds_total
//
// If queueFilter is not a variable, this group is opted-in to centralized observability.
func (executors) NewExecutorQueueGroup(namespace, containerName, queueFilter string) monitoring.Group {
	opts := QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       namespace,
			DescriptionRoot: "Executor jobs",

			// if updating this, also update in NewExecutorProcessorGroup
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "executor",
				MetricDescriptionRoot: "unprocessed executor job",
				Filters:               []string{fmt.Sprintf(`queue=~%q`, queueFilter)},
				By:                    []string{"queue"},
			},
		},

		QueueSize:   NoAlertsOption("none"),
		QueueMaxAge: NoAlertsOption("none"),
		QueueGrowthRate: NoAlertsOption(`
			This value compares the rate of enqueues against the rate of finished jobs for the selected queue.

				- A value < than 1 indicates that process rate > enqueue rate
				- A value = than 1 indicates that process rate = enqueue rate
				- A value > than 1 indicates that process rate < enqueue rate
		`),
	}
	if !strings.Contains(queueFilter, "$") {
		opts.QueueSize = opts.QueueSize.and(MultiInstanceOption())
		opts.QueueMaxAge = opts.QueueMaxAge.and(MultiInstanceOption())
		opts.QueueGrowthRate = opts.QueueGrowthRate.and(MultiInstanceOption())
	}
	return Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, opts)
}

func (executors) NewExecutorMultiqueueGroup(namespace, containerName, queueFilter string) monitoring.Group {
	opts := MultiqueueGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       namespace,
			DescriptionRoot: "Executor jobs",

			// if updating this, also update in NewExecutorProcessorGroup
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "executor",
				MetricDescriptionRoot: "unprocessed executor job",
				Filters:               []string{fmt.Sprintf(`queue=~%q`, queueFilter)},
				By:                    []string{"queue"},
			},
		},
		QueueDequeueCacheSize: NoAlertsOption("none"),
	}
	return Queue.NewMultiqueueGroup(containerName, monitoring.ObservableOwnerCodeIntel, opts)
}
