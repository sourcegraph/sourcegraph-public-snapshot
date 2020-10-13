package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
)

// removeExpiredData removes upload records that have exceeded a threshold age and
// is not an index visible from the head of the default branch.
func (j *Janitor) removeExpiredData(ctx context.Context) error {
	count, err := j.store.SoftDeleteOldDumps(ctx, j.maxDataAge, time.Now())
	if err != nil {
		return err
	}

	if count > 0 {
		log15.Debug("Expired upload records not visible to the tip of the default branch", "count", count)
	}

	return nil
}
