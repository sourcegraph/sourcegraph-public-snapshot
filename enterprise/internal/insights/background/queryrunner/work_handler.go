package queryrunner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

	mu          sync.RWMutex
	seriesCache map[string]*types.InsightSeries

	searchStream func(context.Context, string) (*streaming.TabulationResult, error)

	computeSearchStream    func(context.Context, string) (*streaming.ComputeTabulationResult, error)
	computeTextExtraSearch func(context.Context, string) (*streaming.ComputeTabulationResult, error)
}

type insightsHandler func(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error)

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

// checkSubRepoPermissions returns true if the repo has sub-repo permissions or any error occurred while checking it
// Returns false only if the repo doesn't have sub-repo permissions or these are disabled in settings.
// Note that repo ID is received untyped and being cast to api.RepoID
// err is an upstream error to which any new occurring error is appended
func checkSubRepoPermissions(ctx context.Context, checker authz.SubRepoPermissionChecker, untypedRepoID any, err error) (bool, error) {
	if !authz.SubRepoEnabled(checker) {
		return false, err
	}

	// casting repoID
	var repoID api.RepoID
	switch untypedRepoID := untypedRepoID.(type) {
	case api.RepoID:
		repoID = untypedRepoID
	case string:
		var idErr error
		repoID, idErr = graphqlbackend.UnmarshalRepositoryID(graphql.ID(untypedRepoID))
		if idErr != nil {
			log15.Error("Error during sub-repo permissions check", "repoID", untypedRepoID, "error", "unmarshalling repoID")
			err = errors.Append(err, errors.Wrap(idErr, "Checking sub-repo permissions: UnmarshalRepositoryID"))
			return true, err
		}
	default:
		log15.Error("Error during sub-repo permissions check: Unsupported untypedRepoID type",
			"repoID", untypedRepoID, "type", fmt.Sprintf("%T", untypedRepoID))
		return true, errors.Append(err, errors.Newf("Checking sub-repo permissions for repoID=%v: Unsupported untypedRepoID type=%T",
			untypedRepoID, untypedRepoID))
	}

	// performing the check itself
	enabled, checkErr := authz.SubRepoEnabledForRepoID(ctx, checker, repoID)
	if checkErr != nil {
		log15.Error("Error during sub-repo permissions check", "error", checkErr)
		err = errors.Append(err, errors.Wrap(checkErr, "Checking sub-repo permissions"))
		return true, err
	}
	return enabled, err
}

func ToRecording(record *Job, value float64, recordTime time.Time, repoName string, repoID api.RepoID, capture *string) []store.RecordSeriesPointArgs {
	args := make([]store.RecordSeriesPointArgs, 0, len(record.DependentFrames)+1)
	base := store.RecordSeriesPointArgs{
		SeriesID: record.SeriesID,
		Point: store.SeriesPoint{
			SeriesID: record.SeriesID,
			Time:     recordTime,
			Value:    value,
			Capture:  capture,
		},
		RepoName:    &repoName,
		RepoID:      &repoID,
		PersistMode: store.PersistMode(record.PersistMode),
	}
	args = append(args, base)
	for _, dependent := range record.DependentFrames {
		arg := base
		arg.Point.Time = dependent
		args = append(args, arg)
	}
	return args
}

type streamComputeProvider func(context.Context, string) (*streaming.ComputeTabulationResult, error)

func (r *workHandler) generateComputeRecordingsStream(ctx context.Context, job *Job, recordTime time.Time, provider streamComputeProvider) (_ []store.RecordSeriesPointArgs, err error) {
	streamResults, err := provider(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}
	if len(streamResults.Errors) > 0 {
		return nil, classifiedError(streamResults.Errors, types.SearchCompute)
	}
	if len(streamResults.Alerts) > 0 {
		return nil, errors.Errorf("compute streaming search: alerts: %v", streamResults.Alerts)
	}

	checker := authz.DefaultSubRepoPermsChecker
	var recordings []store.RecordSeriesPointArgs

	for _, match := range streamResults.RepoCounts {
		var subRepoEnabled bool
		subRepoEnabled, err = checkSubRepoPermissions(ctx, checker, match.RepositoryID, err)
		if subRepoEnabled {
			continue
		}

		for capturedValue, count := range match.ValueCounts {
			capture := capturedValue
			if len(capture) == 0 {
				// there seems to be some behavior where empty string values get returned from the compute API. We will just skip them. If there are future changes
				// to fix this, we will automatically pick up any new results without changes here.
				continue
			}
			recordings = append(recordings, ToRecording(job, float64(count), recordTime, match.RepositoryName, api.RepoID(match.RepositoryID), &capture)...)
		}
	}
	return recordings, nil
}

func (r *workHandler) generateSearchRecordingsStream(ctx context.Context, job *Job, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
	tabulationResult, err := r.searchStream(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}

	tr := *tabulationResult

	log15.Info("Search Counts", "streaming", tr.TotalCount)
	if len(tr.SkippedReasons) > 0 {
		log15.Error("insights query issue", "reasons", tr.SkippedReasons, "query", job.SearchQuery)
	}
	if len(tr.Errors) > 0 {
		return nil, classifiedError(tr.Errors, types.Search)
	}
	if len(tr.Alerts) > 0 {
		return nil, errors.Errorf("streaming search: alerts: %v", tr.Alerts)
	}

	checker := authz.DefaultSubRepoPermsChecker
	var recordings []store.RecordSeriesPointArgs

	for _, match := range tr.RepoCounts {
		// sub-repo permissions filtering. If the repo supports it, then it should be excluded from search results
		var subRepoEnabled bool
		repoID := api.RepoID(match.RepositoryID)
		subRepoEnabled, err = checkSubRepoPermissions(ctx, checker, repoID, err)
		if subRepoEnabled {
			continue
		}
		recordings = append(recordings, ToRecording(job, float64(match.MatchCount), recordTime, match.RepositoryName, repoID, nil)...)
	}
	return recordings, nil
}

func (r *workHandler) searchHandler(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
	recordings, err := r.generateSearchRecordingsStream(ctx, job, recordTime)
	if err != nil {
		return nil, errors.Wrapf(err, "searchHandler")
	}
	return recordings, nil
}

func (r *workHandler) computeHandler(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
	computeDelegate := func(ctx context.Context, job *Job, recordTime time.Time) (_ []store.RecordSeriesPointArgs, err error) {
		return r.generateComputeRecordingsStream(ctx, job, recordTime, r.computeSearchStream)
	}
	recordings, err := computeDelegate(ctx, job, recordTime)
	if err != nil {
		return nil, errors.Wrapf(err, "computeHandler")
	}
	return recordings, nil
}

func (r *workHandler) mappingComputeHandler(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
	recordings, err := r.generateComputeRecordingsStream(ctx, job, recordTime, r.computeTextExtraSearch)
	if err != nil {
		return nil, errors.Wrapf(err, "mappingComputeHandler")
	}
	return recordings, err
}

func (r *workHandler) persistRecordings(ctx context.Context, job *Job, series *types.InsightSeries, recordings []store.RecordSeriesPointArgs) (err error) {
	tx, err := r.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if store.PersistMode(job.PersistMode) == store.SnapshotMode {
		// The purpose of the snapshot is for low fidelity but recently updated data points.
		// We store one snapshot of an insight at any time, so we prune the table whenever adding a new series.
		if err := tx.DeleteSnapshots(ctx, series); err != nil {
			return err
		}
	}

	// Newly queued queries should be scoped to correct repos however leaving filtering
	// in place to ensure any older queued jobs get filtered properly. It's a noop for global insights.
	filteredRecordings, err := filterRecordingsBySeriesRepos(ctx, r.repoStore, series, recordings)
	if err != nil {
		return errors.Wrap(err, "filterRecordingsBySeriesRepos")
	}

	if recordErr := tx.RecordSeriesPoints(ctx, filteredRecordings); recordErr != nil {
		err = errors.Append(err, errors.Wrap(recordErr, "RecordSeriesPointsCapture"))
	}
	return err
}

func filterRecordingsBySeriesRepos(ctx context.Context, repoStore discovery.RepoStore, series *types.InsightSeries, recordings []store.RecordSeriesPointArgs) ([]store.RecordSeriesPointArgs, error) {
	// If this series isn't scoped to some repos return all
	if len(series.Repositories) == 0 {
		return recordings, nil
	}

	seriesRepos, err := repoStore.List(ctx, database.ReposListOptions{Names: series.Repositories})
	if err != nil {
		return nil, errors.Wrap(err, "repoStore.List")
	}
	repos := map[api.RepoID]bool{}
	for _, repo := range seriesRepos {
		repos[repo.ID] = true
	}

	filteredRecords := make([]store.RecordSeriesPointArgs, 0, len(series.Repositories))
	for _, record := range recordings {
		if record.RepoID == nil {
			continue
		}
		if included := repos[*record.RepoID]; included == true {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return filteredRecords, nil

}

func (r *workHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
	// 🚨 SECURITY: The request is performed without authentication, we get back results from every
	// repository on Sourcegraph - results will be filtered when users query for insight data based on the
	// repositories they can see.
	ctx = actor.WithInternalActor(ctx)
	defer func() {
		if err != nil {
			log15.Error("insights.queryrunner.workHandler", "error", err)
		}
	}()
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

	handlersByType := map[types.GenerationMethod]insightsHandler{
		types.SearchCompute:  r.computeHandler,
		types.MappingCompute: r.mappingComputeHandler,
		types.Search:         r.searchHandler,
	}

	executableHandler, ok := handlersByType[series.GenerationMethod]
	if !ok {
		return errors.Newf("unable to handle record for series_id: %s and generation_method: %s", series.SeriesID, series.GenerationMethod)
	}

	recordings, err := executableHandler(ctx, job, series, recordTime)
	if err != nil {
		return err
	}
	return r.persistRecordings(ctx, job, series, recordings)
}
