package queryrunner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var _ workerutil.Handler = &workHandler{}

// workHandler implements the dbworker.Handler interface by executing search queries and
// inserting insights about them to the insights database.
type workHandler struct {
	baseWorkerStore *basestore.Store
	insightsStore   *store.Store
	repoStore       discovery.RepoStore
	metadadataStore *store.InsightStore
	limiter         *rate.Limiter

	mu          sync.RWMutex
	seriesCache map[string]*types.InsightSeries

	search       func(context.Context, string) (*query.GqlSearchResponse, error)
	searchStream func(context.Context, string) (*streaming.TabulationResult, error)

	computeSearch       func(context.Context, string) ([]query.ComputeResult, error)
	computeSearchStream func(context.Context, string) (*streaming.ComputeTabulationResult, error)
}

type insightsHandler func(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) error

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

func (r *workHandler) generateComputeRecordings(ctx context.Context, job *Job, recordTime time.Time) (_ []store.RecordSeriesPointArgs, err error) {
	results, err := r.computeSearch(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}

	checker := authz.DefaultSubRepoPermsChecker
	var recordings []store.RecordSeriesPointArgs

	groupedByRepo := query.GroupByRepository(results)
	for repoKey, byRepo := range groupedByRepo {
		groupedByCapture := query.GroupByCaptureMatch(byRepo)
		repoId, idErr := graphqlbackend.UnmarshalRepositoryID(graphql.ID(repoKey))
		if idErr != nil {
			err = errors.Append(err, errors.Wrap(idErr, "UnmarshalRepositoryIDCapture"))
			continue
		}
		// sub-repo permissions filtering. If the repo supports it, then it should be excluded from search results
		var subRepoEnabled bool
		subRepoEnabled, err = checkSubRepoPermissions(ctx, checker, repoId, err)
		if subRepoEnabled {
			continue
		}
		for _, group := range groupedByCapture {
			capture := group.Value
			recordings = append(recordings, ToRecording(job, float64(group.Count), recordTime, byRepo[0].RepoName(), repoId, &capture)...)
		}
	}
	return recordings, nil
}

func (r *workHandler) generateComputeRecordingsStream(ctx context.Context, job *Job, recordTime time.Time) (_ []store.RecordSeriesPointArgs, err error) {
	streamResults, err := r.computeSearchStream(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}
	if len(streamResults.Errors) > 0 {
		return nil, StreamingError{Type: types.SearchCompute, Messages: streamResults.Errors}
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
			recordings = append(recordings, ToRecording(job, float64(count), recordTime, match.RepositoryName, api.RepoID(match.RepositoryID), &capture)...)
		}
	}
	return recordings, nil
}

func (r *workHandler) generateSearchRecordings(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
	results, err := r.search(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}

	if len(results.Errors) > 0 {
		return nil, errors.Errorf("GraphQL errors: %v", results.Errors)
	}
	if alert := results.Data.Search.Results.Alert; alert != nil {
		if alert.Title == "No repositories found" {
			// We got zero results and no repositories matched. This could be for a few reasons:
			//
			// 1. The repo hasn't been cloned by Sourcegraph yet.
			// 2. The repo has been cloned by Sourcegraph, but the user hasn't actually pushed it
			//    to the code host yet so it's empty.
			// 3. This is a search query for backfilling data, and the repository is a fork/archive
			//    which are excluded from search results by default (and the user didn't put `fork:yes`
			//    etc. in their search query.)
			//
			// In any case, this is not a problem - we want to record that we got zero results in
			// general.
		} else {
			// Maybe the user's search query is actually wrong.
			return nil, errors.Errorf("insights query issue: alert: %v query=%q", alert, job.SearchQuery)
		}
	}
	if results.Data.Search.Results.LimitHit {
		log15.Error("insights query issue", "problem", "limit hit", "query", job.SearchQuery)
		dq := types.DirtyQuery{
			Query:   job.SearchQuery,
			ForTime: recordTime,
			Reason:  "limit hit",
		}
		if err := r.metadadataStore.InsertDirtyQuery(ctx, series, &dq); err != nil {
			return nil, errors.Wrap(err, "failed to write dirty query record")
		}
	}
	if cloning := len(results.Data.Search.Results.Cloning); cloning > 0 {
		log15.Error("insights query issue", "cloning_repos", cloning, "query", job.SearchQuery)
	}
	if missing := len(results.Data.Search.Results.Missing); missing > 0 {
		log15.Error("insights query issue", "missing_repos", missing, "query", job.SearchQuery)
	}
	if timedout := len(results.Data.Search.Results.Timedout); timedout > 0 {
		log15.Error("insights query issue", "timedout_repos", timedout, "query", job.SearchQuery)
	}

	checker := authz.DefaultSubRepoPermsChecker
	var recordings []store.RecordSeriesPointArgs

	matchesPerRepo := make(map[string]int, len(results.Data.Search.Results.Results)*4)
	repoNames := make(map[string]string, len(matchesPerRepo))
	for _, result := range results.Data.Search.Results.Results {
		decoded, err := query.DecodeResult(result)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(`for query "%s"`, job.SearchQuery))
		}
		// sub-repo permissions filtering. If the repo supports it, then it should be excluded from search results
		var subRepoEnabled bool
		subRepoEnabled, _ = checkSubRepoPermissions(ctx, checker, decoded.RepoID(), err)
		if subRepoEnabled {
			continue
		}
		repoNames[decoded.RepoID()] = decoded.RepoName()
		matchesPerRepo[decoded.RepoID()] = matchesPerRepo[decoded.RepoID()] + decoded.MatchCount()
	}

	// Record the number of results we got, one data point per-repository.
	for graphQLRepoID, matchCount := range matchesPerRepo {
		dbRepoID, idErr := graphqlbackend.UnmarshalRepositoryID(graphql.ID(graphQLRepoID))
		if idErr != nil {
			err = errors.Append(err, errors.Wrap(idErr, "UnmarshalRepositoryID"))
			continue
		}
		repoName := repoNames[graphQLRepoID]
		if len(repoName) == 0 {
			// this really should never happen, expect if for some reason the gql response is broken
			err = errors.Append(err, errors.Newf("MissingRepositoryName for repo_id: %v", string(dbRepoID)))
			continue
		}

		recordings = append(recordings, ToRecording(job, float64(matchCount), recordTime, repoName, dbRepoID, nil)...)
	}
	return recordings, nil
}

func (r *workHandler) generateSearchRecordingsStream(ctx context.Context, job *Job, _ *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
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
		return nil, StreamingError{Messages: tr.Errors}
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

func (r *workHandler) searchHandler(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) (err error) {
	if series.JustInTime {
		return errors.Newf("just in time series are not eligible for background processing, series_id: %s", series.ID)
	}

	searchDelegate := r.generateSearchRecordingsStream
	useGraphQL := conf.Get().InsightsSearchGraphql
	if useGraphQL != nil && *useGraphQL {
		searchDelegate = r.generateSearchRecordings
	}

	recordings, err := searchDelegate(ctx, job, series, recordTime)
	if err != nil {
		return err
	}

	err = r.persistRecordings(ctx, job, series, recordings)
	return err
}

func (r *workHandler) computeHandler(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) (err error) {
	if series.JustInTime {
		return errors.Newf("just in time series are not eligible for background processing, series_id: %s", series.ID)
	}

	computeDelegate := r.generateComputeRecordingsStream
	useGraphQL := conf.Get().InsightsComputeGraphql
	if useGraphQL != nil && *useGraphQL {
		computeDelegate = r.generateComputeRecordings
	}

	recordings, err := computeDelegate(ctx, job, recordTime)
	if err != nil {
		return err
	}

	err = r.persistRecordings(ctx, job, series, recordings)
	return err
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

	recordTime := time.Now()
	if job.RecordTime != nil {
		recordTime = *job.RecordTime
	}

	handlersByType := map[types.GenerationMethod]insightsHandler{
		types.SearchCompute: r.computeHandler,
		types.Search:        r.searchHandler,
	}

	executableHandler, ok := handlersByType[series.GenerationMethod]
	if !ok {
		return errors.Newf("unable to handle record for series_id: %s and generation_method: %s", series.SeriesID, series.GenerationMethod)
	}
	return executableHandler(ctx, job, series, recordTime)
}
