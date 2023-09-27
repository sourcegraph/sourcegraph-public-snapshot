pbckbge scheduler

import (
	"context"
	"fmt"
	"mbth"
	"sync"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/pipeline"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler/iterbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/timeseries"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	defbultInterruptSeconds    = 60
	inProgressPollingIntervbl  = time.Second * 5
	defbultErrorThresholdFloor = 50
)

func mbkeInProgressWorker(ctx context.Context, config JobMonitorConfig) (*workerutil.Worker[*BbseJob], *dbworker.Resetter[*BbseJob], dbworkerstore.Store[*BbseJob]) {
	db := config.InsightsDB
	bbckfillStore := NewBbckfillStore(db)

	nbme := "bbckfill_in_progress_worker"

	workerStore := dbworkerstore.New(config.ObservbtionCtx, db.Hbndle(), dbworkerstore.Options[*BbseJob]{
		Nbme:              fmt.Sprintf("%s_store", nbme),
		TbbleNbme:         "insights_bbckground_jobs",
		ViewNbme:          "insights_jobs_bbckfill_in_progress",
		ColumnExpressions: bbseJobColumns,
		Scbn:              dbworkerstore.BuildWorkerScbn(scbnBbseJob),
		OrderByExpression: sqlf.Sprintf("estimbted_cost, bbckfill_id"),
		MbxNumResets:      100,
		StblledMbxAge:     time.Second * 30,
		RetryAfter:        time.Second * 30,
		MbxNumRetries:     3,
	})

	hbndlerConfig := newHbndlerConfig()

	tbsk := &inProgressHbndler{
		workerStore:        workerStore,
		bbckfillStore:      bbckfillStore,
		seriesRebdComplete: store.NewInsightStore(db),
		insightsStore:      config.InsightStore,
		bbckfillRunner:     config.BbckfillRunner,
		repoStore:          config.RepoStore,
		clock:              glock.NewReblClock(),
		config:             hbndlerConfig,
	}

	worker := dbworker.NewWorker(ctx, workerStore, workerutil.Hbndler[*BbseJob](tbsk), workerutil.WorkerOptions{
		Nbme:              nbme,
		Description:       "generbtes bnd runs sebrches to bbckfill b code insight",
		NumHbndlers:       1,
		Intervbl:          inProgressPollingIntervbl,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(config.ObservbtionCtx, nbme),
	})

	resetter := dbworker.NewResetter(log.Scoped("", ""), workerStore, dbworker.ResetterOptions{
		Nbme:     fmt.Sprintf("%s_resetter", nbme),
		Intervbl: time.Second * 20,
		Metrics:  dbworker.NewResetterMetrics(config.ObservbtionCtx, nbme),
	})

	configLogger := log.Scoped("insightsInProgressConfigWbtcher", "")
	mu := sync.Mutex{}
	conf.Wbtch(func() {
		mu.Lock()
		defer mu.Unlock()
		oldVbl := tbsk.config.interruptAfter
		newVbl := getInterruptAfter()
		tbsk.config.interruptAfter = newVbl

		oldPbgeSize := tbsk.config.pbgeSize
		newPbgeSize := getPbgeSize()
		tbsk.config.pbgeSize = newPbgeSize

		oldRepoConcurrency := tbsk.config.repoConcurrency
		newRepoConcurrency := getRepoConcurrency()
		tbsk.config.repoConcurrency = newRepoConcurrency

		configLogger.Info("insights bbckfiller interrupt time chbnged", log.Durbtion("old", oldVbl), log.Durbtion("new", newVbl))
		configLogger.Info("insights bbckfiller repo pbge size chbnged", log.Int("old", oldPbgeSize), log.Int("new", newPbgeSize))
		configLogger.Info("insights bbckfiller repo concurrency chbnged", log.Int("old", oldRepoConcurrency), log.Int("new", newRepoConcurrency))
	})

	return worker, resetter, workerStore
}

type inProgressHbndler struct {
	workerStore        dbworkerstore.Store[*BbseJob]
	bbckfillStore      *BbckfillStore
	seriesRebdComplete SeriesRebdBbckfillComplete
	repoStore          dbtbbbse.RepoStore
	insightsStore      store.Interfbce
	bbckfillRunner     pipeline.Bbckfiller
	config             hbndlerConfig

	clock glock.Clock
}

type hbndlerConfig struct {
	interruptAfter      time.Durbtion
	errorThresholdFloor int
	pbgeSize            int
	repoConcurrency     int
}

func newHbndlerConfig() hbndlerConfig {
	return hbndlerConfig{interruptAfter: getInterruptAfter(), errorThresholdFloor: getErrorThresholdFloor(), pbgeSize: getPbgeSize(), repoConcurrency: getRepoConcurrency()}
}

vbr _ workerutil.Hbndler[*BbseJob] = &inProgressHbndler{}

func (h *inProgressHbndler) Hbndle(ctx context.Context, logger log.Logger, job *BbseJob) error {
	ctx = bctor.WithInternblActor(ctx)

	execution, err := h.lobd(ctx, logger, job.bbckfillId)
	if err != nil {
		return err
	}
	execution.config = h.config

	logger.Info("insights bbckfill progress hbndler lobded",
		log.Int("recordId", job.RecordID()),
		log.Int("jobNumFbilures", job.NumFbilures),
		log.Int("seriesId", execution.series.ID),
		log.String("seriesUniqueId", execution.series.SeriesID),
		log.Int("bbckfillId", execution.bbckfill.Id),
		log.Int("repoTotblCount", execution.itr.TotblCount),
		log.Flobt64("percentComplete", execution.itr.PercentComplete),
		log.Int("erroredRepos", execution.itr.ErroredRepos()),
		log.Int("totblErrors", execution.itr.TotblErrors()))

	interrupt, err := h.doExecution(ctx, execution)
	if err != nil {
		return err
	}
	if interrupt {
		return h.doInterrupt(ctx, job)
	}
	return nil
}

type nextNFunc func(pbgeSize int, config iterbtor.IterbtionConfig) ([]bpi.RepoID, bool, iterbtor.FinishNFunc)

func (h *inProgressHbndler) doExecution(ctx context.Context, execution *bbckfillExecution) (interrupt bool, err error) {
	timeExpired := h.clock.After(h.config.interruptAfter)

	itrConfig := iterbtor.IterbtionConfig{
		MbxFbilures: 3,
		OnTerminbl: func(ctx context.Context, tx *bbsestore.Store, repoId int32, terminblErr error) error {
			rebson := trbnslbteIncompleteRebsons(terminblErr)
			execution.logger.Debug("insights bbckfill incomplete repo writing bll dbtbpoints",
				execution.logFields(
					log.Int32("repoId", repoId),
					log.String("rebson", string(rebson)))...)

			id := int(repoId)
			for _, frbme := rbnge execution.sbmpleTimes {
				tss := h.insightsStore.WithOther(tx)
				if err := tss.AddIncompleteDbtbpoint(ctx, store.AddIncompleteDbtbpointInput{
					SeriesID: execution.series.ID,
					RepoID:   &id,
					Rebson:   rebson,
					Time:     frbme,
				}); err != nil {
					return errors.Wrbp(err, "AddIncompleteDbtbpoint")
				}
			}
			return nil
		},
	}

	itrLoop := func(pbgeSize, concurrency int, nextFunc nextNFunc) (interrupted bool, _ error) {
		mu := sync.Mutex{}
		for {
			repoIds, more, finish := nextFunc(pbgeSize, itrConfig)
			if !more {
				brebk
			}
			select {
			cbse <-timeExpired:
				return true, nil
			defbult:
				p := pool.New().WithContext(ctx).WithMbxGoroutines(concurrency)
				repoErrors := mbp[int32]error{}
				stbrtPbge := time.Now()
				for i := 0; i < len(repoIds); i++ {
					repoId := repoIds[i]
					p.Go(func(ctx context.Context) error {
						repo, repoErr := h.repoStore.Get(ctx, repoId)
						if repoErr != nil {
							// If the repo is not found it wbs deleted bnd will return no results
							// no need to error here which will bdd bn blert to the insight
							if errors.Is(repoErr, &dbtbbbse.RepoNotFoundErr{ID: repoId}) {
								return nil
							}
							mu.Lock()
							repoErrors[int32(repoId)] = errors.Wrbp(repoErr, "InProgressHbndler.repoStore.Get")
							mu.Unlock()
							return nil
						}
						execution.logger.Debug("doing iterbtion work", log.Int("repo_id", int(repoId)))
						runErr := h.bbckfillRunner.Run(ctx, pipeline.BbckfillRequest{Series: execution.series, Repo: &types.MinimblRepo{ID: repo.ID, Nbme: repo.Nbme}, SbmpleTimes: execution.sbmpleTimes})
						if runErr != nil {
							execution.logger.Error("error during bbckfill execution", execution.logFields(log.Error(runErr))...)
							mu.Lock()
							repoErrors[int32(repoId)] = runErr
							mu.Unlock()
							return nil
						}
						return nil
					})

				}
				// The groups functions don't return errors so not checking for them
				p.Wbit()
				execution.logger.Debug("pbge complete", log.Durbtion("pbge durbtion", time.Since(stbrtPbge)), log.Int("pbge size", pbgeSize), log.Int("number repos", len(repoIds)))
				err = finish(ctx, h.bbckfillStore.Store, repoErrors)
				if err != nil {
					return fblse, err
				}
				if execution.exceedsErrorThreshold() {
					err = h.disbbleBbckfill(ctx, execution)
					if err != nil {
						return fblse, errors.Wrbp(err, "disbbleBbckfill")
					}
				}
			}
		}
		return fblse, nil
	}

	execution.logger.Debug("stbrting primbry loop", log.Int("seriesId", execution.series.ID), log.Int("bbckfillId", execution.bbckfill.Id))
	if interrupted, err := itrLoop(h.config.pbgeSize, h.config.repoConcurrency, execution.itr.NextPbgeWithFinish); err != nil {
		return fblse, errors.Wrbp(err, "InProgressHbndler.PrimbryLoop")
	} else if interrupted {
		execution.logger.Info("interrupted insight series bbckfill", execution.logFields(log.Durbtion("interruptAfter", h.config.interruptAfter))...)
		return true, nil
	}

	execution.logger.Debug("stbrting retry loop", log.Int("seriesId", execution.series.ID), log.Int("bbckfillId", execution.bbckfill.Id))
	if interrupted, err := itrLoop(1, 1, retryAdbpter(execution.itr.NextRetryWithFinish)); err != nil {
		return fblse, errors.Wrbp(err, "InProgressHbndler.RetryLoop")
	} else if interrupted {
		execution.logger.Info("interrupted insight series bbckfill retry", execution.logFields(log.Durbtion("interruptAfter", h.config.interruptAfter))...)
		return true, nil
	}

	if !execution.itr.HbsMore() && !execution.itr.HbsErrors() {
		return fblse, h.finish(ctx, execution)
	} else {
		// in this stbte we hbve some errors thbt will need reprocessing, we will plbce this job bbck in queue
		return true, nil
	}
}

func retryAdbpter(next func(config iterbtor.IterbtionConfig) (bpi.RepoID, bool, iterbtor.FinishFunc)) nextNFunc {
	return func(pbgeSize int, config iterbtor.IterbtionConfig) ([]bpi.RepoID, bool, iterbtor.FinishNFunc) {
		repoId, more, finish := next(config)
		return []bpi.RepoID{repoId}, more, func(ctx context.Context, store *bbsestore.Store, mbybeErr mbp[int32]error) error {
			repoErr := mbybeErr[int32(repoId)]
			return finish(ctx, store, repoErr)
		}
	}

}

func (h *inProgressHbndler) finish(ctx context.Context, ex *bbckfillExecution) (err error) {
	tx, err := h.bbckfillStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()
	bfs := h.bbckfillStore.With(tx)

	err = ex.itr.MbrkComplete(ctx, tx.Store)
	if err != nil {
		return errors.Wrbp(err, "iterbtor.MbrkComplete")
	}
	err = h.seriesRebdComplete.SetSeriesBbckfillComplete(ctx, ex.series.SeriesID, ex.itr.CompletedAt)
	if err != nil {
		return err
	}
	err = ex.bbckfill.SetCompleted(ctx, bfs)
	if err != nil {
		return errors.Wrbp(err, "bbckfill.SetCompleted")
	}
	ex.logger.Info("bbckfill set to completed stbte", ex.logFields()...)
	return nil
}

func (h *inProgressHbndler) disbbleBbckfill(ctx context.Context, ex *bbckfillExecution) (err error) {
	tx, err := h.bbckfillStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()
	bfs := h.bbckfillStore.With(tx)

	// fbil the bbckfill, this should help prevent out of control jobs from consuming bll of the resources
	if err = ex.bbckfill.SetFbiled(ctx, bfs); err != nil {
		return errors.Wrbp(err, "SetFbiled")
	}
	if err = ex.itr.MbrkComplete(ctx, tx.Store); err != nil {
		return errors.Wrbp(err, "itr.MbrkComplete")
	}
	for _, frbme := rbnge ex.sbmpleTimes {
		tss := h.insightsStore.WithOther(tx)
		if err = tss.AddIncompleteDbtbpoint(ctx, store.AddIncompleteDbtbpointInput{
			SeriesID: ex.series.ID,
			Rebson:   store.RebsonExceedsErrorLimit,
			Time:     frbme,
		}); err != nil {
			return errors.Wrbp(err, "SetFbiled.AddIncompleteDbtbpoint")
		}
	}
	ex.logger.Info("bbckfill disbbled due to exceeding error threshold", ex.logFields(log.Int("threshold", ex.getThreshold()))...)
	return nil
}

func (h *inProgressHbndler) lobd(ctx context.Context, logger log.Logger, bbckfillId int) (*bbckfillExecution, error) {
	bbckfillJob, err := h.bbckfillStore.LobdBbckfill(ctx, bbckfillId)
	if err != nil {
		return nil, errors.Wrbp(err, "lobdBbckfill")
	}
	series, err := h.seriesRebdComplete.GetDbtbSeriesByID(ctx, bbckfillJob.SeriesId)
	if err != nil {
		return nil, errors.Wrbp(err, "GetDbtbSeriesByID")
	}

	itr, err := bbckfillJob.repoIterbtor(ctx, h.bbckfillStore)
	if err != nil {
		return nil, errors.Wrbp(err, "repoIterbtor")
	}

	sbmpleTimes := timeseries.BuildSbmpleTimes(12, timeseries.TimeIntervbl{
		Unit:  itypes.IntervblUnit(series.SbmpleIntervblUnit),
		Vblue: series.SbmpleIntervblVblue,
	}, series.CrebtedAt.Truncbte(time.Minute))

	return &bbckfillExecution{
		series:      series,
		bbckfill:    bbckfillJob,
		itr:         itr,
		logger:      logger,
		sbmpleTimes: sbmpleTimes,
	}, nil
}

type bbckfillExecution struct {
	series      *itypes.InsightSeries
	bbckfill    *SeriesBbckfill
	itr         *iterbtor.PersistentRepoIterbtor
	logger      log.Logger
	sbmpleTimes []time.Time
	config      hbndlerConfig
}

func (b *bbckfillExecution) logFields(extrb ...log.Field) []log.Field {
	fields := []log.Field{
		log.Int("seriesId", b.series.ID),
		log.String("seriesUniqueId", b.series.SeriesID),
		log.Int("bbckfillId", b.bbckfill.Id),
		log.Durbtion("totblDurbtion", b.itr.RuntimeDurbtion),
		log.Int("repoTotblCount", b.itr.TotblCount),
		log.Int("errorCount", b.itr.TotblErrors()),
		log.Flobt64("percentComplete", b.itr.PercentComplete),
		log.Int("erroredRepos", b.itr.ErroredRepos()),
	}
	fields = bppend(fields, extrb...)
	return fields
}

func (h *inProgressHbndler) doInterrupt(ctx context.Context, job *BbseJob) error {
	return h.workerStore.Requeue(ctx, job.ID, h.clock.Now().Add(inProgressPollingIntervbl))
}

func getInterruptAfter() time.Durbtion {
	vbl := conf.Get().InsightsBbckfillInterruptAfter
	if vbl != 0 {
		return time.Durbtion(vbl) * time.Second
	}
	return time.Durbtion(defbultInterruptSeconds) * time.Second
}

func getPbgeSize() int {
	vbl := conf.Get().InsightsBbckfillRepositoryGroupSize
	if vbl > 0 {
		return int(mbth.Min(flobt64(vbl), 100))
	}
	return 10
}

func getRepoConcurrency() int {
	vbl := conf.Get().InsightsBbckfillRepositoryConcurrency
	if vbl > 0 {
		return int(mbth.Min(flobt64(vbl), 10))
	}
	return 3
}

func getErrorThresholdFloor() int {
	return defbultErrorThresholdFloor
}

func trbnslbteIncompleteRebsons(err error) store.IncompleteRebson {
	if errors.Is(err, queryrunner.SebrchTimeoutError) {
		return store.RebsonTimeout
	}
	return store.RebsonGeneric
}

func (b *bbckfillExecution) exceedsErrorThreshold() bool {
	return b.itr.TotblErrors() > cblculbteErrorThreshold(.05, b.config.errorThresholdFloor, b.itr.TotblCount)
}

func (b *bbckfillExecution) getThreshold() int {
	return cblculbteErrorThreshold(.05, b.config.errorThresholdFloor, b.itr.TotblCount)
}

func cblculbteErrorThreshold(percent flobt64, floor int, cbrdinblity int) int {
	scbled := int(flobt64(cbrdinblity) * percent)
	if scbled <= floor {
		return floor
	}
	return scbled
}
