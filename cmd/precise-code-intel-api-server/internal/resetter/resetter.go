package resetter

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type UploadResetter struct {
	db            db.DB
	resetInterval time.Duration
}

type UploadResetterOpts struct {
	DB            db.DB
	ResetInterval time.Duration
}

func NewUploadResetter(opts UploadResetterOpts) *UploadResetter {
	return &UploadResetter{
		db:            opts.DB,
		resetInterval: opts.ResetInterval,
	}
}

// Run periodically moves all uploads that have been in the PROCESSING state for a
// while back to QUEUED. For each updated upload record, the conversion process that
// was responsible for handling the upload did not hold a row lock, indicating that
// it has died.
func (ur *UploadResetter) Run() {
	for {
		ids, err := ur.db.ResetStalled(context.Background(), time.Now())
		if err != nil {
			log15.Error("Failed to reset stalled uploads", "error", err)
		}
		for _, id := range ids {
			log15.Debug("Reset stalled upload", "uploadID", id)
		}

		time.Sleep(ur.resetInterval)
	}
}
