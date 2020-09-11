package janitor

import (
	"context"
	"os"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Janitor struct {
	store              store.Store
	bundleDir          string
	desiredPercentFree int
	janitorInterval    time.Duration
	maxUploadAge       time.Duration
	maxUploadPartAge   time.Duration
	maxDatabasePartAge time.Duration
	metrics            JanitorMetrics
	ctx                context.Context
	cancel             func()
	finished           chan (struct{})
}

func New(
	store store.Store,
	bundleDir string,
	desiredPercentFree int,
	janitorInterval time.Duration,
	maxUploadAge time.Duration,
	maxUploadPartAge time.Duration,
	maxDatabasePartAge time.Duration,
	metrics JanitorMetrics,
) *Janitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Janitor{
		store:              store,
		bundleDir:          bundleDir,
		desiredPercentFree: desiredPercentFree,
		janitorInterval:    janitorInterval,
		maxUploadAge:       maxUploadAge,
		maxUploadPartAge:   maxUploadPartAge,
		maxDatabasePartAge: maxDatabasePartAge,
		metrics:            metrics,
		ctx:                ctx,
		cancel:             cancel,
		finished:           make(chan struct{}),
	}
}

// Run periodically performs a best-effort cleanup process.
func (j *Janitor) Run() {
	defer close(j.finished)

	for {
		if err := j.run(); err != nil {
			j.metrics.Errors.Inc()
			log15.Error("Failed to run janitor process", "err", err)
		}

		select {
		case <-time.After(j.janitorInterval):
		case <-j.ctx.Done():
			return
		}
	}
}

func (j *Janitor) Stop() {
	j.cancel()
	<-j.finished
}

func (j *Janitor) run() error {
	// TODO(efritz) - use cancellable context for API calls

	if err := j.removeOldUploadFiles(); err != nil {
		return errors.Wrap(err, "janitor.removeOldUploadFiles")
	}

	if err := j.removeOldUploadPartFiles(); err != nil {
		return errors.Wrap(err, "janitor.removeOldUploadPartFiles")
	}

	if err := j.removeOldDatabasePartFiles(); err != nil {
		return errors.Wrap(err, "janitor.removeOldDatabasePartFiles")
	}

	if err := j.removeOrphanedUploadFiles(); err != nil {
		return errors.Wrap(err, "janitor.removeOrphanedUploadFiles")
	}

	if err := j.removeOrphanedBundleFiles(); err != nil {
		return errors.Wrap(err, "janitor.removeOrphanedBundleFiles")
	}

	if err := j.removeRecordsForDeletedRepositories(); err != nil {
		return errors.Wrap(err, "janitor.removeRecordsForDeletedRepositories")
	}

	if err := j.removeCompletedRecordsWithoutBundleFile(); err != nil {
		return errors.Wrap(err, "janitor.removeCompletedRecordsWithoutBundleFile")
	}

	if err := j.removeOldUploadingRecords(); err != nil {
		return errors.Wrap(err, "janitor.removeOldUploadingRecords")
	}

	if err := j.freeSpace(); err != nil {
		return errors.Wrap(err, "janitor.freeSpace")
	}

	return nil
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
