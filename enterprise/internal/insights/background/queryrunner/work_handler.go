package queryrunner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/hashicorp/go-multierror"

	"golang.org/x/time/rate"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

var _ workerutil.Handler = &workHandler{}

// workHandler implements the dbworker.Handler interface by executing search queries and
// inserting insights about them to the insights Timescale database.
type workHandler struct {
	baseWorkerStore *basestore.Store
	insightsStore   *store.Store
	metadadataStore *store.InsightStore
	limiter         *rate.Limiter

	mu          sync.RWMutex
	seriesCache map[string]*types.InsightSeries
}

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

func (r *workHandler) Handle(ctx context.Context, record workerutil.Record) (err error) {
	defer func() {
		if err != nil {
			log15.Error("insights.queryrunner.workHandler", "error", err)
		}
	}()

	err = r.limiter.Wait(ctx)
	if err != nil {
		return err
	}
	job, err := dequeueJob(ctx, r.baseWorkerStore, record.RecordID())
	if err != nil {
		return err
	}

	log15.Info("dequeue_job", "job", *job)

	series, err := r.getSeries(ctx, job.SeriesID)
	if err != nil {
		return err
	}

	// Actually perform the search query.
	//
	// 🚨 SECURITY: The request is performed without authentication, we get back results from every
	// repository on Sourcegraph - so we must be careful to only record insightful information that
	// is OK to expose to every user on Sourcegraph (e.g. total result counts are fine, exposing
	// that a repository exists may or may not be fine, exposing individual results is definitely
	// not, etc.)
	var results *gqlSearchResponse
	results, err = search(ctx, job.SearchQuery)
	if err != nil {
		return err
	}

	if len(results.Errors) > 0 {
		return errors.Errorf("GraphQL errors: %v", results.Errors)
	}
	if alert := results.Data.Search.Results.Alert; alert != nil {
		if alert.Title == "No repositories satisfied your repo: filter" {
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
			return errors.Errorf("insights query issue: alert: %v query=%q", alert, job.SearchQuery)
		}
	}
	if results.Data.Search.Results.LimitHit {
		log15.Error("insights query issue", "problem", "limit hit", "query", job.SearchQuery)
		dq := types.DirtyQuery{
			Query:   job.SearchQuery,
			ForTime: *job.RecordTime,
			Reason:  "limit hit",
		}
		if err := r.metadadataStore.InsertDirtyQuery(ctx, series, &dq); err != nil {
			return errors.Wrap(err, "failed to write dirty query record")
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

	// 🚨 SECURITY: The request is performed without authentication, we get back results from every
	// repository on Sourcegraph - so we must be careful to only record insightful information that
	// is OK to expose to every user on Sourcegraph (e.g. total result counts are fine, exposing
	// that a repository exists may just barely be fine, exposing individual results is definitely
	// not, etc.) OR record only data that we later restrict to only users who have access to those
	// repositories.
	recordTime := time.Now()
	if job.RecordTime != nil {
		recordTime = *job.RecordTime
	}

	// Figure out how many matches we got for every unique repository returned in the search
	// results.
	matchesPerRepo := make(map[string]int, len(results.Data.Search.Results.Results)*4)
	repoNames := make(map[string]string, len(matchesPerRepo))
	for _, result := range results.Data.Search.Results.Results {
		decoded, err := decodeResult(result)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf(`for query "%s"`, job.SearchQuery))
		}
		repoNames[decoded.repoID()] = decoded.repoName()
		matchesPerRepo[decoded.repoID()] = matchesPerRepo[decoded.repoID()] + decoded.matchCount()
	}

	tx, err := r.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if job.PersistMode == string(store.SnapshotMode) {
		// The purpose of the snapshot is for low fidelity but recently updated data points.
		// To avoid unbounded growth of the snapshots table we will prune it at the same time as adding new values.
		if err := tx.DeleteSnapshots(ctx, series); err != nil {
			return err
		}
	}

	// Record the number of results we got, one data point per-repository.
	for graphQLRepoID, matchCount := range matchesPerRepo {
		dbRepoID, idErr := graphqlbackend.UnmarshalRepositoryID(graphql.ID(graphQLRepoID))
		if idErr != nil {
			err = multierror.Append(err, errors.Wrap(idErr, "UnmarshalRepositoryID"))
			continue
		}
		repoName := repoNames[graphQLRepoID]
		if len(repoName) == 0 {
			// this really should never happen, expect if for some reason the gql response is broken
			err = multierror.Append(err, errors.Newf("MissingRepositoryName for repo_id: %v", string(dbRepoID)))
			continue
		}

		args := ToRecording(job, float64(matchCount), recordTime, repoName, dbRepoID)
		if recordErr := tx.RecordSeriesPoints(ctx, args); recordErr != nil {
			err = multierror.Append(err, errors.Wrap(recordErr, "RecordSeriesPoints"))
		}
	}
	return err
}

func ToRecording(record *Job, value float64, recordTime time.Time, repoName string, repoID api.RepoID) []store.RecordSeriesPointArgs {
	args := make([]store.RecordSeriesPointArgs, 0, len(record.DependentFrames)+1)
	base := store.RecordSeriesPointArgs{
		SeriesID: record.SeriesID,
		Point: store.SeriesPoint{
			SeriesID: record.SeriesID,
			Time:     recordTime,
			Value:    value,
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
