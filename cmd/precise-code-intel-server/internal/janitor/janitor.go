package janitor

import (
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type Janitor struct {
	db                  db.DB
	bundleManagerClient client.BundleManagerClient
	janitorInterval     time.Duration
	metrics             JanitorMetrics
	done                chan struct{}
	once                sync.Once
}

func New(
	db db.DB,
	bundleManagerClient client.BundleManagerClient,
	janitorInterval time.Duration,
	metrics JanitorMetrics,
) *Janitor {
	return &Janitor{
		db:                  db,
		bundleManagerClient: bundleManagerClient,
		janitorInterval:     janitorInterval,
		metrics:             metrics,
		done:                make(chan struct{}),
	}
}

// Run periodically performs a best-effort cleanup process. See the following methods
// for more specifics: removeProcessedUploadsWithoutBundleFile.
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

	if err := j.removeProcessedUploadsWithoutBundleFile(); err != nil {
		return errors.Wrap(err, "janitor.removeProcessedUploadsWithoutBundle")
	}

	return nil
}
