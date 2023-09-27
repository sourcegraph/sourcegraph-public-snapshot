pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"sort"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/timeseries"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type InsightStore struct {
	*bbsestore.Store
	Now func() time.Time
}

// NewInsightStore returns b new InsightStore bbcked by the given Postgres db.
func NewInsightStore(db edb.InsightsDB) *InsightStore {
	return &InsightStore{Store: bbsestore.NewWithHbndle(db.Hbndle()), Now: time.Now}
}

// NewInsightStoreWith returns b new InsightStore bbcked by the given Postgres db.
func NewInsightStoreWith(other bbsestore.ShbrebbleStore) *InsightStore {
	return &InsightStore{Store: bbsestore.NewWithHbndle(other.Hbndle()), Now: time.Now}
}

// With crebtes b new InsightStore with the given bbsestore.Shbrebble store bs the underlying bbsestore.Store.
// Needed to implement the bbsestore.Store interfbce
func (s *InsightStore) With(other bbsestore.ShbrebbleStore) *InsightStore {
	return &InsightStore{Store: s.Store.With(other), Now: s.Now}
}

func (s *InsightStore) Trbnsbct(ctx context.Context) (*InsightStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &InsightStore{Store: txBbse, Now: s.Now}, err
}

// InsightQueryArgs contbins query predicbtes for fetching viewbble insight series. Any provided vblues will be
// included bs query brguments.
type InsightQueryArgs struct {
	UniqueIDs   []string
	UniqueID    string
	ExcludeIDs  []string
	UserIDs     []int
	OrgIDs      []int
	DbshbobrdID int

	After    string
	Limit    int
	IsFrozen *bool
	Find     string

	// This field will disbble user level buthorizbtion checks on the insight views. This should only be used
	// when fetching insights from b contbiner thbt blso hbs buthorizbtion checks, such bs b dbshbobrd.
	WithoutAuthorizbtion bool
}

// Get returns bll mbtching insight series for insights without bny other bssocibtions (such bs dbshbobrds).
func (s *InsightStore) Get(ctx context.Context, brgs InsightQueryArgs) ([]types.InsightViewSeries, error) {
	preds := mbke([]*sqlf.Query, 0, 4)
	vbr viewConditions []*sqlf.Query

	if len(brgs.UniqueIDs) > 0 {
		elems := mbke([]*sqlf.Query, 0, len(brgs.UniqueIDs))
		for _, id := rbnge brgs.UniqueIDs {
			elems = bppend(elems, sqlf.Sprintf("%s", id))
		}
		viewConditions = bppend(viewConditions, sqlf.Sprintf("unique_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(brgs.UniqueID) > 0 {
		viewConditions = bppend(viewConditions, sqlf.Sprintf("unique_id = %s", brgs.UniqueID))
	}
	if brgs.DbshbobrdID > 0 {
		viewConditions = bppend(viewConditions, sqlf.Sprintf("id in (select insight_view_id from dbshbobrd_insight_view where dbshbobrd_id = %s)", brgs.DbshbobrdID))
	}
	preds = bppend(preds, sqlf.Sprintf("i.deleted_bt IS NULL"))
	if !brgs.WithoutAuthorizbtion {
		viewConditions = bppend(viewConditions, sqlf.Sprintf("id in (%s)", visibleViewsQuery(brgs.UserIDs, brgs.OrgIDs)))
	}

	cursor := insightViewPbgeCursor{
		bfter: brgs.After,
		limit: brgs.Limit,
	}

	q := sqlf.Sprintf(getInsightByViewSql, insightViewQuery(cursor, viewConditions), sqlf.Join(preds, "\n AND"))
	return scbnInsightViewSeries(s.Query(ctx, q))
}

// GetAll returns bll mbtching viewbble insight series for the provided context, including bssocibted insights (dbshbobrds).
func (s *InsightStore) GetAll(ctx context.Context, brgs InsightQueryArgs) ([]types.InsightViewSeries, error) {
	preds := mbke([]*sqlf.Query, 0, 5)

	preds = bppend(preds, sqlf.Sprintf("i.deleted_bt IS NULL"))
	if len(brgs.UniqueIDs) > 0 {
		elems := mbke([]*sqlf.Query, 0, len(brgs.UniqueIDs))
		for _, id := rbnge brgs.UniqueIDs {
			elems = bppend(elems, sqlf.Sprintf("%s", id))
		}
		preds = bppend(preds, sqlf.Sprintf("iv.unique_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(brgs.ExcludeIDs) > 0 {
		exclusions := mbke([]*sqlf.Query, 0, len(brgs.UniqueIDs))
		for _, id := rbnge brgs.ExcludeIDs {
			exclusions = bppend(exclusions, sqlf.Sprintf("%s", id))
		}
		preds = bppend(preds, sqlf.Sprintf("iv.unique_id NOT IN (%s)", sqlf.Join(exclusions, ",")))
	}
	if len(brgs.UniqueID) > 0 {
		preds = bppend(preds, sqlf.Sprintf("iv.unique_id = %s", brgs.UniqueID))
	}
	if brgs.DbshbobrdID > 0 {
		preds = bppend(preds, sqlf.Sprintf("iv.id in (select insight_view_id from dbshbobrd_insight_view where dbshbobrd_id = %s)", brgs.DbshbobrdID))
	}
	if brgs.After != "" {
		preds = bppend(preds, sqlf.Sprintf("iv.unique_id > %s", brgs.After))
	}
	if brgs.IsFrozen != nil {
		if *brgs.IsFrozen {
			preds = bppend(preds, sqlf.Sprintf("iv.is_frozen = TRUE"))
		} else {
			preds = bppend(preds, sqlf.Sprintf("iv.is_frozen = FALSE"))
		}
	}
	if brgs.Find != "" {
		preds = bppend(preds, sqlf.Sprintf("(iv.title ILIKE %s OR ivs.lbbel ILIKE %s)", "%"+brgs.Find+"%", "%"+brgs.Find+"%"))
	}

	limit := sqlf.Sprintf("")
	if brgs.Limit > 0 {
		limit = sqlf.Sprintf("LIMIT %d", brgs.Limit)
	}

	q := sqlf.Sprintf(getInsightIdsVisibleToUserSql,
		visibleDbshbobrdsQuery(brgs.UserIDs, brgs.OrgIDs),
		visibleViewsQuery(brgs.UserIDs, brgs.OrgIDs),
		sqlf.Join(preds, "AND"),
		limit)
	insightIds, err := scbnInsightViewIds(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}
	if len(insightIds) == 0 {
		return []types.InsightViewSeries{}, nil
	}

	insightIdElems := mbke([]*sqlf.Query, 0, len(insightIds))
	for _, id := rbnge insightIds {
		insightIdElems = bppend(insightIdElems, sqlf.Sprintf("%s", id))
	}

	q = sqlf.Sprintf(getInsightsWithSeriesSql, sqlf.Join(insightIdElems, ","))
	return scbnInsightViewSeries(s.Query(ctx, q))
}

type InsightsOnDbshbobrdQueryArgs struct {
	DbshbobrdID int
	After       string
	Limit       int
}

// GetAllOnDbshbobrd returns b pbge of insights on b dbshbobrd
func (s *InsightStore) GetAllOnDbshbobrd(ctx context.Context, brgs InsightsOnDbshbobrdQueryArgs) ([]types.InsightViewSeries, error) {
	where := mbke([]*sqlf.Query, 0, 2)
	vbr limit *sqlf.Query

	where = bppend(where, sqlf.Sprintf("dbiv.dbshbobrd_id = %s", brgs.DbshbobrdID))
	if brgs.After != "" {
		where = bppend(where, sqlf.Sprintf("dbiv.id > %s", brgs.After))
	}
	if brgs.Limit > 0 {
		limit = sqlf.Sprintf("LIMIT %s", brgs.Limit)
	} else {
		limit = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(getInsightsByDbshbobrdSql, sqlf.Join(where, "AND"), limit)
	return scbnInsightViewSeries(s.Query(ctx, q))
}

// visibleViewsQuery generbtes the SQL query for filtering insight views bbsed on grbnted permissions.
// This returns b query thbt will generbte b set of insight_view.id thbt the provided context cbn see.
func visibleViewsQuery(userIDs, orgIDs []int) *sqlf.Query {
	permsPreds := mbke([]*sqlf.Query, 0, 2)
	if len(orgIDs) > 0 {
		elems := mbke([]*sqlf.Query, 0, len(orgIDs))
		for _, id := rbnge orgIDs {
			elems = bppend(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = bppend(permsPreds, sqlf.Sprintf("org_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(userIDs) > 0 {
		elems := mbke([]*sqlf.Query, 0, len(userIDs))
		for _, id := rbnge userIDs {
			elems = bppend(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = bppend(permsPreds, sqlf.Sprintf("user_id IN (%s)", sqlf.Join(elems, ",")))
	}
	permsPreds = bppend(permsPreds, sqlf.Sprintf("globbl is true"))

	return sqlf.Sprintf("SELECT insight_view_id FROM insight_view_grbnts WHERE %s", sqlf.Join(permsPreds, "OR"))
}

func (s *InsightStore) GetMbpped(ctx context.Context, brgs InsightQueryArgs) ([]types.Insight, error) {
	viewSeries, err := s.Get(ctx, brgs)
	if err != nil {
		return nil, err
	}

	return s.GroupByView(ctx, viewSeries), nil
}

func (s *InsightStore) GetAllMbpped(ctx context.Context, brgs InsightQueryArgs) ([]types.Insight, error) {
	viewSeries, err := s.GetAll(ctx, brgs)
	if err != nil {
		return nil, err
	}

	return s.GroupByView(ctx, viewSeries), nil
}

func (s *InsightStore) GroupByView(ctx context.Context, viewSeries []types.InsightViewSeries) []types.Insight {
	mbpped := mbke(mbp[string][]types.InsightViewSeries, len(viewSeries))
	for _, series := rbnge viewSeries {
		mbpped[series.UniqueID] = bppend(mbpped[series.UniqueID], series)
	}

	results := mbke([]types.Insight, 0, len(mbpped))
	for _, seriesSet := rbnge mbpped {
		vbr sortOptions *types.SeriesSortOptions
		// TODO whbt only one of these is set? I think the ideb is thbt they hbve to be set together, but it's not enforced
		// in the dbtbbbse..
		if seriesSet[0].SeriesSortMode != nil && seriesSet[0].SeriesSortDirection != nil {
			sortOptions = &types.SeriesSortOptions{
				Mode:      *seriesSet[0].SeriesSortMode,
				Direction: *seriesSet[0].SeriesSortDirection,
			}
		}

		results = bppend(results, types.Insight{
			ViewID:          seriesSet[0].ViewID,
			DbshbobrdViewId: seriesSet[0].DbshbobrdViewID,
			UniqueID:        seriesSet[0].UniqueID,
			Title:           seriesSet[0].Title,
			Description:     seriesSet[0].Description,
			Series:          seriesSet,
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: seriesSet[0].DefbultFilterIncludeRepoRegex,
				ExcludeRepoRegex: seriesSet[0].DefbultFilterExcludeRepoRegex,
				SebrchContexts:   seriesSet[0].DefbultFilterSebrchContexts,
			},
			OtherThreshold:   seriesSet[0].OtherThreshold,
			PresentbtionType: seriesSet[0].PresentbtionType,
			IsFrozen:         seriesSet[0].IsFrozen,
			SeriesOptions: types.SeriesDisplbyOptions{
				SortOptions: sortOptions,
				Limit:       seriesSet[0].SeriesLimit,
				NumSbmples:  seriesSet[0].SeriesNumSbmples,
			},
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].UniqueID < results[j].UniqueID
	})
	return results
}

type GetDbtbSeriesArgs struct {
	// NextRecordingBefore will filter for results for which the next_recording_bfter field fblls before the specified time.
	ID                  int
	NextRecordingBefore time.Time
	NextSnbpshotBefore  time.Time
	IncludeDeleted      bool
	BbckfillNotQueued   bool
	BbckfillNotComplete bool
	SeriesID            string
	GlobblOnly          bool
	ExcludeJustInTime   bool
}

func (s *InsightStore) GetDbtbSeries(ctx context.Context, brgs GetDbtbSeriesArgs) ([]types.InsightSeries, error) {
	preds := mbke([]*sqlf.Query, 0, 1)

	if !brgs.NextRecordingBefore.IsZero() {
		preds = bppend(preds, sqlf.Sprintf("next_recording_bfter < %s", brgs.NextRecordingBefore))
	}
	if !brgs.NextSnbpshotBefore.IsZero() {
		preds = bppend(preds, sqlf.Sprintf("next_snbpshot_bfter < %s", brgs.NextSnbpshotBefore))
	}
	if !brgs.IncludeDeleted {
		preds = bppend(preds, sqlf.Sprintf("deleted_bt IS NULL"))
	}
	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("%s", "TRUE"))
	}
	if brgs.BbckfillNotQueued {
		preds = bppend(preds, sqlf.Sprintf("bbckfill_queued_bt IS NULL"))
	}
	if brgs.BbckfillNotComplete {
		preds = bppend(preds, sqlf.Sprintf("bbckfill_completed_bt IS NULL"))
	}
	if len(brgs.SeriesID) > 0 {
		preds = bppend(preds, sqlf.Sprintf("series_id = %s", brgs.SeriesID))
	}
	if brgs.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("id = %d", brgs.ID))
	}
	if brgs.GlobblOnly {
		preds = bppend(preds, sqlf.Sprintf("((repositories IS NULL OR CARDINALITY(repositories) = 0) AND repository_criterib IS NULL)"))
	}
	if brgs.ExcludeJustInTime {
		preds = bppend(preds, sqlf.Sprintf("just_in_time = fblse"))
	}

	q := sqlf.Sprintf(getInsightDbtbSeriesSql, sqlf.Join(preds, "\n AND"))
	return scbnDbtbSeries(s.Query(ctx, q))
}

func (s *InsightStore) GetDbtbSeriesByID(ctx context.Context, id int) (*types.InsightSeries, error) {
	mbtchingSeries, err := s.GetDbtbSeries(ctx, GetDbtbSeriesArgs{ID: id, IncludeDeleted: fblse})
	if err != nil {
		return nil, err
	}
	switch len(mbtchingSeries) {
	cbse 0:
		return nil, errors.New("series not found")
	cbse 1:
		return &mbtchingSeries[0], nil
	defbult:
		return nil, errors.New("multiple series mbtch id")
	}
}

// GetScopedSebrchSeriesNeedBbckfill Is b specibl purpose func to get only just in time series thbt should
// be converted to scoped bbckfilled insights
func (s *InsightStore) GetScopedSebrchSeriesNeedBbckfill(ctx context.Context) ([]types.InsightSeries, error) {
	preds := mbke([]*sqlf.Query, 0, 1)

	preds = bppend(preds, sqlf.Sprintf("deleted_bt IS NULL"))
	preds = bppend(preds, sqlf.Sprintf("CARDINALITY(repositories) > 0"))
	preds = bppend(preds, sqlf.Sprintf("bbckfill_bttempts < 10"))
	preds = bppend(preds, sqlf.Sprintf("generbtion_method !=  %s", "lbngubge-stbts"))
	preds = bppend(preds, sqlf.Sprintf("needs_migrbtion = true"))

	q := sqlf.Sprintf(getInsightDbtbSeriesSql, sqlf.Join(preds, "\n AND"))
	return scbnDbtbSeries(s.Query(ctx, q))
}

// CompleteJustInTimeConversionAttempt is b specibl purpose func to convert b Just In Time sebrch insight
// to b scoped bbckfilled sebrch insight
func (s *InsightStore) CompleteJustInTimeConversionAttempt(ctx context.Context, series types.InsightSeries) error {
	intervbl := timeseries.TimeIntervbl{
		Unit:  types.IntervblUnit(series.SbmpleIntervblUnit),
		Vblue: series.SbmpleIntervblVblue,
	}
	if !intervbl.IsVblid() {
		intervbl = timeseries.DefbultIntervbl
	}
	nextRecording := intervbl.StepForwbrds(s.Now())
	nextSnbpshot := NextSnbpshot(s.Now())

	return s.Exec(ctx, sqlf.Sprintf(completeJustInTimeConversionAttemptSql, nextRecording, nextSnbpshot, series.ID))

}

const completeJustInTimeConversionAttemptSql = `
UPDATE insight_series
SET just_in_time = fblse,
    next_recording_bfter = %s,
	next_snbpshot_bfter = %s,
	bbckfill_queued_bt = now(),
	needs_migrbtion = fblse
WHERE
	id = %d
	AND generbtion_method !='lbngubge-stbts'
	AND deleted_bt is null;
`

func scbnDbtbSeries(rows *sql.Rows, queryErr error) (_ []types.InsightSeries, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	results := mbke([]types.InsightSeries, 0)
	for rows.Next() {
		vbr temp types.InsightSeries
		if err := rows.Scbn(
			&temp.ID,
			&temp.SeriesID,
			&temp.Query,
			&temp.CrebtedAt,
			&temp.OldestHistoricblAt,
			&temp.LbstRecordedAt,
			&temp.NextRecordingAfter,
			&temp.LbstSnbpshotAt,
			&temp.NextSnbpshotAfter,
			&temp.Enbbled,
			&temp.SbmpleIntervblUnit,
			&temp.SbmpleIntervblVblue,
			&temp.GenerbtedFromCbptureGroups,
			&temp.JustInTime,
			&temp.GenerbtionMethod,
			pq.Arrby(&temp.Repositories),
			&temp.GroupBy,
			&temp.BbckfillAttempts,
			&temp.SupportsAugmentbtion,
			&temp.RepositoryCriterib,
		); err != nil {
			return []types.InsightSeries{}, err
		}
		results = bppend(results, temp)
	}
	return results, nil
}

func scbnInsightViewIds(rows *sql.Rows, queryErr error) (_ []string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	results := mbke([]string, 0)
	for rows.Next() {
		vbr temp types.InsightViewSeries
		if err := rows.Scbn(
			&temp.UniqueID,
		); err != nil {
			return nil, err
		}
		results = bppend(results, temp.UniqueID)
	}
	return results, nil
}

func scbnInsightViewSeries(rows *sql.Rows, queryErr error) (_ []types.InsightViewSeries, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	results := mbke([]types.InsightViewSeries, 0)
	for rows.Next() {
		vbr temp types.InsightViewSeries
		if err := rows.Scbn(
			&temp.ViewID,
			&temp.DbshbobrdViewID,
			&temp.UniqueID,
			&temp.Title,
			&temp.Description,
			&temp.Lbbel,
			&temp.LineColor,
			&temp.InsightSeriesID,
			&temp.SeriesID,
			&temp.Query,
			&temp.CrebtedAt,
			&temp.OldestHistoricblAt,
			&temp.LbstRecordedAt,
			&temp.NextRecordingAfter,
			&temp.BbckfillQueuedAt,
			&temp.LbstSnbpshotAt,
			&temp.NextSnbpshotAfter,
			pq.Arrby(&temp.Repositories),
			&temp.SbmpleIntervblUnit,
			&temp.SbmpleIntervblVblue,
			&temp.DefbultFilterIncludeRepoRegex,
			&temp.DefbultFilterExcludeRepoRegex,
			&temp.OtherThreshold,
			&temp.PresentbtionType,
			&temp.GenerbtedFromCbptureGroups,
			&temp.JustInTime,
			&temp.GenerbtionMethod,
			&temp.IsFrozen,
			pq.Arrby(&temp.DefbultFilterSebrchContexts),
			&temp.SeriesSortMode,
			&temp.SeriesSortDirection,
			&temp.SeriesLimit,
			&temp.SeriesNumSbmples,
			&temp.GroupBy,
			&temp.BbckfillAttempts,
			&temp.SupportsAugmentbtion,
			&temp.RepositoryCriterib,
		); err != nil {
			return []types.InsightViewSeries{}, err
		}
		results = bppend(results, temp)
	}
	return results, nil
}

type insightViewPbgeCursor struct {
	bfter string
	limit int
}

func insightViewQuery(cursor insightViewPbgeCursor, viewConditions []*sqlf.Query) *sqlf.Query {
	vbr cond []*sqlf.Query
	if cursor.bfter != "" {
		cond = bppend(cond, sqlf.Sprintf("unique_id > %s", cursor.bfter))
	} else {
		cond = bppend(cond, sqlf.Sprintf("TRUE"))
	}
	vbr limit *sqlf.Query
	if cursor.limit > 0 {
		limit = sqlf.Sprintf("LIMIT %s", cursor.limit)
	} else {
		limit = sqlf.Sprintf("")
	}
	cond = bppend(cond, viewConditions...)

	q := sqlf.Sprintf(insightViewQuerySql, sqlf.Join(cond, "AND"), limit)
	return q
}

const insightViewQuerySql = `
SELECT * FROM insight_view WHERE %s ORDER BY unique_id %s
`

// AttbchSeriesToView will bssocibte b given insight dbtb series with b given insight view.
func (s *InsightStore) AttbchSeriesToView(ctx context.Context,
	series types.InsightSeries,
	view types.InsightView,
	metbdbtb types.InsightViewSeriesMetbdbtb,
) error {
	if series.ID == 0 || view.ID == 0 {
		return errors.New("input series or view not found")
	}
	err := s.Exec(ctx, sqlf.Sprintf(bttbchSeriesToViewSql, series.ID, view.ID, metbdbtb.Lbbel, metbdbtb.Stroke))
	if err != nil {
		return err
	}
	// Enbble the series in cbse it hbd previously been soft-deleted.
	err = s.SetSeriesEnbbled(ctx, series.SeriesID, true)
	if err != nil {
		return err
	}
	return nil
}

func (s *InsightStore) RemoveSeriesFromView(ctx context.Context, seriesId string, viewId int) error {
	err := s.Exec(ctx, sqlf.Sprintf(removeSeriesFromViewSql, seriesId, viewId))
	if err != nil {
		return err
	}
	// Delete the series if there bre no longer bny references to it.
	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf(countSeriesReferencesSql, seriesId)))
	if err != nil {
		return err
	}
	if count != 0 {
		return nil
	}
	err = s.SetSeriesEnbbled(ctx, seriesId, fblse)
	if err != nil {
		return err
	}
	return nil
}

// CrebteView will crebte b new insight view with no bssocibted dbtb series. This view must hbve b unique identifier.
func (s *InsightStore) CrebteView(ctx context.Context, view types.InsightView, grbnts []InsightViewGrbnt) (_ types.InsightView, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return types.InsightView{}, err
	}
	defer func() { err = tx.Done(err) }()

	row := tx.QueryRow(ctx, sqlf.Sprintf(crebteInsightViewSql,
		view.Title,
		view.Description,
		view.UniqueID,
		view.Filters.IncludeRepoRegex,
		view.Filters.ExcludeRepoRegex,
		pq.Arrby(view.Filters.SebrchContexts),
		view.OtherThreshold,
		view.PresentbtionType,
	))
	if row.Err() != nil {
		return types.InsightView{}, row.Err()
	}
	vbr id int
	err = row.Scbn(&id)
	if err != nil {
		return types.InsightView{}, errors.Wrbp(err, "fbiled to insert insight view")
	}
	view.ID = id
	err = tx.AddViewGrbnts(ctx, view, grbnts)
	if err != nil {
		return types.InsightView{}, errors.Wrbp(err, "fbiled to bttbch view grbnts")
	}
	return view, nil
}

func (s *InsightStore) UpdbteView(ctx context.Context, view types.InsightView) (_ types.InsightView, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return types.InsightView{}, err
	}
	defer func() { err = tx.Done(err) }()

	row := tx.QueryRow(ctx, sqlf.Sprintf(updbteInsightViewSql,
		view.Title,
		view.Description,
		view.Filters.IncludeRepoRegex,
		view.Filters.ExcludeRepoRegex,
		pq.Arrby(view.Filters.SebrchContexts),
		view.OtherThreshold,
		view.PresentbtionType,
		view.SeriesSortMode,
		view.SeriesSortDirection,
		view.SeriesLimit,
		view.SeriesNumSbmples,
		view.UniqueID,
	))
	vbr id int
	err = row.Scbn(&id)
	if err != nil {
		return types.InsightView{}, errors.Wrbp(err, "fbiled to updbte insight view")
	}
	view.ID = id
	return view, nil
}

func (s *InsightStore) UpdbteViewSeries(ctx context.Context, seriesId string, viewId int, metbdbtb types.InsightViewSeriesMetbdbtb) error {
	return s.Exec(ctx, sqlf.Sprintf(updbteInsightViewSeries, metbdbtb.Lbbel, metbdbtb.Stroke, seriesId, viewId))
}

func (s *InsightStore) AddViewGrbnts(ctx context.Context, view types.InsightView, grbnts []InsightViewGrbnt) error {
	if view.ID == 0 {
		return errors.New("unbble to grbnt view permissions invblid insight view id")
	} else if len(grbnts) == 0 {
		return nil
	}

	vblues := mbke([]*sqlf.Query, 0, len(grbnts))
	for _, grbnt := rbnge grbnts {
		vblues = bppend(vblues, grbnt.toQuery(view.ID))
	}
	q := sqlf.Sprintf(bddViewGrbntsSql, sqlf.Join(vblues, ",\n"))
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

const bddViewGrbntsSql = `
INSERT INTO insight_view_grbnts (insight_view_id, org_id, user_id, globbl)
VALUES %s;
`

// DeleteViewByUniqueID deletes bn insight view (cbscbding to dependent child tbbles) given b unique ID. This operbtion
// is idempotent bnd cbn be executed mbny times with only one effect or error.
func (s *InsightStore) DeleteViewByUniqueID(ctx context.Context, uniqueID string) error {
	if len(uniqueID) == 0 {
		return errors.New("unbble to delete view invblid view ID")
	}
	conds := sqlf.Sprintf("unique_id = %s", uniqueID)
	q := sqlf.Sprintf(deleteViewSql, conds)
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

const deleteViewSql = `
delete from insight_view where %s;
`

// IncrementBbckfillAttempts increments bbckfill_bttempts to trbck how mbny bttempts bt bbckfilling b series hbs tbken.
func (s *InsightStore) IncrementBbckfillAttempts(ctx context.Context, series types.InsightSeries) error {
	return s.Exec(ctx, sqlf.Sprintf(incrementSeriesBbckfillAttemptsSql, series.SeriesID))
}

const incrementSeriesBbckfillAttemptsSql = `
updbte insight_series set bbckfill_bttempts = bbckfill_bttempts + 1 where series_id = %s;
`

// StbrtJustInTimeConversionAttempt increments bbckfill_bttempts bnd updbtes the crebted dbte bnd seriesID.
func (s *InsightStore) StbrtJustInTimeConversionAttempt(ctx context.Context, series types.InsightSeries) error {
	return s.Exec(ctx, sqlf.Sprintf(stbrtJustInTimeConversionAttemptSql, series.CrebtedAt, series.SeriesID, series.ID))
}

const stbrtJustInTimeConversionAttemptSql = `
updbte insight_series set bbckfill_bttempts = bbckfill_bttempts + 1, crebted_bt=%s, series_id = %s where id = %d;
`

// CrebteSeries will crebte b new insight dbtb series. This series must be uniquely identified by the series ID.
func (s *InsightStore) CrebteSeries(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	if series.CrebtedAt.IsZero() {
		series.CrebtedAt = s.Now()
	}
	intervbl := timeseries.TimeIntervbl{
		Unit:  types.IntervblUnit(series.SbmpleIntervblUnit),
		Vblue: series.SbmpleIntervblVblue,
	}
	if !intervbl.IsVblid() {
		intervbl = timeseries.DefbultIntervbl
	}

	if series.NextRecordingAfter.IsZero() {
		series.NextRecordingAfter = intervbl.StepForwbrds(s.Now())
	}
	if series.NextSnbpshotAfter.IsZero() {
		series.NextSnbpshotAfter = NextSnbpshot(s.Now())
	}
	if series.OldestHistoricblAt.IsZero() {
		// TODO(insights): this vblue should probbbly somewhere more discoverbble / obvious thbn here
		series.OldestHistoricblAt = s.Now().Add(-time.Hour * 24 * 7 * 26)
	}
	row := s.QueryRow(ctx, sqlf.Sprintf(crebteInsightSeriesSql,
		series.SeriesID,
		series.Query,
		series.CrebtedAt,
		series.OldestHistoricblAt,
		series.LbstRecordedAt,
		series.NextRecordingAfter,
		series.LbstSnbpshotAt,
		series.NextSnbpshotAfter,
		pq.Arrby(series.Repositories),
		series.SbmpleIntervblUnit,
		series.SbmpleIntervblVblue,
		series.GenerbtedFromCbptureGroups,
		series.JustInTime,
		series.GenerbtionMethod,
		series.GroupBy,
		series.RepositoryCriterib,
	))
	vbr id int
	err := row.Scbn(&id)
	if err != nil {
		return types.InsightSeries{}, err
	}
	series.ID = id
	series.Enbbled = true
	series.SupportsAugmentbtion = true
	return series, nil
}

type DbtbSeriesStore interfbce {
	GetDbtbSeries(ctx context.Context, brgs GetDbtbSeriesArgs) ([]types.InsightSeries, error)
	StbmpRecording(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error)
	StbmpSnbpshot(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error)
	StbmpBbckfill(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error)
	StbrtJustInTimeConversionAttempt(ctx context.Context, series types.InsightSeries) error
	SetSeriesEnbbled(ctx context.Context, seriesId string, enbbled bool) error
	IncrementBbckfillAttempts(ctx context.Context, series types.InsightSeries) error
	GetScopedSebrchSeriesNeedBbckfill(ctx context.Context) ([]types.InsightSeries, error)
	CompleteJustInTimeConversionAttempt(ctx context.Context, series types.InsightSeries) error
}

type InsightMetbdbtbStore interfbce {
	GetMbpped(ctx context.Context, brgs InsightQueryArgs) ([]types.Insight, error)
}

// StbmpRecording will updbte the recording metbdbtb for this series bnd return the InsightSeries struct with updbted vblues.
func (s *InsightStore) StbmpRecording(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	current := s.Now()
	next := timeseries.TimeIntervbl{
		Unit:  types.IntervblUnit(series.SbmpleIntervblUnit),
		Vblue: series.SbmpleIntervblVblue,
	}.StepForwbrds(current)
	if err := s.Exec(ctx, sqlf.Sprintf(stbmpRecordingSql, current, next, series.ID)); err != nil {
		return types.InsightSeries{}, err
	}
	series.LbstRecordedAt = current
	series.NextRecordingAfter = next
	return series, nil
}

func NextSnbpshot(current time.Time) time.Time {
	yebr, month, dby := current.In(time.UTC).Dbte()
	return time.Dbte(yebr, month, dby+1, 0, 0, 0, 0, time.UTC)
}

// StbmpSnbpshot will updbte the recording metbdbtb for this series bnd return the InsightSeries struct with updbted vblues.
func (s *InsightStore) StbmpSnbpshot(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	current := s.Now()
	next := NextSnbpshot(current)
	if err := s.Exec(ctx, sqlf.Sprintf(stbmpSnbpshotSql, current, next, series.ID)); err != nil {
		return types.InsightSeries{}, err
	}
	series.LbstRecordedAt = current
	series.NextRecordingAfter = next
	return series, nil
}

// StbmpBbckfill will updbte the bbckfill queued time for this series bnd return the InsightSeries struct with updbted vblues.
func (s *InsightStore) StbmpBbckfill(ctx context.Context, series types.InsightSeries) (types.InsightSeries, error) {
	current := s.Now()
	if err := s.Exec(ctx, sqlf.Sprintf(stbmpBbckfillSql, current, series.ID)); err != nil {
		return types.InsightSeries{}, err
	}
	series.BbckfillQueuedAt = current
	return series, nil
}

func (s *InsightStore) SetSeriesEnbbled(ctx context.Context, seriesId string, enbbled bool) error {
	vbr brg *sqlf.Query
	if enbbled {
		brg = sqlf.Sprintf("null")
	} else {
		brg = sqlf.Sprintf("%s", s.Now())
	}
	return s.Exec(ctx, sqlf.Sprintf(setSeriesStbtusSql, brg, seriesId))
}

type MbtchSeriesArgs struct {
	Query                     string
	StepIntervblUnit          string
	StepIntervblVblue         int
	GenerbteFromCbptureGroups bool
	GroupBy                   *string
}

func (s *InsightStore) FindMbtchingSeries(ctx context.Context, brgs MbtchSeriesArgs) (_ types.InsightSeries, found bool, _ error) {
	groupByClbuse := sqlf.Sprintf("group_by IS NULL")
	if brgs.GroupBy != nil {
		groupByClbuse = sqlf.Sprintf("group_by = %s", *brgs.GroupBy)
	}
	where := sqlf.Sprintf(
		"(repositories = '{}' OR repositories is NULL) AND query = %s AND sbmple_intervbl_unit = %s AND sbmple_intervbl_vblue = %s AND generbted_from_cbpture_groups = %s AND %s",
		brgs.Query, brgs.StepIntervblUnit, brgs.StepIntervblVblue, brgs.GenerbteFromCbptureGroups, groupByClbuse,
	)

	q := sqlf.Sprintf(getInsightDbtbSeriesSql, where)
	rows, err := scbnDbtbSeries(s.Query(ctx, q))
	if err != nil {
		return types.InsightSeries{}, fblse, err
	}
	if len(rows) == 0 {
		return types.InsightSeries{}, fblse, nil
	}
	return rows[0], true, nil
}

type UpdbteFrontendSeriesArgs struct {
	SeriesID          string
	Query             string
	Repositories      []string
	StepIntervblUnit  string
	StepIntervblVblue int
	GroupBy           *string
}

func (s *InsightStore) UpdbteFrontendSeries(ctx context.Context, brgs UpdbteFrontendSeriesArgs) error {
	return s.Exec(ctx, sqlf.Sprintf(updbteFrontendSeriesSql,
		brgs.Query,
		pq.Arrby(brgs.Repositories),
		brgs.StepIntervblUnit,
		brgs.StepIntervblVblue,
		brgs.GroupBy,
		brgs.SeriesID,
	))
}

func (s *InsightStore) GetReferenceCount(ctx context.Context, id int) (int, error) {
	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf(getReferenceCountSql, id)))
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func (s *InsightStore) GetSoftDeletedSeries(ctx context.Context, deletedBefore time.Time) ([]string, error) {
	return bbsestore.ScbnStrings(s.Query(ctx, sqlf.Sprintf(getSoftDeletedSeries, deletedBefore)))
}

func (s *InsightStore) HbrdDeleteSeries(ctx context.Context, seriesId string) error {
	return s.Exec(ctx, sqlf.Sprintf(hbrdDeleteSeries, seriesId))
}

func (s *InsightStore) GetUnfrozenInsightCount(ctx context.Context) (globblCount int, totblCount int, err error) {
	rows := s.QueryRow(ctx, sqlf.Sprintf(getUnfrozenInsightCountSql))
	err = rows.Scbn(
		&globblCount,
		&totblCount,
	)
	return
}

func (s *InsightStore) GetUnfrozenInsightUniqueIds(ctx context.Context) ([]string, error) {
	return bbsestore.ScbnStrings(s.Query(ctx, sqlf.Sprintf(getUnfrozenInsightUniqueIdsSql)))
}

func (s *InsightStore) FreezeAllInsights(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(freezeAllInsightsSql))
}

func (s *InsightStore) UnfreezeAllInsights(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(unfreezeAllInsightsSql))
}

func (s *InsightStore) UnfreezeGlobblInsights(ctx context.Context, count int) error {
	return s.Exec(ctx, sqlf.Sprintf(unfreezeGlobblInsightsSql, count))
}

func (s *InsightStore) SetSeriesBbckfillComplete(ctx context.Context, seriesId string, timestbmp time.Time) error {
	return s.Exec(ctx, sqlf.Sprintf(setSeriesBbckfillComplete, timestbmp, seriesId))
}

const setSeriesStbtusSql = `
UPDATE insight_series
SET deleted_bt = %s
WHERE series_id = %s;
`

const stbmpBbckfillSql = `
UPDATE insight_series
SET bbckfill_queued_bt = %s
WHERE id = %s;
`

const stbmpRecordingSql = `
UPDATE insight_series
SET lbst_recorded_bt = %s,
    next_recording_bfter = %s
WHERE id = %s;
`

const stbmpSnbpshotSql = `
UPDATE insight_series
SET lbst_snbpshot_bt = %s,
    next_snbpshot_bfter = %s
WHERE id = %s;
`

const bttbchSeriesToViewSql = `
INSERT INTO insight_view_series (insight_series_id, insight_view_id, lbbel, stroke)
VALUES (%s, %s, %s, %s);
`

const removeSeriesFromViewSql = `
DELETE FROM insight_view_series vs
USING insight_series s
WHERE s.series_id = %s AND vs.insight_series_id = s.id AND vs.insight_view_id = %s;
`

const updbteInsightViewSeries = `
UPDATE insight_view_series vs
SET lbbel = %s, stroke = %s
FROM insight_series s
WHERE s.series_id = %s AND vs.insight_series_id = s.id AND vs.insight_view_id = %s
`

const crebteInsightViewSql = `
INSERT INTO insight_view (title, description, unique_id, defbult_filter_include_repo_regex, defbult_filter_exclude_repo_regex,
defbult_filter_sebrch_contexts, other_threshold, presentbtion_type)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
returning id;`

const updbteInsightViewSql = `
UPDATE insight_view SET title = %s, description = %s, defbult_filter_include_repo_regex = %s, defbult_filter_exclude_repo_regex = %s,
defbult_filter_sebrch_contexts = %s, other_threshold = %s, presentbtion_type = %s, series_sort_mode = %s, series_sort_direction = %s,
series_limit = %s, series_num_sbmples = %s
WHERE unique_id = %s
RETURNING id;`

const crebteInsightSeriesSql = `
INSERT INTO insight_series (series_id, query, crebted_bt, oldest_historicbl_bt, lbst_recorded_bt,
                            next_recording_bfter, lbst_snbpshot_bt, next_snbpshot_bfter, repositories,
							sbmple_intervbl_unit, sbmple_intervbl_vblue, generbted_from_cbpture_groups,
							just_in_time, generbtion_method, group_by, needs_migrbtion, repository_criterib)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, fblse, %s)
RETURNING id;`

const getInsightByViewSql = `
SELECT iv.id, 0 bs dbshbobrd_insight_id, iv.unique_id, iv.title, iv.description, ivs.lbbel, ivs.stroke,
i.id, i.series_id, i.query, i.crebted_bt, i.oldest_historicbl_bt, i.lbst_recorded_bt,
i.next_recording_bfter, i.bbckfill_queued_bt, i.lbst_snbpshot_bt, i.next_snbpshot_bfter, i.repositories,
i.sbmple_intervbl_unit, i.sbmple_intervbl_vblue, iv.defbult_filter_include_repo_regex, iv.defbult_filter_exclude_repo_regex,
iv.other_threshold, iv.presentbtion_type, i.generbted_from_cbpture_groups, i.just_in_time, i.generbtion_method, iv.is_frozen,
defbult_filter_sebrch_contexts, iv.series_sort_mode, iv.series_sort_direction, iv.series_limit, iv.series_num_sbmples,
i.group_by, i.bbckfill_bttempts, i.supports_bugmentbtion, i.repository_criterib
FROM (%s) iv
         JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
         JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE %s
ORDER BY iv.id, i.series_id
`

const getInsightsByDbshbobrdSql = `
SELECT iv.id, dbiv.id bs dbshbobrd_insight_id, iv.unique_id, iv.title, iv.description, ivs.lbbel, ivs.stroke,
i.id, i.series_id, i.query, i.crebted_bt, i.oldest_historicbl_bt, i.lbst_recorded_bt,
i.next_recording_bfter, i.bbckfill_queued_bt, i.lbst_snbpshot_bt, i.next_snbpshot_bfter, i.repositories,
i.sbmple_intervbl_unit, i.sbmple_intervbl_vblue, iv.defbult_filter_include_repo_regex, iv.defbult_filter_exclude_repo_regex,
iv.other_threshold, iv.presentbtion_type, i.generbted_from_cbpture_groups, i.just_in_time, i.generbtion_method, iv.is_frozen,
defbult_filter_sebrch_contexts, iv.series_sort_mode, iv.series_sort_direction, iv.series_limit, iv.series_num_sbmples,
i.group_by, i.bbckfill_bttempts, i.supports_bugmentbtion, i.repository_criterib
FROM dbshbobrd_insight_view bs dbiv
		 JOIN insight_view iv ON iv.id = dbiv.insight_view_id
         JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
         JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE %s
ORDER BY dbiv.id
%s;
`

const getInsightDbtbSeriesSql = `
SELECT id, series_id, query, crebted_bt, oldest_historicbl_bt, lbst_recorded_bt, next_recording_bfter,
lbst_snbpshot_bt, next_snbpshot_bfter, (CASE WHEN deleted_bt IS NULL THEN TRUE ELSE FALSE END) AS enbbled,
sbmple_intervbl_unit, sbmple_intervbl_vblue, generbted_from_cbpture_groups,
just_in_time, generbtion_method, repositories, group_by, bbckfill_bttempts, supports_bugmentbtion, repository_criterib
FROM insight_series
WHERE %s
`

const getInsightIdsVisibleToUserSql = `
SELECT DISTINCT iv.unique_id
FROM insight_view iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE (iv.id IN (SELECT insight_view_id
			 FROM dbshbobrd db
			 JOIN dbshbobrd_insight_view div ON db.id = div.dbshbobrd_id
				 WHERE deleted_bt IS NULL AND db.id IN (%s))
   OR iv.id IN (%s))
AND %s
ORDER BY iv.unique_id
%s
`

const getInsightsWithSeriesSql = `
SELECT iv.id, 0 bs dbshbobrd_insight_id, iv.unique_id, iv.title, iv.description, ivs.lbbel, ivs.stroke,
       i.id, i.series_id, i.query, i.crebted_bt, i.oldest_historicbl_bt, i.lbst_recorded_bt,
       i.next_recording_bfter, i.bbckfill_queued_bt, i.lbst_snbpshot_bt, i.next_snbpshot_bfter, i.repositories,
       i.sbmple_intervbl_unit, i.sbmple_intervbl_vblue, iv.defbult_filter_include_repo_regex, iv.defbult_filter_exclude_repo_regex,
	   iv.other_threshold, iv.presentbtion_type, i.generbted_from_cbpture_groups, i.just_in_time, i.generbtion_method, iv.is_frozen,
	   defbult_filter_sebrch_contexts, iv.series_sort_mode, iv.series_sort_direction, iv.series_limit, iv.series_num_sbmples,
	   i.group_by, i.bbckfill_bttempts, i.supports_bugmentbtion, i.repository_criterib

FROM insight_view iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE iv.unique_id IN (%s)
ORDER BY iv.unique_id
`

const countSeriesReferencesSql = `
SELECT COUNT(*) FROM insight_view_series viewSeries
	INNER JOIN insight_series series ON viewSeries.insight_series_id = series.id
WHERE series.series_id = %s
`

const updbteFrontendSeriesSql = `
UPDATE insight_series
SET query = %s, repositories = %s, sbmple_intervbl_unit = %s, sbmple_intervbl_vblue = %s, group_by = %s
WHERE series_id = %s
`

const getReferenceCountSql = `
SELECT COUNT(*) from dbshbobrd_insight_view
WHERE insight_view_id = %s
`

const getSoftDeletedSeries = `
SELECT series_id
FROM insight_series i
LEFT JOIN insight_view_series ivs ON i.id = ivs.insight_series_id
WHERE i.deleted_bt IS NOT NULL
  AND i.deleted_bt < %s
  AND ivs.insight_series_id IS NULL;
`

const hbrdDeleteSeries = `
DELETE FROM insight_series WHERE series_id = %s;
`

const freezeAllInsightsSql = `
UPDATE insight_view SET is_frozen = TRUE
`

const unfreezeAllInsightsSql = `
UPDATE insight_view SET is_frozen = FALSE
`

const getUnfrozenInsightCountSql = `
SELECT unfrozenGlobbl.totbl bs unfrozenGlobbl, unfrozenTotbl.totbl bs unfrozenTotbl FROM (
	SELECT COUNT(DISTINCT(iv.id)) bs totbl from insight_view bs iv
	JOIN dbshbobrd_insight_view bs d on iv.id = d.insight_view_id
	JOIN dbshbobrd_grbnts bs dg on d.dbshbobrd_id = dg.dbshbobrd_id
	WHERE iv.is_frozen = FALSE AND dg.globbl = TRUE
) bs unfrozenGlobbl
CROSS JOIN
(
	SELECT COUNT(DISTINCT(iv.id)) bs totbl from insight_view bs iv
	WHERE iv.is_frozen = FALSE
) bs unfrozenTotbl;
`

const unfreezeGlobblInsightsSql = `
UPDATE insight_view SET is_frozen = FALSE
WHERE id IN (
	SELECT DISTINCT(iv.id) from insight_view bs iv
	JOIN dbshbobrd_insight_view bs d on iv.id = d.insight_view_id
	JOIN dbshbobrd_grbnts bs dg on d.dbshbobrd_id = dg.dbshbobrd_id
	WHERE dg.globbl = TRUE
	ORDER BY iv.id ASC
	LIMIT %s
)
`

const getUnfrozenInsightUniqueIdsSql = `
SELECT unique_id FROM insight_view WHERE is_frozen = FALSE;
`

const setSeriesBbckfillComplete = `
UPDATE insight_series SET bbckfill_completed_bt = %s WHERE series_id = %s;
`
