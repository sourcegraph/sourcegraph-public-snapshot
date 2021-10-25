package resolvers

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"

	"github.com/segmentio/ksuid"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.InsightViewResolver = &insightViewResolver{}
var _ graphqlbackend.LineChartInsightViewPresentation = &lineChartInsightViewPresentation{}
var _ graphqlbackend.LineChartDataSeriesPresentationResolver = &lineChartDataSeriesPresentationResolver{}
var _ graphqlbackend.SearchInsightDataSeriesDefinitionResolver = &searchInsightDataSeriesDefinitionResolver{}
var _ graphqlbackend.InsightRepositoryScopeResolver = &insightRepositoryScopeResolver{}
var _ graphqlbackend.InsightIntervalTimeScope = &insightIntervalTimeScopeResolver{}
var _ graphqlbackend.InsightViewFiltersResolver = &insightViewFiltersResolver{}
var _ graphqlbackend.CreateInsightResultResolver = &createInsightResultResolver{}
var _ graphqlbackend.InsightTimeScope = &insightTimeScopeUnionResolver{}
var _ graphqlbackend.InsightPresentation = &insightPresentationUnionResolver{}
var _ graphqlbackend.InsightDataSeriesDefinition = &insightDataSeriesDefinitionUnionResolver{}

type insightViewResolver struct {
	view *types.Insight

	baseInsightResolver
}

func (i *insightViewResolver) ID() graphql.ID {
	return relay.MarshalID("insight_view", i.view.UniqueID)
}

func (i *insightViewResolver) DefaultFilters(ctx context.Context) (graphqlbackend.InsightViewFiltersResolver, error) {
	return &insightViewFiltersResolver{view: i.view}, nil
}

type insightViewFiltersResolver struct {
	view *types.Insight
}

func (i *insightViewFiltersResolver) IncludeRepoRegex(ctx context.Context) (*string, error) {
	return i.view.Filters.IncludeRepoRegex, nil
}

func (i *insightViewFiltersResolver) ExcludeRepoRegex(ctx context.Context) (*string, error) {
	return i.view.Filters.ExcludeRepoRegex, nil
}

func (i *insightViewResolver) AppliedFilters(ctx context.Context) (graphqlbackend.InsightViewFiltersResolver, error) {
	panic("implement me")
}

func (i *insightViewResolver) DataSeries(ctx context.Context) ([]graphqlbackend.InsightSeriesResolver, error) {
	var resolvers []graphqlbackend.InsightSeriesResolver
	for j := range i.view.Series {
		resolvers = append(resolvers, &insightSeriesResolver{
			insightsStore:   i.timeSeriesStore,
			workerBaseStore: i.workerBaseStore,
			series:          i.view.Series[j],
			metadataStore:   i.insightStore,
		})
	}

	return resolvers, nil
}

func (i *insightViewResolver) Presentation(ctx context.Context) (graphqlbackend.InsightPresentation, error) {
	lineChartPresentation := &lineChartInsightViewPresentation{view: i.view}

	return &insightPresentationUnionResolver{resolver: lineChartPresentation}, nil
}

func (i *insightViewResolver) DataSeriesDefinitions(ctx context.Context) ([]graphqlbackend.InsightDataSeriesDefinition, error) {
	var resolvers []graphqlbackend.InsightDataSeriesDefinition
	for j := range i.view.Series {
		resolvers = append(resolvers, &insightDataSeriesDefinitionUnionResolver{resolver: &searchInsightDataSeriesDefinitionResolver{series: &i.view.Series[j]}})
	}
	return resolvers, nil
}

type searchInsightDataSeriesDefinitionResolver struct {
	series *types.InsightViewSeries
}

func (s *searchInsightDataSeriesDefinitionResolver) SeriesId(ctx context.Context) (string, error) {
	return s.series.SeriesID, nil
}

func (s *searchInsightDataSeriesDefinitionResolver) Query(ctx context.Context) (string, error) {
	return s.series.Query, nil
}

func (s *searchInsightDataSeriesDefinitionResolver) RepositoryScope(ctx context.Context) (graphqlbackend.InsightRepositoryScopeResolver, error) {
	return &insightRepositoryScopeResolver{repositories: s.series.Repositories}, nil
}

func (s *searchInsightDataSeriesDefinitionResolver) TimeScope(ctx context.Context) (graphqlbackend.InsightTimeScope, error) {
	intervalResolver := &insightIntervalTimeScopeResolver{
		unit:  s.series.SampleIntervalUnit,
		value: int32(s.series.SampleIntervalValue),
	}

	return &insightTimeScopeUnionResolver{resolver: intervalResolver}, nil
}

type insightIntervalTimeScopeResolver struct {
	unit  string
	value int32
}

func (i *insightIntervalTimeScopeResolver) Unit(ctx context.Context) (string, error) {
	return i.unit, nil
}

func (i *insightIntervalTimeScopeResolver) Value(ctx context.Context) (int32, error) {
	return i.value, nil
}

type insightRepositoryScopeResolver struct {
	repositories []string
}

func (i *insightRepositoryScopeResolver) Repositories(ctx context.Context) ([]string, error) {
	return i.repositories, nil
}

type lineChartInsightViewPresentation struct {
	view *types.Insight
}

func (l *lineChartInsightViewPresentation) Title(ctx context.Context) (string, error) {
	return l.view.Title, nil
}

func (l *lineChartInsightViewPresentation) SeriesPresentation(ctx context.Context) ([]graphqlbackend.LineChartDataSeriesPresentationResolver, error) {
	var resolvers []graphqlbackend.LineChartDataSeriesPresentationResolver

	for i := range l.view.Series {
		resolvers = append(resolvers, &lineChartDataSeriesPresentationResolver{series: &l.view.Series[i]})
	}

	return resolvers, nil
}

type lineChartDataSeriesPresentationResolver struct {
	series *types.InsightViewSeries
}

func (l *lineChartDataSeriesPresentationResolver) SeriesId(ctx context.Context) (string, error) {
	return l.series.SeriesID, nil
}

func (l *lineChartDataSeriesPresentationResolver) Label(ctx context.Context) (string, error) {
	return l.series.Label, nil
}

func (l *lineChartDataSeriesPresentationResolver) Color(ctx context.Context) (string, error) {
	return l.series.LineColor, nil
}

func (r *Resolver) CreateLineChartSearchInsight(ctx context.Context, args *graphqlbackend.CreateLineChartSearchInsightArgs) (_ graphqlbackend.CreateInsightResultResolver, err error) {
	uid := actor.FromContext(ctx).UID

	tx, err := r.insightStore.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	view, err := tx.CreateView(ctx, types.InsightView{
		Title:    emptyIfNil(args.Input.Options.Title),
		UniqueID: ksuid.New().String(),
		Filters:  types.InsightViewFilters{},
	}, []store.InsightViewGrant{store.UserGrant(int(uid))})
	if err != nil {
		return nil, errors.Wrap(err, "CreateView")
	}

	for _, series := range args.Input.DataSeries {
		created, err := tx.CreateSeries(ctx, types.InsightSeries{
			SeriesID:            ksuid.New().String(), // ignoring sharing data series for now, we will just always generate unique series
			Query:               series.Query,
			CreatedAt:           time.Now(),
			Repositories:        series.RepositoryScope.Repositories,
			SampleIntervalUnit:  series.TimeScope.StepInterval.Unit,
			SampleIntervalValue: int(series.TimeScope.StepInterval.Value),
		})
		if err != nil {
			return nil, errors.Wrap(err, "CreateSeries")
		}
		err = tx.AttachSeriesToView(ctx, created, view, types.InsightViewSeriesMetadata{
			Label:  emptyIfNil(series.Options.Label),
			Stroke: emptyIfNil(series.Options.LineColor),
		})
		if err != nil {
			return nil, errors.Wrap(err, "AttachSeriesToView")
		}
	}
	return &createInsightResultResolver{baseInsightResolver: r.baseInsightResolver, viewId: view.UniqueID}, nil
}

type createInsightResultResolver struct {
	viewId string
	baseInsightResolver
}

func (c *createInsightResultResolver) View(ctx context.Context) (graphqlbackend.InsightViewResolver, error) {
	mapped, err := c.insightStore.GetMapped(ctx, store.InsightQueryArgs{UniqueID: c.viewId, UserID: []int{int(actor.FromContext(ctx).UID)}})
	if err != nil {
		return nil, err
	}
	if len(mapped) < 1 {
		return nil, err
	}
	return &insightViewResolver{view: &mapped[0], baseInsightResolver: c.baseInsightResolver}, nil
}

func emptyIfNil(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}

// A dummy type to represent the GraphQL union InsightTimeScope
type insightTimeScopeUnionResolver struct {
	resolver interface{}
}

// ToInsightIntervalTimeScope is used by the GraphQL library to resolve type fragments for unions
func (r *insightTimeScopeUnionResolver) ToInsightIntervalTimeScope() (graphqlbackend.InsightIntervalTimeScope, bool) {
	res, ok := r.resolver.(*insightIntervalTimeScopeResolver)
	return res, ok
}

// A dummy type to represent the GraphQL union InsightPresentation
type insightPresentationUnionResolver struct {
	resolver interface{}
}

// ToLineChartInsightViewPresentation is used by the GraphQL library to resolve type fragments for unions
func (r *insightPresentationUnionResolver) ToLineChartInsightViewPresentation() (graphqlbackend.LineChartInsightViewPresentation, bool) {
	res, ok := r.resolver.(*lineChartInsightViewPresentation)
	return res, ok
}

// A dummy type to represent the GraphQL union InsightDataSeriesDefinition
type insightDataSeriesDefinitionUnionResolver struct {
	resolver interface{}
}

// ToSearchInsightDataSeriesDefinition is used by the GraphQL library to resolve type fragments for unions
func (r *insightDataSeriesDefinitionUnionResolver) ToSearchInsightDataSeriesDefinition() (graphqlbackend.SearchInsightDataSeriesDefinitionResolver, bool) {
	res, ok := r.resolver.(*searchInsightDataSeriesDefinitionResolver)
	return res, ok
}
