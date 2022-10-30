package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

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
		OrderByExpression: sqlf.Sprintf("id"), // todo
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
	logger.Info("inProgressHandler called", log.Int("recordId", record.RecordID()))

	ctx = actor.WithInternalActor(ctx)
	job := record.(*BaseJob)
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

	for true {
		repoId, more, finish := itr.NextWithFinish()
		if !more {
			break
		}

		repo, err := h.repoStore.Get(ctx, repoId)
		if err != nil {
			// TODO: this repo should be marked as errored and processing should continue
			// revisit when error handling added.
			return err
		}

		logger.Info("doing iteration work", log.Int("repo_id", int(repoId)))
		err = h.backfillRunner.Run(ctx, pipeline.BackfillRequest{Series: series, Repo: &types.MinimalRepo{ID: repo.ID, Name: repo.Name}, Frames: frames})
		if err != nil {
			// TODO: this repo should be marked as errored and processing should continue
			// revisit when error handling added.
			return err
		}

		err = finish(ctx, h.backfillStore.Store, nil)
		if err != nil {
			return err
		}
	}

	if err := h.insightsStore.SetInsightSeriesRecordingTimes(ctx, []itypes.InsightSeriesRecordingTimes{
		{
			InsightSeriesID: series.ID,
			RecordingTimes:  timeseries.MakeRecordingsFromFrames(frames, false),
		},
	}); err != nil {
		return err
	}

	// todo handle errors down here after the main loop https://github.com/sourcegraph/sourcegraph/issues/42724

	err = itr.MarkComplete(ctx, h.backfillStore.Store)
	if err != nil {
		return err
	}
	err = backfillJob.setState(ctx, h.backfillStore, BackfillStateCompleted)
	if err != nil {
		return err
	}

	return nil
}
