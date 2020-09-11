package janitor

import (
	"time"

	"github.com/inconshreveable/log15"
)

// removeRecordsForDeletedRepositories removes all index records for deleted repositories.
func (j *Janitor) removeRecordsForDeletedRepositories() error {
	counts, err := j.store.DeleteIndexesWithoutRepository(j.ctx, time.Now())
	if err != nil {
		return err
	}

	for repoID, count := range counts {
		log15.Debug("Removed index records for a deleted repository", "repository_id", repoID, "count", count)
		j.metrics.IndexRecordsRemoved.Add(float64(count))
	}

	return nil
}
