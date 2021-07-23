package definitions

import (
	"github.com/grafana-tools/sdk"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Executor() *monitoring.Container {
	const (
		containerName      = "(executor|sourcegraph-code-intel-indexers|executor-batches)"
		queueContainerName = "(executor|sourcegraph-code-intel-indexers|executor-batches|executor-queue)"
	)

	filters := []string{`queue=~"${queue:regex}"`}

	return &monitoring.Container{
		Name:        "executor",
		Title:       "Executor",
		Description: `Executes jobs from the executor-queue.`,
		Templates: []sdk.TemplateVar{
			{
				Label:      "Queue name",
				Name:       "queue",
				AllValue:   ".*",
				Current:    sdk.Current{Text: &sdk.StringSliceString{Value: []string{"all"}, Valid: true}, Value: "$__all"},
				IncludeAll: true,
				Options: []sdk.Option{
					{Text: "all", Value: "$__all", Selected: true},
					{Text: "batches", Value: "batches"},
					{Text: "codeintel", Value: "codeintel"},
				},
				Query: "batches,codeintel",
				Type:  "custom",
			},
		},
		Groups: []monitoring.Group{
			// src_executor_total
			// src_executor_processor_total
			shared.Queue.NewGroup(queueContainerName, monitoring.ObservableOwnerCodeIntel, shared.QueueSizeGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "executor",
					DescriptionRoot: "Executor jobs",

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "executor",
						MetricDescriptionRoot: "unprocessed executor job",
						Filters:               filters,
					},
				},

				QueueSize: shared.NoAlertsOption("none"),
				QueueGrowthRate: shared.NoAlertsOption(`
					This value compares the rate of enqueues against the rate of finished jobs for the selected queue.

						- A value < than 1 indicates that process rate > enqueue rate
						- A value = than 1 indicates that process rate = enqueue rate
						- A value > than 1 indicates that process rate < enqueue rate
				`),
			}),

			// src_executor_processor_total
			// src_executor_processor_duration_seconds_bucket
			// src_executor_processor_errors_total
			// src_executor_processor_handlers
			shared.Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.WorkerutilGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "executor",
					DescriptionRoot: "Executor jobs",

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "executor",
						MetricDescriptionRoot: "handler",
						Filters:               filters,
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
				Handlers: shared.NoAlertsOption("none"),
			}),

			// src_apiworker_apiclient_total
			// src_apiworker_apiclient_duration_seconds_bucket
			// src_apiworker_apiclient_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "executor",
					DescriptionRoot: "Queue API client",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "apiworker_apiclient",
						MetricDescriptionRoot: "client",
						Filters:               nil, // note: shared between queues
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_apiworker_command_total
			// src_apiworker_command_duration_seconds_bucket
			// src_apiworker_command_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "executor",
					DescriptionRoot: "Subprocess execution (for job setup)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "apiworker_command",
						MetricDescriptionRoot: "command",
						Filters:               []string{`op=~"setup.*"`}, // note: shared between queues
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_apiworker_command_total
			// src_apiworker_command_duration_seconds_bucket
			// src_apiworker_command_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "executor",
					DescriptionRoot: "Subprocess execution (for job execution)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "apiworker_command",
						MetricDescriptionRoot: "command",
						Filters:               []string{`op=~"exec.*"`}, // note: shared between queues
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_apiworker_command_total
			// src_apiworker_command_duration_seconds_bucket
			// src_apiworker_command_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "executor",
					DescriptionRoot: "Subprocess execution (for job teardown)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "apiworker_command",
						MetricDescriptionRoot: "command",
						Filters:               []string{`op=~"teardown.*"`}, // note: shared between queues
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// Resource monitoring
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
		},
	}
}
