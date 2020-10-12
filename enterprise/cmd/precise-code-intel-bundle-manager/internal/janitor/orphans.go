package janitor

import (
	"context"

	"github.com/inconshreveable/log15"
)

// GetStateBatchSize is the maximum number of bundle ids to request the state of from the
// database at once.
const GetStateBatchSize = 100

// removeOrphanedData removes data from the codeintel database that does not have a
// corresponding upload record in the frontend database.
func (j *Janitor) removeOrphanedData(ctx context.Context) error {
	offset := 0
	for {
		dumpIDs, err := j.lsifStore.DumpIDs(ctx, GetStateBatchSize, offset)
		if err != nil {
			return err
		}

		states, err := j.store.GetStates(ctx, dumpIDs)
		if err != nil {
			return err
		}

		for _, dumpID := range dumpIDs {
			if _, ok := states[dumpID]; !ok {
				if err := j.lsifStore.Clear(ctx, dumpID); err != nil {
					log15.Error("Failed to remove data for dump", "dump_id", dumpID)
				}
			}
		}

		if len(dumpIDs) < GetStateBatchSize {
			break
		}

		offset += GetStateBatchSize
	}

	return nil
}
