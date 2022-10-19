package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newBackfillHandler - Handles backfill that are in the "new" state
// The new state is the initial state post creation of a series.  This handler is responsible only for determining the work
// that needs to be completed to backfill this series.  It then requeues the backfill record into "processing" to perform the actual backfill work.
type newBackfillHandler struct {
	workerStore   dbworkerstore.Store
	backfillStore *BackfillStore
	seriesReader  SeriesReader
	repoIterator  discovery.SeriesRepoIterator
	costAnalyzer  priority.QueryAnalyzer
}

// makeNewBackfillWorker makes a new Worker, Resetter and Store to handle the queue of Backfill jobs that are in the state of "New"
func makeNewBackfillWorker(ctx context.Context, config JobMonitorConfig) (*workerutil.Worker, *dbworker.Resetter, dbworkerstore.Store) {
	insightsDB := config.InsightsDB
	backfillStore := newBackfillStore(insightsDB)

	name := "backfill_new_backfill_worker"

	workerStore := dbworkerstore.NewWithMetrics(insightsDB.Handle(), dbworkerstore.Options{
		Name:              fmt.Sprintf("%s_store", name),
		TableName:         "insights_background_jobs",
		ViewName:          "insights_jobs_backfill_new",
		ColumnExpressions: baseJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanBaseJob),
		OrderByExpression: sqlf.Sprintf("id"), // todo
		MaxNumResets:      100,
		StalledMaxAge:     time.Second * 30,
	}, config.ObsContext)

	task := newBackfillHandler{
		workerStore:   workerStore,
		backfillStore: backfillStore,
		seriesReader:  store.NewInsightStore(insightsDB),
		repoIterator:  discovery.NewSeriesRepoIterator(nil, config.RepoStore), //TODO add in a real all repos iterator
	}

	worker := dbworker.NewWorker(ctx, workerStore, &task, workerutil.WorkerOptions{
		Name:        name,
		NumHandlers: 1,
		Interval:    5 * time.Second,
		Metrics:     workerutil.NewMetrics(config.ObsContext, name),
	})

	resetter := dbworker.NewResetter(log.Scoped("BackfillNewResetter", ""), workerStore, dbworker.ResetterOptions{
		Name:     fmt.Sprintf("%s_resetter", name),
		Interval: time.Second * 20,
		Metrics:  *dbworker.NewMetrics(config.ObsContext, name),
	})

	return worker, resetter, workerStore
}

var _ workerutil.Handler = &newBackfillHandler{}

func (h *newBackfillHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
	logger.Info("newBackfillHandler called", log.Int("recordId", record.RecordID()))
	job, ok := record.(*BaseJob)
	if !ok {
		return errors.New("invalid job received")
	}
	// setup transactions

	tx, err := h.backfillStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		tx.Done(err)
	}()

	// load backfill and series
	backfill, err := tx.loadBackfill(ctx, job.backfillId)
	if err != nil {
		return errors.Wrap(err, "loadBackfill")
	}
	series, err := h.seriesReader.GetDataSeriesByID(ctx, backfill.SeriesId)
	if err != nil {
		return errors.Wrap(err, "GetDataSeriesByID")
	}

	// set backfill repo scope
	repoIds := []int32{}
	reposIterator, err := h.repoIterator.ForSeries(ctx, series)
	if err != nil {
		return errors.Wrap(err, "repoIterator.SeriesRepoIterator")
	}
	err = reposIterator.ForEach(ctx, func(repoName string, id api.RepoID) error {
		repoIds = append(repoIds, int32(id))
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "reposIterator.ForEach")
	}

	//TODO: use query costing
	backfill, err = backfill.SetScope(ctx, tx, repoIds, 0)
	if err != nil {
		return errors.Wrap(err, "backfill.SetScope")
	}

	// update series state
	err = backfill.setState(ctx, tx, BackfillStateProcessing)
	if err != nil {
		return errors.Wrap(err, "backfill.setState")
	}

	// enqueue backfill for next step in processing
	return enqueueBackfill(ctx, tx.Handle(), backfill)
}
