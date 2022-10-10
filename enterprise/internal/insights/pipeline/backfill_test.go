package pipeline

import (
	"context"
	"fmt"
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

// type SearchJobGenerator func(ctx context.Context, req requestContext) (context.Context, *requestContext, []*queryrunner.Job, error)
// type SearchRunner func(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.Job, err error) (context.Context, *requestContext, []store.RecordSeriesPointArgs, error)
// type ResultsPersister func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs, err error) (*requestContext, error)

func makeJobGenerator(numJobs int) SearchJobGenerator {
	return func(ctx context.Context, req requestContext) (context.Context, *requestContext, []*queryrunner.Job, error) {
		jobs := make([]*queryrunner.Job, 0, numJobs)
		for i := 0; i < numJobs; i++ {
			jobs = append(jobs, &queryrunner.Job{
				SeriesID:    req.backfillRequest.Series.SeriesID,
				SearchQuery: fmt.Sprintf("%d", i),
			})
		}
		return ctx, &req, jobs, nil
	}
}

func searchRunner(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.Job, err error) (context.Context, *requestContext, []store.RecordSeriesPointArgs, error) {
	points := make([]store.RecordSeriesPointArgs, 0, len(jobs))
	for _, _ = range jobs {
		points = append(points, store.RecordSeriesPointArgs{Point: store.SeriesPoint{Value: 10}})
	}
	return ctx, reqContext, points, nil
}

type runCounts struct {
	err         error
	resultCount int
	totalCount  int
}

func TestBackfillStepsConnected(t *testing.T) {

	testCases := []struct {
		numJobs int
		want    autogold.Value
	}{
		{10, autogold.Want("With Jobs", runCounts{resultCount: 10, totalCount: 100})},
		{0, autogold.Want("No Jobs", runCounts{})},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := runCounts{}
			countingPersister := func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs, err error) (*requestContext, error) {
				for _, p := range points {
					got.resultCount++
					got.totalCount += int(p.Point.Value)
				}
				return reqContext, nil
			}

			backfiller := NewBackfiller(makeJobGenerator(tc.numJobs), searchRunner, countingPersister)
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
	firstCommit := gitdomain.Commit{ID: "1", Committer: &gitdomain.Signature{}}
	recentCommits := []*gitdomain.Commit{{ID: "1", Committer: &gitdomain.Signature{}}, {ID: "2", Committer: &gitdomain.Signature{}}}
	createdDate := time.Date(2022, time.April, 1, 1, 0, 0, 0, time.UTC)
	backfillReq := &BackfillRequest{
		Series: &types.InsightSeries{
			ID:                  1,
			SeriesID:            "abc",
			Query:               "test query",
			CreatedAt:           createdDate,
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
		},
		Repo: &itypes.MinimalRepo{ID: api.RepoID(1), Name: api.RepoName("testrepo")},
	}
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
			want: autogold.Want("Base case single worker", []string{
				"job recordtime:2022-03-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-01-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2021-12-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2021-11-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2021-10-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2021-09-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2021-08-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2021-07-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2021-06-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2021-05-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
				"error occured: false",
			})},
		{commitClient: basicCommitClient, backfillReq: backfillReq, workers: 5, want: autogold.Want("Base case multiple workers", []string{
			"job recordtime:2021-11-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-10-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-12-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-11-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-11-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-09-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-08-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-06-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-05-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-11-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"job recordtime:2021-06-01T00:00:00Z query:fork:no archived:no patterntype:literal count:99999999 test query repo:^testrepo$@1",
			"error occured: false",
		})},
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
		}, backfillReq: backfillReq, workers: 1, want: autogold.Want("Error in some jobs single worker", []string{"error occured: true"})},
		{commitClient: &fakeCommitClient{
			firstCommit:   basicCommitClient.FirstCommit,
			recentCommits: recentsErrorAfter(6, recentCommits),
		}, backfillReq: backfillReq, workers: 5, want: autogold.Want("Error in some jobs multiple worker", []string{"error occured: true"})},
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
			for _, j := range jobs {
				got = append(got, fmt.Sprintf("job recordtime:%s query:%s", j.RecordTime.Format(time.RFC3339Nano), j.SearchQuery))
			}
			got = append(got, fmt.Sprintf("error occured: %v", err != nil))
			tc.want.Equal(t, got)
		})
	}

}
