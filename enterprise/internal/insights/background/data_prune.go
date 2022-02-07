package background

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// NewInsightsDataPrunerJob will periodically delete recorded data series that have been marked `deleted`.
func NewInsightsDataPrunerJob(ctx context.Context, postgres dbutil.DB, insightsdb dbutil.DB) goroutine.BackgroundRoutine {
	interval := time.Minute * 60

	pruneOlderThan := time.Hour * -1

	insightStore := store.NewInsightStore(insightsdb)
	timeseriesStore := store.New(insightsdb, store.NewInsightPermissionStore(postgres))

	return goroutine.NewPeriodicGoroutine(ctx, interval,
		goroutine.NewHandlerWithErrorMessage("insights_data_prune", func(ctx context.Context) (err error) {
			// select the series that need to be pruned. These will be the series that are currently flagged
			// as "soft deleted" for a given amount of time.
			seriesIds, err := insightStore.GetSoftDeletedSeries(ctx, time.Now().Add(pruneOlderThan))
			if err != nil {
				return errors.Wrap(err, "GetSoftDeletedSeries")
			}

			for _, id := range seriesIds {
				log15.Info("pruning insight series", "seriesId", id)

				// We will always delete the series definition last, such that any possible partial state
				// the series definition will always be referencable and therefore can be re-attempted.
				tx, err := timeseriesStore.Transact(ctx)
				if err != nil {
					return err
				}

				insightStoreTx := insightStore.With(tx)
				err = insightStoreTx.HardDeleteSeries(ctx, id)
				if err != nil {
					return err
				}

				err = tx.Done(err)
			}

			return nil
		}))
}
