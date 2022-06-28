package worker

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"

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
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})
	observationContext := observation.Context{
		Tracer: &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		HoneyDataset: &honey.Dataset{
			Name: "codeintel-worker",
		},
	}

	op := observationContext.Operation(observation.Op{
		Name: "codeintel.uploadHandler",
		ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			return observation.EmitForTraces | observation.EmitForHoney
		},
	})

	handler := &handler{
		dbStore:         dbStore,
		workerStore:     workerStore,
		lsifStore:       lsifStore,
		uploadStore:     uploadStore,
		gitserverClient: gitserverClient,
		enableBudget:    budgetMax > 0,
		budgetRemaining: budgetMax,
		handleOp:        op,
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
