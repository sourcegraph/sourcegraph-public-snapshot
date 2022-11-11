package background

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	codeinteltypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewUploadProcessorWorker(
	uploadSvc UploadService,
	workerStore store.Store,
	uploadStore uploadstore.Store,
	workerConcurrency int,
	workerBudget int64,
	workerPollInterval time.Duration,
	maximumRuntimePerJob time.Duration,
	observationContext *observation.Context,
) *workerutil.Worker {
	rootContext := actor.WithInternalActor(context.Background())

	handler := NewUploadProcessorHandler(
		uploadSvc,
		uploadStore,
		workerConcurrency,
		workerBudget,
		observationContext,
	)

	metrics := workerutil.NewMetrics(observationContext, "codeintel_upload_processor", workerutil.WithSampler(func(job workerutil.Record) bool { return true }))

	return dbworker.NewWorker(rootContext, workerStore, handler, workerutil.WorkerOptions{
		Name:                 "precise_code_intel_upload_worker",
		NumHandlers:          workerConcurrency,
		Interval:             workerPollInterval,
		HeartbeatInterval:    time.Second,
		Metrics:              metrics,
		MaximumRuntimePerJob: maximumRuntimePerJob,
	})
}

type handler struct {
	uploadsSvc      UploadService
	uploadStore     uploadstore.Store
	handleOp        *observation.Operation
	budgetRemaining int64
	enableBudget    bool
	uploadSizeGuage prometheus.Gauge
}

var (
	_ workerutil.Handler        = &handler{}
	_ workerutil.WithPreDequeue = &handler{}
	_ workerutil.WithHooks      = &handler{}
)

func NewUploadProcessorHandler(
	uploadSvc UploadService,
	uploadStore uploadstore.Store,
	numProcessorRoutines int,
	budgetMax int64,
	observationContext *observation.Context,
) workerutil.Handler {
	operations := newOperations(observationContext)

	return &handler{
		uploadsSvc:      uploadSvc,
		uploadStore:     uploadStore,
		handleOp:        operations.uploadProcessor,
		budgetRemaining: budgetMax,
		enableBudget:    budgetMax > 0,
		uploadSizeGuage: operations.uploadSizeGuage,
	}
}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
	upload, ok := record.(codeinteltypes.Upload)
	if !ok {
		return errors.Newf("unexpected record type %T", record)
	}

	var requeued bool

	ctx, otLogger, endObservation := h.handleOp.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{
			LogFields: append(
				createLogFields(upload),
				otlog.Bool("requeued", requeued),
			),
		})
	}()

	requeued, err = h.uploadsSvc.HandleRawUpload(ctx, logger, upload, h.uploadStore, otLogger)

	return err
}

func (h *handler) PreDequeue(ctx context.Context, logger log.Logger) (bool, any, error) {
	if !h.enableBudget {
		return true, nil, nil
	}

	budgetRemaining := atomic.LoadInt64(&h.budgetRemaining)
	if budgetRemaining <= 0 {
		return false, nil, nil
	}

	return true, []*sqlf.Query{sqlf.Sprintf("(upload_size IS NULL OR upload_size <= %s)", budgetRemaining)}, nil
}

func (h *handler) PreHandle(ctx context.Context, logger log.Logger, record workerutil.Record) {
	upload, ok := record.(codeinteltypes.Upload)
	if !ok {
		return
	}

	uncompressedSize := h.getUploadSize(upload.UncompressedSize)
	h.uploadSizeGuage.Add(float64(uncompressedSize))

	gzipSize := h.getUploadSize(upload.UploadSize)
	atomic.AddInt64(&h.budgetRemaining, -gzipSize)
}

func (h *handler) PostHandle(ctx context.Context, logger log.Logger, record workerutil.Record) {
	upload, ok := record.(codeinteltypes.Upload)
	if !ok {
		return
	}

	uncompressedSize := h.getUploadSize(upload.UncompressedSize)
	h.uploadSizeGuage.Sub(float64(uncompressedSize))

	gzipSize := h.getUploadSize(upload.UploadSize)
	atomic.AddInt64(&h.budgetRemaining, +gzipSize)
}

func (h *handler) getUploadSize(field *int64) int64 {
	if field != nil {
		return *field
	}

	return 0
}

func createLogFields(upload codeinteltypes.Upload) []otlog.Field {
	fields := []otlog.Field{
		otlog.Int("uploadID", upload.ID),
		otlog.Int("repositoryID", upload.RepositoryID),
		otlog.String("commit", upload.Commit),
		otlog.String("root", upload.Root),
		otlog.String("indexer", upload.Indexer),
		otlog.Int("queueDuration", int(time.Since(upload.UploadedAt))),
	}

	if upload.UploadSize != nil {
		fields = append(fields, otlog.Int64("uploadSize", *upload.UploadSize))
	}

	return fields
}
