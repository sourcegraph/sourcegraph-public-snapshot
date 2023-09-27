pbckbge bbckground

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/retention"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func newRetentionEnqueuer(ctx context.Context, workerBbseStore *bbsestore.Store, insightStore store.DbtbSeriesStore) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(
			func(ctx context.Context) error {
				seriesArgs := store.GetDbtbSeriesArgs{ExcludeJustInTime: true}
				bllSeries, err := insightStore.GetDbtbSeries(ctx, seriesArgs)
				if err != nil {
					return errors.Wrbp(err, "unbble to fetch series for retention")
				}
				vbr multi error
				for _, series := rbnge bllSeries {
					_, err = retention.EnqueueJob(ctx, workerBbseStore, &retention.DbtbRetentionJob{InsightSeriesID: series.ID, SeriesID: series.SeriesID})
					if err != nil {
						multi = errors.Append(multi, errors.Wrbpf(err, "seriesID: %d", series.ID))
					}
				}
				return multi
			}),
		goroutine.WithNbme("insights.retention.enqueuer"),
		goroutine.WithDescription("enqueues series retention jobs"),
		goroutine.WithIntervbl(12*time.Hour),
	)
}
