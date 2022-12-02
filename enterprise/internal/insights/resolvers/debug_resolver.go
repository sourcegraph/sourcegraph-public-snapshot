package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/pipeline"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	types2 "github.com/sourcegraph/sourcegraph/internal/types"
)

var _ graphqlbackend.InsightBackfillDebugResponseResolver = &insightsBackfillDebugResolver{}
var _ graphqlbackend.InsightBackfillPlanNodeResolver = &insightsBackfillPlanNodeResolver{}
var _ graphqlbackend.InsightBackfillPlanResolver = &insightsBackfillPlanResolver{}

type insightsBackfillDebugResolver struct {
	jobGenerator pipeline.SearchJobGenerator
	baseInsightResolver
}

func newInsightsBackfillDebugResolver(logger log.Logger, base baseInsightResolver) *insightsBackfillDebugResolver {
	searchRateLimiter := limiter.SearchQueryRate()
	historicRateLimiter := limiter.HistoricalWorkRate()
	backfillConfig := pipeline.BackfillerConfig{
		CompressionPlan:         compression.NewHistoricalFilter(true, time.Now().Add(-1*365*24*time.Hour), edb.NewInsightsDBWith(base.timeSeriesStore)),
		SearchHandlers:          queryrunner.GetSearchHandlers(),
		InsightStore:            base.timeSeriesStore,
		CommitClient:            discovery.NewGitCommitClient(base.postgresDB),
		SearchPlanWorkerLimit:   1,
		SearchRunnerWorkerLimit: 5,
		SearchRateLimiter:       searchRateLimiter,
		HistoricRateLimiter:     historicRateLimiter,
	}
	gen := pipeline.NewSearchJobGenerator(logger, backfillConfig)
	return &insightsBackfillDebugResolver{
		baseInsightResolver: base,
		jobGenerator:        gen,
	}
}

func (i *insightsBackfillDebugResolver) Plan(ctx context.Context, args graphqlbackend.InsightBackfillDebugPlanArgs) (graphqlbackend.InsightBackfillPlanResolver, error) {
	// validate user has permissions for repo
	repo, err := i.postgresDB.Repos().GetByName(ctx, api.RepoName(args.Input.Repo))
	if err != nil {
		return nil, err
	}

	frames := timeseries.BuildFrames(12, timeseries.TimeInterval{
		Unit:  types.IntervalUnit(args.Input.TimeScope.StepInterval.Unit),
		Value: int(args.Input.TimeScope.StepInterval.Value),
	}, time.Now().Truncate(time.Hour*24))

	request, jobs, err := i.jobGenerator(ctx, pipeline.RequestContext{BackfillRequest: &pipeline.BackfillRequest{
		Series: &types.InsightSeries{SeriesID: "fake", Query: args.Input.Query},
		Repo:   &types2.MinimalRepo{ID: repo.ID, Name: repo.Name},
		Frames: frames,
	}})
	if err != nil {
		return nil, err
	}

	return &insightsBackfillPlanResolver{percent: request.CompressionSavings, jobs: jobs, repo: args.Input.Repo}, nil
}

type insightsBackfillPlanResolver struct {
	percent float64
	jobs    []*queryrunner.SearchJob
	repo    string
}

func (i *insightsBackfillPlanResolver) Nodes() (nodes []graphqlbackend.InsightBackfillPlanNodeResolver) {
	for _, job := range i.jobs {
		nodes = append(nodes, &insightsBackfillPlanNodeResolver{
			query:      job.SearchQuery,
			time:       *job.RecordTime,
			childTimes: job.DependentFrames,
		})
	}
	return nodes
}

func (i *insightsBackfillPlanResolver) CompressionPercent() float64 {
	return i.percent
}

type insightsBackfillPlanNodeResolver struct {
	query            string
	time             time.Time
	childTimes       []time.Time
	generationMethod types.GenerationMethod
}

func (i *insightsBackfillPlanNodeResolver) Query() string {
	return i.query
}

func (i *insightsBackfillPlanNodeResolver) Time() gqlutil.DateTime {
	return gqlutil.DateTime{Time: i.time}
}

func (i *insightsBackfillPlanNodeResolver) ChildTimes() (gqlTimes []gqlutil.DateTime) {
	for _, childTime := range i.childTimes {
		gqlTimes = append(gqlTimes, gqlutil.DateTime{Time: childTime})
	}
	return gqlTimes
}

func (i *insightsBackfillPlanNodeResolver) GenerationMethod() string {
	return "asdf"
}
