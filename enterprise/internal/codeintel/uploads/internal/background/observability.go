package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/honey"
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
	numReconcileScansFromFrontend      prometheus.Counter
	numReconcileDeletesFromFrontend    prometheus.Counter
	numReconcileScansFromCodeIntelDB   prometheus.Counter
	numReconcileDeletesFromCodeIntelDB prometheus.Counter
}

func newOperations(observationCtx *observation.Context) *operations {
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
		numReconcileScansFromFrontend:      numReconcileScansFromFrontend,
		numReconcileDeletesFromFrontend:    numReconcileDeletesFromFrontend,
		numReconcileScansFromCodeIntelDB:   numReconcileScansFromCodeIntelDB,
		numReconcileDeletesFromCodeIntelDB: numReconcileDeletesFromCodeIntelDB,
	}
}
