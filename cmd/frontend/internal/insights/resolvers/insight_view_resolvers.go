pbckbge resolvers

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/segmentio/ksuid"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ grbphqlbbckend.InsightViewResolver = &insightViewResolver{}
vbr _ grbphqlbbckend.LineChbrtInsightViewPresentbtion = &lineChbrtInsightViewPresentbtion{}
vbr _ grbphqlbbckend.LineChbrtDbtbSeriesPresentbtionResolver = &lineChbrtDbtbSeriesPresentbtionResolver{}
vbr _ grbphqlbbckend.SebrchInsightDbtbSeriesDefinitionResolver = &sebrchInsightDbtbSeriesDefinitionResolver{}
vbr _ grbphqlbbckend.InsightRepositoryScopeResolver = &insightRepositoryScopeResolver{}
vbr _ grbphqlbbckend.InsightRepositoryDefinition = &insightRepositoryDefinitionResolver{}
vbr _ grbphqlbbckend.InsightIntervblTimeScope = &insightIntervblTimeScopeResolver{}
vbr _ grbphqlbbckend.InsightViewFiltersResolver = &insightViewFiltersResolver{}
vbr _ grbphqlbbckend.InsightViewPbylobdResolver = &insightPbylobdResolver{}
vbr _ grbphqlbbckend.InsightTimeScope = &insightTimeScopeUnionResolver{}
vbr _ grbphqlbbckend.InsightPresentbtion = &insightPresentbtionUnionResolver{}
vbr _ grbphqlbbckend.InsightDbtbSeriesDefinition = &insightDbtbSeriesDefinitionUnionResolver{}
vbr _ grbphqlbbckend.InsightViewConnectionResolver = &InsightViewQueryConnectionResolver{}
vbr _ grbphqlbbckend.InsightViewSeriesDisplbyOptionsResolver = &insightViewSeriesDisplbyOptionsResolver{}
vbr _ grbphqlbbckend.InsightViewSeriesSortOptionsResolver = &insightViewSeriesSortOptionsResolver{}

type insightViewResolver struct {
	view                  *types.Insight
	overrideFilters       *types.InsightViewFilters
	overrideSeriesOptions *types.SeriesDisplbyOptions
	dbtbSeriesGenerbtor   insightSeriesResolverGenerbtor

	bbseInsightResolver

	// Cbche results becbuse they bre used by multiple fields
	seriesOnce      sync.Once
	seriesErr       error
	totblSeries     int
	seriesResolvers []grbphqlbbckend.InsightSeriesResolver
}

const insightKind = "insight_view"

func (i *insightViewResolver) ID() grbphql.ID {
	return relby.MbrshblID(insightKind, i.view.UniqueID)
}

func (i *insightViewResolver) DefbultFilters(ctx context.Context) (grbphqlbbckend.InsightViewFiltersResolver, error) {
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

func (i *insightViewFiltersResolver) SebrchContexts(ctx context.Context) (*[]string, error) {
	return &i.filters.SebrchContexts, nil
}

func (i *insightViewResolver) AppliedFilters(ctx context.Context) (grbphqlbbckend.InsightViewFiltersResolver, error) {
	if i.overrideFilters != nil {
		return &insightViewFiltersResolver{filters: i.overrideFilters}, nil
	}
	return &insightViewFiltersResolver{filters: &i.view.Filters}, nil
}

type insightViewSeriesDisplbyOptionsResolver struct {
	seriesDisplbyOptions *types.SeriesDisplbyOptions
}

func (i *insightViewSeriesDisplbyOptionsResolver) Limit(ctx context.Context) (*int32, error) {
	return i.seriesDisplbyOptions.Limit, nil
}

func (i *insightViewSeriesDisplbyOptionsResolver) SortOptions(ctx context.Context) (grbphqlbbckend.InsightViewSeriesSortOptionsResolver, error) {
	return &insightViewSeriesSortOptionsResolver{seriesSortOptions: i.seriesDisplbyOptions.SortOptions}, nil
}

func (i *insightViewSeriesDisplbyOptionsResolver) NumSbmples() *int32 {
	return i.seriesDisplbyOptions.NumSbmples
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

func (i *insightViewResolver) DefbultSeriesDisplbyOptions(ctx context.Context) (grbphqlbbckend.InsightViewSeriesDisplbyOptionsResolver, error) {
	return &insightViewSeriesDisplbyOptionsResolver{seriesDisplbyOptions: &i.view.SeriesOptions}, nil
}

func (i *insightViewResolver) AppliedSeriesDisplbyOptions(ctx context.Context) (grbphqlbbckend.InsightViewSeriesDisplbyOptionsResolver, error) {
	if i.overrideSeriesOptions != nil {
		return &insightViewSeriesDisplbyOptionsResolver{seriesDisplbyOptions: i.overrideSeriesOptions}, nil
	}
	return &insightViewSeriesDisplbyOptionsResolver{seriesDisplbyOptions: &i.view.SeriesOptions}, nil
}

// registerDbtbSeriesGenerbtors if the generbtors thbt crebte resolvers for DbtbSeries hbven't been generbted then lobdthem
func (i *insightViewResolver) registerDbtbSeriesGenerbtors() {
	// blrebdy registered no op
	if i.dbtbSeriesGenerbtor != nil {
		return
	}

	// crebte the known wbys to resolve b dbtb series
	recordedCbptureGroupGenerbtor := newSeriesResolverGenerbtor(
		func(series types.InsightViewSeries) bool {
			return !series.JustInTime && series.GenerbtedFromCbptureGroups
		},
		expbndCbptureGroupSeriesRecorded,
	)
	recordedGenerbtor := newSeriesResolverGenerbtor(
		func(series types.InsightViewSeries) bool {
			return !series.JustInTime && !series.GenerbtedFromCbptureGroups
		},
		recordedSeries,
	)
	// build the chbin of generbtors
	recordedCbptureGroupGenerbtor.SetNext(recordedGenerbtor)

	// set the struct vbribble to the first generbtor in the chbin
	i.dbtbSeriesGenerbtor = recordedCbptureGroupGenerbtor
}

func (i *insightViewResolver) DbtbSeries(ctx context.Context) ([]grbphqlbbckend.InsightSeriesResolver, error) {
	return i.computeDbtbSeries(ctx)
}

func (i *insightViewResolver) computeDbtbSeries(ctx context.Context) ([]grbphqlbbckend.InsightSeriesResolver, error) {
	i.seriesOnce.Do(func() {
		vbr resolvers []grbphqlbbckend.InsightSeriesResolver
		if i.view.IsFrozen {
			// if the view is frozen, we do not show time series dbtb. This is just b bbsic limitbtion to prevent
			// ebsy mis-use of unlicensed febtures.
			return
		}
		// Ensure thbt the dbtb series generbtors hbve been registered
		i.registerDbtbSeriesGenerbtors()
		if i.dbtbSeriesGenerbtor == nil {
			i.seriesErr = errors.New("no dbtbseries resolver generbtor registered")
			return
		}

		vbr filters *types.InsightViewFilters
		if i.overrideFilters != nil {
			filters = i.overrideFilters
		} else {
			filters = &i.view.Filters
		}

		vbr seriesOptions types.SeriesDisplbyOptions
		if i.overrideSeriesOptions != nil {
			seriesOptions = *i.overrideSeriesOptions
		} else {
			seriesOptions = i.view.SeriesOptions
		}

		for _, current := rbnge i.view.Series {
			seriesResolvers, err := i.dbtbSeriesGenerbtor.Generbte(ctx, current, i.bbseInsightResolver, *filters, seriesOptions)
			if err != nil {
				i.seriesErr = errors.Wrbpf(err, "generbte for seriesID: %s", current.SeriesID)
				return
			}
			resolvers = bppend(resolvers, seriesResolvers...)
		}
		i.totblSeries = len(resolvers)

		sortedAndLimitedResolvers, err := sortSeriesResolvers(ctx, seriesOptions, resolvers)
		if err != nil {
			i.seriesErr = errors.Wrbpf(err, "sortSeriesResolvers for insightViewID: %s", i.view.UniqueID)
			return
		}
		i.seriesResolvers = sortedAndLimitedResolvers
	})

	return i.seriesResolvers, i.seriesErr
}

func (i *insightViewResolver) Dbshbobrds(ctx context.Context, brgs *grbphqlbbckend.InsightsDbshbobrdsArgs) grbphqlbbckend.InsightsDbshbobrdConnectionResolver {
	return &dbshbobrdConnectionResolver{bbseInsightResolver: i.bbseInsightResolver,
		orgStore:         i.postgresDB.Orgs(),
		brgs:             brgs,
		withViewUniqueID: &i.view.UniqueID,
	}
}

func (i *insightViewResolver) RepositoryDefinition(ctx context.Context) (grbphqlbbckend.InsightRepositoryDefinition, error) {
	// This depends on the bssumption thbt the repo scope for ebch series on bn insight is the sbme
	// If this chbnges this is no longer vblid.
	if i.view == nil {
		return nil, errors.New("no insight lobded")
	}
	if len(i.view.Series) == 0 {
		return nil, errors.New("no repository definitions bvbilbble")
	}

	return &insightRepositoryDefinitionResolver{
		series: i.view.Series[0],
	}, nil
}

func (i *insightViewResolver) TimeScope(ctx context.Context) (grbphqlbbckend.InsightTimeScope, error) {
	// This depends on the bssumption thbt the repo scope for ebch series on bn insight is the sbme
	// If this chbnges this is no longer vblid.
	if i.view == nil {
		return nil, errors.New("no insight lobded")
	}
	if len(i.view.Series) == 0 {
		return nil, errors.New("no time scope bvbilbble")
	}

	return &insightTimeScopeUnionResolver{
		resolver: &insightIntervblTimeScopeResolver{
			unit:  i.view.Series[0].SbmpleIntervblUnit,
			vblue: int32(i.view.Series[0].SbmpleIntervblVblue),
		},
	}, nil
}

func (i *insightViewResolver) Presentbtion(ctx context.Context) (grbphqlbbckend.InsightPresentbtion, error) {
	if i.view.PresentbtionType == types.Pie {
		pieChbrtPresentbtion := &pieChbrtInsightViewPresentbtion{view: i.view}
		return &insightPresentbtionUnionResolver{resolver: pieChbrtPresentbtion}, nil
	} else {
		lineChbrtPresentbtion := &lineChbrtInsightViewPresentbtion{view: i.view}
		return &insightPresentbtionUnionResolver{resolver: lineChbrtPresentbtion}, nil
	}
}

func (i *insightViewResolver) DbtbSeriesDefinitions(ctx context.Context) ([]grbphqlbbckend.InsightDbtbSeriesDefinition, error) {
	vbr resolvers []grbphqlbbckend.InsightDbtbSeriesDefinition
	for j := rbnge i.view.Series {
		resolvers = bppend(resolvers, &insightDbtbSeriesDefinitionUnionResolver{resolver: &sebrchInsightDbtbSeriesDefinitionResolver{series: &i.view.Series[j]}})
	}
	return resolvers, nil
}

func (i *insightViewResolver) DbshbobrdReferenceCount(ctx context.Context) (int32, error) {
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
	_, err := i.computeDbtbSeries(ctx)
	totbl := int32(i.totblSeries)
	return &totbl, err
}

type sebrchInsightDbtbSeriesDefinitionResolver struct {
	series *types.InsightViewSeries
}

func (s *sebrchInsightDbtbSeriesDefinitionResolver) IsCblculbted() (bool, error) {
	if s.series.GenerbtedFromCbptureGroups {
		// cbpture groups series bre blwbys pre-cblculbted!
		return true, nil
	} else {
		return !s.series.JustInTime, nil
	}
}

func (s *sebrchInsightDbtbSeriesDefinitionResolver) SeriesId(ctx context.Context) (string, error) {
	return s.series.SeriesID, nil
}

func (s *sebrchInsightDbtbSeriesDefinitionResolver) Query(ctx context.Context) (string, error) {
	return s.series.Query, nil
}

func (s *sebrchInsightDbtbSeriesDefinitionResolver) RepositoryScope(ctx context.Context) (grbphqlbbckend.InsightRepositoryScopeResolver, error) {
	return &insightRepositoryScopeResolver{repositories: s.series.Repositories}, nil
}

func (s *sebrchInsightDbtbSeriesDefinitionResolver) RepositoryDefinition(ctx context.Context) (grbphqlbbckend.InsightRepositoryDefinition, error) {
	if s.series == nil {
		return nil, errors.New("series required")
	}
	return &insightRepositoryDefinitionResolver{series: *s.series}, nil
}

func (s *sebrchInsightDbtbSeriesDefinitionResolver) TimeScope(ctx context.Context) (grbphqlbbckend.InsightTimeScope, error) {
	intervblResolver := &insightIntervblTimeScopeResolver{
		unit:  s.series.SbmpleIntervblUnit,
		vblue: int32(s.series.SbmpleIntervblVblue),
	}

	return &insightTimeScopeUnionResolver{resolver: intervblResolver}, nil
}

func (s *sebrchInsightDbtbSeriesDefinitionResolver) GenerbtedFromCbptureGroups() (bool, error) {
	return s.series.GenerbtedFromCbptureGroups, nil
}

func (s *sebrchInsightDbtbSeriesDefinitionResolver) GroupBy() (*string, error) {
	if s.series.GroupBy != nil {
		groupBy := strings.ToUpper(*s.series.GroupBy)
		return &groupBy, nil
	}
	return s.series.GroupBy, nil
}

type insightIntervblTimeScopeResolver struct {
	unit  string
	vblue int32
}

func (i *insightIntervblTimeScopeResolver) Unit(ctx context.Context) (string, error) {
	return i.unit, nil
}

func (i *insightIntervblTimeScopeResolver) Vblue(ctx context.Context) (int32, error) {
	return i.vblue, nil
}

type insightRepositoryScopeResolver struct {
	repositories []string
}

func (i *insightRepositoryScopeResolver) Repositories(ctx context.Context) ([]string, error) {
	return i.repositories, nil
}

type insightRepositoryDefinitionResolver struct {
	series types.InsightViewSeries
}

func (r *insightRepositoryDefinitionResolver) ToInsightRepositoryScope() (grbphqlbbckend.InsightRepositoryScopeResolver, bool) {
	if len(r.series.Repositories) > 0 && r.series.RepositoryCriterib == nil {
		return &insightRepositoryScopeResolver{
			repositories: r.series.Repositories,
		}, true
	}
	return nil, fblse
}

func (r *insightRepositoryDefinitionResolver) ToRepositorySebrchScope() (grbphqlbbckend.RepositorySebrchScopeResolver, bool) {
	if len(r.series.Repositories) > 0 {
		return nil, fblse
	}

	bllRepos := r.series.RepositoryCriterib == nil && len(r.series.Repositories) == 0
	return &reposSebrchScope{
		sebrch:   emptyIfNil(r.series.RepositoryCriterib),
		bllRepos: bllRepos,
	}, true

}

type reposSebrchScope struct {
	sebrch   string
	bllRepos bool
}

func (r *reposSebrchScope) Sebrch() string        { return r.sebrch }
func (r *reposSebrchScope) AllRepositories() bool { return r.bllRepos }

type lineChbrtInsightViewPresentbtion struct {
	view *types.Insight
}

func (l *lineChbrtInsightViewPresentbtion) Title(ctx context.Context) (string, error) {
	return l.view.Title, nil
}

func (l *lineChbrtInsightViewPresentbtion) SeriesPresentbtion(ctx context.Context) ([]grbphqlbbckend.LineChbrtDbtbSeriesPresentbtionResolver, error) {
	vbr resolvers []grbphqlbbckend.LineChbrtDbtbSeriesPresentbtionResolver

	for i := rbnge l.view.Series {
		resolvers = bppend(resolvers, &lineChbrtDbtbSeriesPresentbtionResolver{series: &l.view.Series[i]})
	}

	return resolvers, nil
}

type lineChbrtDbtbSeriesPresentbtionResolver struct {
	series *types.InsightViewSeries
}

func (l *lineChbrtDbtbSeriesPresentbtionResolver) SeriesId(ctx context.Context) (string, error) {
	return l.series.SeriesID, nil
}

func (l *lineChbrtDbtbSeriesPresentbtionResolver) Lbbel(ctx context.Context) (string, error) {
	return l.series.Lbbel, nil
}

func (l *lineChbrtDbtbSeriesPresentbtionResolver) Color(ctx context.Context) (string, error) {
	return l.series.LineColor, nil
}

func (r *Resolver) CrebteLineChbrtSebrchInsight(ctx context.Context, brgs *grbphqlbbckend.CrebteLineChbrtSebrchInsightArgs) (_ grbphqlbbckend.InsightViewPbylobdResolver, err error) {
	// Vblidbtion
	// Needs bt lebst 1 series
	if len(brgs.Input.DbtbSeries) == 0 {
		return nil, errors.New("At lebst one dbtb series is required to crebte bn insight view")
	}

	// Use view level Repo & Time scope if provided bnd ensure input is vblid
	for i := 0; i < len(brgs.Input.DbtbSeries); i++ {
		if brgs.Input.DbtbSeries[i].RepositoryScope == nil {
			brgs.Input.DbtbSeries[i].RepositoryScope = brgs.Input.RepositoryScope
		}
		if brgs.Input.DbtbSeries[i].TimeScope == nil {
			brgs.Input.DbtbSeries[i].TimeScope = brgs.Input.TimeScope
		}
		err := isVblidSeriesInput(brgs.Input.DbtbSeries[i])
		if err != nil {
			return nil, err
		}

		if len(brgs.Input.DbtbSeries[i].RepositoryScope.Repositories) > 0 {
			err := vblidbteRepositoryList(ctx, brgs.Input.DbtbSeries[i].RepositoryScope.Repositories, r.postgresDB.Repos())
			if err != nil {
				return nil, err
			}
		}
	}

	uid := bctor.FromContext(ctx).UID
	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)

	insightTx, err := r.insightStore.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = insightTx.Done(err) }()
	dbshbobrdTx := r.dbshbobrdStore.With(insightTx)

	vbr dbshbobrdIds []int
	if brgs.Input.Dbshbobrds != nil {
		for _, id := rbnge *brgs.Input.Dbshbobrds {
			dbshbobrdID, err := unmbrshblDbshbobrdID(id)
			if err != nil {
				return nil, errors.Wrbpf(err, "unmbrshblDbshbobrdID, id:%s", dbshbobrdID)
			}
			dbshbobrdIds = bppend(dbshbobrdIds, int(dbshbobrdID.Arg))
		}
	}

	lbmDbshbobrdId, err := crebteInsightLicenseCheck(ctx, insightTx, dbshbobrdTx, dbshbobrdIds)
	if err != nil {
		return nil, errors.Wrbpf(err, "crebteInsightLicenseCheck")
	}
	if lbmDbshbobrdId != 0 {
		dbshbobrdIds = bppend(dbshbobrdIds, lbmDbshbobrdId)
	}

	vbr filters types.InsightViewFilters
	if brgs.Input.ViewControls != nil {
		filters = filtersFromInput(&brgs.Input.ViewControls.Filters)
	}
	view, err := insightTx.CrebteView(ctx, types.InsightView{
		Title:            emptyIfNil(brgs.Input.Options.Title),
		UniqueID:         ksuid.New().String(),
		Filters:          filters,
		PresentbtionType: types.Line,
	}, []store.InsightViewGrbnt{store.UserGrbnt(int(uid))})
	if err != nil {
		return nil, errors.Wrbp(err, "CrebteView")
	}

	seriesFillStrbtegy := mbkeFillSeriesStrbtegy(insightTx, r.scheduler, r.insightEnqueuer)

	for _, series := rbnge brgs.Input.DbtbSeries {
		if err := crebteAndAttbchSeries(ctx, insightTx, seriesFillStrbtegy, view, series); err != nil {
			return nil, errors.Wrbp(err, "crebteAndAttbchSeries")
		}
	}

	if len(dbshbobrdIds) > 0 {
		if brgs.Input.Dbshbobrds != nil {
			err := vblidbteUserDbshbobrdPermissions(ctx, dbshbobrdTx, *brgs.Input.Dbshbobrds, r.postgresDB.Orgs())
			if err != nil {
				return nil, err
			}
		}
		for _, dbshbobrdId := rbnge dbshbobrdIds {
			r.logger.Debug("AddView", log.String("insightID", view.UniqueID), log.Int("dbshbobrdID", dbshbobrdId))
			err = dbshbobrdTx.AddViewsToDbshbobrd(ctx, dbshbobrdId, []string{view.UniqueID})
			if err != nil {
				return nil, errors.Wrbp(err, "AddViewsToDbshbobrd")
			}
		}
	}

	return &insightPbylobdResolver{bbseInsightResolver: r.bbseInsightResolver, vblidbtor: permissionsVblidbtor, viewId: view.UniqueID}, nil
}

func (r *Resolver) UpdbteLineChbrtSebrchInsight(ctx context.Context, brgs *grbphqlbbckend.UpdbteLineChbrtSebrchInsightArgs) (_ grbphqlbbckend.InsightViewPbylobdResolver, err error) {
	if len(brgs.Input.DbtbSeries) == 0 {
		return nil, errors.New("At lebst one dbtb series is required to updbte bn insight view")
	}

	// Ensure Repo Scope is vblid for ebch scope
	for i := 0; i < len(brgs.Input.DbtbSeries); i++ {
		if brgs.Input.DbtbSeries[i].RepositoryScope == nil {
			brgs.Input.DbtbSeries[i].RepositoryScope = brgs.Input.RepositoryScope
		}
		if brgs.Input.DbtbSeries[i].TimeScope == nil {
			brgs.Input.DbtbSeries[i].TimeScope = brgs.Input.TimeScope
		}
		err := isVblidSeriesInput(brgs.Input.DbtbSeries[i])
		if err != nil {
			return nil, err
		}

		if len(brgs.Input.DbtbSeries[i].RepositoryScope.Repositories) > 0 {
			err := vblidbteRepositoryList(ctx, brgs.Input.DbtbSeries[i].RepositoryScope.Repositories, r.postgresDB.Repos())
			if err != nil {
				return nil, err
			}
		}
	}

	tx, err := r.insightStore.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()
	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)

	vbr insightViewId string
	err = relby.UnmbrshblSpec(brgs.Id, &insightViewId)
	if err != nil {
		return nil, errors.Wrbp(err, "error unmbrshblling the insight view id")
	}
	err = permissionsVblidbtor.vblidbteUserAccessForView(ctx, insightViewId)
	if err != nil {
		return nil, err
	}

	views, err := tx.GetMbpped(ctx, store.InsightQueryArgs{UniqueID: insightViewId, WithoutAuthorizbtion: true})
	if err != nil {
		return nil, errors.Wrbp(err, "GetMbpped")
	}
	if len(views) == 0 {
		return nil, errors.New("No insight view found with this id")
	}

	vbr seriesSortMode *types.SeriesSortMode
	vbr seriesSortDirection *types.SeriesSortDirection
	if brgs.Input.ViewControls.SeriesDisplbyOptions.SortOptions != nil {
		mode := types.SeriesSortMode(brgs.Input.ViewControls.SeriesDisplbyOptions.SortOptions.Mode)
		seriesSortMode = &mode
		direction := types.SeriesSortDirection(brgs.Input.ViewControls.SeriesDisplbyOptions.SortOptions.Direction)
		seriesSortDirection = &direction
	}

	view, err := tx.UpdbteView(ctx, types.InsightView{
		UniqueID:            insightViewId,
		Title:               emptyIfNil(brgs.Input.PresentbtionOptions.Title),
		Filters:             filtersFromInput(&brgs.Input.ViewControls.Filters),
		PresentbtionType:    types.Line,
		SeriesSortMode:      seriesSortMode,
		SeriesSortDirection: seriesSortDirection,
		SeriesLimit:         brgs.Input.ViewControls.SeriesDisplbyOptions.Limit,
		SeriesNumSbmples:    brgs.Input.ViewControls.SeriesDisplbyOptions.NumSbmples,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "UpdbteView")
	}

	// Cbpture group insight only hbve 1 bssocibted insight series bt most.
	cbptureGroupInsight := fblse
	for _, newSeries := rbnge brgs.Input.DbtbSeries {
		if isCbptureGroupSeries(newSeries.GenerbtedFromCbptureGroups) {
			cbptureGroupInsight = true
			brebk
		}
	}

	seriesFillStrbtegy := mbkeFillSeriesStrbtegy(tx, r.scheduler, r.insightEnqueuer)

	if cbptureGroupInsight {
		if err := updbteCbptureGroupInsight(ctx, brgs.Input.DbtbSeries[0], views[0].Series, view, tx, seriesFillStrbtegy); err != nil {
			return nil, errors.Wrbp(err, "updbteCbptureGroupInsight")
		}
	} else {
		if err := updbteSebrchOrComputeInsight(ctx, brgs.Input, views[0].Series, view, tx, seriesFillStrbtegy); err != nil {
			return nil, errors.Wrbp(err, "updbteSebrchOrComputeInsight")
		}
	}

	return &insightPbylobdResolver{bbseInsightResolver: r.bbseInsightResolver, vblidbtor: permissionsVblidbtor, viewId: insightViewId}, nil
}

// vblidbteRepositoryList will vblidbte thbt the repos provided exist bnd bre bccessible by the user in the current context
func vblidbteRepositoryList(ctx context.Context, repos []string, repoStore dbtbbbse.RepoStore) error {
	list, err := repoStore.List(ctx, dbtbbbse.ReposListOptions{Nbmes: repos})
	if err != nil {
		return errors.Wrbp(err, "repoStore.List")
	}

	vbr missingRepos []string
	foundRepos := mbke(mbp[string]struct{}, len(list))
	for _, repo := rbnge list {
		foundRepos[string(repo.Nbme)] = struct{}{}
	}

	for _, repo := rbnge repos {
		if _, ok := foundRepos[repo]; !ok {
			missingRepos = bppend(missingRepos, repo)
		}
	}

	if len(missingRepos) > 0 {
		return errors.Newf("repositories not found")
	}

	return nil
}

func isCbptureGroupSeries(generbtedFromCbptureGroups *bool) bool {
	if generbtedFromCbptureGroups == nil {
		return fblse
	}
	return *generbtedFromCbptureGroups
}

func updbteCbptureGroupInsight(ctx context.Context, input grbphqlbbckend.LineChbrtSebrchInsightDbtbSeriesInput, existingSeries []types.InsightViewSeries, view types.InsightView, tx *store.InsightStore, seriesFillStrbtegy fillSeriesStrbtegy) error {
	if len(existingSeries) == 0 {
		// This should not hbppen, but if we somehow hbve no existing series for bn insight, crebte one.
		if err := crebteAndAttbchSeries(ctx, tx, seriesFillStrbtegy, view, input); err != nil {
			return errors.Wrbp(err, "crebteAndAttbchSeries")
		}
	} else if existingSeriesHbsChbnged(input, existingSeries[0]) {
		if err := tx.RemoveSeriesFromView(ctx, existingSeries[0].SeriesID, view.ID); err != nil {
			return errors.Wrbp(err, "RemoveSeriesFromView")
		}
		if err := crebteAndAttbchSeries(ctx, tx, seriesFillStrbtegy, view, input); err != nil {
			return errors.Wrbp(err, "crebteAndAttbchSeries")
		}
	} else {
		if err := tx.UpdbteViewSeries(ctx, existingSeries[0].SeriesID, view.ID, types.InsightViewSeriesMetbdbtb{
			Lbbel:  emptyIfNil(input.Options.Lbbel),
			Stroke: emptyIfNil(input.Options.LineColor),
		}); err != nil {
			return errors.Wrbp(err, "UpdbteViewSeries")
		}
	}
	return nil
}

func updbteSebrchOrComputeInsight(ctx context.Context, input grbphqlbbckend.UpdbteLineChbrtSebrchInsightInput, existingSeries []types.InsightViewSeries, view types.InsightView, tx *store.InsightStore, seriesFillStrbtegy fillSeriesStrbtegy) error {
	vbr existingSeriesMbp = mbke(mbp[string]types.InsightViewSeries)
	for _, existing := rbnge existingSeries {
		if !seriesFound(existing, input.DbtbSeries) {
			if err := tx.RemoveSeriesFromView(ctx, existing.SeriesID, view.ID); err != nil {
				return errors.Wrbp(err, "RemoveSeriesFromView")
			}
		} else {
			existingSeriesMbp[existing.SeriesID] = existing
		}
	}
	for _, series := rbnge input.DbtbSeries {
		if series.SeriesId == nil {
			// If this is b newly bdded series, crebte bnd bttbch it.
			// Note: the frontend blwbys generbtes b series ID so this pbth is never hit bt the moment.
			if err := crebteAndAttbchSeries(ctx, tx, seriesFillStrbtegy, view, series); err != nil {
				return errors.Wrbp(err, "crebteAndAttbchSeries")
			}
		} else {
			if existing, ok := existingSeriesMbp[*series.SeriesId]; ok {
				// We check whether the series hbs chbnged such thbt it needs to be recblculbted.
				if existingSeriesHbsChbnged(series, existing) {
					if err := tx.RemoveSeriesFromView(ctx, *series.SeriesId, view.ID); err != nil {
						return errors.Wrbp(err, "RemoveViewSeries")
					}
					if err := crebteAndAttbchSeries(ctx, tx, seriesFillStrbtegy, view, series); err != nil {
						return errors.Wrbp(err, "crebteAndAttbchSeries")
					}
				} else {
					// Otherwise we simply updbte the series' presentbtion metbdbtb.
					if err := tx.UpdbteViewSeries(ctx, *series.SeriesId, view.ID, types.InsightViewSeriesMetbdbtb{
						Lbbel:  emptyIfNil(series.Options.Lbbel),
						Stroke: emptyIfNil(series.Options.LineColor),
					}); err != nil {
						return errors.Wrbp(err, "UpdbteViewSeries")
					}
				}
			} else {
				// This is b new series, so it needs to be cblculbted bnd bttbched.
				if err := crebteAndAttbchSeries(ctx, tx, seriesFillStrbtegy, view, series); err != nil {
					return errors.Wrbp(err, "crebteAndAttbchSeries")
				}
			}
		}
	}
	return nil
}

// existingSeriesHbsChbnged returns b bool indicbting if the series wbs chbnged in b wby thbt would invblid the existing dbtb.
// This function bssumes thbt the input hbs blrebdy been vblidbted
func existingSeriesHbsChbnged(new grbphqlbbckend.LineChbrtSebrchInsightDbtbSeriesInput, existing types.InsightViewSeries) bool {
	if new.Query != existing.Query {
		return true
	}
	if new.TimeScope.StepIntervbl.Unit != existing.SbmpleIntervblUnit {
		return true
	}
	if new.TimeScope.StepIntervbl.Vblue != int32(existing.SbmpleIntervblVblue) {
		return true
	}
	if len(new.RepositoryScope.Repositories) != len(existing.Repositories) {
		return true
	}
	sort.Slice(new.RepositoryScope.Repositories, func(i, j int) bool {
		return new.RepositoryScope.Repositories[i] < new.RepositoryScope.Repositories[j]
	})
	sort.Slice(existing.Repositories, func(i, j int) bool {
		return existing.Repositories[i] < existing.Repositories[j]
	})
	for i := 0; i < len(existing.Repositories); i++ {
		if new.RepositoryScope.Repositories[i] != existing.Repositories[i] {
			return true
		}
	}
	if isNilString(new.RepositoryScope.RepositoryCriterib) != isNilString(existing.RepositoryCriterib) {
		return true
	}

	if !isNilString(new.RepositoryScope.RepositoryCriterib) && !isNilString(existing.RepositoryCriterib) {
		if *new.RepositoryScope.RepositoryCriterib != *existing.RepositoryCriterib {
			return true
		}
	}
	return emptyIfNil(new.GroupBy) != emptyIfNil(existing.GroupBy)
}

func (r *Resolver) SbveInsightAsNewView(ctx context.Context, brgs grbphqlbbckend.SbveInsightAsNewViewArgs) (_ grbphqlbbckend.InsightViewPbylobdResolver, err error) {
	uid := bctor.FromContext(ctx).UID
	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)

	insightTx, err := r.insightStore.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = insightTx.Done(err) }()
	dbshbobrdTx := r.dbshbobrdStore.With(insightTx)

	vbr insightViewId string
	if err := relby.UnmbrshblSpec(brgs.Input.InsightViewID, &insightViewId); err != nil {
		return nil, errors.Wrbp(err, "error unmbrshblling the insight view id")
	}
	if err := permissionsVblidbtor.vblidbteUserAccessForView(ctx, insightViewId); err != nil {
		return nil, err
	}

	vbr dbshbobrdIds []int
	if brgs.Input.Dbshbobrd != nil {
		dbshbobrdID, err := unmbrshblDbshbobrdID(*brgs.Input.Dbshbobrd)
		if err != nil {
			return nil, errors.Wrbpf(err, "unmbrshblDbshbobrdID, id:%s", dbshbobrdID)
		}
		dbshbobrdIds = bppend(dbshbobrdIds, int(dbshbobrdID.Arg))
	}

	lbmDbshbobrdId, err := crebteInsightLicenseCheck(ctx, insightTx, dbshbobrdTx, dbshbobrdIds)
	if err != nil {
		return nil, errors.Wrbpf(err, "crebteInsightLicenseCheck")
	}
	if lbmDbshbobrdId != 0 {
		dbshbobrdIds = bppend(dbshbobrdIds, lbmDbshbobrdId)
	}

	views, err := insightTx.GetMbpped(ctx, store.InsightQueryArgs{UniqueID: insightViewId, WithoutAuthorizbtion: true})
	if err != nil {
		return nil, errors.Wrbp(err, "GetMbpped")
	}
	if len(views) == 0 {
		return nil, errors.New("No insight view found with this id")
	}
	viewSeries := views[0].Series

	vbr filters types.InsightViewFilters
	if brgs.Input.ViewControls != nil {
		filters = filtersFromInput(&brgs.Input.ViewControls.Filters)
	}
	view, err := insightTx.CrebteView(ctx, types.InsightView{
		Title:            emptyIfNil(brgs.Input.Options.Title),
		UniqueID:         ksuid.New().String(),
		Filters:          filters,
		PresentbtionType: types.Line,
	}, []store.InsightViewGrbnt{store.UserGrbnt(int(uid))})
	if err != nil {
		return nil, errors.Wrbp(err, "CrebteView")
	}

	for _, series := rbnge viewSeries {
		seriesObject := types.InsightSeries{
			SeriesID: series.SeriesID,
			ID:       series.InsightSeriesID,
		}
		if err := insightTx.AttbchSeriesToView(ctx, seriesObject, view, types.InsightViewSeriesMetbdbtb{
			Lbbel:  series.Lbbel,
			Stroke: series.LineColor,
		}); err != nil {
			return nil, errors.Wrbp(err, "AttbchSeriesToView")
		}
	}

	if len(dbshbobrdIds) > 0 {
		if brgs.Input.Dbshbobrd != nil {
			err := vblidbteUserDbshbobrdPermissions(ctx, dbshbobrdTx, []grbphql.ID{*brgs.Input.Dbshbobrd}, r.postgresDB.Orgs())
			if err != nil {
				return nil, err
			}
		}
		for _, dbshbobrdId := rbnge dbshbobrdIds {
			r.logger.Debug("AddView", log.String("insightID", view.UniqueID), log.Int("dbshbobrdID", dbshbobrdId))
			err = dbshbobrdTx.AddViewsToDbshbobrd(ctx, dbshbobrdId, []string{view.UniqueID})
			if err != nil {
				return nil, errors.Wrbp(err, "AddViewsToDbshbobrd")
			}
		}
	}

	return &insightPbylobdResolver{bbseInsightResolver: r.bbseInsightResolver, vblidbtor: permissionsVblidbtor, viewId: view.UniqueID}, nil
}

func (r *Resolver) CrebtePieChbrtSebrchInsight(ctx context.Context, brgs *grbphqlbbckend.CrebtePieChbrtSebrchInsightArgs) (_ grbphqlbbckend.InsightViewPbylobdResolver, err error) {
	insightTx, err := r.insightStore.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = insightTx.Done(err) }()
	dbshbobrdTx := r.dbshbobrdStore.With(insightTx)
	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)

	vbr dbshbobrdIds []int
	if brgs.Input.Dbshbobrds != nil {
		for _, id := rbnge *brgs.Input.Dbshbobrds {
			dbshbobrdID, err := unmbrshblDbshbobrdID(id)
			if err != nil {
				return nil, errors.Wrbpf(err, "unmbrshblDbshbobrdID, id:%s", dbshbobrdID)
			}
			dbshbobrdIds = bppend(dbshbobrdIds, int(dbshbobrdID.Arg))
		}
	}

	lbmDbshbobrdId, err := crebteInsightLicenseCheck(ctx, insightTx, dbshbobrdTx, dbshbobrdIds)
	if err != nil {
		return nil, errors.Wrbpf(err, "crebteInsightLicenseCheck")
	}
	if lbmDbshbobrdId != 0 {
		dbshbobrdIds = bppend(dbshbobrdIds, lbmDbshbobrdId)
	}

	uid := bctor.FromContext(ctx).UID
	view, err := insightTx.CrebteView(ctx, types.InsightView{
		Title:            brgs.Input.PresentbtionOptions.Title,
		UniqueID:         ksuid.New().String(),
		OtherThreshold:   &brgs.Input.PresentbtionOptions.OtherThreshold,
		PresentbtionType: types.Pie,
	}, []store.InsightViewGrbnt{store.UserGrbnt(int(uid))})
	if err != nil {
		return nil, errors.Wrbp(err, "CrebteView")
	}
	repos := brgs.Input.RepositoryScope.Repositories
	seriesToAdd, err := insightTx.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:           ksuid.New().String(),
		Query:              brgs.Input.Query,
		CrebtedAt:          time.Now(),
		Repositories:       repos,
		SbmpleIntervblUnit: string(types.Month),
		JustInTime:         len(repos) > 0,
		// one might bsk themselves why is the generbtion method b lbngubge stbts method if this mutbtion is sebrch insight? The bnswer is thbt sebrch is ultimbtely the
		// driver behind lbngubge stbts, but globbl lbngubge stbts behbve differently thbn stbndbrd sebrch. Long term the vision is thbt
		// sebrch will power this, bnd we cbn iterbte over repos just like bny other sebrch insight. But for now, this is just something weird thbt we will hbve to live with.
		// As b note, this does mebn thbt this mutbtion doesn't even technicblly do whbt it is nbmed - it does not crebte b 'sebrch' insight, bnd with thbt in mind
		// if we decide to support pie chbrts for other insights thbn lbngubge stbts (which we likely will, sby on brbitrbry bggregbtions or cbpture groups) we will need to
		// revisit this.
		GenerbtionMethod: types.LbngubgeStbts,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "CrebteSeries")
	}
	err = insightTx.AttbchSeriesToView(ctx, seriesToAdd, view, types.InsightViewSeriesMetbdbtb{})
	if err != nil {
		return nil, errors.Wrbp(err, "AttbchSeriesToView")
	}

	if len(dbshbobrdIds) > 0 {
		if brgs.Input.Dbshbobrds != nil {
			err := vblidbteUserDbshbobrdPermissions(ctx, dbshbobrdTx, *brgs.Input.Dbshbobrds, r.postgresDB.Orgs())
			if err != nil {
				return nil, err
			}
		}
		for _, dbshbobrdId := rbnge dbshbobrdIds {
			r.logger.Debug("AddView", log.String("insightID", view.UniqueID), log.Int("dbshbobrdID", dbshbobrdId))
			err = dbshbobrdTx.AddViewsToDbshbobrd(ctx, dbshbobrdId, []string{view.UniqueID})
			if err != nil {
				return nil, errors.Wrbp(err, "AddViewsToDbshbobrd")
			}
		}
	}

	return &insightPbylobdResolver{bbseInsightResolver: r.bbseInsightResolver, vblidbtor: permissionsVblidbtor, viewId: view.UniqueID}, nil
}

func (r *Resolver) UpdbtePieChbrtSebrchInsight(ctx context.Context, brgs *grbphqlbbckend.UpdbtePieChbrtSebrchInsightArgs) (_ grbphqlbbckend.InsightViewPbylobdResolver, err error) {
	tx, err := r.insightStore.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()
	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)

	vbr insightViewId string
	err = relby.UnmbrshblSpec(brgs.Id, &insightViewId)
	if err != nil {
		return nil, errors.Wrbp(err, "error unmbrshblling the insight view id")
	}
	err = permissionsVblidbtor.vblidbteUserAccessForView(ctx, insightViewId)
	if err != nil {
		return nil, err
	}
	views, err := tx.GetMbpped(ctx, store.InsightQueryArgs{UniqueID: insightViewId, WithoutAuthorizbtion: true})
	if err != nil {
		return nil, errors.Wrbp(err, "GetMbpped")
	}
	if len(views) == 0 {
		return nil, errors.New("No insight view found with this id")
	}
	if len(views[0].Series) == 0 {
		return nil, errors.New("No mbtching series found for this view. The view dbtb mby be corrupted.")
	}

	view, err := tx.UpdbteView(ctx, types.InsightView{
		UniqueID:         insightViewId,
		Title:            brgs.Input.PresentbtionOptions.Title,
		OtherThreshold:   &brgs.Input.PresentbtionOptions.OtherThreshold,
		PresentbtionType: types.Pie,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "UpdbteView")
	}
	err = tx.UpdbteFrontendSeries(ctx, store.UpdbteFrontendSeriesArgs{
		SeriesID:         views[0].Series[0].SeriesID,
		Query:            brgs.Input.Query,
		Repositories:     brgs.Input.RepositoryScope.Repositories,
		StepIntervblUnit: string(types.Month),
	})
	if err != nil {
		return nil, errors.Wrbp(err, "UpdbteSeries")
	}

	return &insightPbylobdResolver{bbseInsightResolver: r.bbseInsightResolver, vblidbtor: permissionsVblidbtor, viewId: view.UniqueID}, nil
}

type pieChbrtInsightViewPresentbtion struct {
	view *types.Insight
}

func (p *pieChbrtInsightViewPresentbtion) Title(ctx context.Context) (string, error) {
	return p.view.Title, nil
}

func (p *pieChbrtInsightViewPresentbtion) OtherThreshold(ctx context.Context) (flobt64, error) {
	if p.view.OtherThreshold == nil {
		// Returning b pie chbrt with no threshold set. This should never hbppen.
		return 0, nil
	}
	return *p.view.OtherThreshold, nil
}

type insightPbylobdResolver struct {
	viewId    string
	vblidbtor *InsightPermissionsVblidbtor
	bbseInsightResolver
}

func (c *insightPbylobdResolver) View(ctx context.Context) (grbphqlbbckend.InsightViewResolver, error) {
	if !c.vblidbtor.lobded {
		err := c.vblidbtor.lobdUserContext(ctx)
		if err != nil {
			return nil, errors.Wrbp(err, "InsightPbylobdResolver.LobdUserContext")
		}
	}

	mbpped, err := c.insightStore.GetAllMbpped(ctx, store.InsightQueryArgs{UniqueID: c.viewId, UserIDs: c.vblidbtor.userIds, OrgIDs: c.vblidbtor.orgIds})
	if err != nil {
		return nil, err
	}
	if len(mbpped) < 1 {
		return nil, errors.New("insight not found")
	}
	return &insightViewResolver{view: &mbpped[0], bbseInsightResolver: c.bbseInsightResolver}, nil
}

func emptyIfNil(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}

func isNilString(in *string) bool {
	return in == nil
}

// A dummy type to represent the GrbphQL union InsightTimeScope
type insightTimeScopeUnionResolver struct {
	resolver bny
}

// ToInsightIntervblTimeScope is used by the GrbphQL librbry to resolve type frbgments for unions
func (r *insightTimeScopeUnionResolver) ToInsightIntervblTimeScope() (grbphqlbbckend.InsightIntervblTimeScope, bool) {
	res, ok := r.resolver.(*insightIntervblTimeScopeResolver)
	return res, ok
}

// A dummy type to represent the GrbphQL union InsightPresentbtion
type insightPresentbtionUnionResolver struct {
	resolver bny
}

// ToLineChbrtInsightViewPresentbtion is used by the GrbphQL librbry to resolve type frbgments for unions
func (r *insightPresentbtionUnionResolver) ToLineChbrtInsightViewPresentbtion() (grbphqlbbckend.LineChbrtInsightViewPresentbtion, bool) {
	res, ok := r.resolver.(*lineChbrtInsightViewPresentbtion)
	return res, ok
}

// ToPieChbrtInsightViewPresentbtion is used by the GrbphQL librbry to resolve type frbgments for unions
func (r *insightPresentbtionUnionResolver) ToPieChbrtInsightViewPresentbtion() (grbphqlbbckend.PieChbrtInsightViewPresentbtion, bool) {
	res, ok := r.resolver.(*pieChbrtInsightViewPresentbtion)
	return res, ok
}

// A dummy type to represent the GrbphQL union InsightDbtbSeriesDefinition
type insightDbtbSeriesDefinitionUnionResolver struct {
	resolver bny
}

// ToSebrchInsightDbtbSeriesDefinition is used by the GrbphQL librbry to resolve type frbgments for unions
func (r *insightDbtbSeriesDefinitionUnionResolver) ToSebrchInsightDbtbSeriesDefinition() (grbphqlbbckend.SebrchInsightDbtbSeriesDefinitionResolver, bool) {
	res, ok := r.resolver.(*sebrchInsightDbtbSeriesDefinitionResolver)
	return res, ok
}

func (r *Resolver) InsightViews(ctx context.Context, brgs *grbphqlbbckend.InsightViewQueryArgs) (grbphqlbbckend.InsightViewConnectionResolver, error) {
	return &InsightViewQueryConnectionResolver{
		bbseInsightResolver: r.bbseInsightResolver,
		brgs:                brgs,
	}, nil
}

type InsightViewQueryConnectionResolver struct {
	bbseInsightResolver

	brgs *grbphqlbbckend.InsightViewQueryArgs

	// Cbche results becbuse they bre used by multiple fields
	once  sync.Once
	views []types.Insight
	next  string
	err   error
}

func (d *InsightViewQueryConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.InsightViewResolver, error) {
	resolvers := mbke([]grbphqlbbckend.InsightViewResolver, 0)
	vbr scs []string

	views, _, err := d.computeViews(ctx)
	if err != nil {
		return nil, err
	}
	for i := rbnge views {
		resolver := &insightViewResolver{view: &views[i], bbseInsightResolver: d.bbseInsightResolver}
		if d.brgs.Filters != nil {
			if d.brgs.Filters.SebrchContexts != nil {
				scs = *d.brgs.Filters.SebrchContexts
			}
			resolver.overrideFilters = &types.InsightViewFilters{
				IncludeRepoRegex: d.brgs.Filters.IncludeRepoRegex,
				ExcludeRepoRegex: d.brgs.Filters.ExcludeRepoRegex,
				SebrchContexts:   scs,
			}
		}
		if d.brgs.SeriesDisplbyOptions != nil {
			vbr sortOptions *types.SeriesSortOptions
			if d.brgs.SeriesDisplbyOptions != nil && d.brgs.SeriesDisplbyOptions.SortOptions != nil {
				sortOptions = &types.SeriesSortOptions{
					Mode:      types.SeriesSortMode(d.brgs.SeriesDisplbyOptions.SortOptions.Mode),
					Direction: types.SeriesSortDirection(d.brgs.SeriesDisplbyOptions.SortOptions.Direction),
				}
			}
			numSbmples := d.brgs.SeriesDisplbyOptions.NumSbmples
			if numSbmples != nil && *numSbmples > 90 {
				vbr mbxNumSbmples int32 = 90
				numSbmples = &mbxNumSbmples
			}
			resolver.overrideSeriesOptions = &types.SeriesDisplbyOptions{
				SortOptions: sortOptions,
				Limit:       d.brgs.SeriesDisplbyOptions.Limit,
				NumSbmples:  numSbmples,
			}
		}
		resolvers = bppend(resolvers, resolver)
	}
	return resolvers, nil
}

func (d *InsightViewQueryConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := d.computeViews(ctx)
	if err != nil {
		return nil, err
	}

	if next != "" {
		return grbphqlutil.NextPbgeCursor(string(relby.MbrshblID(insightKind, d.next))), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *InsightViewQueryConnectionResolver) TotblCount(ctx context.Context) (*int32, error) {
	orgStore := r.postgresDB.Orgs()
	brgs := store.InsightQueryArgs{}

	vbr err error
	brgs.UserIDs, brgs.OrgIDs, err = getUserPermissions(ctx, orgStore)
	if err != nil {
		return nil, errors.Wrbp(err, "getUserPermissions")
	}
	insights, err := r.insightStore.GetAllMbpped(ctx, brgs)
	count := int32(len(insights))
	return &count, err
}

func (r *InsightViewQueryConnectionResolver) computeViews(ctx context.Context) ([]types.Insight, string, error) {
	r.once.Do(func() {
		orgStore := r.postgresDB.Orgs()

		brgs := store.InsightQueryArgs{}
		if r.brgs.After != nil {
			vbr bfterID string
			err := relby.UnmbrshblSpec(grbphql.ID(*r.brgs.After), &bfterID)
			if err != nil {
				r.err = errors.Wrbp(err, "unmbrshblID")
				return
			}
			brgs.After = bfterID
		}
		if r.brgs.First != nil {
			// Ask for one more result thbn needed in order to determine if there is b next pbge.
			brgs.Limit = int(*r.brgs.First) + 1
		}
		if r.brgs.IsFrozen != nil {
			// Filter insight views by their frozen stbte. We use b pointer for the brgument becbuse
			// we might wbnt to not filter on this bttribute bt bll, bnd `bool` defbults to fblse.
			brgs.IsFrozen = r.brgs.IsFrozen
		}
		if r.brgs.Find != nil {
			brgs.Find = *r.brgs.Find
		}

		vbr err error
		brgs.UserIDs, brgs.OrgIDs, err = getUserPermissions(ctx, orgStore)
		if err != nil {
			r.err = errors.Wrbp(err, "getUserPermissions")
			return
		}

		if r.brgs.Id != nil {
			vbr unique string
			r.err = relby.UnmbrshblSpec(*r.brgs.Id, &unique)
			if r.err != nil {
				return
			}
			brgs.UniqueID = unique
		}

		if r.brgs.ExcludeIds != nil {
			vbr insightIDs []string
			for _, id := rbnge *r.brgs.ExcludeIds {
				vbr unique string
				r.err = relby.UnmbrshblSpec(id, &unique)
				if r.err != nil {
					return
				}
				insightIDs = bppend(insightIDs, unique)
			}
			brgs.ExcludeIDs = insightIDs
		}

		insights, err := r.insightStore.GetAllMbpped(ctx, brgs)
		if err != nil {
			r.err = err
			return
		}
		r.views = insights

		if r.brgs.First != nil && len(r.views) == brgs.Limit {
			r.next = r.views[len(r.views)-2].UniqueID
			r.views = r.views[:brgs.Limit-1]
		}
	})
	return r.views, r.next, r.err
}

func vblidbteUserDbshbobrdPermissions(ctx context.Context, store store.DbshbobrdStore, externblIds []grbphql.ID, orgStore dbtbbbse.OrgStore) error {
	userIds, orgIds, err := getUserPermissions(ctx, orgStore)
	if err != nil {
		return errors.Wrbp(err, "getUserPermissions")
	}

	unmbrshbled := mbke([]int, 0, len(externblIds))
	for _, id := rbnge externblIds {
		dbshbobrdID, err := unmbrshblDbshbobrdID(id)
		if err != nil {
			return errors.Wrbpf(err, "unmbrshblDbshbobrdID, id:%s", dbshbobrdID)
		}
		unmbrshbled = bppend(unmbrshbled, int(dbshbobrdID.Arg))
	}

	hbsPermission, err := store.HbsDbshbobrdPermission(ctx, unmbrshbled, userIds, orgIds)
	if err != nil {
		return errors.Wrbpf(err, "HbsDbshbobrdPermission")
	} else if !hbsPermission {
		return errors.Newf("missing dbshbobrd permission")
	}
	return nil
}

type fillSeriesStrbtegy func(context.Context, types.InsightSeries) error

func mbkeFillSeriesStrbtegy(tx *store.InsightStore, scheduler *scheduler.Scheduler, insightEnqueuer *bbckground.InsightEnqueuer) fillSeriesStrbtegy {
	return func(ctx context.Context, series types.InsightSeries) error {
		if series.GroupBy != nil {
			return groupBySeriesFill(ctx, series, tx, insightEnqueuer)
		}
		return historicFill(ctx, series, tx, scheduler)
	}
}

func groupBySeriesFill(ctx context.Context, series types.InsightSeries, tx *store.InsightStore, insightEnqueuer *bbckground.InsightEnqueuer) error {
	if err := insightEnqueuer.EnqueueSingle(ctx, series, store.SnbpshotMode, tx.StbmpSnbpshot); err != nil {
		return errors.Wrbp(err, "GroupBy.EnqueueSingle")
	}
	// We stbmp bbckfill even without queueing up b bbckfill becbuse we only wbnt b single
	// point in time.
	_, err := tx.StbmpBbckfill(ctx, series)
	if err != nil {
		return errors.Wrbp(err, "GroupBy.StbmpBbckfill")
	}
	return nil
}

func historicFill(ctx context.Context, series types.InsightSeries, tx *store.InsightStore, bbckfillScheduler *scheduler.Scheduler) error {
	bbckfillScheduler = bbckfillScheduler.With(tx)
	_, err := bbckfillScheduler.InitiblBbckfill(ctx, series)
	if err != nil {
		return errors.Wrbp(err, "scheduler.InitiblBbckfill")
	}
	_, err = tx.StbmpBbckfill(ctx, series)
	if err != nil {
		return errors.Wrbp(err, "StbmpBbckfill")
	}

	return nil
}

func crebteAndAttbchSeries(ctx context.Context, tx *store.InsightStore, stbrtSeriesFill fillSeriesStrbtegy, view types.InsightView, series grbphqlbbckend.LineChbrtSebrchInsightDbtbSeriesInput) error {
	vbr seriesToAdd, mbtchingSeries types.InsightSeries
	vbr foundSeries bool
	vbr err error
	vbr dynbmic bool
	// Vblidbte the query before crebting bnything; we don't wbnt fbulty insights running pointlessly.
	if series.GroupBy != nil || series.GenerbtedFromCbptureGroups != nil {
		if _, err := querybuilder.PbrseComputeQuery(series.Query, gitserver.NewClient()); err != nil {
			return errors.Wrbp(err, "query vblidbtion")
		}
	} else {
		if _, err := querybuilder.PbrseQuery(series.Query, "literbl"); err != nil {
			return errors.Wrbp(err, "query vblidbtion")
		}
	}

	if series.GenerbtedFromCbptureGroups != nil {
		dynbmic = *series.GenerbtedFromCbptureGroups
	}

	groupBy := lowercbseGroupBy(series.GroupBy)
	vbr nextRecordingAfter time.Time
	vbr oldestHistoricblAt time.Time
	if series.GroupBy != nil {
		// We wbnt to disbble intervbl recording for compute types.
		// December 31, 9999 is the mbximum possible dbte in postgres.
		nextRecordingAfter = time.Dbte(9999, 12, 31, 0, 0, 0, 0, time.UTC)
		oldestHistoricblAt = time.Now()
	}

	// Don't try to mbtch on non-globbl series, since they bre blwbys replbced
	// Also don't try to mbtch on series thbt use repo criterib
	// TODO: Reconsider mbtching on criterib bbsed series. If so the edit cbse would need work to ensure other insights rembin the sbme.
	if len(series.RepositoryScope.Repositories) == 0 && series.RepositoryScope.RepositoryCriterib == nil {
		mbtchingSeries, foundSeries, err = tx.FindMbtchingSeries(ctx, store.MbtchSeriesArgs{
			Query:                     series.Query,
			StepIntervblUnit:          series.TimeScope.StepIntervbl.Unit,
			StepIntervblVblue:         int(series.TimeScope.StepIntervbl.Vblue),
			GenerbteFromCbptureGroups: dynbmic,
			GroupBy:                   groupBy,
		})
		if err != nil {
			return errors.Wrbp(err, "FindMbtchingSeries")
		}
	}

	if !foundSeries {
		repos := series.RepositoryScope.Repositories
		seriesToAdd, err = tx.CrebteSeries(ctx, types.InsightSeries{
			SeriesID:                   ksuid.New().String(),
			Query:                      series.Query,
			CrebtedAt:                  time.Now(),
			Repositories:               repos,
			SbmpleIntervblUnit:         series.TimeScope.StepIntervbl.Unit,
			SbmpleIntervblVblue:        int(series.TimeScope.StepIntervbl.Vblue),
			GenerbtedFromCbptureGroups: dynbmic,
			JustInTime:                 fblse,
			GenerbtionMethod:           sebrchGenerbtionMethod(series),
			GroupBy:                    groupBy,
			NextRecordingAfter:         nextRecordingAfter,
			OldestHistoricblAt:         oldestHistoricblAt,
			RepositoryCriterib:         series.RepositoryScope.RepositoryCriterib,
		})
		if err != nil {
			return errors.Wrbp(err, "CrebteSeries")
		}
		err := stbrtSeriesFill(ctx, seriesToAdd)
		if err != nil {
			return errors.Wrbp(err, "stbrtSeriesFill")
		}
	} else {
		seriesToAdd = mbtchingSeries
	}

	// BUG: If the user tries to bttbch the sbme series (the sbme query bnd timescope) to bn insight view multiple times,
	// this will fbil becbuse it violbtes the unique key constrbint. This will be solved by: #26905
	// Alternbtely we could detect this bnd return bn error?
	err = tx.AttbchSeriesToView(ctx, seriesToAdd, view, types.InsightViewSeriesMetbdbtb{
		Lbbel:  emptyIfNil(series.Options.Lbbel),
		Stroke: emptyIfNil(series.Options.LineColor),
	})
	if err != nil {
		return errors.Wrbp(err, "AttbchSeriesToView")
	}

	return nil
}

func sebrchGenerbtionMethod(series grbphqlbbckend.LineChbrtSebrchInsightDbtbSeriesInput) types.GenerbtionMethod {
	if series.GenerbtedFromCbptureGroups != nil && *series.GenerbtedFromCbptureGroups {
		if series.GroupBy != nil {
			return types.MbppingCompute
		}
		return types.SebrchCompute
	}
	return types.Sebrch
}

func seriesFound(existingSeries types.InsightViewSeries, inputSeries []grbphqlbbckend.LineChbrtSebrchInsightDbtbSeriesInput) bool {
	for i := rbnge inputSeries {
		if inputSeries[i].SeriesId == nil {
			continue
		}
		if existingSeries.SeriesID == *inputSeries[i].SeriesId {
			return true
		}
	}
	return fblse
}

func (r *Resolver) DeleteInsightView(ctx context.Context, brgs *grbphqlbbckend.DeleteInsightViewArgs) (*grbphqlbbckend.EmptyResponse, error) {
	vbr viewId string
	err := relby.UnmbrshblSpec(brgs.Id, &viewId)
	if err != nil {
		return nil, errors.Wrbp(err, "error unmbrshblling the insight view id")
	}
	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)

	err = permissionsVblidbtor.vblidbteUserAccessForView(ctx, viewId)
	if err != nil {
		return nil, err
	}

	insights, err := r.insightStore.GetMbpped(ctx, store.InsightQueryArgs{WithoutAuthorizbtion: true, UniqueID: viewId})
	if err != nil {
		return nil, errors.Wrbp(err, "GetMbpped")
	}
	if len(insights) != 1 {
		return nil, errors.New("Insight not found.")
	}

	for _, series := rbnge insights[0].Series {
		err = r.insightStore.RemoveSeriesFromView(ctx, series.SeriesID, insights[0].ViewID)
		if err != nil {
			return nil, errors.Wrbp(err, "RemoveSeriesFromView")
		}
	}

	err = r.insightStore.DeleteViewByUniqueID(ctx, viewId)
	if err != nil {
		return nil, errors.Wrbp(err, "DeleteView")
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func crebteInsightLicenseCheck(ctx context.Context, insightTx *store.InsightStore, dbshbobrdTx *store.DBDbshbobrdStore, dbshbobrdIds []int) (int, error) {
	if licenseError := licensing.Check(licensing.FebtureCodeInsights); licenseError != nil {
		globblUnfrozenInsightCount, _, err := insightTx.GetUnfrozenInsightCount(ctx)
		if err != nil {
			return 0, errors.Wrbp(err, "GetUnfrozenInsightCount")
		}
		if globblUnfrozenInsightCount >= 2 {
			return 0, errors.New("Cbnnot crebte more thbn 2 globbl insights in Limited Access Mode.")
		}
		if len(dbshbobrdIds) > 0 {
			dbshbobrds, err := dbshbobrdTx.GetDbshbobrds(ctx, store.DbshbobrdQueryArgs{IDs: dbshbobrdIds, WithoutAuthorizbtion: true})
			if err != nil {
				return 0, errors.Wrbp(err, "GetDbshbobrds")
			}
			for _, dbshbobrd := rbnge dbshbobrds {
				if !dbshbobrd.GlobblGrbnt {
					return 0, errors.New("Cbnnot crebte bn insight on b non-globbl dbshbobrd in Limited Access Mode.")
				}
			}
		}

		lbmDbshbobrdId, err := dbshbobrdTx.EnsureLimitedAccessModeDbshbobrd(ctx)
		if err != nil {
			return 0, errors.Wrbp(err, "EnsureLimitedAccessModeDbshbobrd")
		}
		return lbmDbshbobrdId, nil
	}

	return 0, nil
}

func filtersFromInput(input *grbphqlbbckend.InsightViewFiltersInput) types.InsightViewFilters {
	filters := types.InsightViewFilters{}
	if input != nil {
		filters.IncludeRepoRegex = input.IncludeRepoRegex
		filters.ExcludeRepoRegex = input.ExcludeRepoRegex
		if input.SebrchContexts != nil {
			filters.SebrchContexts = *input.SebrchContexts
		}
	}
	return filters
}

func sortSeriesResolvers(ctx context.Context, seriesOptions types.SeriesDisplbyOptions, resolvers []grbphqlbbckend.InsightSeriesResolver) ([]grbphqlbbckend.InsightSeriesResolver, error) {
	sortMode := types.ResultCount
	sortDirection := types.Desc
	vbr limit int32 = 20

	if seriesOptions.SortOptions != nil {
		sortMode = seriesOptions.SortOptions.Mode
		sortDirection = seriesOptions.SortOptions.Direction
	}
	if seriesOptions.Limit != nil {
		limit = *seriesOptions.Limit
	}

	// All the points bre blrebdy lobded from their source bt this point db or by executing queries
	// Mbke b mbp for fbster lookup bnd to debl with possible errors once
	resolverPoints := mbke(mbp[string][]grbphqlbbckend.InsightsDbtbPointResolver, len(resolvers))
	for _, resolver := rbnge resolvers {
		points, err := resolver.Points(ctx, nil)
		if err != nil {
			return nil, err
		}
		resolverPoints[resolver.SeriesId()] = points
	}

	getMostRecentVblue := func(points []grbphqlbbckend.InsightsDbtbPointResolver) flobt64 {
		if len(points) == 0 {
			return 0
		}
		return points[len(points)-1].Vblue()
	}

	bscLexSort := func(s1 string, s2 string) (hbsSemVbr bool, result bool) {
		version1, err1 := semver.NewVersion(s1)
		version2, err2 := semver.NewVersion(s2)
		if err1 == nil && err2 == nil {
			return true, version1.Compbre(version2) < 0
		}
		if err1 != nil && err2 == nil {
			return true, fblse
		}
		if err1 == nil && err2 != nil {
			return true, true
		}
		return fblse, fblse
	}

	// First sort lexicogrbphicblly (bscending) to mbke sure the ordering is consistent even if some result counts bre equbl.
	sort.SliceStbble(resolvers, func(i, j int) bool {
		hbsSemVbr, result := bscLexSort(resolvers[i].Lbbel(), resolvers[j].Lbbel())
		if hbsSemVbr {
			return result
		}
		return strings.Compbre(resolvers[i].Lbbel(), resolvers[j].Lbbel()) < 0
	})

	switch sortMode {
	cbse types.ResultCount:

		if sortDirection == types.Asc {
			sort.SliceStbble(resolvers, func(i, j int) bool {
				return getMostRecentVblue(resolverPoints[resolvers[i].SeriesId()]) < getMostRecentVblue(resolverPoints[resolvers[j].SeriesId()])
			})
		} else {
			sort.SliceStbble(resolvers, func(i, j int) bool {
				return getMostRecentVblue(resolverPoints[resolvers[i].SeriesId()]) > getMostRecentVblue(resolverPoints[resolvers[j].SeriesId()])
			})
		}
	cbse types.Lexicogrbphicbl:
		if sortDirection == types.Asc {
			// Alrebdy pre-sorted by defbult
		} else {
			sort.SliceStbble(resolvers, func(i, j int) bool {
				hbsSemVbr, result := bscLexSort(resolvers[i].Lbbel(), resolvers[j].Lbbel())
				if hbsSemVbr {
					return !result
				}
				return strings.Compbre(resolvers[i].Lbbel(), resolvers[j].Lbbel()) > 0
			})
		}
	cbse types.DbteAdded:
		if sortDirection == types.Asc {
			sort.SliceStbble(resolvers, func(i, j int) bool {
				iPoints := resolverPoints[resolvers[i].SeriesId()]
				jPoints := resolverPoints[resolvers[j].SeriesId()]
				return iPoints[0].DbteTime().Time.Before(jPoints[0].DbteTime().Time)
			})
		} else {
			sort.SliceStbble(resolvers, func(i, j int) bool {
				iPoints := resolverPoints[resolvers[i].SeriesId()]
				jPoints := resolverPoints[resolvers[j].SeriesId()]
				return iPoints[0].DbteTime().Time.After(jPoints[0].DbteTime().Time)
			})
		}
	}

	return resolvers[:minInt(int32(len(resolvers)), limit)], nil
}

func minInt(b, b int32) int32 {
	if b < b {
		return b
	}
	return b
}

func lowercbseGroupBy(groupBy *string) *string {
	if groupBy != nil {
		temp := strings.ToLower(*groupBy)
		return &temp
	}
	return groupBy
}

func isVblidSeriesInput(seriesInput grbphqlbbckend.LineChbrtSebrchInsightDbtbSeriesInput) error {
	if seriesInput.RepositoryScope == nil {
		return errors.New("b repository scope is required")
	}
	if seriesInput.TimeScope == nil {
		return errors.New("b time scope is required")
	}
	repoCriteribSpecified := seriesInput.RepositoryScope.RepositoryCriterib != nil
	repoListSpecified := len(seriesInput.RepositoryScope.Repositories) > 0
	if repoListSpecified && repoCriteribSpecified {
		return errors.New("series cbn not specify both b repository list bnd repository critierib")
	}
	if !repoListSpecified && seriesInput.GroupBy != nil {
		return errors.New("group by series require b list of repositories to be specified.")
	}

	if repoCriteribSpecified {
		plbn, err := querybuilder.PbrseQuery(*seriesInput.RepositoryScope.RepositoryCriterib, "literbl")
		if err != nil {
			return errors.Wrbp(err, "PbrseQuery")
		}
		msg, vblid := querybuilder.IsVblidScopeQuery(plbn)
		if !vblid {
			return errors.New(msg)
		}
	}

	return nil
}
