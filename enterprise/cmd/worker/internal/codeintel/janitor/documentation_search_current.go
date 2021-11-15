package janitor

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type documentationSearchCurrentJanitor struct {
	lsifStore                 LSIFStore
	minimumTimeSinceLastCheck time.Duration
	batchSize                 int
	metrics                   *metrics
}

var _ goroutine.Handler = &documentationSearchCurrentJanitor{}
var _ goroutine.ErrorHandler = &documentationSearchCurrentJanitor{}

// NewDocumentationSearchCurrentJanitor returns a background routine that periodically removes any
// residual lsif_data_docs_search records that are not the most recent for its key, as identified
// by the recent dump_id in the associated lsif_data_docs_search_current table.
func NewDocumentationSearchCurrentJanitor(
	lsifStore LSIFStore,
	minimumTimeSinceLastCheck time.Duration,
	batchSize int,
	interval time.Duration,
	metrics *metrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &documentationSearchCurrentJanitor{
		lsifStore:                 lsifStore,
		minimumTimeSinceLastCheck: minimumTimeSinceLastCheck,
		batchSize:                 batchSize,
		metrics:                   metrics,
	})
}

func (j *documentationSearchCurrentJanitor) Handle(ctx context.Context) (err error) {
	tx, err := j.lsifStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	publicCount, publicErr := tx.DeleteOldPublicSearchRecords(ctx, j.minimumTimeSinceLastCheck, j.batchSize)
	privateCount, privateErr := tx.DeleteOldPrivateSearchRecords(ctx, j.minimumTimeSinceLastCheck, j.batchSize)
	j.metrics.numDocumentSearchRecordsRemoved.Add(float64(publicCount + privateCount))

	if publicErr != nil {
		if privateErr != nil {
			return multierror.Append(publicErr, privateErr)
		}

		return publicErr
	}

	return privateErr
}

func (j *documentationSearchCurrentJanitor) HandleError(err error) {
	j.metrics.numErrors.Inc()
	log15.Error("Failed to remove non-current documentation search records", "error", err)
}
