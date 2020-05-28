package janitor

import (
	"os"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type Janitor struct {
	db                 db.DB
	bundleDir          string
	desiredPercentFree int
	janitorInterval    time.Duration
	maxUploadAge       time.Duration
	maxUploadPartAge   time.Duration
	maxDatabasePartAge time.Duration
	metrics            JanitorMetrics
	done               chan struct{}
	once               sync.Once
}

func New(
	db db.DB,
	bundleDir string,
	desiredPercentFree int,
	janitorInterval time.Duration,
	maxUploadAge time.Duration,
	maxUploadPartAge time.Duration,
	maxDatabasePartAge time.Duration,
	metrics JanitorMetrics,
) *Janitor {
	return &Janitor{
		db:                 db,
		bundleDir:          bundleDir,
		desiredPercentFree: desiredPercentFree,
		janitorInterval:    janitorInterval,
		maxUploadAge:       maxUploadAge,
		maxUploadPartAge:   maxUploadPartAge,
		maxDatabasePartAge: maxDatabasePartAge,
		metrics:            metrics,
		done:               make(chan struct{}),
	}
}

// Run periodically performs a best-effort cleanup process.
func (j *Janitor) Run() {
	for {
		if err := j.run(); err != nil {
			j.metrics.Errors.Inc()
			log15.Error("Failed to run janitor process", "err", err)
		}

		select {
		case <-time.After(j.janitorInterval):
		case <-j.done:
			return
		}
	}
}

func (j *Janitor) Stop() {
	j.once.Do(func() {
		close(j.done)
	})
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

	if err := j.freeSpace(); err != nil {
		return errors.Wrap(err, "janitor.freeSpace")
	}

	if err := j.removeProcessedUploadsWithoutBundleFile(); err != nil {
		return errors.Wrap(err, "janitor.removeProcessedUploadsWithoutBundle")
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
