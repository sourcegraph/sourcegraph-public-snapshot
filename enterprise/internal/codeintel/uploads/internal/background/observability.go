package background

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	updateUploadsVisibleToCommits *observation.Operation

	// Worker metrics
	uploadProcessor *observation.Operation
	uploadSizeGuage prometheus.Gauge

	numReconcileScansFromFrontend      prometheus.Counter
	numReconcileDeletesFromFrontend    prometheus.Counter
	numReconcileScansFromCodeIntelDB   prometheus.Counter
	numReconcileDeletesFromCodeIntelDB prometheus.Counter
}

var once = memo.NewMemoizedConstructorWithArg(func(observationContext *observation.Context) (*metrics.REDMetrics, error) {
	return metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_uploads_background",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	), nil
})

func newOperations(observationContext *observation.Context) *operations {
	m, _ := once.Init(observationContext)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
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

		observationContext.Registerer.MustRegister(counter)
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

	honeyObservationContext := *observationContext
	honeyObservationContext.HoneyDataset = &honey.Dataset{Name: "codeintel-worker"}
	uploadProcessor := honeyObservationContext.Operation(observation.Op{
		Name: "codeintel.uploadHandler",
		ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			return observation.EmitForTraces | observation.EmitForHoney
		},
	})

	uploadSizeGuage := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "src_codeintel_upload_processor_upload_size",
		Help: "The combined size of uploads being processed at this instant by this worker.",
	})
	observationContext.Registerer.MustRegister(uploadSizeGuage)

	return &operations{
		updateUploadsVisibleToCommits: op("UpdateUploadsVisibleToCommits"),

		// Worker metrics
		uploadProcessor: uploadProcessor,
		uploadSizeGuage: uploadSizeGuage,

		numReconcileScansFromFrontend:      numReconcileScansFromFrontend,
		numReconcileDeletesFromFrontend:    numReconcileDeletesFromFrontend,
		numReconcileScansFromCodeIntelDB:   numReconcileScansFromCodeIntelDB,
		numReconcileDeletesFromCodeIntelDB: numReconcileDeletesFromCodeIntelDB,
	}
}
