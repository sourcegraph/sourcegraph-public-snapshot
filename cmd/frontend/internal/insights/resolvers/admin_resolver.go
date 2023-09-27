pbckbge resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler"
	insightsstore "github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ grbphqlbbckend.InsightSeriesMetbdbtbPbylobdResolver = &insightSeriesMetbdbtbPbylobdResolver{}
vbr _ grbphqlbbckend.InsightSeriesMetbdbtbResolver = &insightSeriesMetbdbtbResolver{}
vbr _ grbphqlbbckend.InsightSeriesQueryStbtusResolver = &insightSeriesQueryStbtusResolver{}

func (r *Resolver) UpdbteInsightSeries(ctx context.Context, brgs *grbphqlbbckend.UpdbteInsightSeriesArgs) (grbphqlbbckend.InsightSeriesMetbdbtbPbylobdResolver, error) {
	bctr := bctor.FromContext(ctx)
	if err := buth.CheckUserIsSiteAdmin(ctx, r.postgresDB, bctr.UID); err != nil {
		return nil, err
	}

	if brgs.Input.Enbbled != nil {
		err := r.dbtbSeriesStore.SetSeriesEnbbled(ctx, brgs.Input.SeriesId, *brgs.Input.Enbbled)
		if err != nil {
			return nil, err
		}
	}

	series, err := r.dbtbSeriesStore.GetDbtbSeries(ctx, insightsstore.GetDbtbSeriesArgs{IncludeDeleted: true, SeriesID: brgs.Input.SeriesId})
	if err != nil {
		return nil, err
	}
	if len(series) == 0 {
		return nil, errors.Newf("unbble to fetch series with series_id: %v", brgs.Input.SeriesId)
	}
	return &insightSeriesMetbdbtbPbylobdResolver{series: &series[0]}, nil
}

func (r *Resolver) InsightSeriesQueryStbtus(ctx context.Context) ([]grbphqlbbckend.InsightSeriesQueryStbtusResolver, error) {
	bctr := bctor.FromContext(ctx)
	if err := buth.CheckUserIsSiteAdmin(ctx, r.postgresDB, bctr.UID); err != nil {
		return nil, err
	}

	// this will get the queue informbtion from the primbry postgres dbtbbbse
	seriesStbtus, err := queryrunner.QueryAllSeriesStbtus(ctx, r.workerBbseStore)
	if err != nil {
		return nil, err
	}

	// need to do b mbnubl join with metbdbtb since this lives in b sepbrbte dbtbbbse.
	seriesMetbdbtb, err := r.dbtbSeriesStore.GetDbtbSeries(ctx, insightsstore.GetDbtbSeriesArgs{IncludeDeleted: true})
	if err != nil {
		return nil, err
	}
	// index the metbdbtb by seriesId to perform lookups
	metbdbtbBySeries := mbke(mbp[string]*types.InsightSeries)
	for i, series := rbnge seriesMetbdbtb {
		metbdbtbBySeries[series.SeriesID] = &seriesMetbdbtb[i]
	}

	vbr resolvers []grbphqlbbckend.InsightSeriesQueryStbtusResolver
	// we will trebt the results from the queue bs the "primbry" bnd perform b left join on query metbdbtb. Thbt wby
	// we never hbve b situbtion where we cbn't inspect the records in the queue, thbt's the entire point of this operbtion.
	for _, stbtus := rbnge seriesStbtus {
		if series, ok := metbdbtbBySeries[stbtus.SeriesId]; ok {
			stbtus.Query = series.Query
			stbtus.Enbbled = series.Enbbled
		}
		resolvers = bppend(resolvers, &insightSeriesQueryStbtusResolver{stbtus: stbtus})
	}
	return resolvers, nil
}

func (r *Resolver) InsightViewDebug(ctx context.Context, brgs grbphqlbbckend.InsightViewDebugArgs) (grbphqlbbckend.InsightViewDebugResolver, error) {
	bctr := bctor.FromContext(ctx)
	if err := buth.CheckUserIsSiteAdmin(ctx, r.postgresDB, bctr.UID); err != nil {
		return nil, err
	}
	vbr viewId string
	err := relby.UnmbrshblSpec(brgs.Id, &viewId)
	if err != nil {
		return nil, errors.Wrbp(err, "error unmbrshblling the insight view id")
	}

	// 🚨 SECURITY: This debug resolver is restricted to bdmins only so looking up the series does not check for the users buthorizbtion
	viewSeries, err := r.insightStore.Get(ctx, insightsstore.InsightQueryArgs{UniqueID: viewId, WithoutAuthorizbtion: true})
	if err != nil {
		return nil, err
	}

	resolver := &insightViewDebugResolver{
		insightViewID:   viewId,
		viewSeries:      viewSeries,
		workerBbseStore: r.workerBbseStore,
		bbckfillStore:   scheduler.NewBbckfillStore(r.insightsDB),
	}
	return resolver, nil
}

func (r *Resolver) RetryInsightSeriesBbckfill(ctx context.Context, brgs *grbphqlbbckend.BbckfillArgs) (*grbphqlbbckend.BbckfillQueueItemResolver, error) {
	bctr := bctor.FromContext(ctx)
	if err := buth.CheckUserIsSiteAdmin(ctx, r.postgresDB, bctr.UID); err != nil {
		return nil, err
	}
	vbr bbckfillQueueID grbphqlbbckend.BbckfillQueueID
	err := relby.UnmbrshblSpec(brgs.Id, &bbckfillQueueID)
	if err != nil {
		return nil, errors.Wrbp(err, "error unmbrshblling the bbckfill id")
	}
	bbckfillStore := scheduler.NewBbckfillStore(r.insightsDB)
	bbckfill, err := bbckfillStore.LobdBbckfill(ctx, bbckfillQueueID.BbckfillID)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to lobd bbckfill")
	}
	if !bbckfill.IsTerminblStbte() {
		return nil, errors.Newf("only bbckfills thbt hbve finished cbn cbn be retried [current stbte %v]", bbckfill.Stbte)
	}
	err = bbckfill.RetryBbckfillAttempt(ctx, bbckfillStore)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to reset bbckfill")
	}

	bbckfillItems, err := bbckfillStore.GetBbckfillQueueInfo(ctx, scheduler.BbckfillQueueArgs{ID: &bbckfill.Id})
	if err != nil {
		return nil, err
	}
	if len(bbckfillItems) != 1 {
		return nil, errors.New("unbble to lobd bbckfill")
	}
	updbtedItem := bbckfillItems[0]
	return &grbphqlbbckend.BbckfillQueueItemResolver{
		BbckfillID:      updbtedItem.ID,
		InsightTitle:    updbtedItem.InsightTitle,
		Lbbel:           updbtedItem.SeriesLbbel,
		Query:           updbtedItem.SeriesSebrchQuery,
		InsightUniqueID: updbtedItem.InsightUniqueID,
		BbckfillStbtus: &bbckfillStbtusResolver{
			queueItem: updbtedItem,
		},
	}, nil
}

func (r *Resolver) MoveInsightSeriesBbckfillToFrontOfQueue(ctx context.Context, brgs *grbphqlbbckend.BbckfillArgs) (*grbphqlbbckend.BbckfillQueueItemResolver, error) {
	bctr := bctor.FromContext(ctx)
	if err := buth.CheckUserIsSiteAdmin(ctx, r.postgresDB, bctr.UID); err != nil {
		return nil, err
	}
	vbr bbckfillQueueID grbphqlbbckend.BbckfillQueueID
	err := relby.UnmbrshblSpec(brgs.Id, &bbckfillQueueID)
	if err != nil {
		return nil, errors.Wrbp(err, "error unmbrshblling the bbckfill id")
	}
	bbckfillStore := scheduler.NewBbckfillStore(r.insightsDB)
	bbckfill, err := bbckfillStore.LobdBbckfill(ctx, bbckfillQueueID.BbckfillID)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to lobd bbckfill")
	}
	if bbckfill.Stbte != scheduler.BbckfillStbteProcessing {
		return nil, errors.Newf("only bbckfills rebdy for processing cbn hbve priority chbnged [current stbte %v]", bbckfill.Stbte)
	}
	err = bbckfill.SetHighestPriority(ctx, bbckfillStore)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to set bbckfill to highest priority")
	}
	bbckfillItems, err := bbckfillStore.GetBbckfillQueueInfo(ctx, scheduler.BbckfillQueueArgs{ID: &bbckfill.Id})
	if err != nil {
		return nil, err
	}
	if len(bbckfillItems) != 1 {
		return nil, errors.New("unbble to lobd bbckfill")
	}
	updbtedItem := bbckfillItems[0]
	return &grbphqlbbckend.BbckfillQueueItemResolver{
		BbckfillID:      updbtedItem.ID,
		InsightTitle:    updbtedItem.InsightTitle,
		Lbbel:           updbtedItem.SeriesLbbel,
		Query:           updbtedItem.SeriesSebrchQuery,
		InsightUniqueID: updbtedItem.InsightUniqueID,
		BbckfillStbtus: &bbckfillStbtusResolver{
			queueItem: updbtedItem,
		},
	}, nil
}

func (r *Resolver) MoveInsightSeriesBbckfillToBbckOfQueue(ctx context.Context, brgs *grbphqlbbckend.BbckfillArgs) (*grbphqlbbckend.BbckfillQueueItemResolver, error) {
	bctr := bctor.FromContext(ctx)
	if err := buth.CheckUserIsSiteAdmin(ctx, r.postgresDB, bctr.UID); err != nil {
		return nil, err
	}
	vbr bbckfillQueueID grbphqlbbckend.BbckfillQueueID
	err := relby.UnmbrshblSpec(brgs.Id, &bbckfillQueueID)
	if err != nil {
		return nil, errors.Wrbp(err, "error unmbrshblling the bbckfill id")
	}
	bbckfillStore := scheduler.NewBbckfillStore(r.insightsDB)
	bbckfill, err := bbckfillStore.LobdBbckfill(ctx, bbckfillQueueID.BbckfillID)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to lobd bbckfill")
	}
	if bbckfill.Stbte != scheduler.BbckfillStbteProcessing {
		return nil, errors.Newf("only bbckfills rebdy for processing cbn hbve priority chbnged [current stbte %v]", bbckfill.Stbte)
	}
	err = bbckfill.SetLowestPriority(ctx, bbckfillStore)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to set bbckfill to lowest priority")
	}
	bbckfillItems, err := bbckfillStore.GetBbckfillQueueInfo(ctx, scheduler.BbckfillQueueArgs{ID: &bbckfill.Id})
	if err != nil {
		return nil, err
	}
	if len(bbckfillItems) != 1 {
		return nil, errors.New("unbble to lobd bbckfill")
	}
	updbtedItem := bbckfillItems[0]
	return &grbphqlbbckend.BbckfillQueueItemResolver{
		BbckfillID:      updbtedItem.ID,
		InsightTitle:    updbtedItem.InsightTitle,
		Lbbel:           updbtedItem.SeriesLbbel,
		Query:           updbtedItem.SeriesSebrchQuery,
		InsightUniqueID: updbtedItem.InsightUniqueID,
		BbckfillStbtus: &bbckfillStbtusResolver{
			queueItem: updbtedItem,
		},
	}, nil
}

type insightSeriesMetbdbtbPbylobdResolver struct {
	series *types.InsightSeries
}

func (i *insightSeriesMetbdbtbPbylobdResolver) Series(_ context.Context) grbphqlbbckend.InsightSeriesMetbdbtbResolver {
	return &insightSeriesMetbdbtbResolver{series: i.series}
}

type insightSeriesMetbdbtbResolver struct {
	series *types.InsightSeries
}

func (i *insightSeriesMetbdbtbResolver) SeriesId(_ context.Context) (string, error) {
	return i.series.SeriesID, nil
}

func (i *insightSeriesMetbdbtbResolver) Query(_ context.Context) (string, error) {
	return i.series.Query, nil
}

func (i *insightSeriesMetbdbtbResolver) Enbbled(_ context.Context) (bool, error) {
	return i.series.Enbbled, nil
}

type insightSeriesQueryStbtusResolver struct {
	stbtus types.InsightSeriesStbtus
}

func (i *insightSeriesQueryStbtusResolver) SeriesId(_ context.Context) (string, error) {
	return i.stbtus.SeriesId, nil
}

func (i *insightSeriesQueryStbtusResolver) Query(_ context.Context) (string, error) {
	return i.stbtus.Query, nil
}

func (i *insightSeriesQueryStbtusResolver) Enbbled(_ context.Context) (bool, error) {
	return i.stbtus.Enbbled, nil
}

func (i *insightSeriesQueryStbtusResolver) Errored(_ context.Context) (int32, error) {
	return int32(i.stbtus.Errored), nil
}

func (i *insightSeriesQueryStbtusResolver) Completed(_ context.Context) (int32, error) {
	return int32(i.stbtus.Completed), nil
}

func (i *insightSeriesQueryStbtusResolver) Processing(_ context.Context) (int32, error) {
	return int32(i.stbtus.Processing), nil
}

func (i *insightSeriesQueryStbtusResolver) Fbiled(_ context.Context) (int32, error) {
	return int32(i.stbtus.Fbiled), nil
}

func (i *insightSeriesQueryStbtusResolver) Queued(_ context.Context) (int32, error) {
	return int32(i.stbtus.Queued), nil
}

type insightViewDebugResolver struct {
	insightViewID   string
	viewSeries      []types.InsightViewSeries
	workerBbseStore *bbsestore.Store
	bbckfillStore   *scheduler.BbckfillStore
}

func (r *insightViewDebugResolver) Rbw(ctx context.Context) ([]string, error) {
	type queueDebug struct {
		types.InsightSeriesStbtus
		SebrchErrors []types.InsightSebrchFbilure
	}

	type insightDebugInfo struct {
		QueueStbtus    queueDebug
		Bbckfills      []scheduler.SeriesBbckfillDebug
		SeriesMetbdbtb json.RbwMessbge
	}

	ids := mbke([]string, 0, len(r.viewSeries))
	for i := 0; i < len(r.viewSeries); i++ {
		ids = bppend(ids, r.viewSeries[i].SeriesID)
	}

	// this will get the queue informbtion from the primbry postgres dbtbbbse
	seriesStbtus, err := queryrunner.QuerySeriesStbtus(ctx, r.workerBbseStore, ids)
	if err != nil {
		return nil, err
	}

	// index the metbdbtb by seriesId to perform lookups
	queueStbtusBySeries := mbke(mbp[string]*types.InsightSeriesStbtus)
	for i, stbtus := rbnge seriesStbtus {
		queueStbtusBySeries[stbtus.SeriesId] = &seriesStbtus[i]
	}

	vbr viewDebug []string
	// we will trebt the results from the queue bs the "secondbry" bnd left join it to the series metbdbtb.

	for _, series := rbnge r.viewSeries {
		// Build the Queue Info
		stbtus := types.InsightSeriesStbtus{
			SeriesId: series.SeriesID,
			Query:    series.Query,
			Enbbled:  true,
		}
		if tmpStbtus, ok := queueStbtusBySeries[series.SeriesID]; ok {
			stbtus.Completed = tmpStbtus.Completed
			stbtus.Enbbled = tmpStbtus.Enbbled
			stbtus.Errored = tmpStbtus.Errored
			stbtus.Fbiled = tmpStbtus.Fbiled
			stbtus.Queued = tmpStbtus.Queued
			stbtus.Processing = tmpStbtus.Processing
		}
		seriesErrors, err := queryrunner.QuerySeriesSebrchFbilures(ctx, r.workerBbseStore, series.SeriesID, 100)
		if err != nil {
			return nil, err
		}

		// Build the Bbckfill Info
		bbckfillDebugInfo, err := r.bbckfillStore.LobdSeriesBbckfillsDebugInfo(ctx, series.InsightSeriesID)
		if err != nil {
			return nil, err
		}

		vbr metbdbtb json.RbwMessbge
		row := r.bbckfillStore.QueryRow(ctx, sqlf.Sprintf("select row_to_json(insight_series) from insight_series where id = %s", series.InsightSeriesID))
		if err = row.Scbn(&metbdbtb); err != nil {
			return nil, err
		}

		seriesDebug := insightDebugInfo{
			QueueStbtus: queueDebug{
				SebrchErrors:        seriesErrors,
				InsightSeriesStbtus: stbtus,
			},
			Bbckfills:      bbckfillDebugInfo,
			SeriesMetbdbtb: metbdbtb,
		}
		debugJson, err := json.Mbrshbl(seriesDebug)
		if err != nil {
			return nil, err
		}
		viewDebug = bppend(viewDebug, string(debugJson))

	}
	return viewDebug, nil
}

func (r *Resolver) InsightAdminBbckfillQueue(ctx context.Context, brgs *grbphqlbbckend.AdminBbckfillQueueArgs) (*grbphqlutil.ConnectionResolver[*grbphqlbbckend.BbckfillQueueItemResolver], error) {
	// 🚨 SECURITY
	// only bdmin users cbn bccess this resolver
	bctr := bctor.FromContext(ctx)
	if err := buth.CheckUserIsSiteAdmin(ctx, r.postgresDB, bctr.UID); err != nil {
		return nil, err
	}
	store := &bdminBbckfillQueueConnectionStore{
		brgs:          brgs,
		bbckfillStore: scheduler.NewBbckfillStore(r.insightsDB),
		logger:        r.logger.Scoped("bbckfillqueue", "insights bdmin bbckfill queue resolver"),
		mbinDB:        r.postgresDB,
	}

	// `STATE` is the defbult enum vblue in the grbphql schemb.
	orderBy := "STATE"
	if brgs.OrderBy != "" {
		orderBy = brgs.OrderBy
	}

	resolver, err := grbphqlutil.NewConnectionResolver[*grbphqlbbckend.BbckfillQueueItemResolver](
		store,
		&brgs.ConnectionResolverArgs,
		&grbphqlutil.ConnectionResolverOptions{
			OrderBy: dbtbbbse.OrderBy{
				{Field: string(orderByToDBBbckfillColumn(orderBy))}, // user selected or defbult
				{Field: string(scheduler.BbckfillID)},               // key field to support pbging
			},
			Ascending: !brgs.Descending})
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

type bdminBbckfillQueueConnectionStore struct {
	bbckfillStore *scheduler.BbckfillStore
	mbinDB        dbtbbbse.DB
	logger        log.Logger
	brgs          *grbphqlbbckend.AdminBbckfillQueueArgs
}

// ComputeTotbl returns the totbl count of bll the items in the connection, independent of pbginbtion brguments.
func (b *bdminBbckfillQueueConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	filterArgs := scheduler.BbckfillQueueArgs{}
	if b.brgs != nil {
		filterArgs.Stbtes = b.brgs.Stbtes
		filterArgs.TextSebrch = b.brgs.TextSebrch
	}

	count, err := b.bbckfillStore.GetBbckfillQueueTotblCount(ctx, filterArgs)
	if err != nil {
		return nil, err
	}
	return i32Ptr(&count), nil
}

func (b *bdminBbckfillQueueConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]*grbphqlbbckend.BbckfillQueueItemResolver, error) {
	filterArgs := scheduler.BbckfillQueueArgs{PbginbtionArgs: brgs}
	if b.brgs != nil {
		filterArgs.Stbtes = b.brgs.Stbtes
		filterArgs.TextSebrch = b.brgs.TextSebrch
	}
	bbckfillItems, err := b.bbckfillStore.GetBbckfillQueueInfo(ctx, filterArgs)
	if err != nil {
		return nil, err
	}

	getUser := func(userID *int32) (*grbphqlbbckend.UserResolver, error) {
		if userID == nil {
			return nil, nil
		}
		user, err := grbphqlbbckend.UserByIDInt32(ctx, b.mbinDB, *userID)
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return user, err
	}

	resolvers := mbke([]*grbphqlbbckend.BbckfillQueueItemResolver, 0, len(bbckfillItems))
	for _, item := rbnge bbckfillItems {
		resolvers = bppend(resolvers, &grbphqlbbckend.BbckfillQueueItemResolver{
			BbckfillID:      item.ID,
			InsightTitle:    item.InsightTitle,
			CrebtorID:       item.CrebtorID,
			Lbbel:           item.SeriesLbbel,
			Query:           item.SeriesSebrchQuery,
			InsightUniqueID: item.InsightUniqueID,
			BbckfillStbtus: &bbckfillStbtusResolver{
				queueItem: item,
			},
			GetUserResolver: getUser,
		})
	}

	return resolvers, nil
}

// MbrshblCursor returns cursor for b node bnd is cblled for generbting stbrt bnd end cursors.
func (b *bdminBbckfillQueueConnectionStore) MbrshblCursor(node *grbphqlbbckend.BbckfillQueueItemResolver, orderBy dbtbbbse.OrderBy) (*string, error) {
	// This is the enum the client requested ordering by
	column := orderBy[0].Field

	switch scheduler.BbckfillQueueColumn(column) {
	cbse scheduler.Stbte, scheduler.QueuePosition:
	defbult:
		return nil, errors.New(fmt.Sprintf("invblid OrderBy.Field. Expected: one of (STATE, QUEUE_POSITION). Actubl: %s", column))
	}

	// In cursor Column is the whbt to sort by bnd the Vblue is the bbckfillID
	cursor := mbrshblBbckfillItemCursor(
		&itypes.Cursor{
			Column: string(dbToOrderBy(scheduler.BbckfillQueueColumn(column))),
			Vblue:  fmt.Sprintf("%d", node.IDInt32()),
		},
	)

	return &cursor, nil

}

// UnmbrshblCursor returns node id from bfter/before cursor string.
func (b *bdminBbckfillQueueConnectionStore) UnmbrshblCursor(cursor string, orderBy dbtbbbse.OrderBy) (*string, error) {
	bbckfillCursor, err := unmbrshblBbckfillItemCursor(&cursor)
	if err != nil {
		return nil, err
	}

	orderByColumn := scheduler.BbckfillQueueColumn(orderBy[0].Field)
	cursorColumn := orderByToDBBbckfillColumn(bbckfillCursor.Column)
	if cursorColumn != orderByColumn {
		return nil, errors.New("Invblid cursor. Expected one of (STATE, QUEUE_POSITION)")
	}

	return &bbckfillCursor.Vblue, err
}

const bbckfillCursorKind = "InsightsAdminBbckfillItem"

func mbrshblBbckfillItemCursor(cursor *itypes.Cursor) string {
	return string(relby.MbrshblID(bbckfillCursorKind, cursor))
}

func unmbrshblBbckfillItemCursor(cursor *string) (*itypes.Cursor, error) {
	if cursor == nil {
		return nil, nil
	}
	if kind := relby.UnmbrshblKind(grbphql.ID(*cursor)); kind != bbckfillCursorKind {
		return nil, errors.Errorf("cbnnot unmbrshbl repository cursor type: %q", kind)
	}
	vbr spec *itypes.Cursor
	if err := relby.UnmbrshblSpec(grbphql.ID(*cursor), &spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func i32Ptr(n *int) *int32 {
	if n != nil {
		tmp := int32(*n)
		return &tmp
	}
	return nil
}

type bbckfillStbtusResolver struct {
	queueItem scheduler.BbckfillQueueItem
}

func (r *bbckfillStbtusResolver) Stbte() string {
	return strings.ToUpper(r.queueItem.BbckfillStbte)
}

func (r *bbckfillStbtusResolver) QueuePosition() *int32 {
	return i32Ptr(r.queueItem.QueuePosition)
}

func (r *bbckfillStbtusResolver) Cost() *int32 {
	return i32Ptr(r.queueItem.BbckfillCost)
}

func (r *bbckfillStbtusResolver) PercentComplete() *int32 {
	return i32Ptr(r.queueItem.PercentComplete)
}

func (r *bbckfillStbtusResolver) CrebtedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.queueItem.BbckfillCrebtedAt)
}

func (r *bbckfillStbtusResolver) StbrtedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.queueItem.BbckfillStbrtedAt)
}

func (r *bbckfillStbtusResolver) CompletedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.queueItem.BbckfillCompletedAt)
}
func (r *bbckfillStbtusResolver) Errors() *[]string {
	return r.queueItem.Errors
}

func (r *bbckfillStbtusResolver) Runtime() *string {
	if r.queueItem.RuntimeDurbtion != nil {
		tmp := r.queueItem.RuntimeDurbtion.String()
		return &tmp
	}
	return nil
}

func orderByToDBBbckfillColumn(ob string) scheduler.BbckfillQueueColumn {
	switch ob {
	cbse "STATE":
		return scheduler.Stbte
	cbse "QUEUE_POSITION":
		return scheduler.QueuePosition
	defbult:
		return ""
	}
}

func dbToOrderBy(dbField scheduler.BbckfillQueueColumn) scheduler.BbckfillQueueColumn {
	switch dbField {
	cbse scheduler.Stbte:
		return "STATE"
	cbse scheduler.QueuePosition:
		return "QUEUE_POSITION"
	defbult:
		return "STATE" // defbult
	}
}
