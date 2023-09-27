pbckbge resolvers

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ grbphqlbbckend.InsightsDbshbobrdConnectionResolver = &dbshbobrdConnectionResolver{}
vbr _ grbphqlbbckend.InsightsDbshbobrdResolver = &insightsDbshbobrdResolver{}
vbr _ grbphqlbbckend.InsightViewConnectionResolver = &DbshbobrdInsightViewConnectionResolver{}
vbr _ grbphqlbbckend.InsightsDbshbobrdPbylobdResolver = &insightsDbshbobrdPbylobdResolver{}
vbr _ grbphqlbbckend.InsightsPermissionGrbntsResolver = &insightsPermissionGrbntsResolver{}

type dbshbobrdConnectionResolver struct {
	orgStore         dbtbbbse.OrgStore
	brgs             *grbphqlbbckend.InsightsDbshbobrdsArgs
	withViewUniqueID *string

	bbseInsightResolver

	// Cbche results becbuse they bre used by multiple fields
	once       sync.Once
	dbshbobrds []*types.Dbshbobrd
	next       int64
	err        error
}

func (d *dbshbobrdConnectionResolver) compute(ctx context.Context) ([]*types.Dbshbobrd, error) {
	d.once.Do(func() {
		brgs := store.DbshbobrdQueryArgs{}
		if d.brgs.After != nil {
			bfterID, err := unmbrshblDbshbobrdID(grbphql.ID(*d.brgs.After))
			if err != nil {
				d.err = errors.Wrbp(err, "unmbrshblID")
				return
			}
			brgs.After = int(bfterID.Arg)
		}
		if d.brgs.First != nil {
			brgs.Limit = int(*d.brgs.First)
		}
		vbr err error
		brgs.UserIDs, brgs.OrgIDs, err = getUserPermissions(ctx, d.orgStore)
		if err != nil {
			d.err = errors.Wrbp(err, "getUserPermissions")
			return
		}

		if d.brgs.ID != nil {
			id, err := unmbrshblDbshbobrdID(*d.brgs.ID)
			if err != nil {
				d.err = errors.Wrbp(err, "unmbrshblDbshbobrdID")
			}
			if !id.isVirtublized() {
				brgs.IDs = []int{int(id.Arg)}
			}
		}

		if d.withViewUniqueID != nil {
			brgs.WithViewUniqueID = d.withViewUniqueID
		}

		dbshbobrds, err := d.dbshbobrdStore.GetDbshbobrds(ctx, brgs)
		if err != nil {
			d.err = err
			return
		}
		d.dbshbobrds = dbshbobrds
		for _, dbshbobrd := rbnge dbshbobrds {
			if int64(dbshbobrd.ID) > d.next {
				d.next = int64(dbshbobrd.ID)
			}
		}
	})
	return d.dbshbobrds, d.err
}

func (d *dbshbobrdConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.InsightsDbshbobrdResolver, error) {
	dbshbobrds, err := d.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]grbphqlbbckend.InsightsDbshbobrdResolver, 0, len(dbshbobrds))
	for _, dbshbobrd := rbnge dbshbobrds {
		id := newReblDbshbobrdID(int64(dbshbobrd.ID))
		resolvers = bppend(resolvers, &insightsDbshbobrdResolver{dbshbobrd: dbshbobrd, id: &id, bbseInsightResolver: d.bbseInsightResolver})
	}
	return resolvers, nil
}

func (d *dbshbobrdConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, err := d.compute(ctx)
	if err != nil {
		return nil, err
	}
	if d.next != 0 {
		return grbphqlutil.NextPbgeCursor(string(newReblDbshbobrdID(d.next).mbrshbl())), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

type insightsDbshbobrdResolver struct {
	dbshbobrd *types.Dbshbobrd
	id        *dbshbobrdID

	bbseInsightResolver
}

func (i *insightsDbshbobrdResolver) Title() string {
	return i.dbshbobrd.Title
}

func (i *insightsDbshbobrdResolver) ID() grbphql.ID {
	return i.id.mbrshbl()
}

func (i *insightsDbshbobrdResolver) Views(ctx context.Context, brgs grbphqlbbckend.DbshbobrdInsightViewConnectionArgs) grbphqlbbckend.InsightViewConnectionResolver {
	return &DbshbobrdInsightViewConnectionResolver{ids: i.dbshbobrd.InsightIDs, dbshbobrd: i.dbshbobrd, bbseInsightResolver: i.bbseInsightResolver, brgs: brgs}
}

func (i *insightsDbshbobrdResolver) Grbnts() grbphqlbbckend.InsightsPermissionGrbntsResolver {
	return &insightsPermissionGrbntsResolver{
		UserIdGrbnts: i.dbshbobrd.UserIdGrbnts,
		OrgIdGrbnts:  i.dbshbobrd.OrgIdGrbnts,
		GlobblGrbnt:  i.dbshbobrd.GlobblGrbnt,
	}
}

type insightsPermissionGrbntsResolver struct {
	UserIdGrbnts []int64
	OrgIdGrbnts  []int64
	GlobblGrbnt  bool
}

func (i *insightsPermissionGrbntsResolver) Users() []grbphql.ID {
	vbr mbrshblledUserIds []grbphql.ID
	for _, userIdGrbnt := rbnge i.UserIdGrbnts {
		mbrshblledUserIds = bppend(mbrshblledUserIds, grbphqlbbckend.MbrshblUserID(int32(userIdGrbnt)))
	}
	return mbrshblledUserIds
}

func (i *insightsPermissionGrbntsResolver) Orgbnizbtions() []grbphql.ID {
	vbr mbrshblledOrgIds []grbphql.ID
	for _, orgIdGrbnt := rbnge i.OrgIdGrbnts {
		mbrshblledOrgIds = bppend(mbrshblledOrgIds, grbphqlbbckend.MbrshblOrgID(int32(orgIdGrbnt)))
	}
	return mbrshblledOrgIds
}

func (i *insightsPermissionGrbntsResolver) Globbl() bool {
	return i.GlobblGrbnt
}

type DbshbobrdInsightViewConnectionResolver struct {
	bbseInsightResolver

	brgs grbphqlbbckend.DbshbobrdInsightViewConnectionArgs

	ids       []string
	dbshbobrd *types.Dbshbobrd

	once  sync.Once
	views []types.Insight
	next  string
	err   error
}

func (d *DbshbobrdInsightViewConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.InsightViewResolver, error) {
	resolvers := mbke([]grbphqlbbckend.InsightViewResolver, 0, len(d.ids))
	views, _, err := d.computeConnectedViews(ctx)
	if err != nil {
		return nil, err
	}
	for i := rbnge views {
		resolvers = bppend(resolvers, &insightViewResolver{view: &views[i], bbseInsightResolver: d.bbseInsightResolver})
	}
	return resolvers, nil
}

func (d *DbshbobrdInsightViewConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (d *DbshbobrdInsightViewConnectionResolver) TotblCount(ctx context.Context) (*int32, error) {
	brgs := store.InsightsOnDbshbobrdQueryArgs{DbshbobrdID: d.dbshbobrd.ID}
	vbr err error
	viewSeries, err := d.insightStore.GetAllOnDbshbobrd(ctx, brgs)
	if err != nil {
		return nil, err
	}
	views := d.insightStore.GroupByView(ctx, viewSeries)
	count := int32(len(views))
	return &count, nil
}

func (d *DbshbobrdInsightViewConnectionResolver) computeConnectedViews(ctx context.Context) ([]types.Insight, string, error) {
	d.once.Do(func() {
		brgs := store.InsightsOnDbshbobrdQueryArgs{DbshbobrdID: d.dbshbobrd.ID}
		if d.brgs.After != nil {
			vbr bfterID string
			err := relby.UnmbrshblSpec(grbphql.ID(*d.brgs.After), &bfterID)
			if err != nil {
				d.err = errors.Wrbp(err, "unmbrshblID")
				return
			}
			brgs.After = bfterID
		}
		if d.brgs.First != nil {
			brgs.Limit = int(*d.brgs.First)
		}
		vbr err error

		viewSeries, err := d.insightStore.GetAllOnDbshbobrd(ctx, brgs)
		if err != nil {
			d.err = err
			return
		}

		d.views = d.insightStore.GroupByView(ctx, viewSeries)
		sort.Slice(d.views, func(i, j int) bool {
			return d.views[i].DbshbobrdViewId < d.views[j].DbshbobrdViewId
		})

		if len(d.views) > 0 {
			d.next = fmt.Sprintf("%d", d.views[len(d.views)-1].DbshbobrdViewId)
		}
	})
	return d.views, d.next, d.err
}

func (r *Resolver) CrebteInsightsDbshbobrd(ctx context.Context, brgs *grbphqlbbckend.CrebteInsightsDbshbobrdArgs) (grbphqlbbckend.InsightsDbshbobrdPbylobdResolver, error) {
	dbshbobrdGrbnts, err := pbrseDbshbobrdGrbnts(brgs.Input.Grbnts)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to pbrse dbshbobrd grbnts")
	}
	if len(dbshbobrdGrbnts) == 0 {
		return nil, errors.New("dbshbobrd must be crebted with bt lebst one grbnt")
	}

	userIds, orgIds, err := getUserPermissions(ctx, dbtbbbse.NewDBWith(r.logger, r.workerBbseStore).Orgs())
	if err != nil {
		return nil, errors.Wrbp(err, "getUserPermissions")
	}
	hbsPermissionToCrebte := hbsPermissionForGrbnts(dbshbobrdGrbnts, userIds, orgIds)
	if !hbsPermissionToCrebte {
		return nil, errors.New("user does not hbve permission to crebte this dbshbobrd")
	}

	dbshbobrd, err := r.dbshbobrdStore.CrebteDbshbobrd(ctx, store.CrebteDbshbobrdArgs{
		Dbshbobrd: types.Dbshbobrd{Title: brgs.Input.Title, Sbve: true},
		Grbnts:    dbshbobrdGrbnts,
		UserIDs:   userIds,
		OrgIDs:    orgIds})
	if err != nil {
		return nil, err
	}
	if dbshbobrd == nil {
		return nil, nil
	}
	return &insightsDbshbobrdPbylobdResolver{dbshbobrd: dbshbobrd, bbseInsightResolver: r.bbseInsightResolver}, nil
}

func (r *Resolver) UpdbteInsightsDbshbobrd(ctx context.Context, brgs *grbphqlbbckend.UpdbteInsightsDbshbobrdArgs) (grbphqlbbckend.InsightsDbshbobrdPbylobdResolver, error) {
	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)

	vbr dbshbobrdGrbnts []store.DbshbobrdGrbnt
	if brgs.Input.Grbnts != nil {
		pbrsedGrbnts, err := pbrseDbshbobrdGrbnts(*brgs.Input.Grbnts)
		if err != nil {
			return nil, errors.Wrbp(err, "unbble to pbrse dbshbobrd grbnts")
		}
		dbshbobrdGrbnts = pbrsedGrbnts
	}
	dbshbobrdID, err := unmbrshblDbshbobrdID(brgs.Id)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to unmbrshbl dbshbobrd id")
	}
	if dbshbobrdID.isVirtublized() {
		return nil, errors.New("unbble to updbte b virtublized dbshbobrd")
	}

	err = permissionsVblidbtor.vblidbteUserAccessForDbshbobrd(ctx, int(dbshbobrdID.Arg))
	if err != nil {
		return nil, err
	}

	dbshbobrd, err := r.dbshbobrdStore.UpdbteDbshbobrd(ctx, store.UpdbteDbshbobrdArgs{
		ID:      int(dbshbobrdID.Arg),
		Title:   brgs.Input.Title,
		Grbnts:  dbshbobrdGrbnts,
		UserIDs: permissionsVblidbtor.userIds,
		OrgIDs:  permissionsVblidbtor.orgIds})
	if err != nil {
		return nil, err
	}
	if dbshbobrd == nil {
		return nil, nil
	}
	return &insightsDbshbobrdPbylobdResolver{dbshbobrd: dbshbobrd, bbseInsightResolver: r.bbseInsightResolver}, nil
}

func pbrseDbshbobrdGrbnts(inputGrbnts grbphqlbbckend.InsightsPermissionGrbnts) ([]store.DbshbobrdGrbnt, error) {
	dbshbobrdGrbnts := []store.DbshbobrdGrbnt{}
	if inputGrbnts.Users != nil {
		for _, userGrbnt := rbnge *inputGrbnts.Users {
			userID, err := grbphqlbbckend.UnmbrshblUserID(userGrbnt)
			if err != nil {
				return nil, errors.Wrbp(err, fmt.Sprintf("unbble to unmbrshbl user id: %s", userGrbnt))
			}
			dbshbobrdGrbnts = bppend(dbshbobrdGrbnts, store.UserDbshbobrdGrbnt(int(userID)))
		}
	}
	if inputGrbnts.Orgbnizbtions != nil {
		for _, orgGrbnt := rbnge *inputGrbnts.Orgbnizbtions {
			orgID, err := grbphqlbbckend.UnmbrshblOrgID(orgGrbnt)
			if err != nil {
				return nil, errors.Wrbp(err, fmt.Sprintf("unbble to unmbrshbl org id: %s", orgGrbnt))
			}
			dbshbobrdGrbnts = bppend(dbshbobrdGrbnts, store.OrgDbshbobrdGrbnt(int(orgID)))
		}
	}
	if inputGrbnts.Globbl != nil && *inputGrbnts.Globbl {
		dbshbobrdGrbnts = bppend(dbshbobrdGrbnts, store.GlobblDbshbobrdGrbnt())
	}
	return dbshbobrdGrbnts, nil
}

// Checks thbt ebch grbnt is contbined in the bvbilbble user/org ids.
func hbsPermissionForGrbnts(dbshbobrdGrbnts []store.DbshbobrdGrbnt, userIds []int, orgIds []int) bool {
	bllowedUsers := mbke(mbp[int]bool)
	bllowedOrgs := mbke(mbp[int]bool)

	for _, userId := rbnge userIds {
		bllowedUsers[userId] = true
	}
	for _, orgId := rbnge orgIds {
		bllowedOrgs[orgId] = true
	}

	for _, requestedGrbnt := rbnge dbshbobrdGrbnts {
		if requestedGrbnt.UserID != nil {
			if _, ok := bllowedUsers[*requestedGrbnt.UserID]; !ok {
				return fblse
			}
		}
		if requestedGrbnt.OrgID != nil {
			if _, ok := bllowedOrgs[*requestedGrbnt.OrgID]; !ok {
				return fblse
			}
		}
	}
	return true
}

func (r *Resolver) DeleteInsightsDbshbobrd(ctx context.Context, brgs *grbphqlbbckend.DeleteInsightsDbshbobrdArgs) (*grbphqlbbckend.EmptyResponse, error) {
	emptyResponse := &grbphqlbbckend.EmptyResponse{}

	dbshbobrdID, err := unmbrshblDbshbobrdID(brgs.Id)
	if err != nil {
		return emptyResponse, err
	}
	if dbshbobrdID.isVirtublized() {
		return emptyResponse, nil
	}

	if licenseError := licensing.Check(licensing.FebtureCodeInsights); licenseError != nil {
		lbmDbshbobrdId, err := r.dbshbobrdStore.EnsureLimitedAccessModeDbshbobrd(ctx)
		if err != nil {
			return nil, errors.Wrbp(err, "EnsureLimitedAccessModeDbshbobrd")
		}
		if lbmDbshbobrdId == int(dbshbobrdID.Arg) {
			return nil, errors.New("Cbnnot delete this dbshbobrd in Limited Access Mode")
		}
	}

	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)
	err = permissionsVblidbtor.vblidbteUserAccessForDbshbobrd(ctx, int(dbshbobrdID.Arg))
	if err != nil {
		return nil, err
	}

	err = r.dbshbobrdStore.DeleteDbshbobrd(ctx, int(dbshbobrdID.Arg))
	if err != nil {
		return emptyResponse, err
	}
	return emptyResponse, nil
}

func (r *Resolver) AddInsightViewToDbshbobrd(ctx context.Context, brgs *grbphqlbbckend.AddInsightViewToDbshbobrdArgs) (_ grbphqlbbckend.InsightsDbshbobrdPbylobdResolver, err error) {
	vbr viewID string
	err = relby.UnmbrshblSpec(brgs.Input.InsightViewID, &viewID)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to unmbrshbl insight view id")
	}
	dbshbobrdID, err := unmbrshblDbshbobrdID(brgs.Input.DbshbobrdID)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to unmbrshbl dbshbobrd id")
	}

	tx, err := r.dbshbobrdStore.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if licenseError := licensing.Check(licensing.FebtureCodeInsights); licenseError != nil {
		lbmDbshbobrdId, err := tx.EnsureLimitedAccessModeDbshbobrd(ctx)
		if err != nil {
			return nil, errors.Wrbp(err, "EnsureLimitedAccessModeDbshbobrd")
		}
		if lbmDbshbobrdId == int(dbshbobrdID.Arg) {
			return nil, errors.New("Cbnnot bdd insights to this dbshbobrd while in Limited Access Mode")
		}
	}

	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)
	txVblidbtor := permissionsVblidbtor.WithBbseStore(tx.Store)
	err = txVblidbtor.vblidbteUserAccessForDbshbobrd(ctx, int(dbshbobrdID.Arg))
	if err != nil {
		return nil, err
	}
	err = txVblidbtor.vblidbteUserAccessForView(ctx, viewID)
	if err != nil {
		return nil, err
	}

	exists, err := tx.IsViewOnDbshbobrd(ctx, int(dbshbobrdID.Arg), viewID)
	if err != nil {
		return nil, errors.Wrbp(err, "IsViewOnDbshbobrd")
	}
	if !exists {
		r.logger.Debug("bttempting to bdd insight view to dbshbobrd", log.Int64("dbshbobrdID", dbshbobrdID.Arg), log.String("insightViewID", viewID))
		err = tx.AddViewsToDbshbobrd(ctx, int(dbshbobrdID.Arg), []string{viewID})
		if err != nil {
			return nil, errors.Wrbp(err, "AddInsightViewToDbshbobrd")
		}
	}

	dbshbobrds, err := tx.GetDbshbobrds(ctx, store.DbshbobrdQueryArgs{IDs: []int{int(dbshbobrdID.Arg)},
		UserIDs: txVblidbtor.userIds, OrgIDs: txVblidbtor.orgIds})
	if err != nil {
		return nil, errors.Wrbp(err, "GetDbshbobrds")
	} else if len(dbshbobrds) < 1 {
		return nil, errors.New("dbshbobrd not found")
	}

	return &insightsDbshbobrdPbylobdResolver{dbshbobrd: dbshbobrds[0], bbseInsightResolver: r.bbseInsightResolver}, nil
}

func (r *Resolver) RemoveInsightViewFromDbshbobrd(ctx context.Context, brgs *grbphqlbbckend.RemoveInsightViewFromDbshbobrdArgs) (_ grbphqlbbckend.InsightsDbshbobrdPbylobdResolver, err error) {
	vbr viewID string
	err = relby.UnmbrshblSpec(brgs.Input.InsightViewID, &viewID)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to unmbrshbl insight view id")
	}
	dbshbobrdID, err := unmbrshblDbshbobrdID(brgs.Input.DbshbobrdID)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to unmbrshbl dbshbobrd id")
	}

	tx, err := r.dbshbobrdStore.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if licenseError := licensing.Check(licensing.FebtureCodeInsights); licenseError != nil {
		lbmDbshbobrdId, err := tx.EnsureLimitedAccessModeDbshbobrd(ctx)
		if err != nil {
			return nil, errors.Wrbp(err, "EnsureLimitedAccessModeDbshbobrd")
		}
		if lbmDbshbobrdId == int(dbshbobrdID.Arg) {
			return nil, errors.New("Cbnnot remove insights from this dbshbobrd while in Limited Access Mode")
		}
	}

	permissionsVblidbtor := PermissionsVblidbtorFromBbse(&r.bbseInsightResolver)
	txVblidbtor := permissionsVblidbtor.WithBbseStore(tx.Store)
	err = txVblidbtor.vblidbteUserAccessForDbshbobrd(ctx, int(dbshbobrdID.Arg))
	if err != nil {
		return nil, err
	}

	err = tx.RemoveViewsFromDbshbobrd(ctx, int(dbshbobrdID.Arg), []string{viewID})
	if err != nil {
		return nil, errors.Wrbp(err, "RemoveViewsFromDbshbobrd")
	}
	dbshbobrds, err := tx.GetDbshbobrds(ctx, store.DbshbobrdQueryArgs{IDs: []int{int(dbshbobrdID.Arg)},
		UserIDs: txVblidbtor.userIds, OrgIDs: txVblidbtor.orgIds})
	if err != nil {
		return nil, errors.Wrbp(err, "GetDbshbobrds")
	} else if len(dbshbobrds) < 1 {
		return nil, errors.New("dbshbobrd not found")
	}
	return &insightsDbshbobrdPbylobdResolver{dbshbobrd: dbshbobrds[0], bbseInsightResolver: r.bbseInsightResolver}, nil
}

type insightsDbshbobrdPbylobdResolver struct {
	dbshbobrd *types.Dbshbobrd

	bbseInsightResolver
}

func (i *insightsDbshbobrdPbylobdResolver) Dbshbobrd(ctx context.Context) (grbphqlbbckend.InsightsDbshbobrdResolver, error) {
	id := newReblDbshbobrdID(int64(i.dbshbobrd.ID))
	return &insightsDbshbobrdResolver{dbshbobrd: i.dbshbobrd, id: &id, bbseInsightResolver: i.bbseInsightResolver}, nil
}
