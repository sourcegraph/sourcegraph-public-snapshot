package pipeline

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func makeTestJobGenerator(numJobs int) SearchJobGenerator {
	return func(ctx context.Context, req requestContext) (context.Context, *requestContext, []*queryrunner.SearchJob, error) {
		jobs := make([]*queryrunner.SearchJob, 0, numJobs)
		for i := 0; i < numJobs; i++ {
			jobs = append(jobs, &queryrunner.SearchJob{
				SeriesID:    req.backfillRequest.Series.SeriesID,
				SearchQuery: "test search",
			})
		}
		return ctx, &req, jobs, nil
	}
}

func testSearchRunner(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.SearchJob, err error) (context.Context, *requestContext, []store.RecordSeriesPointArgs, error) {
	points := make([]store.RecordSeriesPointArgs, 0, len(jobs))
	for range jobs {
		points = append(points, store.RecordSeriesPointArgs{Point: store.SeriesPoint{Value: 10}})
	}
	return ctx, reqContext, points, nil
}

type testRunCounts struct {
	err         error
	resultCount int
	totalCount  int
}

func TestBackfillStepsConnected(t *testing.T) {

	testCases := []struct {
		numJobs int
		want    autogold.Value
	}{
		{10, autogold.Want("With Jobs", testRunCounts{resultCount: 10, totalCount: 100})},
		{0, autogold.Want("No Jobs", testRunCounts{})},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := testRunCounts{}
			countingPersister := func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs, err error) (*requestContext, error) {
				for _, p := range points {
					got.resultCount++
					got.totalCount += int(p.Point.Value)
				}
				return reqContext, nil
			}

			backfiller := NewBackfiller(makeTestJobGenerator(tc.numJobs), testSearchRunner, countingPersister)
			got.err = backfiller.Run(context.Background(), BackfillRequest{Series: &types.InsightSeries{SeriesID: "1"}})
			tc.want.Equal(t, got)
		})
	}
}

type fakeCommitClient struct {
	firstCommit   func(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error)
	recentCommits func(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error)
}

func (f *fakeCommitClient) FirstCommit(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error) {
	return f.firstCommit(ctx, repoName)
}
func (f *fakeCommitClient) RecentCommits(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error) {
	return f.recentCommits(ctx, repoName, target)
}

func newFakeCommitClient(first *gitdomain.Commit, recents []*gitdomain.Commit) gitCommitClient {
	return &fakeCommitClient{
		firstCommit: func(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error) { return first, nil },
		recentCommits: func(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error) {
			return recents, nil
		},
	}
}

func TestMakeSearchJobs(t *testing.T) {
	// Setup
	threeWeeks := 24 * 21 * time.Hour
	createdDate := time.Date(2022, time.April, 1, 1, 0, 0, 0, time.UTC)
	firstCommit := gitdomain.Commit{ID: "1", Committer: &gitdomain.Signature{}}
	recentFirstCommit := gitdomain.Commit{ID: "1", Committer: &gitdomain.Signature{}, Author: gitdomain.Signature{Date: createdDate.Add(-1 * threeWeeks)}}
	recentCommits := []*gitdomain.Commit{{ID: "1", Committer: &gitdomain.Signature{}}, {ID: "2", Committer: &gitdomain.Signature{}}}

	backfillReq := &BackfillRequest{
		Series: &types.InsightSeries{
			ID:                  1,
			SeriesID:            "abc",
			Query:               "test query",
			CreatedAt:           createdDate,
			SampleIntervalUnit:  string(types.Week),
			SampleIntervalValue: 1,
		},
		Repo: &itypes.MinimalRepo{ID: api.RepoID(1), Name: api.RepoName("testrepo")},
	}

	backfillReqInvalidQuery := backfillReq
	backfillReqInvalidQuery.Series.Query = "patterntype:regexp i++"

	backfillReqRepoQuery := backfillReq
	backfillReqRepoQuery.Series.Query = "test query repo:repoA"

	basicCommitClient := newFakeCommitClient(&firstCommit, recentCommits)
	// used to simulate a single call to recent commits failing
	recentsErrorAfter := func(times int, commits []*gitdomain.Commit) func(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error) {
		var called *int
		called = new(int)
		var mu sync.Mutex
		return func(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error) {
			mu.Lock()
			defer mu.Unlock()
			if *called >= times {
				return nil, errors.New("error hit")
			}
			*called++
			return commits, nil
		}
	}

	testCases := []struct {
		commitClient gitCommitClient
		backfillReq  *BackfillRequest
		workers      int
		cancled      bool
		want         autogold.Value
	}{
		{commitClient: basicCommitClient, backfillReq: backfillReq, workers: 1,
			want: autogold.Want("Base case single worker", []string{"error occured: false"})},
		{commitClient: basicCommitClient, backfillReq: backfillReq, workers: 5, want: autogold.Want("Base case multiple workers", []string{"error occured: false"})},
		{commitClient: newFakeCommitClient(&recentFirstCommit, recentCommits), backfillReq: backfillReq, workers: 1, want: autogold.Want("First commit during backfill period", []string{"error occured: false"})},
		{commitClient: newFakeCommitClient(&recentFirstCommit, recentCommits), backfillReq: backfillReq, workers: 5, want: autogold.Want("First commit during backfill period multiple workers", []string{"error occured: false"})},
		{commitClient: basicCommitClient, backfillReq: backfillReq, workers: 1, cancled: true, want: autogold.Want("Cancled case single worker", []string{"error occured: true"})},
		{commitClient: basicCommitClient, backfillReq: backfillReq, workers: 5, cancled: true, want: autogold.Want("Cancled case multiple workers", []string{"error occured: true"})},
		{commitClient: &fakeCommitClient{
			firstCommit: func(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error) {
				return nil, errors.New("somethings wrong")
			},
			recentCommits: basicCommitClient.RecentCommits,
		}, backfillReq: backfillReq, workers: 1, want: autogold.Want("First commit error", []string{"error occured: true"})},
		{commitClient: &fakeCommitClient{
			firstCommit: func(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error) {
				return nil, discovery.EmptyRepoErr
			},
			recentCommits: basicCommitClient.RecentCommits,
		}, backfillReq: backfillReq, workers: 1, want: autogold.Want("Empty repo", []string{"error occured: false"})},
		{commitClient: &fakeCommitClient{
			firstCommit:   basicCommitClient.FirstCommit,
			recentCommits: recentsErrorAfter(6, recentCommits),
		}, backfillReq: backfillReq, workers: 1, want: autogold.Want("Error in some jobs single worker", []string{"error occured: false"})},
		{commitClient: &fakeCommitClient{
			firstCommit:   basicCommitClient.FirstCommit,
			recentCommits: recentsErrorAfter(6, recentCommits),
		}, backfillReq: backfillReq, workers: 5, want: autogold.Want("Error in some jobs multiple worker", []string{"error occured: false"})},
		{commitClient: basicCommitClient, backfillReq: backfillReqInvalidQuery, workers: 1, want: autogold.Want("Invalid query", []string{"error occured: false"})},
		{commitClient: basicCommitClient, backfillReq: backfillReqRepoQuery, workers: 1, want: autogold.Want("Query with repo: in it ", []string{"error occured: false"})},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			testCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancled {
				cancel()
			}
			jobsFunc := makeSearchJobsFunc(logtest.NoOp(t), tc.commitClient, &compression.NoopFilter{}, tc.workers)
			_, _, jobs, err := jobsFunc(testCtx, requestContext{backfillRequest: tc.backfillReq})
			got := []string{}
			// sorted jobs to make test stable
			sort.SliceStable(jobs, func(i, j int) bool {
				return jobs[i].RecordTime.After(*jobs[j].RecordTime)
			})
			for _, j := range jobs {
				got = append(got, fmt.Sprintf("job recordtime:%s query:%s", j.RecordTime.Format(time.RFC3339Nano), j.SearchQuery))
			}
			got = append(got, fmt.Sprintf("error occured: %v", err != nil))
			tc.want.Equal(t, got)
		})
	}

}
