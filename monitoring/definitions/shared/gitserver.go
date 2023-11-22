package shared

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

// GitServer exports available shared observable and group constructors related to gitserver and
// the client. Some of these panels are useful from multiple container contexts, so we maintain
// this struct as a place of authority over team alert definitions.
var GitServer gitServer

// gitServer provides `GitServer` implementations.
type gitServer struct{}

// src_gitserver_api_total
// src_gitserver_api_duration_seconds_bucket
// src_gitserver_api_errors_total
func (gitServer) NewAPIGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerSource, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "gitserver",
			DescriptionRoot: "Gitserver API (powered by internal/observation)",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "gitserver_api",
				MetricDescriptionRoot: "graphql",
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

// src_gitserver_client_total
// src_gitserver_client_duration_seconds_bucket
// src_gitserver_client_errors_total
func (gitServer) NewClientGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerSource, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "gitserver",
			DescriptionRoot: "Gitserver Client",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "gitserver_client",
				MetricDescriptionRoot: "graphql",
				By:                    []string{"op", "scope"},
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

// src_batch_log_semaphore_wait_duration_seconds_bucket
func (gitServer) NewBatchLogSemaphoreWait(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  "Global operation semaphores",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				NoAlertsOption("none")(Observation.Duration(ObservableConstructorOptions{
					MetricNameRoot:        "batch_log_semaphore_wait",
					MetricDescriptionRoot: "batch log semaphore",
				})(containerName, monitoring.ObservableOwnerSource)).Observable(),
			},
		},
	}
}
