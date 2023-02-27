package background

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type workerOperations struct {
	uploadProcessor *observation.Operation
	uploadSizeGauge prometheus.Gauge
}

func newWorkerOperations(observationCtx *observation.Context) *workerOperations {
	honeyObservationCtx := *observationCtx
	honeyObservationCtx.HoneyDataset = &honey.Dataset{Name: "codeintel-worker"}
	uploadProcessor := honeyObservationCtx.Operation(observation.Op{
		Name: "codeintel.uploadHandler",
		ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			return observation.EmitForTraces | observation.EmitForHoney
		},
	})

	uploadSizeGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "src_codeintel_upload_processor_upload_size",
		Help: "The combined size of uploads being processed at this instant by this worker.",
	})
	observationCtx.Registerer.MustRegister(uploadSizeGauge)

	return &workerOperations{
		uploadProcessor: uploadProcessor,
		uploadSizeGauge: uploadSizeGauge,
	}
}

type operations struct {
	updateUploadsVisibleToCommits *observation.Operation

	numReconcileScansFromFrontend      prometheus.Counter
	numReconcileDeletesFromFrontend    prometheus.Counter
	numReconcileScansFromCodeIntelDB   prometheus.Counter
	numReconcileDeletesFromCodeIntelDB prometheus.Counter
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_uploads_background",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.background.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numReconcileScansFromFrontend := counter(
		"src_codeintel_uploads_frontend_reconciliation_records_scanned_total",
		"The number of upload records read from the frontend db for reconciliation.",
	)
	numReconcileDeletesFromFrontend := counter(
		"src_codeintel_uploads_frontend_reconciliation_records_deleted_total",
		"The number of abandoned uploads deleted from the frontend db.",
	)
	numReconcileScansFromCodeIntelDB := counter(
		"src_codeintel_uploads_codeinteldb_reconciliation_records_scanned_total",
		"The number of upload records read from the codeintel-db for reconciliation.",
	)
	numReconcileDeletesFromCodeIntelDB := counter(
		"src_codeintel_uploads_codeinteldb_reconciliation_records_deleted_total",
		"The number of abandoned uploads deleted from the codeintel-db.",
	)

	return &operations{
		updateUploadsVisibleToCommits: op("UpdateUploadsVisibleToCommits"),

		numReconcileScansFromFrontend:      numReconcileScansFromFrontend,
		numReconcileDeletesFromFrontend:    numReconcileDeletesFromFrontend,
		numReconcileScansFromCodeIntelDB:   numReconcileScansFromCodeIntelDB,
		numReconcileDeletesFromCodeIntelDB: numReconcileDeletesFromCodeIntelDB,
	}
}

type rankMappingOperations struct {
	numReferenceRecordsProcessed prometheus.Counter
	numInputsInserted            prometheus.Counter
}

func newRankMappingOperations(observationCtx *observation.Context) *rankMappingOperations {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numReferenceRecordsProcessed := counter(
		"src_codeintel_ranking_reference_records_processed_total",
		"The number of reference rows processed.",
	)
	numInputsInserted := counter(
		"src_codeintel_ranking_inputs_inserted_total",
		"The number of input rows inserted.",
	)

	return &rankMappingOperations{
		numReferenceRecordsProcessed: numReferenceRecordsProcessed,
		numInputsInserted:            numInputsInserted,
	}
}

type rankReducingOperations struct {
	numPathCountsInputsRowsProcessed prometheus.Counter
	numPathRanksInserted             prometheus.Counter
}

func newRankReducingOperations(observationCtx *observation.Context) *rankReducingOperations {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numPathCountInputsRowsProcessed := counter(
		"src_codeintel_ranking_path_count_inputs_rows_processed_total",
		"The number of input rows processed.",
	)
	numPathRanksInserted := counter(
		"src_codeintel_ranking_path_ranks_inserted_total",
		"The number of path ranks inserted.",
	)

	return &rankReducingOperations{
		numPathCountsInputsRowsProcessed: numPathCountInputsRowsProcessed,
		numPathRanksInserted:             numPathRanksInserted,
	}
}
