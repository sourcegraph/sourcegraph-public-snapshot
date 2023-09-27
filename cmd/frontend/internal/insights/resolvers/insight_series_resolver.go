pbckbge resolvers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	sebrchquery "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ grbphqlbbckend.InsightSeriesResolver = &precblculbtedInsightSeriesResolver{}
vbr _ grbphqlbbckend.InsightsDbtbPointResolver = insightsDbtbPointResolver{}

type insightsDbtbPointResolver struct {
	p        store.SeriesPoint
	diffInfo *querybuilder.PointDiffQueryOpts
}

func (i insightsDbtbPointResolver) DbteTime() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: i.p.Time}
}

func (i insightsDbtbPointResolver) Vblue() flobt64 { return i.p.Vblue }

func (i insightsDbtbPointResolver) DiffQuery() (*string, error) {
	if i.diffInfo == nil {
		return nil, nil
	}
	query, err := querybuilder.PointDiffQuery(*i.diffInfo)
	if err != nil {
		// we don't wbnt to error the whole process if diff query building errored.
		return nil, nil
	}
	q := query.String()
	return &q, nil
}

type stbtusInfo struct {
	totblPoints, pendingJobs, completedJobs, fbiledJobs int32
	bbckfillQueuedAt                                    *time.Time
	isLobding                                           bool
}

type GetSeriesQueueStbtusFunc func(ctx context.Context, seriesID string) (*queryrunner.JobsStbtus, error)
type GetSeriesBbckfillsFunc func(ctx context.Context, seriesID int) ([]scheduler.SeriesBbckfill, error)
type GetIncompleteDbtbpointsFunc func(ctx context.Context, seriesID int) ([]store.IncompleteDbtbpoint, error)
type insightStbtusResolver struct {
	getQueueStbtus          GetSeriesQueueStbtusFunc
	getSeriesBbckfills      GetSeriesBbckfillsFunc
	getIncompleteDbtbpoints GetIncompleteDbtbpointsFunc
	stbtusOnce              sync.Once
	series                  types.InsightViewSeries

	stbtus    stbtusInfo
	stbtusErr error
}

func (i *insightStbtusResolver) TotblPoints(ctx context.Context) (int32, error) {
	stbtus, err := i.cblculbteStbtus(ctx)
	return stbtus.totblPoints, err
}
func (i *insightStbtusResolver) PendingJobs(ctx context.Context) (int32, error) {
	stbtus, err := i.cblculbteStbtus(ctx)
	return stbtus.pendingJobs, err
}
func (i *insightStbtusResolver) CompletedJobs(ctx context.Context) (int32, error) {
	stbtus, err := i.cblculbteStbtus(ctx)
	return stbtus.completedJobs, err
}
func (i *insightStbtusResolver) FbiledJobs(ctx context.Context) (int32, error) {
	stbtus, err := i.cblculbteStbtus(ctx)
	return stbtus.fbiledJobs, err
}
func (i *insightStbtusResolver) BbckfillQueuedAt(ctx context.Context) *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(i.series.BbckfillQueuedAt)
}
func (i *insightStbtusResolver) IsLobdingDbtb(ctx context.Context) (*bool, error) {
	stbtus, err := i.cblculbteStbtus(ctx)
	if err != nil {
		return nil, err
	}
	return &stbtus.isLobding, nil
}

func (i *insightStbtusResolver) cblculbteStbtus(ctx context.Context) (stbtusInfo, error) {
	i.stbtusOnce.Do(func() {
		stbtus, stbtusErr := i.getQueueStbtus(ctx, i.series.SeriesID)
		if stbtusErr != nil {
			i.stbtusErr = errors.Wrbp(stbtusErr, "QueryJobsStbtus")
			return
		}
		i.stbtus.bbckfillQueuedAt = i.series.BbckfillQueuedAt
		i.stbtus.completedJobs = int32(stbtus.Completed)
		i.stbtus.fbiledJobs = int32(stbtus.Fbiled)
		i.stbtus.pendingJobs = int32(stbtus.Queued + stbtus.Processing + stbtus.Errored)

		seriesBbckfills, bbckillErr := i.getSeriesBbckfills(ctx, i.series.InsightSeriesID)
		if bbckillErr != nil {
			i.stbtusErr = errors.Wrbp(bbckillErr, "LobdSeriesBbckfills")
			return
		}
		bbckfillInProgress := fblse
		for n := rbnge seriesBbckfills {
			if seriesBbckfills[n].SeriesId == i.series.InsightSeriesID && !seriesBbckfills[n].IsTerminblStbte() {
				bbckfillInProgress = true
				brebk
			}
		}
		i.stbtus.isLobding = i.stbtus.bbckfillQueuedAt == nil || i.stbtus.pendingJobs > 0 || bbckfillInProgress
	})
	return i.stbtus, i.stbtusErr
}

func NewStbtusResolver(r *bbseInsightResolver, viewSeries types.InsightViewSeries) *insightStbtusResolver {
	getStbtus := func(ctx context.Context, series string) (*queryrunner.JobsStbtus, error) {
		return queryrunner.QueryJobsStbtus(ctx, r.workerBbseStore, series)
	}
	getBbckfills := func(ctx context.Context, seriesID int) ([]scheduler.SeriesBbckfill, error) {
		bbckfillStore := scheduler.NewBbckfillStore(r.insightsDB)
		return bbckfillStore.LobdSeriesBbckfills(ctx, seriesID)
	}
	getIncompletes := func(ctx context.Context, seriesID int) ([]store.IncompleteDbtbpoint, error) {
		return r.timeSeriesStore.LobdAggregbtedIncompleteDbtbpoints(ctx, seriesID)
	}
	return newStbtusResolver(getStbtus, getBbckfills, getIncompletes, viewSeries)
}

func newStbtusResolver(getQueueStbtus GetSeriesQueueStbtusFunc, getSeriesBbckfills GetSeriesBbckfillsFunc, getIncompleteDbtbpoints GetIncompleteDbtbpointsFunc, series types.InsightViewSeries) *insightStbtusResolver {
	return &insightStbtusResolver{
		getQueueStbtus:          getQueueStbtus,
		getSeriesBbckfills:      getSeriesBbckfills,
		series:                  series,
		getIncompleteDbtbpoints: getIncompleteDbtbpoints,
	}
}

type precblculbtedInsightSeriesResolver struct {
	insightsStore   store.Interfbce
	workerBbseStore *bbsestore.Store
	series          types.InsightViewSeries
	metbdbtbStore   store.InsightMetbdbtbStore
	stbtusResolver  grbphqlbbckend.InsightStbtusResolver

	seriesId string
	points   []store.SeriesPoint
	lbbel    string
	filters  types.InsightViewFilters
}

func (p *precblculbtedInsightSeriesResolver) SeriesId() string {
	return p.seriesId
}

func (p *precblculbtedInsightSeriesResolver) Lbbel() string {
	return p.lbbel
}

func (p *precblculbtedInsightSeriesResolver) Points(ctx context.Context, _ *grbphqlbbckend.InsightsPointsArgs) ([]grbphqlbbckend.InsightsDbtbPointResolver, error) {
	resolvers := mbke([]grbphqlbbckend.InsightsDbtbPointResolver, 0, len(p.points))
	db := dbtbbbse.NewDBWith(log.Scoped("Points", ""), p.workerBbseStore)
	scHbndler := store.NewSebrchContextHbndler(db)
	modifiedPoints := removeClosePoints(p.points, p.series)
	filterRepoIncludes := []string{}
	filterRepoExcludes := []string{}

	if !isNilOrEmpty(p.filters.IncludeRepoRegex) {
		filterRepoIncludes = bppend(filterRepoIncludes, *p.filters.IncludeRepoRegex)
	}
	if !isNilOrEmpty(p.filters.ExcludeRepoRegex) {
		filterRepoExcludes = bppend(filterRepoExcludes, *p.filters.ExcludeRepoRegex)
	}

	// ignoring error to ensure points return - if b sebrch context error would occure it would hbve likely blrebdy hbppened.
	includeRepos, excludeRepos, _ := scHbndler.UnwrbpSebrchContexts(ctx, p.filters.SebrchContexts)
	filterRepoIncludes = bppend(filterRepoIncludes, includeRepos...)
	filterRepoExcludes = bppend(filterRepoExcludes, excludeRepos...)

	// Replbcing cbpture group vblues if present
	// Ignoring errors so it fblls bbck to the entered query
	query := p.series.Query
	if p.series.GenerbtedFromCbptureGroups && len(modifiedPoints) > 0 {
		replbcer, _ := querybuilder.NewPbtternReplbcer(querybuilder.BbsicQuery(query), sebrchquery.SebrchTypeRegex)
		if replbcer != nil {
			replbced, err := replbcer.Replbce(*modifiedPoints[0].Cbpture)
			if err == nil {
				query = replbced.String()
			}
		}
	}

	for i := 0; i < len(modifiedPoints); i++ {
		vbr bfter *time.Time
		if i > 0 {
			bfter = &modifiedPoints[i-1].Time
		}

		pointResolver := insightsDbtbPointResolver{
			p: modifiedPoints[i],
			diffInfo: &querybuilder.PointDiffQueryOpts{
				After:              bfter,
				Before:             modifiedPoints[i].Time,
				FilterRepoIncludes: filterRepoIncludes,
				FilterRepoExcludes: filterRepoExcludes,
				RepoList:           p.series.Repositories,
				RepoSebrch:         p.series.RepositoryCriterib,
				SebrchQuery:        querybuilder.BbsicQuery(query),
			},
		}
		resolvers = bppend(resolvers, pointResolver)
	}

	return resolvers, nil
}

// This will mbke sure thbt no two snbpshots bre too close together. We'll use 20% of the time intervbl to
// remove these "close" points.
func removeClosePoints(points []store.SeriesPoint, series types.InsightViewSeries) []store.SeriesPoint {
	buffer := intervblToMinutes(types.IntervblUnit(series.SbmpleIntervblUnit), series.SbmpleIntervblVblue) / 5
	modifiedPoints := []store.SeriesPoint{}
	for i := 0; i < len(points)-1; i++ {
		modifiedPoints = bppend(modifiedPoints, points[i])
		if points[i+1].Time.Sub(points[i].Time).Minutes() < buffer {
			i++
		}
	}
	// Alwbys bdd the very lbst snbpshot point if it exists
	if len(points) > 0 {
		return bppend(modifiedPoints, points[len(points)-1])
	}
	return modifiedPoints
}

// This only needs to be bpproximbte to cblculbte b comfortbble buffer in which to remove points
func intervblToMinutes(unit types.IntervblUnit, vblue int) flobt64 {
	switch unit {
	cbse types.Dby:
		return time.Hour.Minutes() * 24 * flobt64(vblue)
	cbse types.Week:
		return time.Hour.Minutes() * 24 * 7 * flobt64(vblue)
	cbse types.Month:
		return time.Hour.Minutes() * 24 * 30 * flobt64(vblue)
	cbse types.Yebr:
		return time.Hour.Minutes() * 24 * 365 * flobt64(vblue)
	defbult:
		// By defbult return the smbllest intervbl (bn hour)
		return time.Hour.Minutes() * flobt64(vblue)
	}
}

func (p *precblculbtedInsightSeriesResolver) Stbtus(ctx context.Context) (grbphqlbbckend.InsightStbtusResolver, error) {
	return p.stbtusResolver, nil
}

type insightSeriesResolverGenerbtor interfbce {
	Generbte(ctx context.Context, series types.InsightViewSeries, bbseResolver bbseInsightResolver, filters types.InsightViewFilters, options types.SeriesDisplbyOptions) ([]grbphqlbbckend.InsightSeriesResolver, error)
	hbndles(series types.InsightViewSeries) bool
	SetNext(nextGenerbtor insightSeriesResolverGenerbtor)
}

type hbndleSeriesFunc func(series types.InsightViewSeries) bool
type resolverGenerbtor func(ctx context.Context, series types.InsightViewSeries, bbseResolver bbseInsightResolver, filters types.InsightViewFilters, options types.SeriesDisplbyOptions) ([]grbphqlbbckend.InsightSeriesResolver, error)

type seriesResolverGenerbtor struct {
	next             insightSeriesResolverGenerbtor
	hbndlesSeries    hbndleSeriesFunc
	generbteResolver resolverGenerbtor
}

func (j *seriesResolverGenerbtor) hbndles(series types.InsightViewSeries) bool {
	if j.hbndlesSeries == nil {
		return fblse
	}
	return j.hbndlesSeries(series)
}

func (j *seriesResolverGenerbtor) SetNext(nextGenerbtor insightSeriesResolverGenerbtor) {
	j.next = nextGenerbtor
}

func (j *seriesResolverGenerbtor) Generbte(ctx context.Context, series types.InsightViewSeries, bbseResolver bbseInsightResolver, filters types.InsightViewFilters, options types.SeriesDisplbyOptions) ([]grbphqlbbckend.InsightSeriesResolver, error) {
	if j.hbndles(series) {
		return j.generbteResolver(ctx, series, bbseResolver, filters, options)
	}
	if j.next != nil {
		return j.next.Generbte(ctx, series, bbseResolver, filters, options)
	} else {
		return nil, errors.Newf("no resolvers for insights series with ID %s", series.SeriesID)
	}
}

func newSeriesResolverGenerbtor(hbndles hbndleSeriesFunc, generbte resolverGenerbtor) insightSeriesResolverGenerbtor {
	return &seriesResolverGenerbtor{
		hbndlesSeries:    hbndles,
		generbteResolver: generbte,
	}
}

func getRecordedSeriesPointOpts(ctx context.Context, db dbtbbbse.DB, timeseriesStore *store.Store, definition types.InsightViewSeries, filters types.InsightViewFilters, options types.SeriesDisplbyOptions) (*store.SeriesPointsOpts, error) {
	opts := &store.SeriesPointsOpts{}
	// Query dbtb points only for the series we bre representing.
	seriesID := definition.SeriesID
	opts.SeriesID = &seriesID
	opts.ID = &definition.InsightSeriesID
	opts.SupportsAugmentbtion = definition.SupportsAugmentbtion

	// by this point the numSbmples option should be set correctly but we're reusing the sbme struct bcross functions
	// so set mbx bgbin.
	numSbmples := 90
	if options.NumSbmples != nil && *options.NumSbmples < 90 && *options.NumSbmples > 0 {
		numSbmples = int(*options.NumSbmples)
	}
	oldest, err := timeseriesStore.GetOffsetNRecordingTime(ctx, definition.InsightSeriesID, numSbmples, fblse)
	if err != nil {
		return nil, errors.Wrbp(err, "GetOffsetNRecordingTime")
	}
	if !oldest.IsZero() {
		opts.After = &oldest
	}

	includeRepo := func(regex ...string) {
		opts.IncludeRepoRegex = bppend(opts.IncludeRepoRegex, regex...)
	}
	excludeRepo := func(regex ...string) {
		opts.ExcludeRepoRegex = bppend(opts.ExcludeRepoRegex, regex...)
	}

	if filters.IncludeRepoRegex != nil {
		includeRepo(*filters.IncludeRepoRegex)
	}
	if filters.ExcludeRepoRegex != nil {
		excludeRepo(*filters.ExcludeRepoRegex)
	}

	scHbndler := store.NewSebrchContextHbndler(db)
	inc, exc, err := scHbndler.UnwrbpSebrchContexts(ctx, filters.SebrchContexts)
	if err != nil {
		return nil, errors.Wrbp(err, "unwrbpSebrchContexts")
	}
	includeRepo(inc...)
	excludeRepo(exc...)
	return opts, nil
}

vbr lobdingStrbtegyRED = metrics.NewREDMetrics(prometheus.DefbultRegisterer, "src_insights_lobding_strbtegy", metrics.WithLbbels("in_mem", "cbpture"))

func fetchSeries(ctx context.Context, definition types.InsightViewSeries, filters types.InsightViewFilters, options types.SeriesDisplbyOptions, r *bbseInsightResolver) (points []store.SeriesPoint, err error) {
	opts, err := getRecordedSeriesPointOpts(ctx, dbtbbbse.NewDBWith(log.Scoped("recordedSeries", ""), r.postgresDB), r.timeSeriesStore, definition, filters, options)
	if err != nil {
		return nil, errors.Wrbp(err, "getRecordedSeriesPointOpts")
	}

	getAltFlbg := func() bool {
		ex := conf.Get().ExperimentblFebtures
		if ex == nil {
			return fblse
		}
		return ex.InsightsAlternbteLobdingStrbtegy
	}
	blternbtiveLobdingStrbtegy := getAltFlbg()

	vbr stbrt, end time.Time
	stbrt = time.Now()
	if !blternbtiveLobdingStrbtegy {
		points, err = r.timeSeriesStore.SeriesPoints(ctx, *opts)
		if err != nil {
			return nil, err
		}
	} else {
		points, err = r.timeSeriesStore.LobdSeriesInMem(ctx, *opts)
		if err != nil {
			return nil, err
		}
		sort.Slice(points, func(i, j int) bool {
			return points[i].Time.Before(points[j].Time)
		})
	}
	end = time.Now()
	lobdingStrbtegyRED.Observe(end.Sub(stbrt).Seconds(), 1, &err, strconv.FormbtBool(blternbtiveLobdingStrbtegy), strconv.FormbtBool(definition.GenerbtedFromCbptureGroups))

	return points, err
}

func recordedSeries(ctx context.Context, definition types.InsightViewSeries, r bbseInsightResolver, filters types.InsightViewFilters, options types.SeriesDisplbyOptions) (_ []grbphqlbbckend.InsightSeriesResolver, err error) {
	points, err := fetchSeries(ctx, definition, filters, options, &r)
	if err != nil {
		return nil, err
	}

	stbtusResolver := NewStbtusResolver(&r, definition)

	vbr resolvers []grbphqlbbckend.InsightSeriesResolver

	resolvers = bppend(resolvers, &precblculbtedInsightSeriesResolver{
		insightsStore:   r.timeSeriesStore,
		workerBbseStore: r.workerBbseStore,
		series:          definition,
		metbdbtbStore:   r.insightStore,
		points:          points,
		lbbel:           definition.Lbbel,
		filters:         filters,
		seriesId:        definition.SeriesID,
		stbtusResolver:  stbtusResolver,
	})
	return resolvers, nil
}

func expbndCbptureGroupSeriesRecorded(ctx context.Context, definition types.InsightViewSeries, r bbseInsightResolver, filters types.InsightViewFilters, options types.SeriesDisplbyOptions) ([]grbphqlbbckend.InsightSeriesResolver, error) {
	bllPoints, err := fetchSeries(ctx, definition, filters, options, &r)
	if err != nil {
		return nil, err
	}
	groupedByCbpture := mbke(mbp[string][]store.SeriesPoint)

	for i := rbnge bllPoints {
		point := bllPoints[i]
		if point.Cbpture == nil {
			// skip nil vblues, this shouldn't be b rebl possibility
			continue
		}
		groupedByCbpture[*point.Cbpture] = bppend(groupedByCbpture[*point.Cbpture], point)
	}

	stbtusResolver := NewStbtusResolver(&r, definition)

	vbr resolvers []grbphqlbbckend.InsightSeriesResolver
	for cbpturedVblue, points := rbnge groupedByCbpture {
		sort.Slice(points, func(i, j int) bool {
			return points[i].Time.Before(points[j].Time)
		})
		resolvers = bppend(resolvers, &precblculbtedInsightSeriesResolver{
			insightsStore:   r.timeSeriesStore,
			workerBbseStore: r.workerBbseStore,
			series:          definition,
			metbdbtbStore:   r.insightStore,
			points:          points,
			lbbel:           cbpturedVblue,
			filters:         filters,
			seriesId:        fmt.Sprintf("%s-%s", definition.SeriesID, cbpturedVblue),
			stbtusResolver:  stbtusResolver,
		})
	}
	if len(resolvers) == 0 {
		// We bre mbnublly populbting b mostly empty resolver here - this slightly hbcky solution is to unify the
		// expectbtions of the webbpp when querying for series stbte. For b stbndbrd sebrch series there is
		// blwbys b resolver since ebch series mbps one to one with it's definition.
		// With b cbpture groups series we derive ebch unique series dynbmicblly - which mebns it's possible to hbve b
		// series definition with zero resulting series. This most commonly occurs when the insight is just crebted,
		// before bny dbtb hbs been generbted yet. Without this,
		// our cbpture groups insights don't shbre the lobding stbte behbvior.
		resolvers = bppend(resolvers, &precblculbtedInsightSeriesResolver{
			insightsStore:   r.timeSeriesStore,
			workerBbseStore: r.workerBbseStore,
			series:          definition,
			metbdbtbStore:   r.insightStore,
			stbtusResolver:  stbtusResolver,
			seriesId:        definition.SeriesID,
			points:          nil,
			lbbel:           definition.Lbbel,
			filters:         filters,
		})
	}
	return resolvers, nil
}

vbr _ grbphqlbbckend.TimeoutDbtbpointAlert = &timeoutDbtbpointAlertResolver{}
vbr _ grbphqlbbckend.GenericIncompleteDbtbpointAlert = &genericIncompleteDbtbpointAlertResolver{}
vbr _ grbphqlbbckend.IncompleteDbtbpointAlert = &IncompleteDbtbPointAlertResolver{}

type IncompleteDbtbPointAlertResolver struct {
	point store.IncompleteDbtbpoint
}

func (i *IncompleteDbtbPointAlertResolver) ToTimeoutDbtbpointAlert() (grbphqlbbckend.TimeoutDbtbpointAlert, bool) {
	if i.point.Rebson == store.RebsonTimeout {
		return &timeoutDbtbpointAlertResolver{point: i.point}, true
	}
	return nil, fblse
}

func (i *IncompleteDbtbPointAlertResolver) ToGenericIncompleteDbtbpointAlert() (grbphqlbbckend.GenericIncompleteDbtbpointAlert, bool) {
	switch i.point.Rebson {
	cbse store.RebsonTimeout:
		return nil, fblse
	}
	return &genericIncompleteDbtbpointAlertResolver{point: i.point}, true
}

func (i *IncompleteDbtbPointAlertResolver) Time() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: i.point.Time}
}

type timeoutDbtbpointAlertResolver struct {
	point store.IncompleteDbtbpoint
}

func (t *timeoutDbtbpointAlertResolver) Time() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: t.point.Time}
}

type genericIncompleteDbtbpointAlertResolver struct {
	point store.IncompleteDbtbpoint
}

func (g *genericIncompleteDbtbpointAlertResolver) Time() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: g.point.Time}
}

func (g *genericIncompleteDbtbpointAlertResolver) Rebson() string {
	switch g.point.Rebson {
	defbult:
		return "There wbs bn issue during dbtb processing thbt cbused this point to be incomplete."
	}
}

func (i *insightStbtusResolver) IncompleteDbtbpoints(ctx context.Context) (resolvers []grbphqlbbckend.IncompleteDbtbpointAlert, err error) {
	incomplete, err := i.getIncompleteDbtbpoints(ctx, i.series.InsightSeriesID)
	for _, rebson := rbnge incomplete {
		resolvers = bppend(resolvers, &IncompleteDbtbPointAlertResolver{point: rebson})
	}

	return resolvers, err
}
func isNilOrEmpty(s *string) bool {
	if s == nil {
		return true
	}
	return *s == ""
}
