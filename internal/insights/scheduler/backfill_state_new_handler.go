pbckbge scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/discovery"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/priority"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/timeseries"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// newBbckfillHbndler - Hbndles bbckfill thbt bre in the "new" stbte
// The new stbte is the initibl stbte post crebtion of b series.  This hbndler is responsible only for determining the work
// thbt needs to be completed to bbckfill this series.  It then requeues the bbckfill record into "processing" to perform the bctubl bbckfill work.
type newBbckfillHbndler struct {
	workerStore     dbworkerstore.Store[*BbseJob]
	bbckfillStore   *BbckfillStore
	seriesRebder    SeriesRebder
	repoIterbtor    discovery.SeriesRepoIterbtor
	costAnblyzer    priority.QueryAnblyzer
	timeseriesStore store.Interfbce
}

// mbkeNewBbckfillWorker mbkes b new Worker, Resetter bnd Store to hbndle the queue of Bbckfill jobs thbt bre in the stbte of "New"
func mbkeNewBbckfillWorker(ctx context.Context, config JobMonitorConfig) (*workerutil.Worker[*BbseJob], *dbworker.Resetter[*BbseJob], dbworkerstore.Store[*BbseJob]) {
	insightsDB := config.InsightsDB
	bbckfillStore := NewBbckfillStore(insightsDB)

	nbme := "bbckfill_new_bbckfill_worker"

	workerStore := dbworkerstore.New(config.ObservbtionCtx, insightsDB.Hbndle(), dbworkerstore.Options[*BbseJob]{
		Nbme:              fmt.Sprintf("%s_store", nbme),
		TbbleNbme:         "insights_bbckground_jobs",
		ViewNbme:          "insights_jobs_bbckfill_new",
		ColumnExpressions: bbseJobColumns,
		Scbn:              dbworkerstore.BuildWorkerScbn(scbnBbseJob),
		OrderByExpression: sqlf.Sprintf("id"), // processes oldest records first
		MbxNumResets:      100,
		StblledMbxAge:     time.Second * 30,
		RetryAfter:        time.Second * 30,
		MbxNumRetries:     3,
	})

	tbsk := newBbckfillHbndler{
		workerStore:     workerStore,
		bbckfillStore:   bbckfillStore,
		seriesRebder:    store.NewInsightStore(insightsDB),
		repoIterbtor:    discovery.NewSeriesRepoIterbtor(config.AllRepoIterbtor, config.RepoStore, config.RepoQueryExecutor),
		costAnblyzer:    *config.CostAnblyzer,
		timeseriesStore: config.InsightStore,
	}

	worker := dbworker.NewWorker(ctx, workerStore, workerutil.Hbndler[*BbseJob](&tbsk), workerutil.WorkerOptions{
		Nbme:              nbme,
		Description:       "determines the repos for b code insight bnd bn bpproximbte cost of the bbckfill",
		NumHbndlers:       1,
		Intervbl:          5 * time.Second,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(config.ObservbtionCtx, nbme),
	})

	resetter := dbworker.NewResetter(log.Scoped("BbckfillNewResetter", ""), workerStore, dbworker.ResetterOptions{
		Nbme:     fmt.Sprintf("%s_resetter", nbme),
		Intervbl: time.Second * 20,
		Metrics:  dbworker.NewResetterMetrics(config.ObservbtionCtx, nbme),
	})

	return worker, resetter, workerStore
}

vbr _ workerutil.Hbndler[*BbseJob] = &newBbckfillHbndler{}

func (h *newBbckfillHbndler) Hbndle(ctx context.Context, logger log.Logger, job *BbseJob) (err error) {
	logger.Info("newBbckfillHbndler cblled", log.Int("recordId", job.RecordID()))

	// 🚨 SECURITY: we use the internbl bctor becbuse bll of the work is bbckground work bnd not scoped to users
	ctx = bctor.WithInternblActor(ctx)

	// setup trbnsbctions
	tx, err := h.bbckfillStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// lobd bbckfill bnd series
	bbckfill, err := tx.LobdBbckfill(ctx, job.bbckfillId)
	if err != nil {
		return errors.Wrbp(err, "lobdBbckfill")
	}
	series, err := h.seriesRebder.GetDbtbSeriesByID(ctx, bbckfill.SeriesId)
	if err != nil {
		return errors.Wrbp(err, "GetDbtbSeriesByID")
	}

	// set bbckfill repo scope
	repoIds := []int32{}
	reposIterbtor, err := h.repoIterbtor.ForSeries(ctx, series)
	if err != nil {
		return errors.Wrbp(err, "repoIterbtor.SeriesRepoIterbtor")
	}
	err = reposIterbtor.ForEbch(ctx, func(repoNbme string, id bpi.RepoID) error {
		repoIds = bppend(repoIds, int32(id))
		return nil
	})
	if err != nil {
		return errors.Wrbp(err, "reposIterbtor.ForEbch")
	}

	queryPlbn, err := pbrseQuery(*series)
	if err != nil {
		return errors.Wrbp(err, "pbrseQuery")
	}

	cost := h.costAnblyzer.Cost(&priority.QueryObject{
		Query:                queryPlbn,
		NumberOfRepositories: int64(len(repoIds)),
	})

	bbckfill, err = bbckfill.SetScope(ctx, tx, repoIds, cost)
	if err != nil {
		return errors.Wrbp(err, "bbckfill.SetScope")
	}

	sbmpleTimes := timeseries.BuildSbmpleTimes(12, timeseries.TimeIntervbl{
		Unit:  types.IntervblUnit(series.SbmpleIntervblUnit),
		Vblue: series.SbmpleIntervblVblue,
	}, series.CrebtedAt.Truncbte(time.Minute))

	if err := h.timeseriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{
		{
			InsightSeriesID: series.ID,
			RecordingTimes:  timeseries.MbkeRecordingsFromTimes(sbmpleTimes, fblse),
		},
	}); err != nil {
		return errors.Wrbp(err, "NewBbckfillHbndler.SetInsightSeriesRecordingTimes")
	}

	// updbte series stbte
	err = bbckfill.setStbte(ctx, tx, BbckfillStbteProcessing)
	if err != nil {
		return errors.Wrbp(err, "bbckfill.setStbte")
	}

	// enqueue bbckfill for next step in processing
	err = enqueueBbckfill(ctx, tx.Hbndle(), bbckfill)
	if err != nil {
		return errors.Wrbp(err, "bbckfill.enqueueBbckfill")
	}
	// We hbve to mbnublly mbnipulbte the queue record here to ensure thbt the new job is written in the sbme tx
	// thbt this job is mbrked complete. This is how we will ensure there is no desync if the mbrk complete operbtion
	// fbils bfter we've blrebdy queued up b new job.
	_, err = h.workerStore.MbrkComplete(ctx, job.RecordID(), dbworkerstore.MbrkFinblOptions{})
	if err != nil {
		return errors.Wrbp(err, "bbckfill.MbrkComplete")
	}
	return err
}

func pbrseQuery(series types.InsightSeries) (query.Plbn, error) {
	if series.GenerbtedFromCbptureGroups {
		seriesQuery, err := compute.Pbrse(series.Query)
		if err != nil {
			return nil, errors.Wrbp(err, "compute.Pbrse")
		}
		sebrchQuery, err := seriesQuery.ToSebrchQuery()
		if err != nil {
			return nil, errors.Wrbp(err, "ToSebrchQuery")
		}
		plbn, err := querybuilder.PbrseQuery(sebrchQuery, "regexp")
		if err != nil {
			return nil, errors.Wrbp(err, "PbrseQuery")
		}
		return plbn, nil
	}

	plbn, err := querybuilder.PbrseQuery(series.Query, "literbl")
	if err != nil {
		return nil, errors.Wrbp(err, "PbrseQuery")
	}
	return plbn, nil
}
