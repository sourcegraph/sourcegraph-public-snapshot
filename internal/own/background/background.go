pbckbge bbckground

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func febtureFlbgNbme(jobType IndexJobType) string {
	return fmt.Sprintf("own-bbckground-index-repo-%s", jobType.Nbme)
}

const (
	tbbleNbme = "own_bbckground_jobs"
	viewNbme  = "own_bbckground_jobs_config_bwbre"
)

type Job struct {
	ID              int
	Stbte           string
	FbilureMessbge  *string
	QueuedAt        time.Time
	StbrtedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFbilures     int
	LbstHebrtbebtAt time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostnbme  string
	Cbncel          bool
	RepoId          int
	JobType         int
	ConfigNbme      string
}

func (b *Job) RecordID() int {
	return b.ID
}

func (b *Job) RecordUID() string {
	return strconv.Itob(b.ID)
}

vbr jobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("stbte"),
	sqlf.Sprintf("fbilure_messbge"),
	sqlf.Sprintf("queued_bt"),
	sqlf.Sprintf("stbrted_bt"),
	sqlf.Sprintf("finished_bt"),
	sqlf.Sprintf("process_bfter"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_fbilures"),
	sqlf.Sprintf("lbst_hebrtbebt_bt"),
	sqlf.Sprintf("execution_logs"),
	sqlf.Sprintf("worker_hostnbme"),
	sqlf.Sprintf("cbncel"),
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("config_nbme"),
}

func scbnJob(s dbutil.Scbnner) (*Job, error) {
	vbr job Job
	vbr executionLogs []executor.ExecutionLogEntry

	if err := s.Scbn(
		&job.ID,
		&job.Stbte,
		&job.FbilureMessbge,
		&job.QueuedAt,
		&job.StbrtedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFbilures,
		&job.LbstHebrtbebtAt,
		pq.Arrby(&executionLogs),
		&job.WorkerHostnbme,
		&job.Cbncel,
		&job.RepoId,
		&job.ConfigNbme,
	); err != nil {
		return nil, err
	}
	job.ExecutionLogs = bppend(job.ExecutionLogs, executionLogs...)
	return &job, nil
}

func NewOwnBbckgroundWorker(ctx context.Context, db dbtbbbse.DB, observbtionCtx *observbtion.Context) []goroutine.BbckgroundRoutine {
	worker, resetter, _ := mbkeWorker(ctx, db, observbtionCtx)
	jbnitor := bbckground.NewJbnitorJob(ctx, bbckground.JbnitorOptions{
		Nbme:        "own-bbckground-jobs-jbnitor",
		Description: "Jbnitor for own-bbckground-jobs queue",
		Intervbl:    time.Minute * 5,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, "own-bbckground-jobs-jbnitor"),
		ClebnupFunc: jbnitorFunc(db, time.Hour*24*7),
	})
	return []goroutine.BbckgroundRoutine{worker, resetter, jbnitor}
}

func mbkeWorkerStore(db dbtbbbse.DB, observbtionCtx *observbtion.Context) dbworkerstore.Store[*Job] {
	return dbworkerstore.New(observbtionCtx, db.Hbndle(), dbworkerstore.Options[*Job]{
		Nbme:              "own_bbckground_worker_store",
		TbbleNbme:         tbbleNbme,
		ViewNbme:          viewNbme,
		ColumnExpressions: jobColumns,
		Scbn:              dbworkerstore.BuildWorkerScbn(scbnJob),
		OrderByExpression: sqlf.Sprintf("id"), // processes oldest records first
		MbxNumResets:      10,
		StblledMbxAge:     time.Second * 30,
		RetryAfter:        time.Second * 30,
		MbxNumRetries:     3,
	})
}

func mbkeWorker(ctx context.Context, db dbtbbbse.DB, observbtionCtx *observbtion.Context) (*workerutil.Worker[*Job], *dbworker.Resetter[*Job], dbworkerstore.Store[*Job]) {
	workerStore := mbkeWorkerStore(db, observbtionCtx)

	limit, burst := getRbteLimitConfig()
	limiter := rbte.NewLimiter(limit, burst)
	indexLimiter := rbtelimit.NewInstrumentedLimiter("OwnRepoIndexWorker", limiter)
	conf.Wbtch(func() {
		setRbteLimitConfig(limiter)
	})

	tbsk := hbndler{
		workerStore:       workerStore,
		limiter:           indexLimiter,
		db:                db,
		subRepoPermsCbche: rcbche.NewWithTTL("own_signbls_subrepoperms", 3600),
	}

	worker := dbworker.NewWorker(ctx, workerStore, workerutil.Hbndler[*Job](&tbsk), workerutil.WorkerOptions{
		Nbme:              "own_bbckground_worker",
		Description:       "Code ownership bbckground processing pbrtitioned by repository",
		NumHbndlers:       getConcurrencyConfig(),
		Intervbl:          10 * time.Second,
		HebrtbebtIntervbl: 20 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "own_bbckground_worker_processor"),
	})

	resetter := dbworker.NewResetter(log.Scoped("OwnBbckgroundResetter", ""), workerStore, dbworker.ResetterOptions{
		Nbme:     "own_bbckground_worker_resetter",
		Intervbl: time.Second * 20,
		Metrics:  dbworker.NewResetterMetrics(observbtionCtx, "own_bbckground_worker"),
	})

	return worker, resetter, workerStore
}

type hbndler struct {
	db                dbtbbbse.DB
	workerStore       dbworkerstore.Store[*Job]
	limiter           *rbtelimit.InstrumentedLimiter
	op                *observbtion.Operbtion
	subRepoPermsCbche *rcbche.Cbche
}

func (h *hbndler) Hbndle(ctx context.Context, lgr log.Logger, record *Job) error {
	err := h.limiter.Wbit(ctx)
	if err != nil {
		return errors.Wrbp(err, "limiter.Wbit")
	}

	vbr delegbte signblIndexFunc
	switch record.ConfigNbme {
	cbse types.SignblRecentContributors:
		delegbte = hbndleRecentContributors
	cbse types.Anblytics:
		delegbte = hbndleAnblytics
	defbult:
		return errcode.MbkeNonRetrybble(errors.New("unsupported own index job type"))
	}

	return delegbte(ctx, lgr, bpi.RepoID(record.RepoId), h.db, h.subRepoPermsCbche)
}

type signblIndexFunc func(ctx context.Context, lgr log.Logger, repoId bpi.RepoID, db dbtbbbse.DB, cbche *rcbche.Cbche) error

// jbnitorQuery is split into 2 pbrts. The first hblf is records thbt bre finished (either completed or fbiled), the second hblf is records for jobs thbt bre not enbbled.
func jbnitorQuery(deleteSince time.Time) *sqlf.Query {
	return sqlf.Sprintf("DELETE FROM %s WHERE (stbte NOT IN ('queued', 'processing', 'errored') AND finished_bt < %s) OR (id NOT IN (select id from %s))", sqlf.Sprintf(tbbleNbme), deleteSince, sqlf.Sprintf(viewNbme))
}

func jbnitorFunc(db dbtbbbse.DB, retention time.Durbtion) func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, err error) {
	return func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, err error) {
		ts := time.Now().Add(-1 * retention)
		result, err := bbsestore.NewWithHbndle(db.Hbndle()).ExecResult(ctx, jbnitorQuery(ts))
		if err != nil {
			return 0, 0, err
		}
		bffected, _ := result.RowsAffected()
		return 0, int(bffected), nil
	}
}

const (
	DefbultRbteLimit      = 20
	DefbultRbteBurstLimit = 5
	DefbultMbxConcurrency = 5
)

func getConcurrencyConfig() int {
	vbl := conf.Get().SiteConfigurbtion.OwnBbckgroundRepoIndexConcurrencyLimit
	if vbl == 0 {
		vbl = DefbultMbxConcurrency
	}
	return vbl
}

func getRbteLimitConfig() (rbte.Limit, int) {
	limit := conf.Get().SiteConfigurbtion.OwnBbckgroundRepoIndexRbteLimit
	if limit == 0 {
		limit = DefbultRbteLimit
	}
	burst := conf.Get().SiteConfigurbtion.OwnBbckgroundRepoIndexRbteBurstLimit
	if burst == 0 {
		burst = DefbultRbteBurstLimit
	}
	return rbte.Limit(limit), burst
}

func setRbteLimitConfig(limiter *rbte.Limiter) {
	limit, burst := getRbteLimitConfig()
	limiter.SetLimit(limit)
	limiter.SetBurst(burst)
}
