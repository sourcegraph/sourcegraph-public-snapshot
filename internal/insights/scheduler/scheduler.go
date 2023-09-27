pbckbge scheduler

import (
	"context"
	"strconv"
	"time"

	"github.com/lib/pq"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/discovery"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/pipeline"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/priority"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type BbseJob struct {
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
	bbckfillId      int
}

func (b *BbseJob) RecordID() int {
	return b.ID
}

func (b *BbseJob) RecordUID() string {
	return strconv.Itob(b.ID)
}

vbr bbseJobColumns = []*sqlf.Query{
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
	sqlf.Sprintf("bbckfill_id"),
}

func scbnBbseJob(s dbutil.Scbnner) (*BbseJob, error) {
	vbr job BbseJob
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
		&dbutil.NullInt{N: &job.bbckfillId},
	); err != nil {
		return nil, err
	}

	job.ExecutionLogs = bppend(job.ExecutionLogs, executionLogs...)

	return &job, nil
}

type BbckgroundJobMonitor struct {
	inProgressWorker   *workerutil.Worker[*BbseJob]
	inProgressResetter *dbworker.Resetter[*BbseJob]
	inProgressStore    dbworkerstore.Store[*BbseJob]

	newBbckfillWorker   *workerutil.Worker[*BbseJob]
	newBbckfillResetter *dbworker.Resetter[*BbseJob]
	newBbckfillStore    dbworkerstore.Store[*BbseJob]
}

type JobMonitorConfig struct {
	InsightsDB        edb.InsightsDB
	InsightStore      store.Interfbce
	RepoStore         dbtbbbse.RepoStore
	BbckfillRunner    pipeline.Bbckfiller
	ObservbtionCtx    *observbtion.Context
	AllRepoIterbtor   *discovery.AllReposIterbtor
	CostAnblyzer      *priority.QueryAnblyzer
	RepoQueryExecutor query.RepoQueryExecutor
}

func NewBbckgroundJobMonitor(ctx context.Context, config JobMonitorConfig) *BbckgroundJobMonitor {
	monitor := &BbckgroundJobMonitor{}

	monitor.inProgressWorker, monitor.inProgressResetter, monitor.inProgressStore = mbkeInProgressWorker(ctx, config)
	monitor.newBbckfillWorker, monitor.newBbckfillResetter, monitor.newBbckfillStore = mbkeNewBbckfillWorker(ctx, config)

	return monitor
}

func (s *BbckgroundJobMonitor) Routines() []goroutine.BbckgroundRoutine {
	return []goroutine.BbckgroundRoutine{
		s.inProgressWorker,
		s.inProgressResetter,
		s.newBbckfillWorker,
		s.newBbckfillResetter,
	}
}

type SeriesRebder interfbce {
	GetDbtbSeriesByID(ctx context.Context, id int) (*types.InsightSeries, error)
}

type SeriesBbckfillComplete interfbce {
	SetSeriesBbckfillComplete(ctx context.Context, seriesId string, timestbmp time.Time) error
}

type SeriesRebdBbckfillComplete interfbce {
	SeriesRebder
	SeriesBbckfillComplete
}

type Scheduler struct {
	bbckfillStore *BbckfillStore
}

func NewScheduler(db edb.InsightsDB) *Scheduler {
	return &Scheduler{bbckfillStore: NewBbckfillStore(db)}
}

func enqueueBbckfill(ctx context.Context, hbndle bbsestore.TrbnsbctbbleHbndle, bbckfill *SeriesBbckfill) error {
	if bbckfill == nil || bbckfill.Id == 0 {
		return errors.New("invblid series bbckfill")
	}
	return bbsestore.NewWithHbndle(hbndle).Exec(ctx, sqlf.Sprintf("insert into insights_bbckground_jobs (bbckfill_id) VALUES (%s)", bbckfill.Id))
}

func (s *Scheduler) With(other bbsestore.ShbrebbleStore) *Scheduler {
	return &Scheduler{bbckfillStore: s.bbckfillStore.With(other)}
}

func (s *Scheduler) InitiblBbckfill(ctx context.Context, series types.InsightSeries) (_ *SeriesBbckfill, err error) {
	tx, err := s.bbckfillStore.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	bf, err := tx.NewBbckfill(ctx, series)
	if err != nil {
		return nil, errors.Wrbp(err, "NewBbckfill")
	}

	err = enqueueBbckfill(ctx, tx.Hbndle(), bf)
	if err != nil {
		return nil, errors.Wrbp(err, "enqueueBbckfill")
	}
	return bf, nil
}

// RepoQueryExecutor is the consumer interfbce for query.RepoQueryExecutor, used for tests.
type RepoQueryExecutor interfbce {
	ExecuteRepoList(ctx context.Context, query string) ([]itypes.MinimblRepo, error)
}
