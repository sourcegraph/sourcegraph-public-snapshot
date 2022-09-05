package worker

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// UploadHeartbeatInterval is the duration between heartbeat updates to the upload job records.
const UploadHeartbeatInterval = time.Second

func NewWorker(
	dbStore DBStore,
	workerStore dbworkerstore.Store,
	lsifStore LSIFStore,
	uploadStore uploadstore.Store,
	gitserverClient GitserverClient,
	pollInterval time.Duration,
	numProcessorRoutines int,
	budgetMax int64,
	maximumRuntimePerJob time.Duration,
	workerMetrics workerutil.WorkerMetrics,
) *workerutil.Worker {
	rootContext := actor.WithInternalActor(context.Background())
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

	handler := &handler{
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

	return dbworker.NewWorker(rootContext, workerStore, handler, workerutil.WorkerOptions{
		Name:                 "precise_code_intel_upload_worker",
		NumHandlers:          numProcessorRoutines,
		Interval:             pollInterval,
		HeartbeatInterval:    UploadHeartbeatInterval,
		Metrics:              workerMetrics,
		MaximumRuntimePerJob: maximumRuntimePerJob,
	})
}
