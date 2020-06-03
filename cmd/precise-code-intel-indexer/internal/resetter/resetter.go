package resetter

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type IndexResetter struct {
	DB            db.DB
	ResetInterval time.Duration
	Metrics       ResetterMetrics
}

// Run periodically moves all indexes that have been in the PROCESSING state for a
// while back to QUEUED. For each updated index record, the indexer process that
// was responsible for handling the index did not hold a row lock, indicating that
// it has died.
func (ur *IndexResetter) Run() {
	for {
		ids, err := ur.DB.ResetStalledIndexes(context.Background(), time.Now())
		if err != nil {
			ur.Metrics.Errors.Inc()
			log15.Error("Failed to reset stalled indexes", "error", err)
		}
		for _, id := range ids {
			log15.Debug("Reset stalled index", "indexID", id)
		}

		ur.Metrics.Count.Add(float64(len(ids)))
		time.Sleep(ur.ResetInterval)
	}
}
