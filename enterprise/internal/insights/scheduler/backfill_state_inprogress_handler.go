package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/scheduler/iterator"
	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/pipeline"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	itypes "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func makeInProgressWorker(ctx context.Context, config JobMonitorConfig) (*workerutil.Worker, *dbworker.Resetter, dbworkerstore.Store) {
	db := config.InsightsDB
	backfillStore := NewBackfillStore(db)

	name := "backfill_in_progress_worker"

	workerStore := dbworkerstore.NewWithMetrics(db.Handle(), dbworkerstore.Options{
		Name:              fmt.Sprintf("%s_store", name),
		TableName:         "insights_background_jobs",
		ViewName:          "insights_jobs_backfill_in_progress",
		ColumnExpressions: baseJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanBaseJob),
		OrderByExpression: sqlf.Sprintf("cost_bucket, id"), // take the oldest item in the group of least work
		MaxNumResets:      100,
		StalledMaxAge:     time.Second * 30,
		RetryAfter:        time.Second * 30,
		MaxNumRetries:     3,
	}, config.ObsContext)

	task := inProgressHandler{
		workerStore:    workerStore,
		backfillStore:  backfillStore,
		seriesReader:   store.NewInsightStore(db),
		insightsStore:  config.InsightStore,
		backfillRunner: config.BackfillRunner,
		repoStore:      config.RepoStore,
	}

	worker := dbworker.NewWorker(ctx, workerStore, &task, workerutil.WorkerOptions{
		Name:              name,
		NumHandlers:       1,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(config.ObsContext, name),
	})

	resetter := dbworker.NewResetter(log.Scoped("", ""), workerStore, dbworker.ResetterOptions{
		Name:     fmt.Sprintf("%s_resetter", name),
		Interval: time.Second * 20,
		Metrics:  *dbworker.NewMetrics(config.ObsContext, name),
	})

	return worker, resetter, workerStore
}

type inProgressHandler struct {
	workerStore    dbworkerstore.Store
	backfillStore  *BackfillStore
	seriesReader   SeriesReader
	repoStore      database.RepoStore
	insightsStore  store.Interface
	backfillRunner pipeline.Backfiller
}

var _ workerutil.Handler = &inProgressHandler{}

func (h *inProgressHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	ctx = actor.WithInternalActor(ctx)
	job, ok := record.(*BaseJob)
	if !ok {
		return errors.New("unable to convert to backfill inprogress job")
	}
	backfillJob, err := h.backfillStore.loadBackfill(ctx, job.backfillId)
	if err != nil {
		return errors.Wrap(err, "loadBackfill")
	}
	series, err := h.seriesReader.GetDataSeriesByID(ctx, backfillJob.SeriesId)
	if err != nil {
		return errors.Wrap(err, "GetDataSeriesByID")
	}

	itr, err := backfillJob.repoIterator(ctx, h.backfillStore)
	if err != nil {
		return errors.Wrap(err, "repoIterator")
	}

	frames := timeseries.BuildFrames(12, timeseries.TimeInterval{
		Unit:  itypes.IntervalUnit(series.SampleIntervalUnit),
		Value: series.SampleIntervalValue,
	}, series.CreatedAt.Truncate(time.Hour*24))

	logger.Info("insights backfill progress handler loaded",
		log.Int("recordId", record.RecordID()),
		log.Int("jobNumFailures", job.NumFailures),
		log.Int("seriesId", series.ID),
		log.String("seriesUniqueId", series.SeriesID),
		log.Int("backfillId", backfillJob.Id),
		log.Int("repoTotalCount", itr.TotalCount),
		log.Float64("percentComplete", itr.PercentComplete),
		log.Int("erroredRepos", itr.ErroredRepos()),
		log.Int("totalErrors", itr.TotalErrors()))

	type nextFunc func() (api.RepoID, bool, iterator.FinishFunc)
	itrLoop := func(nextFunc nextFunc) error {
		for {
			repoId, more, finish := nextFunc()
			if !more {
				break
			}

			repo, repoErr := h.repoStore.Get(ctx, repoId)
			if repoErr != nil {
				err = finish(ctx, h.backfillStore.Store, errors.Wrap(repoErr, "InProgressHandler.repoStore.Get"))
				if err != nil {
					return err
				}
				continue
			}

			logger.Debug("doing iteration work", log.Int("repo_id", int(repoId)))
			runErr := h.backfillRunner.Run(ctx, pipeline.BackfillRequest{Series: series, Repo: &types.MinimalRepo{ID: repo.ID, Name: repo.Name}, Frames: frames})
			if runErr != nil {
				logger.Error("error during backfill execution", log.Int("seriesId", series.ID), log.Int("backfillId", backfillJob.Id), log.Error(runErr))
			}
			err = finish(ctx, h.backfillStore.Store, runErr)
			if err != nil {
				return err
			}
		}
		return nil
	}

	logger.Debug("starting primary loop", log.Int("seriesId", series.ID), log.Int("backfillId", backfillJob.Id))
	if err := itrLoop(itr.NextWithFinish); err != nil {
		return errors.Wrap(err, "InProgressHandler.PrimaryLoop")
	}

	logger.Debug("starting retry loop", log.Int("seriesId", series.ID), log.Int("backfillId", backfillJob.Id))
	if err := itrLoop(itr.NextRetryWithFinish); err != nil {
		return errors.Wrap(err, "InProgressHandler.RetryLoop")
	}

	if !itr.HasMore() && !itr.HasErrors() {
		logger.Info("setting backfill to completed state", log.Int("seriesId", series.ID), log.String("seriesUniqueId", series.SeriesID), log.Int("backfillId", backfillJob.Id), log.Duration("totalDuration", itr.RuntimeDuration))
		err = itr.MarkComplete(ctx, h.backfillStore.Store)
		if err != nil {
			return err
		}
		err = backfillJob.setState(ctx, h.backfillStore, BackfillStateCompleted)
		if err != nil {
			return err
		}
	} else {
		// this is a rudimentary way of getting this job to retry. Eventually we should manually queue up work so that
		// we aren't bound by the retry limits placed on the queue, but for now this will work.
		return incompleteBackfillErr
	}

	return nil
}

var incompleteBackfillErr error = errors.New("incomplete backfill")
