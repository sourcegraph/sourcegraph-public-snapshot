package queryrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func GetSearchHandlers() map[types.GenerationMethod]InsightsHandler {
	searchStream := func(ctx context.Context, query string) (*streaming.TabulationResult, error) {
		tr, ctx := trace.New(ctx, "CodeInsightsSearch.searchStream")
		defer tr.End()

		decoder, streamResults := streaming.TabulationDecoder()
		err := streaming.Search(ctx, query, nil, decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.Search")
		}
		tr.AddEvent("search results", attribute.Int("count", streamResults.TotalCount), attribute.Bool("timeout", streamResults.DidTimeout), attribute.Int("repo_count", len(streamResults.RepoCounts)))
		return streamResults, nil
	}

	computeSearchStream := func(ctx context.Context, query string) (*streaming.ComputeTabulationResult, error) {
		decoder, streamResults := streaming.MatchContextComputeDecoder()
		tr, ctx := trace.New(ctx, "CodeInsightsSearch.computeMatchContextSearchStream")
		defer tr.End()

		err := streaming.ComputeMatchContextStream(ctx, query, decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.Compute")
		}
		tr.AddEvent("compute match context results", attribute.Int("count", streamResults.TotalCount), attribute.Bool("timeout", streamResults.DidTimeout), attribute.Int("repo_count", len(streamResults.RepoCounts)))
		return streamResults, nil
	}

	computeTextExtraSearch := func(ctx context.Context, query string) (*streaming.ComputeTabulationResult, error) {
		decoder, streamResults := streaming.ComputeTextDecoder()
		err := streaming.ComputeTextExtraStream(ctx, query, decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.ComputeText")
		}
		return streamResults, nil
	}

	return map[types.GenerationMethod]InsightsHandler{
		types.MappingCompute: makeMappingComputeHandler(computeTextExtraSearch),
		types.SearchCompute:  makeComputeHandler(computeSearchStream),
		types.Search:         makeSearchHandler(searchStream),
	}

}

func toRecording(record *SearchJob, value float64, recordTime time.Time, repoName string, repoID api.RepoID, capture *string) []store.RecordSeriesPointArgs {
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
type streamSearchProvider func(context.Context, string) (*streaming.TabulationResult, error)

func generateComputeRecordingsStream(ctx context.Context, job *SearchJob, recordTime time.Time, provider streamComputeProvider, logger log.Logger) (_ []store.RecordSeriesPointArgs, err error) {
	streamResults, err := provider(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}
	if len(streamResults.SkippedReasons) > 0 {
		logger.Error("compute search encountered skipped events", log.String("seriesID", job.SeriesID), log.String("reasons", fmt.Sprintf("%v", streamResults.SkippedReasons)), log.String("query", job.SearchQuery))
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
		subRepoEnabled, subRepoErr := authz.SubRepoEnabledForRepoID(ctx, checker, api.RepoID(match.RepositoryID))
		if subRepoErr != nil {
			logger.Error("sub-repo permissions check errored", log.String("seriesID", job.SeriesID), log.String("repo", match.RepositoryName), log.Error(subRepoErr))
			continue
		}
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
			recordings = append(recordings, toRecording(job, float64(count), recordTime, match.RepositoryName, api.RepoID(match.RepositoryID), &capture)...)
		}
	}

	return recordings, nil
}

func generateSearchRecordingsStream(ctx context.Context, job *SearchJob, recordTime time.Time, provider streamSearchProvider, logger log.Logger) ([]store.RecordSeriesPointArgs, error) {
	tabulationResult, err := provider(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}

	tr := *tabulationResult
	if len(tr.SkippedReasons) > 0 {
		logger.Error("search encountered skipped events", log.String("seriesID", job.SeriesID), log.String("reasons", fmt.Sprintf("%v", tr.SkippedReasons)), log.String("query", job.SearchQuery))
	}
	if len(tr.Errors) > 0 {
		return nil, classifiedError(tr.Errors, types.Search)
	}
	if tr.DidTimeout {
		return nil, SearchTimeoutError
	}
	if len(tr.Alerts) > 0 {
		return nil, errors.Errorf("streaming search: alerts: %v", tr.Alerts)
	}

	checker := authz.DefaultSubRepoPermsChecker
	var recordings []store.RecordSeriesPointArgs

	for _, match := range tr.RepoCounts {
		// sub-repo permissions filtering. If the repo supports it, then it should be excluded from search results
		repoID := api.RepoID(match.RepositoryID)
		subRepoEnabled, subRepoErr := authz.SubRepoEnabledForRepoID(ctx, checker, repoID)
		if subRepoErr != nil {
			logger.Error("sub-repo permissions check errored", log.String("seriesID", job.SeriesID), log.String("repo", match.RepositoryName), log.Error(subRepoErr))
			continue
		}
		if subRepoEnabled {
			continue
		}
		recordings = append(recordings, toRecording(job, float64(match.MatchCount), recordTime, match.RepositoryName, repoID, nil)...)
	}

	return recordings, nil
}

func makeSearchHandler(provider streamSearchProvider) InsightsHandler {
	return func(ctx context.Context, job *SearchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		recordings, err := generateSearchRecordingsStream(ctx, job, recordTime, provider, log.Scoped("SearchRecordingsGenerator"))
		if err != nil {
			return nil, errors.Wrapf(err, "searchHandler")
		}
		return recordings, nil
	}
}

func makeComputeHandler(provider streamComputeProvider) InsightsHandler {
	return func(ctx context.Context, job *SearchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		computeDelegate := func(ctx context.Context, job *SearchJob, recordTime time.Time, logger log.Logger) (_ []store.RecordSeriesPointArgs, err error) {
			return generateComputeRecordingsStream(ctx, job, recordTime, provider, logger)
		}
		recordings, err := computeDelegate(ctx, job, recordTime, log.Scoped("ComputeRecordingsGenerator"))
		if err != nil {
			return nil, errors.Wrapf(err, "computeHandler")
		}
		return recordings, nil
	}
}

func makeMappingComputeHandler(provider streamComputeProvider) InsightsHandler {
	return func(ctx context.Context, job *SearchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		recordings, err := generateComputeRecordingsStream(ctx, job, recordTime, provider, log.Scoped("ComputeMappingRecordingsGenerator"))
		if err != nil {
			return nil, errors.Wrapf(err, "mappingComputeHandler")
		}
		return recordings, err
	}
}

func (r *workHandler) persistRecordings(ctx context.Context, job *SearchJob, series *types.InsightSeries, recordings []store.RecordSeriesPointArgs, recordTime time.Time) (err error) {
	tx, err := r.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	seriesRecordingTimes := types.InsightSeriesRecordingTimes{
		InsightSeriesID: series.ID,
	}
	snapshot := false
	if store.PersistMode(job.PersistMode) == store.SnapshotMode {
		// The purpose of the snapshot is for low fidelity but recently updated data points.
		// We store one snapshot of an insight at any time, so we prune the table whenever adding a new series.
		if err := tx.DeleteSnapshots(ctx, series); err != nil {
			return errors.Wrap(err, "DeleteSnapshots")
		}
		snapshot = true
	}
	seriesRecordingTimes.RecordingTimes = append(seriesRecordingTimes.RecordingTimes, types.RecordingTime{recordTime, snapshot})

	// Newly queued queries should be scoped to correct repos however leaving filtering
	// in place to ensure any older queued jobs get filtered properly. It's a noop for global insights.
	filteredRecordings, err := filterRecordingsBySeriesRepos(ctx, r.repoStore, series, recordings)
	if err != nil {
		return errors.Wrap(err, "filterRecordingsBySeriesRepos")
	}

	if recordErr := tx.RecordSeriesPointsAndRecordingTimes(ctx, filteredRecordings, seriesRecordingTimes); recordErr != nil {
		err = errors.Append(err, errors.Wrap(recordErr, "RecordSeriesPointsAndRecordingTimes"))
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
		if included := repos[*record.RepoID]; included {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return filteredRecords, nil

}
