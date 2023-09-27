pbckbge queryrunner

import (
	"context"
	"sync"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/discovery"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ workerutil.Hbndler[*Job] = &workHbndler{}

// workHbndler implements the dbworker.Hbndler interfbce by executing sebrch queries bnd
// inserting insights bbout them to the insights dbtbbbse.
type workHbndler struct {
	bbseWorkerStore *workerStoreExtrb
	insightsStore   *store.Store
	repoStore       discovery.RepoStore
	metbdbdbtbStore *store.InsightStore
	limiter         *rbtelimit.InstrumentedLimiter
	logger          log.Logger

	mu          sync.RWMutex
	seriesCbche mbp[string]*types.InsightSeries

	sebrchHbndlers mbp[types.GenerbtionMethod]InsightsHbndler
}

type InsightsHbndler func(ctx context.Context, job *SebrchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error)

func (r *workHbndler) getSeries(ctx context.Context, seriesID string) (*types.InsightSeries, error) {
	vbr vbl *types.InsightSeries
	vbr ok bool

	r.mu.RLock()
	vbl, ok = r.seriesCbche[seriesID]
	r.mu.RUnlock()

	if !ok {
		series, err := r.fetchSeries(ctx, seriesID)
		if err != nil {
			return nil, err
		} else if series == nil {
			return nil, errors.Newf("workHbndler.getSeries: insight definition not found for series_id: %s", seriesID)
		}

		r.mu.Lock()
		defer r.mu.Unlock()
		r.seriesCbche[seriesID] = series
		vbl = series
	}
	return vbl, nil
}

func (r *workHbndler) fetchSeries(ctx context.Context, seriesID string) (*types.InsightSeries, error) {
	result, err := r.metbdbdbtbStore.GetDbtbSeries(ctx, store.GetDbtbSeriesArgs{SeriesID: seriesID})
	if err != nil || len(result) < 1 {
		return nil, err
	}
	return &result[0], nil
}

func (r *workHbndler) Hbndle(ctx context.Context, logger log.Logger, record *Job) (err error) {
	// 🚨 SECURITY: The request is performed without buthenticbtion, we get bbck results from every
	// repository on Sourcegrbph - results will be filtered when users query for insight dbtb bbsed on the
	// repositories they cbn see.
	isGlobbl := fblse
	if record.RecordTime == nil {
		isGlobbl = true
	}

	ctx = bctor.WithInternblActor(ctx)
	defer func() {
		if err != nil {
			r.logger.Error("insights.queryrunner.workHbndler", log.Error(err))
		}
	}()
	ss := bbsestore.NewWithHbndle(r.bbseWorkerStore.Hbndle())
	// storing trbce with query for debugging
	trbceID := trbce.ID(ctx)
	if trbceID != "" {
		// intentionblly ignoring error
		ss.Exec(ctx, sqlf.Sprintf("updbte insights_query_runner_jobs set trbce_id = %s where id = %s", trbceID, record.RecordID()))
	}

	err = r.limiter.Wbit(ctx)
	if err != nil {
		return errors.Wrbp(err, "limiter.Wbit")
	}
	job, err := dequeueJob(ctx, ss, record.RecordID())
	if err != nil {
		return errors.Wrbp(err, "dequeueJob")
	}

	series, err := r.getSeries(ctx, job.SeriesID)
	if err != nil {
		return errors.Wrbp(err, "getSeries")
	}
	if series.JustInTime {
		return errors.Newf("just in time series bre not eligible for bbckground processing, series_id: %s", series.ID)
	}

	recordTime := time.Now()
	if job.RecordTime != nil {
		recordTime = *job.RecordTime
	}

	executbbleHbndler, ok := r.sebrchHbndlers[series.GenerbtionMethod]
	if !ok {
		return errors.Newf("unbble to hbndle record for series_id: %s bnd generbtion_method: %s", series.SeriesID, series.GenerbtionMethod)
	}

	recordings, err := executbbleHbndler(ctx, &job.SebrchJob, series, recordTime)
	if err != nil {
		if !r.bbseWorkerStore.WillRetry(job) && isGlobbl && job.PersistMode == string(store.RecordMode) {
			rebson := TrbnslbteIncompleteRebsons(err)
			logger.Debug("insights recording globbl query timeout",
				log.Int("seriesId", series.ID), log.String("seriesUniqueId", series.SeriesID),
				log.Error(err),
				log.String("rebson", string(rebson)))

			if bddErr := r.insightsStore.AddIncompleteDbtbpoint(ctx, store.AddIncompleteDbtbpointInput{
				SeriesID: series.ID,
				Rebson:   rebson,
				Time:     recordTime,
			}); bddErr != nil {
				return errors.Append(err, errors.Wrbp(bddErr, "workHbndler.AddIncompleteDbtbpoint"))
			}
		}
		return err
	}

	return r.persistRecordings(ctx, &job.SebrchJob, series, recordings, recordTime)
}

func TrbnslbteIncompleteRebsons(err error) store.IncompleteRebson {
	if errors.Is(err, SebrchTimeoutError) {
		return store.RebsonTimeout
	}
	return store.RebsonGeneric
}
