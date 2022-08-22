package resolvers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/grafana/regexp"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/inconshreveable/log15"
	"github.com/segmentio/ksuid"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.InsightViewResolver = &insightViewResolver{}
var _ graphqlbackend.LineChartInsightViewPresentation = &lineChartInsightViewPresentation{}
var _ graphqlbackend.LineChartDataSeriesPresentationResolver = &lineChartDataSeriesPresentationResolver{}
var _ graphqlbackend.SearchInsightDataSeriesDefinitionResolver = &searchInsightDataSeriesDefinitionResolver{}
var _ graphqlbackend.InsightRepositoryScopeResolver = &insightRepositoryScopeResolver{}
var _ graphqlbackend.InsightIntervalTimeScope = &insightIntervalTimeScopeResolver{}
var _ graphqlbackend.InsightViewFiltersResolver = &insightViewFiltersResolver{}
var _ graphqlbackend.InsightViewPayloadResolver = &insightPayloadResolver{}
var _ graphqlbackend.InsightTimeScope = &insightTimeScopeUnionResolver{}
var _ graphqlbackend.InsightPresentation = &insightPresentationUnionResolver{}
var _ graphqlbackend.InsightDataSeriesDefinition = &insightDataSeriesDefinitionUnionResolver{}
var _ graphqlbackend.InsightViewConnectionResolver = &InsightViewQueryConnectionResolver{}
var _ graphqlbackend.InsightViewSeriesDisplayOptionsResolver = &insightViewSeriesDisplayOptionsResolver{}
var _ graphqlbackend.InsightViewSeriesSortOptionsResolver = &insightViewSeriesSortOptionsResolver{}

type insightViewResolver struct {
	view                  *types.Insight
	overrideFilters       *types.InsightViewFilters
	overrideSeriesOptions *types.SeriesDisplayOptions
	dataSeriesGenerator   insightSeriesResolverGenerator

	baseInsightResolver

	// Cache results because they are used by multiple fields
	seriesOnce      sync.Once
	seriesErr       error
	totalSeries     int
	seriesResolvers []graphqlbackend.InsightSeriesResolver
}

const insightKind = "insight_view"

func (i *insightViewResolver) ID() graphql.ID {
	return relay.MarshalID(insightKind, i.view.UniqueID)
}

func (i *insightViewResolver) DefaultFilters(ctx context.Context) (graphqlbackend.InsightViewFiltersResolver, error) {
	return &insightViewFiltersResolver{filters: &i.view.Filters}, nil
}

type insightViewFiltersResolver struct {
	filters *types.InsightViewFilters
}

func (i *insightViewFiltersResolver) IncludeRepoRegex(ctx context.Context) (*string, error) {
	return i.filters.IncludeRepoRegex, nil
}

func (i *insightViewFiltersResolver) ExcludeRepoRegex(ctx context.Context) (*string, error) {
	return i.filters.ExcludeRepoRegex, nil
}

func (i *insightViewFiltersResolver) SearchContexts(ctx context.Context) (*[]string, error) {
	return &i.filters.SearchContexts, nil
}

func (i *insightViewResolver) AppliedFilters(ctx context.Context) (graphqlbackend.InsightViewFiltersResolver, error) {
	if i.overrideFilters != nil {
		return &insightViewFiltersResolver{filters: i.overrideFilters}, nil
	}
	return &insightViewFiltersResolver{filters: &i.view.Filters}, nil
}

type insightViewSeriesDisplayOptionsResolver struct {
	seriesDisplayOptions *types.SeriesDisplayOptions
}

func (i *insightViewSeriesDisplayOptionsResolver) Limit(ctx context.Context) (*int32, error) {
	return i.seriesDisplayOptions.Limit, nil
}

func (i *insightViewSeriesDisplayOptionsResolver) SortOptions(ctx context.Context) (graphqlbackend.InsightViewSeriesSortOptionsResolver, error) {
	return &insightViewSeriesSortOptionsResolver{seriesSortOptions: i.seriesDisplayOptions.SortOptions}, nil
}

type insightViewSeriesSortOptionsResolver struct {
	seriesSortOptions *types.SeriesSortOptions
}

func (i *insightViewSeriesSortOptionsResolver) Mode(ctx context.Context) (*string, error) {
	if i.seriesSortOptions != nil {
		return (*string)(&i.seriesSortOptions.Mode), nil
	}
	return nil, nil
}

func (i *insightViewSeriesSortOptionsResolver) Direction(ctx context.Context) (*string, error) {
	if i.seriesSortOptions != nil {
		return (*string)(&i.seriesSortOptions.Direction), nil
	}
	return nil, nil
}

func (i *insightViewResolver) DefaultSeriesDisplayOptions(ctx context.Context) (graphqlbackend.InsightViewSeriesDisplayOptionsResolver, error) {
	return &insightViewSeriesDisplayOptionsResolver{seriesDisplayOptions: &i.view.SeriesOptions}, nil
}

func (i *insightViewResolver) AppliedSeriesDisplayOptions(ctx context.Context) (graphqlbackend.InsightViewSeriesDisplayOptionsResolver, error) {
	if i.overrideSeriesOptions != nil {
		return &insightViewSeriesDisplayOptionsResolver{seriesDisplayOptions: i.overrideSeriesOptions}, nil
	}
	return &insightViewSeriesDisplayOptionsResolver{seriesDisplayOptions: &i.view.SeriesOptions}, nil
}

// registerDataSeriesGenerators if the generators that create resolvers for DataSeries haven't been generated then loadthem
func (i *insightViewResolver) registerDataSeriesGenerators() {
	// already registered no op
	if i.dataSeriesGenerator != nil {
		return
	}

	// create the known ways to resolve a data series
	jitCaptureGroupGenerator := newSeriesResolverGenerator(
		func(series types.InsightViewSeries) bool {
			return series.JustInTime && series.GeneratedFromCaptureGroups
		},
		expandCaptureGroupSeriesJustInTime,
	)
	recordedCaptureGroupGenerator := newSeriesResolverGenerator(
		func(series types.InsightViewSeries) bool {
			return !series.JustInTime && series.GeneratedFromCaptureGroups
		},
		expandCaptureGroupSeriesRecorded,
	)
	recordedGenerator := newSeriesResolverGenerator(
		func(series types.InsightViewSeries) bool {
			return !series.JustInTime && !series.GeneratedFromCaptureGroups
		},
		recordedSeries,
	)

	jitStreamingGenerator := newSeriesResolverGenerator(
		func(series types.InsightViewSeries) bool {
			return series.JustInTime && !series.GeneratedFromCaptureGroups
		},
		streamingSeriesJustInTime,
	)

	// build the chain of generators
	jitCaptureGroupGenerator.SetNext(recordedCaptureGroupGenerator)
	recordedCaptureGroupGenerator.SetNext(recordedGenerator)
	recordedGenerator.SetNext(jitStreamingGenerator)

	// set the struct variable to the first generator in the chain
	i.dataSeriesGenerator = jitCaptureGroupGenerator
}

func (i *insightViewResolver) DataSeries(ctx context.Context) ([]graphqlbackend.InsightSeriesResolver, error) {
	return i.computeDataSeries(ctx)
}

func (i *insightViewResolver) computeDataSeries(ctx context.Context) ([]graphqlbackend.InsightSeriesResolver, error) {
	i.seriesOnce.Do(func() {
		var resolvers []graphqlbackend.InsightSeriesResolver
		if i.view.IsFrozen {
			// if the view is frozen, we do not show time series data. This is just a basic limitation to prevent
			// easy mis-use of unlicensed features.
			return
		}
		// Ensure that the data series generators have been registered
		i.registerDataSeriesGenerators()
		if i.dataSeriesGenerator == nil {
			i.seriesErr = errors.New("no dataseries resolver generator registered")
			return
		}

		var filters *types.InsightViewFilters
		if i.overrideFilters != nil {
			filters = i.overrideFilters
		} else {
			filters = &i.view.Filters
		}

		var seriesOptions types.SeriesDisplayOptions
		if i.overrideSeriesOptions != nil {
			seriesOptions = *i.overrideSeriesOptions
		} else {
			seriesOptions = i.view.SeriesOptions
		}

		for _, current := range i.view.Series {
			seriesResolvers, err := i.dataSeriesGenerator.Generate(ctx, current, i.baseInsightResolver, *filters)
			if err != nil {
				i.seriesErr = errors.Wrapf(err, "generate for seriesID: %s", current.SeriesID)
				return
			}
			resolvers = append(resolvers, seriesResolvers...)
		}
		i.totalSeries = len(resolvers)

		sortedAndLimitedResovlers, err := sortSeriesResolvers(ctx, seriesOptions, resolvers)
		if err != nil {
			i.seriesErr = errors.Wrapf(err, "sortSeriesResolvers for insightViewID: %s", i.view.UniqueID)
			return
		}
		i.seriesResolvers = sortedAndLimitedResovlers
	})

	return i.seriesResolvers, i.seriesErr
}

func (i *insightViewResolver) Dashboards(ctx context.Context, args *graphqlbackend.InsightsDashboardsArgs) graphqlbackend.InsightsDashboardConnectionResolver {
	return &dashboardConnectionResolver{baseInsightResolver: i.baseInsightResolver,
		orgStore:         i.postgresDB.Orgs(),
		args:             args,
		withViewUniqueID: &i.view.UniqueID,
	}
}

func filterRepositories(ctx context.Context, filters types.InsightViewFilters, repositories []string, scLoader SearchContextLoader) ([]string, error) {
	matches := make(map[string]any)

	// we need to "unwrap" the search contexts and extract the regexps that compose
	// the Sourcegraph query filters. Then we will union these sets of included /
	// excluded regexps with the standalone filter regexp strings that are
	// available.
	var includedWrapped []string
	var excludedWrapped []string
	inc, exc, err := unwrapSearchContexts(ctx, scLoader, filters.SearchContexts)
	if err != nil {
		return nil, errors.Wrap(err, "unwrapSearchContexts")
	}
	includedWrapped = append(includedWrapped, inc...)
	if filters.IncludeRepoRegex != nil && *filters.IncludeRepoRegex != "" {
		includedWrapped = append(includedWrapped, *filters.IncludeRepoRegex)
	}

	// we have to wrap the excluded ones in a non-capturing group so that we can
	// apply regexp OR semantics across the entire set
	wrapRegexp := func(r string) string {
		return fmt.Sprintf("(?:%s)", r)
	}
	for _, s := range exc {
		excludedWrapped = append(excludedWrapped, wrapRegexp(s))
	}
	if filters.ExcludeRepoRegex != nil && *filters.ExcludeRepoRegex != "" {
		excludedWrapped = append(excludedWrapped, wrapRegexp(*filters.ExcludeRepoRegex))
	}

	// first we process the exclude regexps and remove these repos from the original search results
	if len(excludedWrapped) > 0 {
		excludeRegexp, err := regexp.Compile(strings.Join(excludedWrapped, "|"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to compile ExcludeRepoRegex")
		}
		for _, repository := range repositories {
			if !excludeRegexp.MatchString(repository) {
				matches[repository] = struct{}{}
			}
		}
	} else {
		for _, repository := range repositories {
			matches[repository] = struct{}{}
		}
	}

	// finally we process the included regexps over the set of repos that remain after we exclude
	if len(includedWrapped) > 0 {
		// we are iterating through a set of patterns and trying to match all of them to
		// replicate forward lookahead semantics, which are not available in RE2. This
		// means that every single provided regexp gets an opportunity to match and
		// consume against the entire input string, and we only consider success if we
		// can match every single input pattern.
		matchAll := func(patterns []*regexp.Regexp, input string) bool {
			for _, pattern := range patterns {
				if !pattern.MatchString(input) {
					return false
				}
			}
			return true
		}

		patterns := make([]*regexp.Regexp, 0, len(includedWrapped))
		for _, wrapped := range includedWrapped {
			compile, err := regexp.Compile(wrapped)
			if err != nil {
				return nil, errors.Wrap(err, "regexp.Compile for included regexp set")
			}
			patterns = append(patterns, compile)
		}
		for match := range matches {
			if !matchAll(patterns, match) {
				delete(matches, match)
			}
		}
	}

	results := make([]string, 0, len(matches))
	for match := range matches {
		results = append(results, match)
	}
	return results, nil
}

func (i *insightViewResolver) Presentation(ctx context.Context) (graphqlbackend.InsightPresentation, error) {
	if i.view.PresentationType == types.Pie {
		pieChartPresentation := &pieChartInsightViewPresentation{view: i.view}
		return &insightPresentationUnionResolver{resolver: pieChartPresentation}, nil
	} else {
		lineChartPresentation := &lineChartInsightViewPresentation{view: i.view}
		return &insightPresentationUnionResolver{resolver: lineChartPresentation}, nil
	}
}

func (i *insightViewResolver) DataSeriesDefinitions(ctx context.Context) ([]graphqlbackend.InsightDataSeriesDefinition, error) {
	var resolvers []graphqlbackend.InsightDataSeriesDefinition
	for j := range i.view.Series {
		resolvers = append(resolvers, &insightDataSeriesDefinitionUnionResolver{resolver: &searchInsightDataSeriesDefinitionResolver{series: &i.view.Series[j]}})
	}
	return resolvers, nil
}

func (i *insightViewResolver) DashboardReferenceCount(ctx context.Context) (int32, error) {
	referenceCount, err := i.insightStore.GetReferenceCount(ctx, i.view.ViewID)
	if err != nil {
		return 0, err
	}
	return int32(referenceCount), nil
}

func (i *insightViewResolver) IsFrozen(ctx context.Context) (bool, error) {
	return i.view.IsFrozen, nil
}

func (i *insightViewResolver) SeriesCount(ctx context.Context) (*int32, error) {
	_, err := i.computeDataSeries(ctx)
	total := int32(i.totalSeries)
	return &total, err
}

type searchInsightDataSeriesDefinitionResolver struct {
	series *types.InsightViewSeries
}

func (s *searchInsightDataSeriesDefinitionResolver) IsCalculated() (bool, error) {
	if s.series.GeneratedFromCaptureGroups {
		// capture groups series are always pre-calculated!
		return true, nil
	} else {
		return !s.series.JustInTime, nil
	}
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

func (s *searchInsightDataSeriesDefinitionResolver) GeneratedFromCaptureGroups() (bool, error) {
	return s.series.GeneratedFromCaptureGroups, nil
}

func (s *searchInsightDataSeriesDefinitionResolver) GroupBy() (*string, error) {
	if s.series.GroupBy != nil {
		groupBy := strings.ToUpper(*s.series.GroupBy)
		return &groupBy, nil
	}
	return s.series.GroupBy, nil
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

func (r *Resolver) CreateLineChartSearchInsight(ctx context.Context, args *graphqlbackend.CreateLineChartSearchInsightArgs) (_ graphqlbackend.InsightViewPayloadResolver, err error) {
	if len(args.Input.DataSeries) == 0 {
		return nil, errors.New("At least one data series is required to create an insight view")
	}

	uid := actor.FromContext(ctx).UID
	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)

	insightTx, err := r.insightStore.Transact(ctx)
	dashboardTx := r.dashboardStore.With(insightTx)
	if err != nil {
		return nil, err
	}
	defer func() { err = insightTx.Done(err) }()

	var dashboardIds []int
	if args.Input.Dashboards != nil {
		for _, id := range *args.Input.Dashboards {
			dashboardID, err := unmarshalDashboardID(id)
			if err != nil {
				return nil, errors.Wrapf(err, "unmarshalDashboardID, id:%s", dashboardID)
			}
			dashboardIds = append(dashboardIds, int(dashboardID.Arg))
		}
	}

	lamDashboardId, err := createInsightLicenseCheck(ctx, insightTx, dashboardTx, dashboardIds)
	if err != nil {
		return nil, errors.Wrapf(err, "createInsightLicenseCheck")
	}
	if lamDashboardId != 0 {
		dashboardIds = append(dashboardIds, lamDashboardId)
	}

	var filters types.InsightViewFilters
	if args.Input.ViewControls != nil {
		filters = filtersFromInput(&args.Input.ViewControls.Filters)
	}
	view, err := insightTx.CreateView(ctx, types.InsightView{
		Title:            emptyIfNil(args.Input.Options.Title),
		UniqueID:         ksuid.New().String(),
		Filters:          filters,
		PresentationType: types.Line,
	}, []store.InsightViewGrant{store.UserGrant(int(uid))})
	if err != nil {
		return nil, errors.Wrap(err, "CreateView")
	}

	var scoped []types.InsightSeries
	for _, series := range args.Input.DataSeries {
		c, err := createAndAttachSeries(ctx, insightTx, r.backfiller, r.insightEnqueuer, view, series)
		if err != nil {
			return nil, errors.Wrap(err, "createAndAttachSeries")
		}
		if len(c.Repositories) > 0 {
			scoped = append(scoped, *c)
		}
	}

	if len(dashboardIds) > 0 {
		if args.Input.Dashboards != nil {
			err := validateUserDashboardPermissions(ctx, dashboardTx, *args.Input.Dashboards, r.postgresDB.Orgs())
			if err != nil {
				return nil, err
			}
		}
		for _, dashboardId := range dashboardIds {
			log15.Debug("AddView", "insightId", view.UniqueID, "dashboardId", dashboardId)
			err = dashboardTx.AddViewsToDashboard(ctx, dashboardId, []string{view.UniqueID})
			if err != nil {
				return nil, errors.Wrap(err, "AddViewsToDashboard")
			}
		}
	}

	return &insightPayloadResolver{baseInsightResolver: r.baseInsightResolver, validator: permissionsValidator, viewId: view.UniqueID}, nil
}

func (r *Resolver) UpdateLineChartSearchInsight(ctx context.Context, args *graphqlbackend.UpdateLineChartSearchInsightArgs) (_ graphqlbackend.InsightViewPayloadResolver, err error) {
	if len(args.Input.DataSeries) == 0 {
		return nil, errors.New("At least one data series is required to update an insight view")
	}

	tx, err := r.insightStore.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()
	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)

	var insightViewId string
	err = relay.UnmarshalSpec(args.Id, &insightViewId)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling the insight view id")
	}
	err = permissionsValidator.validateUserAccessForView(ctx, insightViewId)
	if err != nil {
		return nil, err
	}

	views, err := tx.GetMapped(ctx, store.InsightQueryArgs{UniqueID: insightViewId, WithoutAuthorization: true})
	if err != nil {
		return nil, errors.Wrap(err, "GetMapped")
	}
	if len(views) == 0 {
		return nil, errors.New("No insight view found with this id")
	}

	var seriesSortMode *types.SeriesSortMode
	var seriesSortDirection *types.SeriesSortDirection
	if args.Input.ViewControls.SeriesDisplayOptions.SortOptions != nil {
		mode := types.SeriesSortMode(args.Input.ViewControls.SeriesDisplayOptions.SortOptions.Mode)
		seriesSortMode = &mode
		direction := types.SeriesSortDirection(args.Input.ViewControls.SeriesDisplayOptions.SortOptions.Direction)
		seriesSortDirection = &direction
	}

	view, err := tx.UpdateView(ctx, types.InsightView{
		UniqueID:            insightViewId,
		Title:               emptyIfNil(args.Input.PresentationOptions.Title),
		Filters:             filtersFromInput(&args.Input.ViewControls.Filters),
		PresentationType:    types.Line,
		SeriesSortMode:      seriesSortMode,
		SeriesSortDirection: seriesSortDirection,
		SeriesLimit:         args.Input.ViewControls.SeriesDisplayOptions.Limit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "UpdateView")
	}

	for _, existingSeries := range views[0].Series {
		if !seriesFound(existingSeries, args.Input.DataSeries) {
			err = tx.RemoveSeriesFromView(ctx, existingSeries.SeriesID, view.ID)
			if err != nil {
				return nil, errors.Wrap(err, "RemoveSeriesFromView")
			}
		}
	}

	for _, series := range args.Input.DataSeries {
		if series.SeriesId == nil {
			_, err = createAndAttachSeries(ctx, tx, r.backfiller, r.insightEnqueuer, view, series)
			if err != nil {
				return nil, errors.Wrap(err, "createAndAttachSeries")
			}
		} else {
			err = tx.RemoveSeriesFromView(ctx, *series.SeriesId, view.ID)
			if err != nil {
				return nil, errors.Wrap(err, "RemoveViewSeries")
			}
			_, err = createAndAttachSeries(ctx, tx, r.backfiller, r.insightEnqueuer, view, series)
			if err != nil {
				return nil, errors.Wrap(err, "createAndAttachSeries")
			}

			err = tx.UpdateViewSeries(ctx, *series.SeriesId, view.ID, types.InsightViewSeriesMetadata{
				Label:  emptyIfNil(series.Options.Label),
				Stroke: emptyIfNil(series.Options.LineColor),
			})
			if err != nil {
				return nil, errors.Wrap(err, "UpdateViewSeries")
			}
		}
	}
	return &insightPayloadResolver{baseInsightResolver: r.baseInsightResolver, validator: permissionsValidator, viewId: insightViewId}, nil
}

func (r *Resolver) CreatePieChartSearchInsight(ctx context.Context, args *graphqlbackend.CreatePieChartSearchInsightArgs) (_ graphqlbackend.InsightViewPayloadResolver, err error) {
	insightTx, err := r.insightStore.Transact(ctx)
	dashboardTx := r.dashboardStore.With(insightTx)
	if err != nil {
		return nil, err
	}
	defer func() { err = insightTx.Done(err) }()
	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)

	var dashboardIds []int
	if args.Input.Dashboards != nil {
		for _, id := range *args.Input.Dashboards {
			dashboardID, err := unmarshalDashboardID(id)
			if err != nil {
				return nil, errors.Wrapf(err, "unmarshalDashboardID, id:%s", dashboardID)
			}
			dashboardIds = append(dashboardIds, int(dashboardID.Arg))
		}
	}

	lamDashboardId, err := createInsightLicenseCheck(ctx, insightTx, dashboardTx, dashboardIds)
	if err != nil {
		return nil, errors.Wrapf(err, "createInsightLicenseCheck")
	}
	if lamDashboardId != 0 {
		dashboardIds = append(dashboardIds, lamDashboardId)
	}

	uid := actor.FromContext(ctx).UID
	view, err := insightTx.CreateView(ctx, types.InsightView{
		Title:            args.Input.PresentationOptions.Title,
		UniqueID:         ksuid.New().String(),
		OtherThreshold:   &args.Input.PresentationOptions.OtherThreshold,
		PresentationType: types.Pie,
	}, []store.InsightViewGrant{store.UserGrant(int(uid))})
	if err != nil {
		return nil, errors.Wrap(err, "CreateView")
	}
	repos := args.Input.RepositoryScope.Repositories
	seriesToAdd, err := insightTx.CreateSeries(ctx, types.InsightSeries{
		SeriesID:           ksuid.New().String(),
		Query:              args.Input.Query,
		CreatedAt:          time.Now(),
		Repositories:       repos,
		SampleIntervalUnit: string(types.Month),
		JustInTime:         len(repos) > 0,
		// one might ask themselves why is the generation method a language stats method if this mutation is search insight? The answer is that search is ultimately the
		// driver behind language stats, but global language stats behave differently than standard search. Long term the vision is that
		// search will power this, and we can iterate over repos just like any other search insight. But for now, this is just something weird that we will have to live with.
		// As a note, this does mean that this mutation doesn't even technically do what it is named - it does not create a 'search' insight, and with that in mind
		// if we decide to support pie charts for other insights than language stats (which we likely will, say on arbitrary aggregations or capture groups) we will need to
		// revisit this.
		GenerationMethod: types.LanguageStats,
	})
	if err != nil {
		return nil, errors.Wrap(err, "CreateSeries")
	}
	err = insightTx.AttachSeriesToView(ctx, seriesToAdd, view, types.InsightViewSeriesMetadata{})
	if err != nil {
		return nil, errors.Wrap(err, "AttachSeriesToView")
	}

	if len(dashboardIds) > 0 {
		if args.Input.Dashboards != nil {
			err := validateUserDashboardPermissions(ctx, dashboardTx, *args.Input.Dashboards, r.postgresDB.Orgs())
			if err != nil {
				return nil, err
			}
		}
		for _, dashboardId := range dashboardIds {
			log15.Debug("AddView", "insightId", view.UniqueID, "dashboardId", dashboardId)
			err = dashboardTx.AddViewsToDashboard(ctx, dashboardId, []string{view.UniqueID})
			if err != nil {
				return nil, errors.Wrap(err, "AddViewsToDashboard")
			}
		}
	}

	return &insightPayloadResolver{baseInsightResolver: r.baseInsightResolver, validator: permissionsValidator, viewId: view.UniqueID}, nil
}

func (r *Resolver) UpdatePieChartSearchInsight(ctx context.Context, args *graphqlbackend.UpdatePieChartSearchInsightArgs) (_ graphqlbackend.InsightViewPayloadResolver, err error) {
	tx, err := r.insightStore.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()
	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)

	var insightViewId string
	err = relay.UnmarshalSpec(args.Id, &insightViewId)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling the insight view id")
	}
	err = permissionsValidator.validateUserAccessForView(ctx, insightViewId)
	if err != nil {
		return nil, err
	}
	views, err := tx.GetMapped(ctx, store.InsightQueryArgs{UniqueID: insightViewId, WithoutAuthorization: true})
	if err != nil {
		return nil, errors.Wrap(err, "GetMapped")
	}
	if len(views) == 0 {
		return nil, errors.New("No insight view found with this id")
	}
	if len(views[0].Series) == 0 {
		return nil, errors.New("No matching series found for this view. The view data may be corrupted.")
	}

	view, err := tx.UpdateView(ctx, types.InsightView{
		UniqueID:         insightViewId,
		Title:            args.Input.PresentationOptions.Title,
		OtherThreshold:   &args.Input.PresentationOptions.OtherThreshold,
		PresentationType: types.Pie,
	})
	if err != nil {
		return nil, errors.Wrap(err, "UpdateView")
	}
	err = tx.UpdateFrontendSeries(ctx, store.UpdateFrontendSeriesArgs{
		SeriesID:         views[0].Series[0].SeriesID,
		Query:            args.Input.Query,
		Repositories:     args.Input.RepositoryScope.Repositories,
		StepIntervalUnit: string(types.Month),
	})
	if err != nil {
		return nil, errors.Wrap(err, "UpdateSeries")
	}

	return &insightPayloadResolver{baseInsightResolver: r.baseInsightResolver, validator: permissionsValidator, viewId: view.UniqueID}, nil
}

type pieChartInsightViewPresentation struct {
	view *types.Insight
}

func (p *pieChartInsightViewPresentation) Title(ctx context.Context) (string, error) {
	return p.view.Title, nil
}

func (p *pieChartInsightViewPresentation) OtherThreshold(ctx context.Context) (float64, error) {
	if p.view.OtherThreshold == nil {
		log15.Warn("Returning a pie chart with no threshold set. This should never happen!", "id", p.view.UniqueID)
		return 0, nil
	}
	return *p.view.OtherThreshold, nil
}

type insightPayloadResolver struct {
	viewId    string
	validator *InsightPermissionsValidator
	baseInsightResolver
}

func (c *insightPayloadResolver) View(ctx context.Context) (graphqlbackend.InsightViewResolver, error) {
	if !c.validator.loaded {
		err := c.validator.loadUserContext(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "InsightPayloadResolver.LoadUserContext")
		}
	}

	mapped, err := c.insightStore.GetAllMapped(ctx, store.InsightQueryArgs{UniqueID: c.viewId, UserID: c.validator.userIds, OrgID: c.validator.orgIds})
	if err != nil {
		return nil, err
	}
	if len(mapped) < 1 {
		return nil, errors.New("insight not found")
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
	resolver any
}

// ToInsightIntervalTimeScope is used by the GraphQL library to resolve type fragments for unions
func (r *insightTimeScopeUnionResolver) ToInsightIntervalTimeScope() (graphqlbackend.InsightIntervalTimeScope, bool) {
	res, ok := r.resolver.(*insightIntervalTimeScopeResolver)
	return res, ok
}

// A dummy type to represent the GraphQL union InsightPresentation
type insightPresentationUnionResolver struct {
	resolver any
}

// ToLineChartInsightViewPresentation is used by the GraphQL library to resolve type fragments for unions
func (r *insightPresentationUnionResolver) ToLineChartInsightViewPresentation() (graphqlbackend.LineChartInsightViewPresentation, bool) {
	res, ok := r.resolver.(*lineChartInsightViewPresentation)
	return res, ok
}

// ToPieChartInsightViewPresentation is used by the GraphQL library to resolve type fragments for unions
func (r *insightPresentationUnionResolver) ToPieChartInsightViewPresentation() (graphqlbackend.PieChartInsightViewPresentation, bool) {
	res, ok := r.resolver.(*pieChartInsightViewPresentation)
	return res, ok
}

// A dummy type to represent the GraphQL union InsightDataSeriesDefinition
type insightDataSeriesDefinitionUnionResolver struct {
	resolver any
}

// ToSearchInsightDataSeriesDefinition is used by the GraphQL library to resolve type fragments for unions
func (r *insightDataSeriesDefinitionUnionResolver) ToSearchInsightDataSeriesDefinition() (graphqlbackend.SearchInsightDataSeriesDefinitionResolver, bool) {
	res, ok := r.resolver.(*searchInsightDataSeriesDefinitionResolver)
	return res, ok
}

func (r *Resolver) InsightViews(ctx context.Context, args *graphqlbackend.InsightViewQueryArgs) (graphqlbackend.InsightViewConnectionResolver, error) {
	return &InsightViewQueryConnectionResolver{
		baseInsightResolver: r.baseInsightResolver,
		args:                args,
	}, nil
}

type InsightViewQueryConnectionResolver struct {
	baseInsightResolver

	args *graphqlbackend.InsightViewQueryArgs

	// Cache results because they are used by multiple fields
	once  sync.Once
	views []types.Insight
	next  string
	err   error
}

func (d *InsightViewQueryConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightViewResolver, error) {
	resolvers := make([]graphqlbackend.InsightViewResolver, 0)
	var scs []string

	views, _, err := d.computeViews(ctx)
	if err != nil {
		return nil, err
	}
	for i := range views {
		resolver := &insightViewResolver{view: &views[i], baseInsightResolver: d.baseInsightResolver}
		if d.args.Filters != nil {
			if d.args.Filters.SearchContexts != nil {
				scs = *d.args.Filters.SearchContexts
			}
			resolver.overrideFilters = &types.InsightViewFilters{
				IncludeRepoRegex: d.args.Filters.IncludeRepoRegex,
				ExcludeRepoRegex: d.args.Filters.ExcludeRepoRegex,
				SearchContexts:   scs,
			}
		}
		if d.args.SeriesDisplayOptions != nil {
			var sortOptions *types.SeriesSortOptions
			if d.args.SeriesDisplayOptions != nil && d.args.SeriesDisplayOptions.SortOptions != nil {
				sortOptions = &types.SeriesSortOptions{
					Mode:      types.SeriesSortMode(d.args.SeriesDisplayOptions.SortOptions.Mode),
					Direction: types.SeriesSortDirection(d.args.SeriesDisplayOptions.SortOptions.Direction),
				}
			}
			resolver.overrideSeriesOptions = &types.SeriesDisplayOptions{
				SortOptions: sortOptions,
				Limit:       d.args.SeriesDisplayOptions.Limit,
			}
		}
		resolvers = append(resolvers, resolver)
	}
	return resolvers, nil
}

func (d *InsightViewQueryConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := d.computeViews(ctx)
	if err != nil {
		return nil, err
	}

	if next != "" {
		return graphqlutil.NextPageCursor(string(relay.MarshalID(insightKind, d.next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *InsightViewQueryConnectionResolver) computeViews(ctx context.Context) ([]types.Insight, string, error) {
	r.once.Do(func() {
		orgStore := r.postgresDB.Orgs()

		args := store.InsightQueryArgs{}
		if r.args.After != nil {
			var afterID string
			err := relay.UnmarshalSpec(graphql.ID(*r.args.After), &afterID)
			if err != nil {
				r.err = errors.Wrap(err, "unmarshalID")
				return
			}
			args.After = afterID
		}
		if r.args.First != nil {
			// Ask for one more result than needed in order to determine if there is a next page.
			args.Limit = int(*r.args.First) + 1
		}
		if r.args.IsFrozen != nil {
			// Filter insight views by their frozen state. We use a pointer for the argument because
			// we might want to not filter on this attribute at all, and `bool` defaults to false.
			args.IsFrozen = r.args.IsFrozen
		}
		var err error
		args.UserID, args.OrgID, err = getUserPermissions(ctx, orgStore)
		if err != nil {
			r.err = errors.Wrap(err, "getUserPermissions")
			return
		}

		if r.args.Id != nil {
			var unique string
			r.err = relay.UnmarshalSpec(*r.args.Id, &unique)
			if r.err != nil {
				return
			}
			log15.Info("unique_id", "id", unique)
			args.UniqueID = unique
		}

		viewSeries, err := r.insightStore.GetAll(ctx, args)
		if err != nil {
			r.err = err
			return
		}

		r.views = r.insightStore.GroupByView(ctx, viewSeries)

		if r.args.First != nil && len(r.views) == args.Limit {
			r.next = r.views[len(r.views)-2].UniqueID
			r.views = r.views[:args.Limit-1]
		}
	})
	return r.views, r.next, r.err
}

func validateUserDashboardPermissions(ctx context.Context, store store.DashboardStore, externalIds []graphql.ID, orgStore database.OrgStore) error {
	userIds, orgIds, err := getUserPermissions(ctx, orgStore)
	if err != nil {
		return errors.Wrap(err, "getUserPermissions")
	}

	unmarshaled := make([]int, 0, len(externalIds))
	for _, id := range externalIds {
		dashboardID, err := unmarshalDashboardID(id)
		if err != nil {
			return errors.Wrapf(err, "unmarshalDashboardID, id:%s", dashboardID)
		}
		unmarshaled = append(unmarshaled, int(dashboardID.Arg))
	}

	hasPermission, err := store.HasDashboardPermission(ctx, unmarshaled, userIds, orgIds)
	if err != nil {
		return errors.Wrapf(err, "HasDashboardPermission")
	} else if !hasPermission {
		return errors.Newf("missing dashboard permission")
	}
	return nil
}

func createAndAttachSeries(ctx context.Context, tx *store.InsightStore, scopedBackfiller *background.ScopedBackfiller, insightEnqueuer *background.InsightEnqueuer, view types.InsightView, series graphqlbackend.LineChartSearchInsightDataSeriesInput) (*types.InsightSeries, error) {
	var seriesToAdd, matchingSeries types.InsightSeries
	var foundSeries bool
	var err error
	var dynamic bool
	// Validate the query before creating anything; we don't want faulty insights running pointlessly.
	if series.GroupBy != nil || series.GeneratedFromCaptureGroups != nil {
		if _, err := querybuilder.ParseComputeQuery(series.Query); err != nil {
			return nil, errors.Wrap(err, "query validation")
		}
	} else {
		if _, err := querybuilder.ParseQuery(series.Query, "literal"); err != nil {
			return nil, errors.Wrap(err, "query validation")
		}
	}

	if series.GeneratedFromCaptureGroups != nil {
		dynamic = *series.GeneratedFromCaptureGroups
	}

	groupBy := lowercaseGroupBy(series.GroupBy)
	var nextRecordingAfter time.Time
	var oldestHistoricalAt time.Time
	if series.GroupBy != nil {
		// We want to disable interval recording for compute types.
		// December 31, 9999 is the maximum possible date in postgres.
		nextRecordingAfter = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
		oldestHistoricalAt = time.Now()
	}

	// Don't try to match on non-global series, since they are always replaced
	if len(series.RepositoryScope.Repositories) == 0 {
		matchingSeries, foundSeries, err = tx.FindMatchingSeries(ctx, store.MatchSeriesArgs{
			Query:                     series.Query,
			StepIntervalUnit:          series.TimeScope.StepInterval.Unit,
			StepIntervalValue:         int(series.TimeScope.StepInterval.Value),
			GenerateFromCaptureGroups: dynamic,
			GroupBy:                   groupBy,
		})
		if err != nil {
			return nil, errors.Wrap(err, "FindMatchingSeries")
		}
	}

	flags := featureflag.FromContext(ctx)
	deprecateJustInTime := flags.GetBoolOr("code_insights_deprecate_jit", true)

	if !foundSeries {
		repos := series.RepositoryScope.Repositories
		seriesToAdd, err = tx.CreateSeries(ctx, types.InsightSeries{
			SeriesID:                   ksuid.New().String(),
			Query:                      series.Query,
			CreatedAt:                  time.Now(),
			Repositories:               repos,
			SampleIntervalUnit:         series.TimeScope.StepInterval.Unit,
			SampleIntervalValue:        int(series.TimeScope.StepInterval.Value),
			GeneratedFromCaptureGroups: dynamic,
			JustInTime:                 len(repos) > 0 && !deprecateJustInTime,
			GenerationMethod:           searchGenerationMethod(series),
			GroupBy:                    groupBy,
			NextRecordingAfter:         nextRecordingAfter,
			OldestHistoricalAt:         oldestHistoricalAt,
		})
		if err != nil {
			return nil, errors.Wrap(err, "CreateSeries")
		}
		if groupBy != nil {
			if err := insightEnqueuer.EnqueueSingle(ctx, seriesToAdd, store.SnapshotMode, tx.StampSnapshot); err != nil {
				return nil, errors.Wrap(err, "GroupBy.EnqueueSingle")
			}
			// We stamp backfill even without queueing up a backfill because we only want a single
			// point in time.
			_, err = tx.StampBackfill(ctx, seriesToAdd)
			if err != nil {
				return nil, errors.Wrap(err, "GroupBy.StampBackfill")
			}
		} else if len(seriesToAdd.Repositories) > 0 && deprecateJustInTime {
			err := scopedBackfiller.ScopedBackfill(ctx, []types.InsightSeries{seriesToAdd})
			if err != nil {
				return nil, errors.Wrap(err, "ScopedBackfill")
			}
			_, err = tx.StampBackfill(ctx, seriesToAdd) // note that this isn't transactional with the backfill above until the queue is migrated to the insights DB
			if err != nil {
				return nil, errors.Wrap(err, "StampBackfill")
			}
		}
	} else {
		seriesToAdd = matchingSeries
	}

	// BUG: If the user tries to attach the same series (the same query and timescope) to an insight view multiple times,
	// this will fail because it violates the unique key constraint. This will be solved by: #26905
	// Alternately we could detect this and return an error?
	err = tx.AttachSeriesToView(ctx, seriesToAdd, view, types.InsightViewSeriesMetadata{
		Label:  emptyIfNil(series.Options.Label),
		Stroke: emptyIfNil(series.Options.LineColor),
	})
	if err != nil {
		return nil, errors.Wrap(err, "AttachSeriesToView")
	}

	return &seriesToAdd, nil
}

func searchGenerationMethod(series graphqlbackend.LineChartSearchInsightDataSeriesInput) types.GenerationMethod {
	if series.GeneratedFromCaptureGroups != nil && *series.GeneratedFromCaptureGroups {
		if series.GroupBy != nil {
			return types.MappingCompute
		}
		return types.SearchCompute
	}
	return types.Search
}

func seriesFound(existingSeries types.InsightViewSeries, inputSeries []graphqlbackend.LineChartSearchInsightDataSeriesInput) bool {
	for i := range inputSeries {
		if inputSeries[i].SeriesId == nil {
			continue
		}
		if existingSeries.SeriesID == *inputSeries[i].SeriesId {
			return true
		}
	}
	return false
}

func getExistingSeriesRepositories(seriesId string, existingSeries []types.InsightViewSeries) []string {
	for i := range existingSeries {
		if existingSeries[i].SeriesID == seriesId {
			return existingSeries[i].Repositories
		}
	}
	return nil
}

func (r *Resolver) DeleteInsightView(ctx context.Context, args *graphqlbackend.DeleteInsightViewArgs) (*graphqlbackend.EmptyResponse, error) {
	var viewId string
	err := relay.UnmarshalSpec(args.Id, &viewId)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling the insight view id")
	}
	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)

	err = permissionsValidator.validateUserAccessForView(ctx, viewId)
	if err != nil {
		return nil, err
	}

	insights, err := r.insightStore.GetMapped(ctx, store.InsightQueryArgs{WithoutAuthorization: true, UniqueID: viewId})
	if err != nil {
		return nil, errors.Wrap(err, "GetMapped")
	}
	if len(insights) != 1 {
		return nil, errors.New("Insight not found.")
	}

	for _, series := range insights[0].Series {
		err = r.insightStore.RemoveSeriesFromView(ctx, series.SeriesID, insights[0].ViewID)
		if err != nil {
			return nil, errors.Wrap(err, "RemoveSeriesFromView")
		}
	}

	err = r.insightStore.DeleteViewByUniqueID(ctx, viewId)
	if err != nil {
		return nil, errors.Wrap(err, "DeleteView")
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func createInsightLicenseCheck(ctx context.Context, insightTx *store.InsightStore, dashboardTx *store.DBDashboardStore, dashboardIds []int) (int, error) {
	licenseError := licensing.Check(licensing.FeatureCodeInsights)
	if licenseError != nil {
		globalUnfrozenInsightCount, _, err := insightTx.GetUnfrozenInsightCount(ctx)
		if err != nil {
			return 0, errors.Wrap(err, "GetUnfrozenInsightCount")
		}
		if globalUnfrozenInsightCount >= 2 {
			return 0, errors.New("Cannot create more than 2 global insights in Limited Access Mode.")
		}
		if len(dashboardIds) > 0 {
			dashboards, err := dashboardTx.GetDashboards(ctx, store.DashboardQueryArgs{ID: dashboardIds, WithoutAuthorization: true})
			if err != nil {
				return 0, errors.Wrap(err, "GetDashboards")
			}
			for _, dashboard := range dashboards {
				if !dashboard.GlobalGrant {
					return 0, errors.New("Cannot create an insight on a non-global dashboard in Limited Access Mode.")
				}
			}
		}

		lamDashboardId, err := dashboardTx.EnsureLimitedAccessModeDashboard(ctx)
		if err != nil {
			return 0, errors.Wrap(err, "EnsureLimitedAccessModeDashboard")
		}
		return lamDashboardId, nil
	}

	return 0, nil
}

func filtersFromInput(input *graphqlbackend.InsightViewFiltersInput) types.InsightViewFilters {
	filters := types.InsightViewFilters{}
	if input != nil {
		filters.IncludeRepoRegex = input.IncludeRepoRegex
		filters.ExcludeRepoRegex = input.ExcludeRepoRegex
		if input.SearchContexts != nil {
			filters.SearchContexts = *input.SearchContexts
		}
	}
	return filters
}

func sortSeriesResolvers(ctx context.Context, seriesOptions types.SeriesDisplayOptions, resolvers []graphqlbackend.InsightSeriesResolver) ([]graphqlbackend.InsightSeriesResolver, error) {
	var sortMode types.SeriesSortMode = types.ResultCount
	var sortDirection types.SeriesSortDirection = types.Desc
	var limit int32 = 20

	if seriesOptions.SortOptions != nil {
		sortMode = seriesOptions.SortOptions.Mode
		sortDirection = seriesOptions.SortOptions.Direction
	}
	if seriesOptions.Limit != nil {
		limit = *seriesOptions.Limit
	}

	// All the points are already loaded from their source at this point db or by executing queries
	// Make a map for faster lookup and to deal with possible errors once
	resolverPoints := make(map[string][]graphqlbackend.InsightsDataPointResolver, len(resolvers))
	for _, resolver := range resolvers {
		points, err := resolver.Points(ctx, nil)
		if err != nil {
			return nil, err
		}
		resolverPoints[resolver.SeriesId()] = points
	}

	getMostRecentValue := func(points []graphqlbackend.InsightsDataPointResolver) float64 {
		if len(points) == 0 {
			return 0
		}
		return points[len(points)-1].Value()
	}

	ascLexSort := func(s1 string, s2 string) (hasSemVar bool, result bool) {
		version1, err1 := semver.NewVersion(s1)
		version2, err2 := semver.NewVersion(s2)
		if err1 == nil && err2 == nil {
			return true, version1.Compare(version2) < 0
		}
		if err1 != nil && err2 == nil {
			return true, false
		}
		if err1 == nil && err2 != nil {
			return true, true
		}
		return false, false
	}

	// First sort lexicographically (ascending) to make sure the ordering is consistent even if some result counts are equal.
	sort.SliceStable(resolvers, func(i, j int) bool {
		hasSemVar, result := ascLexSort(resolvers[i].Label(), resolvers[j].Label())
		if hasSemVar == true {
			return result
		}
		return strings.Compare(resolvers[i].Label(), resolvers[j].Label()) < 0
	})

	switch sortMode {
	case types.ResultCount:

		if sortDirection == types.Asc {
			sort.SliceStable(resolvers, func(i, j int) bool {
				return getMostRecentValue(resolverPoints[resolvers[i].SeriesId()]) < getMostRecentValue(resolverPoints[resolvers[j].SeriesId()])
			})
		} else {
			sort.SliceStable(resolvers, func(i, j int) bool {
				return getMostRecentValue(resolverPoints[resolvers[i].SeriesId()]) > getMostRecentValue(resolverPoints[resolvers[j].SeriesId()])
			})
		}
	case types.Lexicographical:
		if sortDirection == types.Asc {
			// Already pre-sorted by default
		} else {
			sort.SliceStable(resolvers, func(i, j int) bool {
				hasSemVar, result := ascLexSort(resolvers[i].Label(), resolvers[j].Label())
				if hasSemVar == true {
					return !result
				}
				return strings.Compare(resolvers[i].Label(), resolvers[j].Label()) > 0
			})
		}
	case types.DateAdded:
		if sortDirection == types.Asc {
			sort.SliceStable(resolvers, func(i, j int) bool {
				iPoints := resolverPoints[resolvers[i].SeriesId()]
				jPoints := resolverPoints[resolvers[j].SeriesId()]
				return iPoints[0].DateTime().Time.Before(jPoints[0].DateTime().Time)
			})
		} else {
			sort.SliceStable(resolvers, func(i, j int) bool {
				iPoints := resolverPoints[resolvers[i].SeriesId()]
				jPoints := resolverPoints[resolvers[j].SeriesId()]
				return iPoints[0].DateTime().Time.After(jPoints[0].DateTime().Time)
			})
		}
	}

	return resolvers[:minInt(int32(len(resolvers)), limit)], nil
}

func minInt(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func lowercaseGroupBy(groupBy *string) *string {
	if groupBy != nil {
		temp := strings.ToLower(*groupBy)
		return &temp
	}
	return groupBy
}
