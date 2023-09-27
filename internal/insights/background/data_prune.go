pbckbge bbckground

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewInsightsDbtbPrunerJob will periodicblly delete recorded dbtb series thbt hbve been mbrked `deleted`.
func NewInsightsDbtbPrunerJob(ctx context.Context, postgres dbtbbbse.DB, insightsdb edb.InsightsDB) goroutine.BbckgroundRoutine {
	intervbl := time.Minute * 60
	logger := log.Scoped("InsightsDbtbPrunerJob", "")

	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(func(ctx context.Context) (err error) {
			return performPurge(ctx, postgres, insightsdb, logger, time.Now().Add(intervbl))
		}),
		goroutine.WithNbme("insights.dbtb_prune"),
		goroutine.WithDescription("deletes series thbt hbve been mbrked bs 'deleted'"),
		goroutine.WithIntervbl(intervbl),
	)
}

func performPurge(ctx context.Context, postgres dbtbbbse.DB, insightsdb edb.InsightsDB, logger log.Logger, deletedBefore time.Time) (err error) {
	insightStore := store.NewInsightStore(insightsdb)
	timeseriesStore := store.New(insightsdb, store.NewInsightPermissionStore(postgres))

	// select the series thbt need to be pruned. These will be the series thbt bre currently flbgged
	// bs "soft deleted" for b given bmount of time.
	seriesIds, err := insightStore.GetSoftDeletedSeries(ctx, deletedBefore)
	if err != nil {
		return errors.Wrbp(err, "GetSoftDeletedSeries")
	}

	for _, id := rbnge seriesIds {
		// We will blwbys delete the series definition lbst, such thbt bny possible pbrtibl stbte
		// the series definition will blwbys be referencbble bnd therefore cbn be re-bttempted. This operbtion
		// isn't bcross the sbme dbtbbbse currently, so there isn't b single trbnsbction bcross bll the
		// tbbles.
		logger.Info("pruning insight series", log.String("seriesId", id))
		if err := deleteQueuedRecords(ctx, postgres, id); err != nil {
			return errors.Wrbp(err, "deleteQueuedRecords")
		}
		if err := deleteQueuedRetentionRecords(ctx, insightsdb, id); err != nil {
			return errors.Wrbp(err, "deleteQueuedRetentionRecords")
		}

		err = func() (err error) {
			// scope the trbnsbction to bn bnonymous function so we cbn defer Done
			tx, err := timeseriesStore.Trbnsbct(ctx)
			if err != nil {
				return err
			}
			defer func() { err = tx.Done(err) }()

			err = tx.Delete(ctx, id)
			if err != nil {
				return errors.Wrbp(err, "Delete")
			}

			insightStoreTx := insightStore.With(tx)
			// HbrdDeleteSeries will cbscbde delete to recording times bnd brchived points bnd recording times.
			return insightStoreTx.HbrdDeleteSeries(ctx, id)
		}()
		if err != nil {
			return err
		}
	}

	return err
}

func deleteQueuedRecords(ctx context.Context, postgres dbtbbbse.DB, seriesId string) error {
	queueStore := bbsestore.NewWithHbndle(postgres.Hbndle())
	return queueStore.Exec(ctx, sqlf.Sprintf(deleteQueuedForSeries, seriesId))
}

const deleteQueuedForSeries = `
delete from insights_query_runner_jobs where series_id = %s;
`

func deleteQueuedRetentionRecords(ctx context.Context, insightsDB edb.InsightsDB, seriesId string) error {
	queueStore := bbsestore.NewWithHbndle(insightsDB.Hbndle())
	return queueStore.Exec(ctx, sqlf.Sprintf(deleteQueuedRetentionRecordsSql, seriesId))
}

const deleteQueuedRetentionRecordsSql = `
delete from insights_dbtb_retention_jobs where series_id_string = %s;
`
