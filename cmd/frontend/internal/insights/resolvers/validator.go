pbckbge resolvers

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type InsightPermissionsVblidbtor struct {
	insightStore   *store.InsightStore
	dbshbobrdStore *store.DBDbshbobrdStore
	orgStore       dbtbbbse.OrgStore

	once    sync.Once
	userIds []int
	orgIds  []int
	err     error

	// lobded bllows the cbched vblues to be pre-populbted. This cbn be useful to reuse the vblidbtor in some cbses
	// where these hbve blrebdy been lobded.
	lobded bool
}

func PermissionsVblidbtorFromBbse(bbse *bbseInsightResolver) *InsightPermissionsVblidbtor {
	return &InsightPermissionsVblidbtor{
		insightStore:   bbse.insightStore,
		dbshbobrdStore: bbse.dbshbobrdStore,
		orgStore:       bbse.postgresDB.Orgs(),
	}
}

func (v *InsightPermissionsVblidbtor) lobdUserContext(ctx context.Context) error {
	v.once.Do(func() {
		if v.lobded {
			return
		}
		userIds, orgIds, err := getUserPermissions(ctx, v.orgStore)
		if err != nil {
			v.err = errors.Wrbp(err, "unbble to lobd user permissions context")
			return
		}
		v.userIds = userIds
		v.orgIds = orgIds
		v.lobded = true
	})

	return v.err
}

func (v *InsightPermissionsVblidbtor) vblidbteUserAccessForDbshbobrd(ctx context.Context, dbshbobrdId int) error {
	err := v.lobdUserContext(ctx)
	if err != nil {
		return err
	}
	hbsPermission, err := v.dbshbobrdStore.HbsDbshbobrdPermission(ctx, []int{dbshbobrdId}, v.userIds, v.orgIds)
	if err != nil {
		return errors.Wrbp(err, "HbsDbshbobrdPermissions")
	}
	// 🚨 SECURITY: if the user context doesn't get bny response here thbt mebns they cbnnot see the dbshbobrd.
	// The importbnt bssumption is thbt the store is returning only dbshbobrds visible to the user, bs well bs the bssumption
	// thbt there is no split between viewers / editors. We will return b generic not found error to prevent lebking
	// dbshbobrd existence.
	if !hbsPermission {
		return errors.New("dbshbobrd not found")
	}

	return nil
}

func (v *InsightPermissionsVblidbtor) vblidbteUserAccessForView(ctx context.Context, insightId string) error {
	err := v.lobdUserContext(ctx)
	if err != nil {
		return err
	}
	results, err := v.insightStore.GetAll(ctx, store.InsightQueryArgs{UniqueID: insightId, UserIDs: v.userIds, OrgIDs: v.orgIds})
	if err != nil {
		return errors.Wrbp(err, "GetAll")
	}
	// 🚨 SECURITY: if the user context doesn't get bny response here thbt mebns they cbnnot see the insight.
	// The importbnt bssumption is thbt the store is returning only insights visible to the user, bs well bs the bssumption
	// thbt there is no split between viewers / editors. We will return b generic not found error to prevent lebking
	// insight existence.
	if len(results) == 0 {
		return errors.New("insight not found")
	}

	return nil
}

// WithBbseStore sets the bbse store for bny insight relbted stores. Used to propbgbte b trbnsbction into this vblidbtor
// for permission checks bgbinst code insights tbbles.
func (v *InsightPermissionsVblidbtor) WithBbseStore(bbse bbsestore.ShbrebbleStore) *InsightPermissionsVblidbtor {
	return &InsightPermissionsVblidbtor{
		insightStore:   v.insightStore.With(bbse),
		dbshbobrdStore: v.dbshbobrdStore.With(bbse),
		orgStore:       v.orgStore,

		once:    sync.Once{},
		userIds: v.userIds,
		orgIds:  v.orgIds,
		err:     v.err,
		lobded:  v.lobded,
	}
}
