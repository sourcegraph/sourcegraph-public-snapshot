package janitor

import (
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

type Janitor struct {
	bundleDir          string
	desiredPercentFree int
	janitorInterval    time.Duration
	maxUploadAge       time.Duration
	metrics            JanitorMetrics
	done               chan struct{}
	once               sync.Once
}

func New(
	bundleDir string,
	desiredPercentFree int,
	janitorInterval time.Duration,
	maxUploadAge time.Duration,
	metrics JanitorMetrics,
) *Janitor {
	return &Janitor{
		bundleDir:          bundleDir,
		desiredPercentFree: desiredPercentFree,
		janitorInterval:    janitorInterval,
		maxUploadAge:       maxUploadAge,
		metrics:            metrics,
		done:               make(chan struct{}),
	}
}

// step performs a best-effort cleanup. See the following methods for more specifics.
// Run periodically performs a best-effort cleanup process. See the following methods
// for more specifics: removeOldUploadFiles, removeOrphanedBundleFiles, and freeSpace.
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
	// TODO(efritz) - should also remove orphaned upload files

	if err := j.removeOldUploadFiles(); err != nil {
		return errors.Wrap(err, "janitor.removeOldUploadFiles")
	}

	if err := j.removeOrphanedBundleFiles(defaultStatesFn); err != nil {
		return errors.Wrap(err, "janitor.removeOrphanedBundleFiles")
	}

	if err := j.freeSpace(defaultPruneFn); err != nil {
		return errors.Wrap(err, "janitor.freeSpace")
	}

	return nil
}
