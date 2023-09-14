package shared

import (
	"time"

	"github.com/prometheus/common/model"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

var SourcegraphOwn sourcegraphOwn

var ownNamespace = "own"

// sourcegraphOwn provides `SourcegraphOwn` implementations.
type sourcegraphOwn struct{}

func (sourcegraphOwn) NewOwnRepoIndexerStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerOwn, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       ownNamespace,
			DescriptionRoot: "repo indexer dbstore",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "workerutil_dbworker_store_own_background_worker_store",
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

// src_own_background_worker_total
// src_own_background_worker_duration_seconds_bucket
// src_own_background_worker_errors_total
// src_own_background_worker_handlers
func (sourcegraphOwn) NewOwnRepoIndexerWorkerGroup(containerName string) monitoring.Group {
	return Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeInsights, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       ownNamespace,
			DescriptionRoot: "repo indexer worker queue",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "own_background_worker",
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

// src_own_background_worker_resets_total
// src_own_background_worker_reset_failures_total
// src_own_background_worker_reset_errors_total
func (sourcegraphOwn) NewOwnRepoIndexerResetterGroup(containerName string) monitoring.Group {

	return WorkerutilResetter.NewGroup(containerName, monitoring.ObservableOwnerCodeInsights, ResetterGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       ownNamespace,
			DescriptionRoot: "own repo indexer record resetter",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "own_background_worker",
				MetricDescriptionRoot: "own repo indexer queue",
			},
		},

		RecordResets:        NoAlertsOption("none"),
		RecordResetFailures: NoAlertsOption("none"),
		Errors:              NoAlertsOption("none"),
	})
}

// src_own_background_index_scheduler_total{op=”}
func (sourcegraphOwn) NewOwnRepoIndexerSchedulerGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       ownNamespace,
			DescriptionRoot: "index job scheduler",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "own_background_index_scheduler",
				MetricDescriptionRoot: "own index job scheduler",
				RangeWindow:           model.Duration(time.Minute) * 10,
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}
