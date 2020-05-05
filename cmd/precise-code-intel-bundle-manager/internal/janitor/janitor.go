package janitor

import (
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

type Janitor struct {
	bundleDir          string
	desiredPercentFree int
	janitorInterval    time.Duration
	maxUploadAge       time.Duration
}

type JanitorOpts struct {
	BundleDir          string
	DesiredPercentFree int
	JanitorInterval    time.Duration
	MaxUploadAge       time.Duration
}

func NewJanitor(opts JanitorOpts) *Janitor {
	return &Janitor{
		bundleDir:          opts.BundleDir,
		desiredPercentFree: opts.DesiredPercentFree,
		janitorInterval:    opts.JanitorInterval,
		maxUploadAge:       opts.MaxUploadAge,
	}
}

// step performs a best-effort cleanup. See the following methods for more specifics.
// Run periodically performs a best-effort cleanup process. See the following methods
// for more specifics: cleanOldUploads, removeDeadDumps, and freeSpace.
func (j *Janitor) Run() {
	for {
		if err := j.run(); err != nil {
			log15.Error("Failed to run janitor process", "err", err)
		}

		time.Sleep(j.janitorInterval)
	}
}

func (j *Janitor) run() error {
	if err := j.cleanOldUploads(); err != nil {
		return errors.Wrap(err, "janitor.cleanOldUploads")
	}

	if err := j.removeDeadDumps(defaultStatesFn); err != nil {
		return errors.Wrap(err, "janitor.removeDeadDumps")
	}

	if err := j.freeSpace(defaultPruneFn); err != nil {
		return errors.Wrap(err, "janitor.freeSpace")
	}

	return nil
}
