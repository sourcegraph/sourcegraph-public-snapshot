package janitor

import (
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

type Janitor struct {
	BundleDir          string
	DesiredPercentFree int
	JanitorInterval    time.Duration
	MaxUploadAge       time.Duration
	Metrics            JanitorMetrics
}

// step performs a best-effort cleanup. See the following methods for more specifics.
// Run periodically performs a best-effort cleanup process. See the following methods
// for more specifics: cleanOldUploads, removeOrphanedDumps, and freeSpace.
func (j *Janitor) Run() {
	for {
		if err := j.run(); err != nil {
			j.Metrics.Errors.Inc()
			log15.Error("Failed to run janitor process", "err", err)
		}

		time.Sleep(j.JanitorInterval)
	}
}

func (j *Janitor) run() error {
	if err := j.cleanOldUploads(); err != nil {
		return errors.Wrap(err, "janitor.cleanOldUploads")
	}

	if err := j.removeOrphanedDumps(defaultStatesFn); err != nil {
		return errors.Wrap(err, "janitor.removeOrphanedDumps")
	}

	if err := j.freeSpace(defaultPruneFn); err != nil {
		return errors.Wrap(err, "janitor.freeSpace")
	}

	return nil
}
