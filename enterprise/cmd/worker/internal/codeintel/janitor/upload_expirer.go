package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type uploadExpirer struct {
	dbStore                DBStore
	gitserverClient        GitserverClient
	metrics                *metrics
	repositoryProcessDelay time.Duration
	repositoryBatchSize    int
	uploadProcessDelay     time.Duration
	uploadBatchSize        int
}

var _ goroutine.Handler = &uploadExpirer{}
var _ goroutine.ErrorHandler = &uploadExpirer{}

// NewUploadExpirer returns a background routine that periodically compares the age of upload records against
// the age of uploads protected by global and repository specific data retention policies.
//
// Uploads that are older than the protected retention age are marked as expired. Expired records with no
// dependents will be removed by the expiredUploadDeleter.
func NewUploadExpirer(
	dbStore DBStore,
	gitserverClient GitserverClient,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	uploadProcessDelay time.Duration,
	uploadBatchSize int,
	interval time.Duration,
	metrics *metrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &uploadExpirer{
		dbStore:                dbStore,
		gitserverClient:        gitserverClient,
		metrics:                metrics,
		repositoryProcessDelay: repositoryProcessDelay,
		repositoryBatchSize:    repositoryBatchSize,
		uploadProcessDelay:     uploadProcessDelay,
		uploadBatchSize:        uploadBatchSize,
	})
}

func (e *uploadExpirer) Handle(ctx context.Context) (err error) {
	// TODO - implement
	return nil
}

func (e *uploadExpirer) HandleError(err error) {
	e.metrics.numErrors.Inc()
	log15.Error("Failed to expire old codeintel records", "error", err)
}
