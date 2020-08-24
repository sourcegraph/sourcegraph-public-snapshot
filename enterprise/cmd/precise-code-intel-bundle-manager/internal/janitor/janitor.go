package janitor

import (
	"context"
	"os"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type Janitor struct {
	store              store.Store
	bundleDir          string
	desiredPercentFree int
	maxUploadAge       time.Duration
	maxUploadPartAge   time.Duration
	maxDatabasePartAge time.Duration
	metrics            JanitorMetrics
}

var _ goroutine.Handler = &Janitor{}

func New(
	store store.Store,
	bundleDir string,
	desiredPercentFree int,
	janitorInterval time.Duration,
	maxUploadAge time.Duration,
	maxUploadPartAge time.Duration,
	maxDatabasePartAge time.Duration,
	metrics JanitorMetrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), janitorInterval, &Janitor{
		store:              store,
		bundleDir:          bundleDir,
		desiredPercentFree: desiredPercentFree,
		maxUploadAge:       maxUploadAge,
		maxUploadPartAge:   maxUploadPartAge,
		maxDatabasePartAge: maxDatabasePartAge,
		metrics:            metrics,
	})
}

// Handle performs a best-effort cleanup process.
func (j *Janitor) Handle(ctx context.Context) error {
	if err := j.removeOldUploadFiles(ctx); err != nil {
		return errors.Wrap(err, "janitor.removeOldUploadFiles")
	}

	if err := j.removeOldUploadPartFiles(ctx); err != nil {
		return errors.Wrap(err, "janitor.removeOldUploadPartFiles")
	}

	if err := j.removeOldDatabasePartFiles(ctx); err != nil {
		return errors.Wrap(err, "janitor.removeOldDatabasePartFiles")
	}

	if err := j.removeOrphanedUploadFiles(ctx); err != nil {
		return errors.Wrap(err, "janitor.removeOrphanedUploadFiles")
	}

	if err := j.removeOrphanedBundleFiles(ctx); err != nil {
		return errors.Wrap(err, "janitor.removeOrphanedBundleFiles")
	}

	if err := j.removeRecordsForDeletedRepositories(ctx); err != nil {
		return errors.Wrap(err, "janitor.removeRecordsForDeletedRepositories")
	}

	if err := j.removeCompletedRecordsWithoutBundleFile(ctx); err != nil {
		return errors.Wrap(err, "janitor.removeCompletedRecordsWithoutBundleFile")
	}

	if err := j.removeOldUploadingRecords(ctx); err != nil {
		return errors.Wrap(err, "janitor.removeOldUploadingRecords")
	}

	if err := j.freeSpace(ctx); err != nil {
		return errors.Wrap(err, "janitor.freeSpace")
	}

	return nil
}

func (j *Janitor) HandleError(err error) {
	j.metrics.Errors.Inc()
	log15.Error("Failed to run janitor process", "err", err)
}

// remove unlinks the file or directory at the given path. Returns a boolean indicating
// success. If unsuccessful, the path and error will be logged and the error counter will
// be incremented.
func (j *Janitor) remove(path string) bool {
	if err := os.RemoveAll(path); err != nil {
		j.metrics.Errors.Inc()
		log15.Error("Failed to remove path", "path", path, "err", err)
		return false
	}

	return true
}
