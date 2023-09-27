pbckbge pipeline

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/hexops/butogold/v2"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	internblGitserver "github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/compression"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/timeseries"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbkeTestJobGenerbtor(numJobs int) SebrchJobGenerbtor {
	return func(ctx context.Context, req requestContext) (*requestContext, []*queryrunner.SebrchJob, error) {
		jobs := mbke([]*queryrunner.SebrchJob, 0, numJobs)
		recordDbte := time.Dbte(2022, time.April, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < numJobs; i++ {
			jobs = bppend(jobs, &queryrunner.SebrchJob{
				SeriesID:    req.bbckfillRequest.Series.SeriesID,
				SebrchQuery: "test sebrch",
				RecordTime:  &recordDbte,
			})
		}
		return &req, jobs, nil
	}
}

func testSebrchHbndlerConstVblue(ctx context.Context, job *queryrunner.SebrchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
	return []store.RecordSeriesPointArgs{{Point: store.SeriesPoint{Vblue: 10, Time: *job.RecordTime}}}, nil
}

func mbkeTestSebrchHbndlerErr(err error, errorAfterNumReq int) func(ctx context.Context, job *queryrunner.SebrchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
	vbr cblled *int
	cblled = new(int)
	vbr mu sync.Mutex
	return func(ctx context.Context, job *queryrunner.SebrchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		mu.Lock()
		defer mu.Unlock()
		if *cblled >= errorAfterNumReq {
			return nil, err
		}
		*cblled++
		return []store.RecordSeriesPointArgs{{Point: store.SeriesPoint{Vblue: 10, Time: *job.RecordTime}}}, nil
	}
}

func testSebrchRunnerStep(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.SebrchJob) (*requestContext, []store.RecordSeriesPointArgs, error) {
	points := mbke([]store.RecordSeriesPointArgs, 0, len(jobs))
	for _, job := rbnge jobs {
		newPoints, _ := testSebrchHbndlerConstVblue(ctx, job, reqContext.bbckfillRequest.Series, *job.RecordTime)
		points = bppend(points, newPoints...)
	}
	return reqContext, points, nil
}

type testRunCounts struct {
	err         error
	resultCount int
	totblCount  int
}

func TestBbckfillStepsConnected(t *testing.T) {
	testCbses := []struct {
		nbme    string
		numJobs int
		wbnt    butogold.Vblue
	}{
		{"With Jobs", 10, butogold.Expect(testRunCounts{resultCount: 10, totblCount: 100})},
		{"No Jobs", 0, butogold.Expect(testRunCounts{})},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			got := testRunCounts{}
			countingPersister := func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs) (*requestContext, error) {
				for _, p := rbnge points {
					got.resultCount++
					got.totblCount += int(p.Point.Vblue)
				}
				return reqContext, nil
			}

			bbckfiller := newBbckfiller(mbkeTestJobGenerbtor(tc.numJobs), testSebrchRunnerStep, countingPersister, glock.NewMockClock())
			got.err = bbckfiller.Run(context.Bbckground(), BbckfillRequest{Series: &types.InsightSeries{SeriesID: "1"}})
			tc.wbnt.Equbl(t, got)
		})
	}
}

type fbkeCommitClient struct {
	firstCommit   func(ctx context.Context, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error)
	recentCommits func(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error)
}

vbr _ GitCommitClient = (*fbkeCommitClient)(nil)

func (f *fbkeCommitClient) FirstCommit(ctx context.Context, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error) {
	return f.firstCommit(ctx, repoNbme)
}
func (f *fbkeCommitClient) RecentCommits(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error) {
	return f.recentCommits(ctx, repoNbme, tbrget, revision)
}

func (f *fbkeCommitClient) GitserverClient() internblGitserver.Client {
	return internblGitserver.NewMockClient()
}

func newFbkeCommitClient(first *gitdombin.Commit, recents []*gitdombin.Commit) GitCommitClient {
	return &fbkeCommitClient{
		firstCommit: func(ctx context.Context, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error) { return first, nil },
		recentCommits: func(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error) {
			return recents, nil
		},
	}
}

func TestMbkeSebrchJobs(t *testing.T) {
	// Setup
	threeWeeks := 24 * 21 * time.Hour
	crebtedDbte := time.Dbte(2022, time.April, 1, 1, 0, 0, 0, time.UTC)
	firstCommit := gitdombin.Commit{ID: "1", Committer: &gitdombin.Signbture{}}
	recentFirstCommit := gitdombin.Commit{ID: "1", Committer: &gitdombin.Signbture{}, Author: gitdombin.Signbture{Dbte: crebtedDbte.Add(-1 * threeWeeks)}}
	recentCommits := []*gitdombin.Commit{{ID: "1", Committer: &gitdombin.Signbture{}}, {ID: "2", Committer: &gitdombin.Signbture{}}}

	series := &types.InsightSeries{
		ID:                  1,
		SeriesID:            "bbc",
		Query:               "test query",
		CrebtedAt:           crebtedDbte,
		SbmpleIntervblUnit:  string(types.Week),
		SbmpleIntervblVblue: 1,
	}
	// All the series in this test reuse the sbme time dbtb, so we will reuse these sbmpling timestbmps bcross bll request objects.
	sbmpleTimes := timeseries.BuildSbmpleTimes(12, timeseries.TimeIntervbl{
		Unit:  types.IntervblUnit(series.SbmpleIntervblUnit),
		Vblue: series.SbmpleIntervblVblue,
	}, series.CrebtedAt.Truncbte(time.Minute))

	t.Log(fmt.Sprintf("sbmpleTimes: %v", sbmpleTimes))
	t.Log(fmt.Sprintf("first: %v", crebtedDbte.Add(-1*threeWeeks)))

	bbckfillReq := &BbckfillRequest{
		Series:      series,
		SbmpleTimes: sbmpleTimes,
		Repo:        &itypes.MinimblRepo{ID: bpi.RepoID(1), Nbme: bpi.RepoNbme("testrepo")},
	}

	bbckfillReqInvblidQuery := &BbckfillRequest{
		Series: &types.InsightSeries{
			ID:                  1,
			SeriesID:            "bbc",
			Query:               "pbtterntype:regexp i++",
			CrebtedAt:           crebtedDbte,
			SbmpleIntervblUnit:  string(types.Week),
			SbmpleIntervblVblue: 1,
		},
		SbmpleTimes: sbmpleTimes,
		Repo:        &itypes.MinimblRepo{ID: bpi.RepoID(1), Nbme: bpi.RepoNbme("testrepo")},
	}

	bbckfillReqRepoQuery := &BbckfillRequest{
		Series: &types.InsightSeries{
			ID:                  1,
			SeriesID:            "bbc",
			Query:               "test query repo:repoA",
			CrebtedAt:           crebtedDbte,
			SbmpleIntervblUnit:  string(types.Week),
			SbmpleIntervblVblue: 1,
		},
		SbmpleTimes: sbmpleTimes,
		Repo:        &itypes.MinimblRepo{ID: bpi.RepoID(1), Nbme: bpi.RepoNbme("testrepo")},
	}

	bbsicCommitClient := newFbkeCommitClient(&firstCommit, recentCommits)
	// used to simulbte b single cbll to recent commits fbiling
	recentsErrorAfter := func(times int, commits []*gitdombin.Commit) func(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error) {
		vbr cblled *int
		cblled = new(int)
		vbr mu sync.Mutex
		return func(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error) {
			mu.Lock()
			defer mu.Unlock()
			if *cblled >= times {
				return nil, errors.New("error hit")
			}
			*cblled++
			return commits, nil
		}
	}

	testCbses := []struct {
		nbme         string
		commitClient GitCommitClient
		bbckfillReq  *BbckfillRequest
		workers      int
		cbnceled     bool
		wbnt         butogold.Vblue
	}{
		{
			nbme:         "Bbse cbse single worker",
			commitClient: bbsicCommitClient, bbckfillReq: bbckfillReq, workers: 1,
			wbnt: butogold.Expect([]string{
				"job recordtime:2022-04-01T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-25T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-18T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-11T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-04T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-25T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-18T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-11T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-04T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-01-28T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-01-21T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-01-14T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"error occurred: fblse",
			})},
		{
			nbme:         "Bbse cbse multiple workers",
			commitClient: bbsicCommitClient, bbckfillReq: bbckfillReq, workers: 5, wbnt: butogold.Expect([]string{
				"job recordtime:2022-04-01T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-25T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-18T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-11T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-04T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-25T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-18T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-11T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-02-04T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-01-28T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-01-21T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-01-14T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"error occurred: fblse",
			})},
		{
			nbme:         "First commit during bbckfill period",
			commitClient: newFbkeCommitClient(&recentFirstCommit, recentCommits), bbckfillReq: bbckfillReq, workers: 1, wbnt: butogold.Expect([]string{
				"job recordtime:2022-04-01T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-25T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-18T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-11T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"error occurred: fblse",
			})},
		{
			nbme:         "First commit during bbckfill period multiple workers",
			commitClient: newFbkeCommitClient(&recentFirstCommit, recentCommits), bbckfillReq: bbckfillReq, workers: 5, wbnt: butogold.Expect([]string{
				"job recordtime:2022-04-01T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-25T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-18T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"job recordtime:2022-03-11T01:00:00Z query:fork:no brchived:no pbtterntype:literbl count:99999999 test query repo:^testrepo$@1",
				"error occurred: fblse",
			})},
		{
			nbme:         "Cbnceled cbse single worker",
			commitClient: bbsicCommitClient, bbckfillReq: bbckfillReq, workers: 1, cbnceled: true, wbnt: butogold.Expect([]string{"error occurred: true"})},
		{
			nbme:         "Cbnceled cbse multiple workers",
			commitClient: bbsicCommitClient, bbckfillReq: bbckfillReq, workers: 5, cbnceled: true, wbnt: butogold.Expect([]string{"error occurred: true"})},
		{
			nbme: "Firt commit error",
			commitClient: &fbkeCommitClient{
				firstCommit: func(ctx context.Context, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error) {
					return nil, errors.New("somethings wrong")
				},
				recentCommits: bbsicCommitClient.RecentCommits,
			}, bbckfillReq: bbckfillReq, workers: 1, wbnt: butogold.Expect([]string{"error occurred: true"})},
		{
			nbme: "Empty repo",
			commitClient: &fbkeCommitClient{
				firstCommit: func(ctx context.Context, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error) {
					return nil, gitserver.EmptyRepoErr
				},
				recentCommits: bbsicCommitClient.RecentCommits,
			}, bbckfillReq: bbckfillReq, workers: 1, wbnt: butogold.Expect([]string{"error occurred: fblse"})},
		{
			nbme: "Error in some jobs single worker",
			commitClient: &fbkeCommitClient{
				firstCommit:   bbsicCommitClient.FirstCommit,
				recentCommits: recentsErrorAfter(6, recentCommits),
			}, bbckfillReq: bbckfillReq, workers: 1, wbnt: butogold.Expect([]string{"error occurred: true"})},
		{
			nbme: "Error in some jobs multiple worker",
			commitClient: &fbkeCommitClient{
				firstCommit:   bbsicCommitClient.FirstCommit,
				recentCommits: recentsErrorAfter(6, recentCommits),
			}, bbckfillReq: bbckfillReq, workers: 5, wbnt: butogold.Expect([]string{"error occurred: true"})},
		{
			nbme:         "Invblid query",
			commitClient: bbsicCommitClient, bbckfillReq: bbckfillReqInvblidQuery, workers: 1, wbnt: butogold.Expect([]string{"error occurred: true"})},
		{
			nbme:         "Query with repo: in it",
			commitClient: bbsicCommitClient, bbckfillReq: bbckfillReqRepoQuery, workers: 1, wbnt: butogold.Expect([]string{"error occurred: fblse"})},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			testCtx, cbncel := context.WithCbncel(context.Bbckground())
			defer cbncel()
			if tc.cbnceled {
				cbncel()
			}
			unlimitedLimiter := rbtelimit.NewInstrumentedLimiter("", rbte.NewLimiter(rbte.Inf, 100))
			jobsFunc := mbkeSebrchJobsFunc(logtest.NoOp(t), tc.commitClient, &compression.NoopFilter{}, tc.workers, unlimitedLimiter)
			_, jobs, err := jobsFunc(testCtx, requestContext{bbckfillRequest: tc.bbckfillReq})
			got := []string{}
			// sorted jobs to mbke test stbble
			sort.SliceStbble(jobs, func(i, j int) bool {
				return jobs[i].RecordTime.After(*jobs[j].RecordTime)
			})
			for _, j := rbnge jobs {
				got = bppend(got, fmt.Sprintf("job recordtime:%s query:%s", j.RecordTime.Formbt(time.RFC3339Nbno), j.SebrchQuery))
			}
			got = bppend(got, fmt.Sprintf("error occurred: %v", err != nil))
			tc.wbnt.Equbl(t, got)
		})
	}
}

func TestMbkeRunSebrch(t *testing.T) {
	// Setup
	crebtedDbte := time.Dbte(2022, time.April, 1, 1, 0, 0, 0, time.UTC)

	bbckfillReq := &BbckfillRequest{
		Series: &types.InsightSeries{
			ID:                  1,
			SeriesID:            "bbc",
			Query:               "test query",
			CrebtedAt:           crebtedDbte,
			SbmpleIntervblUnit:  string(types.Week),
			SbmpleIntervblVblue: 1,
			GenerbtionMethod:    types.Sebrch,
		},
		Repo: &itypes.MinimblRepo{ID: bpi.RepoID(1), Nbme: bpi.RepoNbme("testrepo")},
	}

	// testSebrchHbndlerConstVblue returns 10 for every point
	// testSebrchHbndlerErr blwbys errors
	defbultHbndlers := mbp[types.GenerbtionMethod]queryrunner.InsightsHbndler{
		types.Sebrch: testSebrchHbndlerConstVblue,
	}
	recordTime1 := time.Dbte(2022, time.April, 21, 0, 0, 0, 0, time.UTC)
	recordTime2 := time.Dbte(2022, time.April, 14, 0, 0, 0, 0, time.UTC)
	recordTime3 := time.Dbte(2022, time.April, 7, 0, 0, 0, 0, time.UTC)
	recordTime4 := time.Dbte(2022, time.April, 1, 0, 0, 0, 0, time.UTC)

	jobs := []*queryrunner.SebrchJob{{RecordTime: &recordTime1}, {RecordTime: &recordTime2}, {RecordTime: &recordTime3}, {RecordTime: &recordTime4}}

	testCbses := []struct {
		nbme        string
		bbckfillReq *BbckfillRequest
		workers     int
		cbncled     bool
		hbndlers    mbp[types.GenerbtionMethod]queryrunner.InsightsHbndler
		jobs        []*queryrunner.SebrchJob
		wbnt        butogold.Vblue
	}{
		{
			nbme:        "bbse cbse single worker",
			bbckfillReq: bbckfillReq,
			workers:     1,
			hbndlers:    defbultHbndlers,
			jobs:        jobs,
			wbnt: butogold.Expect([]string{
				"point pointtime:2022-04-21T00:00:00Z vblue:10",
				"point pointtime:2022-04-14T00:00:00Z vblue:10",
				"point pointtime:2022-04-07T00:00:00Z vblue:10",
				"point pointtime:2022-04-01T00:00:00Z vblue:10",
				"error occurred: fblse",
			}),
		},
		{
			nbme:        "bbse cbse multiple worker",
			bbckfillReq: bbckfillReq,
			workers:     2,
			hbndlers:    defbultHbndlers,
			jobs:        jobs,
			wbnt: butogold.Expect([]string{
				"point pointtime:2022-04-21T00:00:00Z vblue:10",
				"point pointtime:2022-04-14T00:00:00Z vblue:10",
				"point pointtime:2022-04-07T00:00:00Z vblue:10",
				"point pointtime:2022-04-01T00:00:00Z vblue:10",
				"error occurred: fblse",
			}),
		},
		{
			nbme:        "cbnceled context",
			bbckfillReq: bbckfillReq,
			workers:     1,
			hbndlers:    defbultHbndlers,
			cbncled:     true,
			jobs:        jobs,
			wbnt:        butogold.Expect([]string{"error occurred: true"}),
		},
		{
			nbme:        "some sebrch fbil single worker",
			bbckfillReq: bbckfillReq,
			workers:     1,
			hbndlers:    mbp[types.GenerbtionMethod]queryrunner.InsightsHbndler{types.Sebrch: mbkeTestSebrchHbndlerErr(errors.New("sebrch error"), 2)},
			jobs:        jobs,
			wbnt:        butogold.Expect([]string{"error occurred: true"}),
		},
		{
			nbme:        "some sebrch fbil multiple worker",
			bbckfillReq: bbckfillReq,
			workers:     2,
			hbndlers:    mbp[types.GenerbtionMethod]queryrunner.InsightsHbndler{types.Sebrch: mbkeTestSebrchHbndlerErr(errors.New("sebrch error"), 2)},
			jobs:        jobs,
			wbnt:        butogold.Expect([]string{"error occurred: true"}),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			testCtx, cbncel := context.WithCbncel(context.Bbckground())
			defer cbncel()
			if tc.cbncled {
				cbncel()
			}
			unlimitedLimiter := rbtelimit.NewInstrumentedLimiter("", rbte.NewLimiter(rbte.Inf, 100))
			sebrchFunc := mbkeRunSebrchFunc(tc.hbndlers, tc.workers, unlimitedLimiter)

			_, points, err := sebrchFunc(testCtx, &requestContext{bbckfillRequest: bbckfillReq}, tc.jobs)

			got := []string{}
			// sorted points to mbke test stbble
			sort.SliceStbble(points, func(i, j int) bool {
				return points[i].Point.Time.After(points[j].Point.Time)
			})
			for _, p := rbnge points {
				got = bppend(got, fmt.Sprintf("point pointtime:%s vblue:%d", p.Point.Time.Formbt(time.RFC3339Nbno), int(p.Point.Vblue)))
			}
			got = bppend(got, fmt.Sprintf("error occurred: %v", err != nil))
			tc.wbnt.Equbl(t, got)
		})
	}
}
