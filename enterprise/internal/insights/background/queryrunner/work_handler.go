package queryrunner

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"
)

var _ workerutil.Handler = &workHandler{}

// workHandler implements the dbworker.Handler interface by executing search queries and
// inserting insights about them to the insights database.
type workHandler struct {
	baseWorkerStore *basestore.Store
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

func (r *workHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
	// 🚨 SECURITY: The request is performed without authentication, we get back results from every
	// repository on Sourcegraph - results will be filtered when users query for insight data based on the
	// repositories they can see.

	ctx = actor.WithInternalActor(ctx)
	defer func() {
		if err != nil {
			r.logger.Error("insights.queryrunner.workHandler", log.Error(err))
		}
	}()

	// storing trace with query for debugging
	traceID := trace.ID(ctx)
	if traceID != "" {
		// intentionally ignoring error
		r.baseWorkerStore.Exec(ctx, sqlf.Sprintf("update insights_query_runner_jobs set trace_id = %s where id = %s", traceID, record.RecordID()))
	}

	err = r.limiter.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "limiter.Wait")
	}
	job, err := dequeueJob(ctx, r.baseWorkerStore, record.RecordID())
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
		return err
	}
	return r.persistRecordings(ctx, &job.SearchJob, series, recordings, recordTime)
}

type zoektHealthChecker struct {
	samples []int
	index   int
	healthy bool

	client zoekt.Streamer
	mu     sync.RWMutex
	logger log.Logger

	gauge prometheus.Gauge
}

func newZoektHealthChecker(client zoekt.Streamer, logger log.Logger, bufferSize int) *zoektHealthChecker {
	buf := make([]int, bufferSize)

	for i := range buf {
		buf[i] = -1
	}
	return &zoektHealthChecker{client: client, logger: logger, samples: buf}
}

func (z *zoektHealthChecker) isHealthy() bool {
	z.mu.RLock()
	defer z.mu.RUnlock()
	return z.healthy
}

func (z *zoektHealthChecker) updateHealthy() {
	current := z.samples[0]
	for i := 1; i < len(z.samples); i++ {
		if z.samples[i] != current {
			z.healthy = false
			return
		}
		current = z.samples[i]
	}
	z.healthy = true
	z.setGauge()
}

func (z *zoektHealthChecker) setGauge() {
	if z.gauge == nil {
		return
	}
	var val float64
	if z.healthy {
		val = 1.0
	}
	z.gauge.Set(val)
}

func (z *zoektHealthChecker) next() int {
	z.index = (z.index + 1) % len(z.samples)
	return z.index
}

func (z *zoektHealthChecker) sample(ctx context.Context) {
	idx := z.next()
	z.mu.Lock()
	defer z.mu.Unlock()

	results, err := z.client.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Minimal: true})
	if err != nil {
		z.samples[idx] = -1
		z.updateHealthy()
		return
	}
	statz, _ := json.Marshal(results.Stats)
	z.samples[idx] = results.Stats.Shards
	z.updateHealthy()
	z.logger.Info("code insights zoekt health check", log.Bool("healthy", z.healthy), log.String("stats", string(statz)))
}

func (z *zoektHealthChecker) Handle(ctx context.Context) error {
	z.sample(ctx)
	return nil
}
