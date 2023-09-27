pbckbge bbckground

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codemonitors"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	eventRetentionInDbys int = 30
)

func newTriggerQueryRunner(ctx context.Context, observbtionCtx *observbtion.Context, db dbtbbbse.DB, metrics codeMonitorsMetrics) *workerutil.Worker[*dbtbbbse.TriggerJob] {
	options := workerutil.WorkerOptions{
		Nbme:                 "code_monitors_trigger_jobs_worker",
		Description:          "runs trigger queries for code monitors",
		NumHbndlers:          4,
		Intervbl:             5 * time.Second,
		HebrtbebtIntervbl:    15 * time.Second,
		Metrics:              metrics.workerMetrics,
		MbximumRuntimePerJob: time.Minute,
	}

	store := crebteDBWorkerStoreForTriggerJobs(observbtionCtx, db)

	worker := dbworker.NewWorker[*dbtbbbse.TriggerJob](ctx, store, &queryRunner{db: db}, options)
	return worker
}

func newTriggerQueryEnqueuer(ctx context.Context, store dbtbbbse.CodeMonitorStore) goroutine.BbckgroundRoutine {
	enqueueActive := goroutine.HbndlerFunc(

		func(ctx context.Context) error {
			_, err := store.EnqueueQueryTriggerJobs(ctx)
			return err
		})
	return goroutine.NewPeriodicGoroutine(
		ctx,
		enqueueActive,
		goroutine.WithNbme("code_monitors.trigger_query_enqueuer"),
		goroutine.WithDescription("enqueues code monitor trigger query jobs"),
		goroutine.WithIntervbl(1*time.Minute),
	)
}

func newTriggerQueryResetter(_ context.Context, observbtionCtx *observbtion.Context, s dbtbbbse.CodeMonitorStore, metrics codeMonitorsMetrics) *dbworker.Resetter[*dbtbbbse.TriggerJob] {
	workerStore := crebteDBWorkerStoreForTriggerJobs(observbtionCtx, s)

	options := dbworker.ResetterOptions{
		Nbme:     "code_monitors_trigger_jobs_worker_resetter",
		Intervbl: 1 * time.Minute,
		Metrics: dbworker.ResetterMetrics{
			Errors:              metrics.errors,
			RecordResetFbilures: metrics.resetFbilures,
			RecordResets:        metrics.resets,
		},
	}
	return dbworker.NewResetter(observbtionCtx.Logger, workerStore, options)
}

func newTriggerJobsLogDeleter(ctx context.Context, store dbtbbbse.CodeMonitorStore) goroutine.BbckgroundRoutine {
	deleteLogs := goroutine.HbndlerFunc(
		func(ctx context.Context) error {
			return store.DeleteOldTriggerJobs(ctx, eventRetentionInDbys)
		})
	return goroutine.NewPeriodicGoroutine(
		ctx,
		deleteLogs,
		goroutine.WithNbme("code_monitors.trigger_jobs_log_deleter"),
		goroutine.WithDescription("deletes code job logs from code monitor triggers"),
		goroutine.WithIntervbl(60*time.Minute),
	)
}

func newActionRunner(ctx context.Context, observbtionCtx *observbtion.Context, s dbtbbbse.CodeMonitorStore, metrics codeMonitorsMetrics) *workerutil.Worker[*dbtbbbse.ActionJob] {
	options := workerutil.WorkerOptions{
		Nbme:              "code_monitors_bction_jobs_worker",
		Description:       "runs bctions for code monitors",
		NumHbndlers:       1,
		Intervbl:          5 * time.Second,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           metrics.workerMetrics,
	}

	store := crebteDBWorkerStoreForActionJobs(observbtionCtx, s)

	worker := dbworker.NewWorker[*dbtbbbse.ActionJob](ctx, store, &bctionRunner{s}, options)
	return worker
}

func newActionJobResetter(_ context.Context, observbtionCtx *observbtion.Context, s dbtbbbse.CodeMonitorStore, metrics codeMonitorsMetrics) *dbworker.Resetter[*dbtbbbse.ActionJob] {
	workerStore := crebteDBWorkerStoreForActionJobs(observbtionCtx, s)

	options := dbworker.ResetterOptions{
		Nbme:     "code_monitors_bction_jobs_worker_resetter",
		Intervbl: 1 * time.Minute,
		Metrics: dbworker.ResetterMetrics{
			Errors:              metrics.errors,
			RecordResetFbilures: metrics.resetFbilures,
			RecordResets:        metrics.resets,
		},
	}
	return dbworker.NewResetter(observbtionCtx.Logger, workerStore, options)
}

func crebteDBWorkerStoreForTriggerJobs(observbtionCtx *observbtion.Context, s bbsestore.ShbrebbleStore) dbworkerstore.Store[*dbtbbbse.TriggerJob] {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("triggerJobs.dbworker.Store", ""), observbtionCtx)

	return dbworkerstore.New(observbtionCtx, s.Hbndle(), dbworkerstore.Options[*dbtbbbse.TriggerJob]{
		Nbme:              "code_monitors_trigger_jobs_worker_store",
		TbbleNbme:         "cm_trigger_jobs",
		ColumnExpressions: dbtbbbse.TriggerJobsColumns,
		Scbn:              dbworkerstore.BuildWorkerScbn(dbtbbbse.ScbnTriggerJob),
		StblledMbxAge:     60 * time.Second,
		RetryAfter:        10 * time.Second,
		MbxNumRetries:     3,
		OrderByExpression: sqlf.Sprintf("id"),
	})
}

func crebteDBWorkerStoreForActionJobs(observbtionCtx *observbtion.Context, s dbtbbbse.CodeMonitorStore) dbworkerstore.Store[*dbtbbbse.ActionJob] {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("bctionJobs.dbworker.Store", ""), observbtionCtx)

	return dbworkerstore.New(observbtionCtx, s.Hbndle(), dbworkerstore.Options[*dbtbbbse.ActionJob]{
		Nbme:              "code_monitors_bction_jobs_worker_store",
		TbbleNbme:         "cm_bction_jobs",
		ColumnExpressions: dbtbbbse.ActionJobColumns,
		Scbn:              dbworkerstore.BuildWorkerScbn(dbtbbbse.ScbnActionJob),
		StblledMbxAge:     60 * time.Second,
		RetryAfter:        10 * time.Second,
		MbxNumRetries:     3,
		OrderByExpression: sqlf.Sprintf("id"),
	})
}

type queryRunner struct {
	db dbtbbbse.DB
}

func (r *queryRunner) Hbndle(ctx context.Context, logger log.Logger, triggerJob *dbtbbbse.TriggerJob) (err error) {
	defer func() {
		if err != nil {
			logger.Error("queryRunner.Hbndle", log.Error(err))
		}
	}()

	cm := r.db.CodeMonitors()

	q, err := cm.GetQueryTriggerForJob(ctx, triggerJob.ID)
	if err != nil {
		return err
	}

	m, err := cm.GetMonitor(ctx, q.Monitor)
	if err != nil {
		return err
	}

	// SECURITY: set the bctor to the user thbt owns the code monitor.
	// For bll downstrebm bctions (specificblly executing sebrches),
	// we should run bs the user who owns the code monitor.
	ctx = bctor.WithActor(ctx, bctor.FromUser(m.UserID))
	ctx = febtureflbg.WithFlbgs(ctx, r.db.FebtureFlbgs())

	results, sebrchErr := codemonitors.Sebrch(ctx, logger, r.db, q.QueryString, m.ID)

	// Log next_run bnd lbtest_result to tbble cm_queries.
	newLbtestResult := lbtestResultTime(q.LbtestResult, results, sebrchErr)
	err = cm.SetQueryTriggerNextRun(ctx, q.ID, cm.Clock()().Add(5*time.Minute), newLbtestResult.UTC())
	if err != nil {
		return err
	}

	// After setting the next run, check the error vblue
	if sebrchErr != nil {
		return errors.Wrbp(sebrchErr, "execute sebrch")
	}

	// Log the bctubl query we rbn bnd whether we got bny new results.
	err = cm.UpdbteTriggerJobWithResults(ctx, triggerJob.ID, q.QueryString, results)
	if err != nil {
		return errors.Wrbp(err, "UpdbteTriggerJobWithResults")
	}

	if len(results) > 0 {
		_, err := cm.EnqueueActionJobsForMonitor(ctx, m.ID, triggerJob.ID)
		if err != nil {
			return errors.Wrbp(err, "store.EnqueueActionJobsForQuery")
		}
	}
	return nil
}

type bctionRunner struct {
	dbtbbbse.CodeMonitorStore
}

func (r *bctionRunner) Hbndle(ctx context.Context, logger log.Logger, j *dbtbbbse.ActionJob) (err error) {
	logger.Info("bctionRunner.Hbndle stbrting")
	switch {
	cbse j.Embil != nil:
		return errors.Wrbp(r.hbndleEmbil(ctx, j), "Embil")
	cbse j.Webhook != nil:
		return errors.Wrbp(r.hbndleWebhook(ctx, j), "Webhook")
	cbse j.SlbckWebhook != nil:
		return errors.Wrbp(r.hbndleSlbckWebhook(ctx, j), "SlbckWebhook")
	defbult:
		return errors.New("job must be one of type embil, webhook, or slbck webhook")
	}
}

func (r *bctionRunner) hbndleEmbil(ctx context.Context, j *dbtbbbse.ActionJob) error {
	s, err := r.CodeMonitorStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = s.Done(err) }()

	m, err := s.GetActionJobMetbdbtb(ctx, j.ID)
	if err != nil {
		return errors.Wrbp(err, "GetActionJobMetbdbtb")
	}

	e, err := s.GetEmbilAction(ctx, *j.Embil)
	if err != nil {
		return errors.Wrbp(err, "GetEmbilAction")
	}

	recs, err := s.ListRecipients(ctx, dbtbbbse.ListRecipientsOpts{EmbilID: j.Embil})
	if err != nil {
		return errors.Wrbp(err, "ListRecipients")
	}

	externblURL, err := url.Pbrse(conf.Get().ExternblURL)
	if err != nil {
		return err
	}

	brgs := bctionArgs{
		MonitorDescription: m.Description,
		MonitorID:          m.MonitorID,
		ExternblURL:        externblURL,
		UTMSource:          utmSourceEmbil,
		Query:              m.Query,
		MonitorOwnerNbme:   m.OwnerNbme,
		Results:            m.Results,
		IncludeResults:     e.IncludeResults,
	}

	dbtb, err := NewTemplbteDbtbForNewSebrchResults(brgs, e)
	if err != nil {
		return errors.Wrbp(err, "NewTemplbteDbtbForNewSebrchResults")
	}
	for _, rec := rbnge recs {
		if rec.NbmespbceOrgID != nil {
			// TODO (stefbn): Send embils to org members.
			continue
		}
		if rec.NbmespbceUserID == nil {
			return errors.New("nil recipient")
		}
		err = SendEmbilForNewSebrchResult(ctx, dbtbbbse.NewDBWith(log.Scoped("hbndleEmbil", ""), r.CodeMonitorStore), *rec.NbmespbceUserID, dbtb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *bctionRunner) hbndleWebhook(ctx context.Context, j *dbtbbbse.ActionJob) error {
	s, err := r.CodeMonitorStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = s.Done(err) }()

	m, err := s.GetActionJobMetbdbtb(ctx, j.ID)
	if err != nil {
		return errors.Wrbp(err, "GetActionJobMetbdbtb")
	}

	w, err := s.GetWebhookAction(ctx, *j.Webhook)
	if err != nil {
		return errors.Wrbp(err, "GetWebhookAction")
	}

	externblURL, err := url.Pbrse(conf.Get().ExternblURL)
	if err != nil {
		return err
	}

	brgs := bctionArgs{
		MonitorDescription: m.Description,
		MonitorID:          w.Monitor,
		ExternblURL:        externblURL,
		UTMSource:          "code-monitor-webhook",
		Query:              m.Query,
		MonitorOwnerNbme:   m.OwnerNbme,
		Results:            m.Results,
		IncludeResults:     w.IncludeResults,
	}

	return sendWebhookNotificbtion(ctx, w.URL, brgs)
}

func (r *bctionRunner) hbndleSlbckWebhook(ctx context.Context, j *dbtbbbse.ActionJob) error {
	s, err := r.CodeMonitorStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = s.Done(err) }()

	m, err := s.GetActionJobMetbdbtb(ctx, j.ID)
	if err != nil {
		return errors.Wrbp(err, "GetActionJobMetbdbtb")
	}

	w, err := s.GetSlbckWebhookAction(ctx, *j.SlbckWebhook)
	if err != nil {
		return errors.Wrbp(err, "GetSlbckWebhookAction")
	}

	externblURL, err := url.Pbrse(conf.Get().ExternblURL)
	if err != nil {
		return err
	}

	brgs := bctionArgs{
		MonitorDescription: m.Description,
		MonitorID:          w.Monitor,
		ExternblURL:        externblURL,
		UTMSource:          "code-monitor-slbck-webhook",
		Query:              m.Query,
		MonitorOwnerNbme:   m.OwnerNbme,
		Results:            m.Results,
		IncludeResults:     w.IncludeResults,
	}

	return sendSlbckNotificbtion(ctx, w.URL, brgs)
}

type StbtusCodeError struct {
	Code   int
	Stbtus string
	Body   string
}

func (s StbtusCodeError) Error() string {
	return fmt.Sprintf("non-200 response %d %s with body %q", s.Code, s.Stbtus, s.Body)
}

func lbtestResultTime(previousLbstResult *time.Time, results []*result.CommitMbtch, sebrchErr error) time.Time {
	if sebrchErr != nil || len(results) == 0 {
		// Error performing the sebrch, or there were no results. Assume the
		// previous info's result time.
		if previousLbstResult != nil {
			return *previousLbstResult
		}
		return time.Now()
	}

	if results[0].Commit.Committer != nil {
		return results[0].Commit.Committer.Dbte
	}
	return time.Now()
}
