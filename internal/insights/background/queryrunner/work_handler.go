package queryrunner

import (
	"context"
	"sync"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ workerutil.Handler[*Job] = &workHandler{}

// workHandler implements the dbworker.Handler interface by executing search queries and
// inserting insights about them to the insights database.
type workHandler struct {
	baseWorkerStore *workerStoreExtra
	insightsStore   *store.Store
	repoStore       discovery.RepoStore
	metadadataStore *store.InsightStore
	limiter         *ratelimit.InstrumentedLimiter
	logger          log.Logger

	mu          sync.RWMutex
	seriesCache map[string]*types.InsightSeries

	searchHandlers map[types.GenerationMethod]InsightsHandler
}

type InsightsHandler func(ctx context.Context, job *SearchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error)

func (r *workHandler) getSeries(ctx context.Context, seriesID string) (*types.InsightSeries, error) {
	var val *types.InsightSeries
	var ok bool

	r.mu.RLock()
	val, ok = r.seriesCache[seriesID]
	r.mu.RUnlock()

	if !ok {
		series, err := r.fetchSeries(ctx, seriesID)
		if err != nil {
			return nil, err
		} else if series == nil {
			return nil, errors.Newf("workHandler.getSeries: insight definition not found for series_id: %s", seriesID)
		}

		r.mu.Lock()
		defer r.mu.Unlock()
		r.seriesCache[seriesID] = series
		val = series
	}
	return val, nil
}

func (r *workHandler) fetchSeries(ctx context.Context, seriesID string) (*types.InsightSeries, error) {
	result, err := r.metadadataStore.GetDataSeries(ctx, store.GetDataSeriesArgs{SeriesID: seriesID})
	if err != nil || len(result) < 1 {
		return nil, err
	}
	return &result[0], nil
}

func (r *workHandler) Handle(ctx context.Context, logger log.Logger, record *Job) (err error) {
	// 🚨 SECURITY: The request is performed without authentication, we get back results from every
	// repository on Sourcegraph - results will be filtered when users query for insight data based on the
	// repositories they can see.
	isGlobal := false
	if record.RecordTime == nil {
		isGlobal = true
	}

	ctx = actor.WithInternalActor(ctx)
	defer func() {
		if err != nil {
			r.logger.Error("insights.queryrunner.workHandler", log.Error(err))
		}
	}()
	ss := basestore.NewWithHandle(r.baseWorkerStore.Handle())
	// storing trace with query for debugging
	traceID := trace.ID(ctx)
	if traceID != "" {
		// intentionally ignoring error
		ss.Exec(ctx, sqlf.Sprintf("update insights_query_runner_jobs set trace_id = %s where id = %s", traceID, record.RecordID()))
	}

	err = r.limiter.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "limiter.Wait")
	}
	job, err := dequeueJob(ctx, ss, record.RecordID())
	if err != nil {
		return errors.Wrap(err, "dequeueJob")
	}

	series, err := r.getSeries(ctx, job.SeriesID)
	if err != nil {
		return errors.Wrap(err, "getSeries")
	}
	if series.JustInTime {
		return errors.Newf("just in time series are not eligible for background processing, series_id: %s", series.ID)
	}

	recordTime := time.Now()
	if job.RecordTime != nil {
		recordTime = *job.RecordTime
	}

	executableHandler, ok := r.searchHandlers[series.GenerationMethod]
	if !ok {
		return errors.Newf("unable to handle record for series_id: %s and generation_method: %s", series.SeriesID, series.GenerationMethod)
	}

	recordings, err := executableHandler(ctx, &job.SearchJob, series, recordTime)
	if err != nil {
		if !r.baseWorkerStore.WillRetry(job) && isGlobal && job.PersistMode == string(store.RecordMode) {
			reason := TranslateIncompleteReasons(err)
			logger.Debug("insights recording global query timeout",
				log.Int("seriesId", series.ID), log.String("seriesUniqueId", series.SeriesID),
				log.Error(err),
				log.String("reason", string(reason)))

			if addErr := r.insightsStore.AddIncompleteDatapoint(ctx, store.AddIncompleteDatapointInput{
				SeriesID: series.ID,
				Reason:   reason,
				Time:     recordTime,
			}); addErr != nil {
				return errors.Append(err, errors.Wrap(addErr, "workHandler.AddIncompleteDatapoint"))
			}
		}
		return err
	}

	return r.persistRecordings(ctx, &job.SearchJob, series, recordings, recordTime)
}

func TranslateIncompleteReasons(err error) store.IncompleteReason {
	if errors.Is(err, SearchTimeoutError) {
		return store.ReasonTimeout
	}
	return store.ReasonGeneric
}
