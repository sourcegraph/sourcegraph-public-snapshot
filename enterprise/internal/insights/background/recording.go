package background

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type InsightRecorder struct {
	insightsDB            dbutil.DB
	mainDB                dbutil.DB
	enqueueQueryRunnerJob func(ctx context.Context, job *queryrunner.Job) error
	// stampFunc             func(ctx context.Context, insightSeries types.InsightSeries) (types.InsightSeries, error)
	queryRunnerEnqueueJob func(ctx context.Context, job *queryrunner.Job) error
}

func NewInsightRecorder(insightsDB dbutil.DB, mainDB dbutil.DB) *InsightRecorder {
	return &InsightRecorder{insightsDB: insightsDB, mainDB: mainDB, queryRunnerEnqueueJob: func(ctx context.Context, job *queryrunner.Job) error {
		_, err := queryrunner.EnqueueJob(ctx, basestore.NewWithDB(mainDB, sql.TxOptions{}), job)
		return err
	}}
}

type ManualRecordingArgs struct {
	Time     time.Time
	Value    float64
	RepoId   api.RepoID
	RepoName api.RepoName
}

func (i *InsightRecorder) ManualRecording(ctx context.Context, seriesId string, args ManualRecordingArgs) error {
	insightStore := store.NewInsightStore(i.insightsDB)
	timeSeriesStore := store.New(i.insightsDB, store.NewInsightPermissionStore(i.mainDB))

	seriesList, err := insightStore.GetDataSeries(ctx, store.GetDataSeriesArgs{SeriesID: seriesId})
	if err != nil {
		return errors.Wrap(err, "GetDataSeries")
	} else if len(seriesList) == 0 {
		return errors.Newf("unable to load missing series, series_id: %s", seriesId)
	}

	rn := string(args.RepoName)
	return timeSeriesStore.RecordSeriesPoint(ctx, store.RecordSeriesPointArgs{
		SeriesID: seriesId,
		Point: store.SeriesPoint{
			SeriesID: seriesId,
			Time:     args.Time,
			Value:    args.Value,
		},
		RepoName:    &rn,
		RepoID:      &args.RepoId,
		PersistMode: store.RecordMode,
	})
}

func (i *InsightRecorder) EnqueueGlobalRecording(ctx context.Context, seriesId string) error {
	return i.enqueueGlobal(ctx, seriesId, store.RecordMode)
}

func (i *InsightRecorder) EnqueueGlobalSnapshot(ctx context.Context, seriesId string) error {
	return i.enqueueGlobal(ctx, seriesId, store.SnapshotMode)
}

func (i *InsightRecorder) enqueueGlobal(ctx context.Context, seriesId string, mode store.PersistMode) (err error) {
	insightStore := store.NewInsightStore(i.insightsDB)
	seriesList, err := insightStore.GetDataSeries(ctx, store.GetDataSeriesArgs{SeriesID: seriesId})
	if err != nil {
		return errors.Wrap(err, "GetDataSeries")
	} else if len(seriesList) == 0 {
		return errors.Newf("unable to load missing series, series_id: %s", seriesId)
	}

	series := seriesList[0]
	err = i.enqueueQueryRunnerJob(ctx, &queryrunner.Job{
		SeriesID:    seriesId,
		SearchQuery: withCountUnlimited(series.Query),
		State:       "queued",
		Priority:    int(priority.High),
		Cost:        int(priority.Indexed),
		PersistMode: string(mode),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to enqueue insight series_id: %s", seriesId)
	}
	// The timestamp update can't be transactional because this is a separate database currently, so we will use
	// at-least-once semantics by waiting until the queue transaction is complete and without error.
	_, err = insightStore.StampRecording(ctx, series)
	if err != nil {
		return errors.Wrapf(err, "failed to stamp insight series_id: %s", seriesId)
	}
	return nil
}
