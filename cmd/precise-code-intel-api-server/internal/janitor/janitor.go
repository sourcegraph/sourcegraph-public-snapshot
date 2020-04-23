package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/db"
)

type Janitor struct {
	db              db.DB
	janitorInterval time.Duration
}

type JanitorOpts struct {
	DB              db.DB
	JanitorInterval time.Duration
}

func NewJanitor(opts JanitorOpts) *Janitor {
	return &Janitor{
		db:              opts.DB,
		janitorInterval: opts.JanitorInterval,
	}
}

func (j *Janitor) Start() {
	for {
		if err := j.step(); err != nil {
			log15.Error("Failed to run janitor process", "error", err)
		}

		time.Sleep(j.janitorInterval)
	}
}

// Run performs a best-effort cleanup. See the following methods for more specifics.
//   - resetStalled
func (j *Janitor) step() error {
	cleanupFns := []func() error{
		j.resetStalled,
	}

	for _, fn := range cleanupFns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// resetStalled moves all uploads that have been in the PROCESSING state for a while back
// to QUEUED. For each updated upload record, the conversion process that was responsible
// for handling the upload did not hold a row lock, indicating that it has died.
func (j *Janitor) resetStalled() error {
	ids, err := j.db.ResetStalled(context.Background(), time.Now())
	if err != nil {
		return err
	}

	for _, id := range ids {
		log15.Debug("Reset stalled upload", "uploadID", id)
	}

	return nil
}
