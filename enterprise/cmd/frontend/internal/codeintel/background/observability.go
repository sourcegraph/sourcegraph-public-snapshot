package background

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	commitUpdate            *observation.Operation
	numUploadRecordsRemoved prometheus.Counter
	numIndexRecordsRemoved  prometheus.Counter
	numUploadsPurged        prometheus.Counter
	numUploadResets         prometheus.Counter
	numUploadResetFailures  prometheus.Counter
	numIndexResets          prometheus.Counter
	numIndexResetFailures   prometheus.Counter
	numErrors               prometheus.Counter
}

var NewOperations = newOperations

func newOperations(dbStore DBStore, observationContext *observation.Context) *operations {
	//
	// Operations

	commitUpdate := observationContext.Operation(observation.Op{
		Name: "codeintel.commitUpdater",
		Metrics: metrics.NewOperationMetrics(
			observationContext.Registerer,
			"codeintel_commit_graph_updater",
			metrics.WithCountHelp("Total number of method invocations."),
		),
	})

	//
	// Counters

	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationContext.Registerer.MustRegister(counter)
		return counter
	}

	numUploadRecordsRemoved := counter(
		"src_codeintel_background_upload_records_removed_total",
		"The number of codeintel upload records removed.",
	)
	numIndexRecordsRemoved := counter(
		"src_codeintel_background_index_records_removed_total",
		"The number of codeintel index records removed.",
	)
	numUploadsPurged := counter(
		"src_codeintel_background_uploads_purged_total",
		"The number of uploads for which records in the codeintel db were removed.",
	)
	numUploadResets := counter(
		"src_codeintel_background_upload_resets_total",
		"The number of upload record resets.",
	)
	numUploadResetFailures := counter(
		"src_codeintel_background_upload_reset_failures_total",
		"The number of upload reset failures.",
	)
	numIndexResets := counter(
		"src_codeintel_background_index_resets_total",
		"The number of index records reset.",
	)
	numIndexResetFailures := counter(
		"src_codeintel_background_index_reset_failures_total",
		"The number of index reset failures.",
	)
	numErrors := counter(
		"src_codeintel_background_errors_total",
		"The number of errors that occur during a codeintel background job.",
	)

	//
	// Periodic metrics

	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_dirty_repositories_total",
		Help: "Total number of repositories with stale commit graphs.",
	}, func() float64 {
		dirtyRepositories, err := dbStore.DirtyRepositories(context.Background())
		if err != nil {
			log15.Error("Failed to determine number of dirty repositories", "err", err)
		}

		return float64(len(dirtyRepositories))
	}))

	//
	//

	return &operations{
		commitUpdate:            commitUpdate,
		numUploadRecordsRemoved: numUploadRecordsRemoved,
		numIndexRecordsRemoved:  numIndexRecordsRemoved,
		numUploadsPurged:        numUploadsPurged,
		numUploadResets:         numUploadResets,
		numUploadResetFailures:  numUploadResetFailures,
		numIndexResets:          numIndexResets,
		numIndexResetFailures:   numIndexResetFailures,
		numErrors:               numErrors,
	}
}
