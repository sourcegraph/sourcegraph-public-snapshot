package queryrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var _ dbworker.Handler = &workHandler{}

// workHandler implements the dbworker.Handler interface by executing search queries and
// inserting insights about them to the insights Timescale database.
type workHandler struct {
	workerBaseStore *basestore.Store
	insightsStore   *store.Store
}

func (r *workHandler) Handle(ctx context.Context, workerStore dbworkerstore.Store, record workerutil.Record) (err error) {
	defer func() {
		if err != nil {
			log15.Error("insights.queryrunner.workHandler", "error", err)
		}
	}()

	// Dequeue the job to get information about it, like what search query to perform.
	job, err := dequeueJob(ctx, r.workerBaseStore, record.RecordID())
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

	// TODO(slimsag): future: Logs are not a good way to surface these errors to users.
	if len(results.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %v", results.Errors)
	}
	if alert := results.Data.Search.Results.Alert; alert != nil {
		return fmt.Errorf("insights query issue: alert: %v query=%q", alert, job.SearchQuery)
	}
	if results.Data.Search.Results.LimitHit {
		log15.Error("insights query issue", "problem", "limit hit", "query", job.SearchQuery)
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

	// Record the match count to the insights DB.
	var matchCount int
	if results != nil {
		matchCount = results.Data.Search.Results.MatchCount
	}

	log15.Info("insights query runner", "found matches", matchCount, "query", job.SearchQuery)

	// 🚨 SECURITY: The request is performed without authentication, we get back results from every
	// repository on Sourcegraph - so we must be careful to only record insightful information that
	// is OK to expose to every user on Sourcegraph (e.g. total result counts are fine, exposing
	// that a repository exists may or may not be fine, exposing individual results is definitely
	// not, etc.)
	return r.insightsStore.RecordSeriesPoint(ctx, &store.RecordSeriesPointArgs{
		SeriesID: job.SeriesID,
		Point: store.SeriesPoint{
			Time:  time.Now(),
			Value: float64(matchCount),
		},
		// TODO(slimsag): future: determine match count per repository, and store RepoID/RepoName (and maybe Metadata?)
	})
}
