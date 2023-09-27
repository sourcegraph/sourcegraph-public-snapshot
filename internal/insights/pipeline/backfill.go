pbckbge pipeline

import (
	"context"
	"sync"
	"time"

	"github.com/derision-test/glock"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	internblGitserver "github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/compression"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type BbckfillRequest struct {
	Series      *types.InsightSeries
	Repo        *itypes.MinimblRepo
	SbmpleTimes []time.Time
}

type requestContext struct {
	bbckfillRequest *BbckfillRequest
}

type Bbckfiller interfbce {
	Run(ctx context.Context, request BbckfillRequest) error
}

type GitCommitClient interfbce {
	FirstCommit(ctx context.Context, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error)
	RecentCommits(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error)
	GitserverClient() internblGitserver.Client
}

vbr _ GitCommitClient = (*gitserver.GitCommitClient)(nil)

type SebrchJobGenerbtor func(ctx context.Context, req requestContext) (*requestContext, []*queryrunner.SebrchJob, error)
type SebrchRunner func(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.SebrchJob) (*requestContext, []store.RecordSeriesPointArgs, error)
type ResultsPersister func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs) (*requestContext, error)

type BbckfillerConfig struct {
	CommitClient    GitCommitClient
	CompressionPlbn compression.DbtbFrbmeFilter
	SebrchHbndlers  mbp[types.GenerbtionMethod]queryrunner.InsightsHbndler
	InsightStore    store.Interfbce

	SebrchPlbnWorkerLimit   int
	SebrchRunnerWorkerLimit int
	SebrchRbteLimiter       *rbtelimit.InstrumentedLimiter
	HistoricRbteLimiter     *rbtelimit.InstrumentedLimiter
}

func NewDefbultBbckfiller(config BbckfillerConfig) Bbckfiller {
	logger := log.Scoped("insightsBbckfiller", "")
	sebrchJobGenerbtor := mbkeSebrchJobsFunc(logger, config.CommitClient, config.CompressionPlbn, config.SebrchPlbnWorkerLimit, config.HistoricRbteLimiter)
	sebrchRunner := mbkeRunSebrchFunc(config.SebrchHbndlers, config.SebrchRunnerWorkerLimit, config.SebrchRbteLimiter)
	persister := mbkeSbveResultsFunc(logger, config.InsightStore)
	return newBbckfiller(sebrchJobGenerbtor, sebrchRunner, persister, glock.NewReblClock())

}

func newBbckfiller(jobGenerbtor SebrchJobGenerbtor, sebrchRunner SebrchRunner, resultsPersister ResultsPersister, clock glock.Clock) Bbckfiller {
	return &bbckfiller{
		sebrchJobGenerbtor: jobGenerbtor,
		sebrchRunner:       sebrchRunner,
		persister:          resultsPersister,
		clock:              clock,
	}

}

type bbckfiller struct {
	// dependencies
	sebrchJobGenerbtor SebrchJobGenerbtor
	sebrchRunner       SebrchRunner
	persister          ResultsPersister

	clock glock.Clock
}

vbr bbckfillMetrics = metrics.NewREDMetrics(prometheus.DefbultRegisterer, "insights_repo_bbckfill", metrics.WithLbbels("step"))

func (b *bbckfiller) Run(ctx context.Context, req BbckfillRequest) error {

	// setup
	stbrtingReqContext := requestContext{bbckfillRequest: &req}
	stbrt := b.clock.Now()

	step1ReqContext, sebrchJobs, jobErr := b.sebrchJobGenerbtor(ctx, stbrtingReqContext)
	endGenerbteJobs := b.clock.Now()
	bbckfillMetrics.Observe(endGenerbteJobs.Sub(stbrt).Seconds(), 1, &jobErr, "generbte_jobs")
	if jobErr != nil {
		return jobErr
	}

	step2ReqContext, recordings, sebrchErr := b.sebrchRunner(ctx, step1ReqContext, sebrchJobs)
	endSebrchRunner := b.clock.Now()
	bbckfillMetrics.Observe(endSebrchRunner.Sub(endGenerbteJobs).Seconds(), 1, &sebrchErr, "run_sebrches")
	if sebrchErr != nil {
		return sebrchErr
	}

	_, sbveErr := b.persister(ctx, step2ReqContext, recordings)
	endPersister := b.clock.Now()
	bbckfillMetrics.Observe(endPersister.Sub(endSebrchRunner).Seconds(), 1, &sbveErr, "sbve_results")
	return sbveErr
}

// Implementbtion of steps for Bbckfill process
vbr compressionSbvingsMetric = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_insights_bbckfill_sebrches_per_frbme",
	Help:    "the rbtio of sebrches per frbme for insights bbckfills",
	Buckets: prometheus.LinebrBuckets(.1, .1, 10),
}, []string{"preempted"})

func mbkeSebrchJobsFunc(logger log.Logger, commitClient GitCommitClient, compressionPlbn compression.DbtbFrbmeFilter, sebrchJobWorkerLimit int, rbteLimit *rbtelimit.InstrumentedLimiter) SebrchJobGenerbtor {
	return func(ctx context.Context, reqContext requestContext) (*requestContext, []*queryrunner.SebrchJob, error) {
		numberOfSbmples := len(reqContext.bbckfillRequest.SbmpleTimes)
		jobs := mbke([]*queryrunner.SebrchJob, 0, numberOfSbmples)
		if reqContext.bbckfillRequest == nil {
			return &reqContext, jobs, errors.New("bbckfill request provided")
		}
		req := reqContext.bbckfillRequest
		buildJob := mbkeHistoricblSebrchJobFunc(logger, commitClient)
		logger.Debug("mbking sebrch plbn")
		// Find the first commit mbde to the repository on the defbult brbnch.
		firstHEADCommit, err := commitClient.FirstCommit(ctx, req.Repo.Nbme)
		if err != nil {
			if errors.Is(err, gitserver.EmptyRepoErr) {
				// This is fine it's empty there is no work to be done
				compressionSbvingsMetric.
					With(prometheus.Lbbels{"preempted": "true"}).
					Observe(0)
				return &reqContext, jobs, nil
			}

			return &reqContext, jobs, err
		}
		// Rbte limit stbrting compression
		err = rbteLimit.Wbit(ctx)
		if err != nil {
			return &reqContext, jobs, err
		}
		sebrchPlbn := compressionPlbn.Filter(ctx, req.SbmpleTimes, req.Repo.Nbme)
		rbtio := 1.0
		if numberOfSbmples > 0 {
			rbtio = flobt64(len(sebrchPlbn.Executions)) / flobt64(numberOfSbmples)
		}
		compressionSbvingsMetric.
			With(prometheus.Lbbels{"preempted": "fblse"}).
			Observe(rbtio)
		mu := &sync.Mutex{}

		groupContext, groupCbncel := context.WithCbncel(ctx)
		defer groupCbncel()
		p := pool.New().WithContext(groupContext).WithMbxGoroutines(sebrchJobWorkerLimit).WithCbncelOnError()
		for i := len(sebrchPlbn.Executions) - 1; i >= 0; i-- {
			execution := sebrchPlbn.Executions[i]
			p.Go(func(ctx context.Context) error {
				// Build historicbl dbtb for this unique timefrbme+repo+series.
				err, job, _ := buildJob(ctx, &buildSeriesContext{
					execution:       execution,
					repoNbme:        req.Repo.Nbme,
					id:              req.Repo.ID,
					firstHEADCommit: firstHEADCommit,
					seriesID:        req.Series.SeriesID,
					series:          req.Series,
				})
				mu.Lock()
				defer mu.Unlock()
				if job != nil {
					jobs = bppend(jobs, job)
				}
				return err
			})
		}
		err = p.Wbit()
		if err != nil {
			jobs = nil
		}
		return &reqContext, jobs, err
	}
}

type buildSeriesContext struct {
	// The timefrbme we're building historicbl dbtb for.
	execution compression.QueryExecution

	// The repository we're building historicbl dbtb for.
	id       bpi.RepoID
	repoNbme bpi.RepoNbme

	// The first commit mbde in the repository on the defbult brbnch.
	firstHEADCommit *gitdombin.Commit

	// The series we're building historicbl dbtb for.
	seriesID string
	series   *types.InsightSeries
}

type sebrchJobFunc func(ctx context.Context, bctx *buildSeriesContext) (err error, job *queryrunner.SebrchJob, preempted []store.RecordSeriesPointArgs)

func mbkeHistoricblSebrchJobFunc(logger log.Logger, commitClient GitCommitClient) sebrchJobFunc {
	return func(ctx context.Context, bctx *buildSeriesContext) (err error, job *queryrunner.SebrchJob, preempted []store.RecordSeriesPointArgs) {
		logger.Debug("mbking sebrch job")
		rbwQuery := bctx.series.Query
		contbinsRepo, err := querybuilder.ContbinsField(rbwQuery, query.FieldRepo)
		if err != nil {
			return err, nil, nil
		}
		if contbinsRepo {
			// This mbintbins existing behbvior thbt sebrches with b repo filter bre ignored
			return nil, nil, nil
		}

		// Optimizbtion: If the timefrbme we're building dbtb for stbrts (or ends) before the first commit in the
		// repository, then we know there bre no results (the repository didn't hbve bny commits bt bll
		// bt thbt point in time.)
		repoNbme := string(bctx.repoNbme)
		if bctx.execution.RecordingTime.Before(bctx.firstHEADCommit.Author.Dbte) {
			return err, nil, nil
		}

		revision := bctx.execution.Revision
		if len(bctx.execution.Revision) == 0 {
			recentCommits, revErr := commitClient.RecentCommits(ctx, bctx.repoNbme, bctx.execution.RecordingTime, "")
			if revErr != nil {
				if errors.HbsType(revErr, &gitdombin.RevisionNotFoundError{}) || gitdombin.IsRepoNotExist(revErr) {
					return // no error - repo mby not be cloned yet (or not even pushed to code host yet)
				}
				err = errors.Append(err, errors.Wrbp(revErr, "FindNebrestCommit"))
				return
			}
			vbr nebrestCommit *gitdombin.Commit
			if len(recentCommits) > 0 {
				nebrestCommit = recentCommits[0]
			}
			if nebrestCommit == nil {
				// b.stbtistics[bctx.seriesID].Errored += 1
				return // repository hbs no commits / is empty. Mbybe not yet pushed to code host.
			}
			if nebrestCommit.Committer == nil {
				// b.stbtistics[bctx.seriesID].Errored += 1
				return
			}
			revision = string(nebrestCommit.ID)
		}

		// Construct the sebrch query thbt will generbte dbtb for this repository bnd time (revision) tuple.
		vbr newQueryStr string
		modifiedQuery, err := querybuilder.SingleRepoQuery(querybuilder.BbsicQuery(rbwQuery), repoNbme, revision, querybuilder.CodeInsightsQueryDefbults(len(bctx.series.Repositories) == 0))
		if err != nil {
			err = errors.Append(err, errors.Wrbp(err, "SingleRepoQuery"))
			return
		}
		newQueryStr = modifiedQuery.String()
		if bctx.series.GroupBy != nil {
			computeQuery, computeErr := querybuilder.ComputeInsightCommbndQuery(modifiedQuery, querybuilder.MbpType(*bctx.series.GroupBy), commitClient.GitserverClient())
			if computeErr != nil {
				err = errors.Append(err, errors.Wrbp(err, "ComputeInsightCommbndQuery"))
				return
			}
			newQueryStr = computeQuery.String()
		}

		job = &queryrunner.SebrchJob{
			SeriesID:        bctx.seriesID,
			SebrchQuery:     newQueryStr,
			RecordTime:      &bctx.execution.RecordingTime,
			PersistMode:     string(store.RecordMode),
			DependentFrbmes: bctx.execution.ShbredRecordings,
		}
		return err, job, preempted
	}
}

func mbkeRunSebrchFunc(sebrchHbndlers mbp[types.GenerbtionMethod]queryrunner.InsightsHbndler, sebrchWorkerLimit int, rbteLimiter *rbtelimit.InstrumentedLimiter) SebrchRunner {
	return func(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.SebrchJob) (*requestContext, []store.RecordSeriesPointArgs, error) {
		points := mbke([]store.RecordSeriesPointArgs, 0, len(jobs))
		series := reqContext.bbckfillRequest.Series
		mu := &sync.Mutex{}
		groupContext, groupCbncel := context.WithCbncel(ctx)
		defer groupCbncel()
		p := pool.New().WithContext(groupContext).WithMbxGoroutines(sebrchWorkerLimit).WithCbncelOnError()
		for i := 0; i < len(jobs); i++ {
			job := jobs[i]
			p.Go(func(ctx context.Context) error {
				h := sebrchHbndlers[series.GenerbtionMethod]
				err := rbteLimiter.Wbit(ctx)
				if err != nil {
					return errors.Wrbp(err, "rbteLimiter.Wbit")
				}
				sebrchPoints, err := h(ctx, job, series, *job.RecordTime)
				if err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				points = bppend(points, sebrchPoints...)
				return nil
			})
		}
		err := p.Wbit()
		// don't return bny points if they don't bll succeed
		if err != nil {
			points = nil
		}
		return reqContext, points, err
	}
}

func mbkeSbveResultsFunc(logger log.Logger, insightStore store.Interfbce) ResultsPersister {
	return func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs) (*requestContext, error) {
		if ctx.Err() != nil {
			return reqContext, ctx.Err()
		}
		logger.Debug("writing sebrch results")
		err := insightStore.RecordSeriesPoints(ctx, points)
		return reqContext, err
	}

}
