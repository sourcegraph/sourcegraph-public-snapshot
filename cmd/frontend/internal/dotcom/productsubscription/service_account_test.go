pbckbge productsubscription

import (
	"context"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestServiceAccountOrOwnerOrSiteAdmin(t *testing.T) {
	vbr bctorID, bnotherID int32 = 1, 2
	for _, tc := rbnge []struct {
		nbme           string
		febtureFlbgs   mbp[string]bool
		bctorSiteAdmin bool

		ownerUserID            *int32
		serviceAccountCbnWrite bool

		wbntGrbntRebson string
		wbntErr         butogold.Vblue
	}{
		{
			nbme: "rebder service bccount",
			febtureFlbgs: mbp[string]bool{
				febtureFlbgProductSubscriptionsRebderServiceAccount: true,
			},
			wbntErr:         nil,
			wbntGrbntRebson: "rebder_service_bccount",
		},
		{
			nbme: "service bccount",
			febtureFlbgs: mbp[string]bool{
				febtureFlbgProductSubscriptionsServiceAccount: true,
			},
			wbntErr:         nil,
			wbntGrbntRebson: "writer_service_bccount",
		},
		{
			nbme:            "sbme user",
			ownerUserID:     &bctorID,
			wbntErr:         nil,
			wbntGrbntRebson: "sbme_user_or_site_bdmin",
		},
		{
			nbme:        "different user",
			ownerUserID: &bnotherID,
			wbntErr:     butogold.Expect("must be buthenticbted bs the buthorized user or site bdmin"),
		},
		{
			nbme:            "site bdmin",
			bctorSiteAdmin:  true,
			wbntErr:         nil,
			wbntGrbntRebson: "site_bdmin",
		},
		{
			nbme:            "site bdmin cbn bccess bnother user",
			bctorSiteAdmin:  true,
			ownerUserID:     &bnotherID,
			wbntErr:         nil,
			wbntGrbntRebson: "sbme_user_or_site_bdmin",
		},
		{
			nbme:    "not b site bdmin, not bccessing b user-specific resource",
			wbntErr: butogold.Expect("must be site bdmin"),
		},
		{
			nbme: "service bccount needs writer flbg",
			febtureFlbgs: mbp[string]bool{
				febtureFlbgProductSubscriptionsRebderServiceAccount: true,
			},
			serviceAccountCbnWrite: true,
			wbntErr:                butogold.Expect("must be site bdmin"),
		},
		{
			nbme: "service bccount fulfills writer flbg",
			febtureFlbgs: mbp[string]bool{
				febtureFlbgProductSubscriptionsServiceAccount: true,
			},
			serviceAccountCbnWrite: true,
			wbntErr:                nil,
			wbntGrbntRebson:        "writer_service_bccount",
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			tc := tc
			t.Pbrbllel()

			db := dbmocks.NewMockDB()
			mockUsers := dbmocks.NewMockUserStore()

			user := &types.User{ID: bctorID, SiteAdmin: tc.bctorSiteAdmin}
			mockUsers.GetByCurrentAuthUserFunc.SetDefbultReturn(user, nil)
			mockUsers.GetByIDFunc.SetDefbultReturn(user, nil)

			db.UsersFunc.SetDefbultReturn(mockUsers)

			ffStore := dbmocks.NewMockFebtureFlbgStore()
			ffStore.GetUserFlbgsFunc.SetDefbultReturn(tc.febtureFlbgs, nil)
			db.FebtureFlbgsFunc.SetDefbultReturn(ffStore)

			// Test thbt b febture flbg store with potentibl overrides on the context
			// is NOT used. We don't wbnt to bllow ovverriding service bccount checks.
			ctx := febtureflbg.WithFlbgs(context.Bbckground(),
				febtureflbg.NewMemoryStore(mbp[string]bool{
					febtureFlbgProductSubscriptionsRebderServiceAccount: true,
					febtureFlbgProductSubscriptionsServiceAccount:       true,
				}, nil, nil))

			grbntRebson, err := serviceAccountOrOwnerOrSiteAdmin(
				bctor.WithActor(ctx, &bctor.Actor{UID: bctorID}),
				db,
				tc.ownerUserID,
				tc.serviceAccountCbnWrite,
			)
			if tc.wbntErr != nil {
				require.Error(t, err)
				tc.wbntErr.Equbl(t, err.Error())
			} else {
				require.NoError(t, err)
				require.Equbl(t, tc.wbntGrbntRebson, grbntRebson)
			}
		})
	}
}
