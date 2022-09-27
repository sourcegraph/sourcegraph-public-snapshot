package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewHandler(
	dbStore DBStore,
	workerStore dbworkerstore.Store,
	lsifStore LSIFStore,
	uploadStore uploadstore.Store,
	gitserverClient GitserverClient,
	numProcessorRoutines int,
	budgetMax int64,
) workerutil.Handler {
	observationContext := observation.Context{
		Tracer: &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		HoneyDataset: &honey.Dataset{
			Name: "codeintel-worker",
		},
		Registerer: prometheus.DefaultRegisterer,
	}

	op := observationContext.Operation(observation.Op{
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

	return &handler{
		dbStore:           dbStore,
		workerStore:       workerStore,
		lsifStore:         lsifStore,
		uploadStore:       uploadStore,
		gitserverClient:   gitserverClient,
		handleOp:          op,
		budgetRemaining:   budgetMax,
		enableBudget:      budgetMax > 0,
		uncompressedSizes: make(map[int]uint64, numProcessorRoutines),
		uploadSizeGuage:   uploadSizeGuage,
	}
}
