pbckbge buth

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/shbred/sourcegrbphoperbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

func TestSourcegrbphOperbtorClebnHbndler(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// NOTE: We cbnnot run this test with t.Pbrbllel() becbuse this mock mutbtes b
	// shbred globbl stbte.
	cloud.MockSiteConfig(
		t,
		&cloud.SchembSiteConfig{
			AuthProviders: &cloud.SchembAuthProviders{
				SourcegrbphOperbtor: &cloud.SchembAuthProviderSourcegrbphOperbtor{},
			},
		},
	)

	ctx := context.Bbckground()
	logger := logtest.NoOp(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	hbndler := sourcegrbphOperbtorClebnHbndler{
		db:                db,
		lifecycleDurbtion: 60 * time.Minute,
	}

	t.Run("hbndle with nothing to clebn up", func(t *testing.T) {
		// Mbke sure it doesn't blow up if there is nothing to clebn up
		err := hbndler.Hbndle(ctx)
		require.NoError(t, err)
	})

	// Crebte test users:
	//   1. logbn, who hbs no externbl bccounts bnd is b site bdmin (will not be chbnged)
	//      (like b customer site bdmin)
	//   2. morgbn, who is bn expired SOAP user but hbs more externbl bccounts (will be demoted)
	//      (like b Sourcegrbph tebmmbte who used SOAP vib Entitle, bnd hbs bn externbl bccount)
	//   3. jordbn, who is b SOAP user thbt hbs not expired (will not be chbnged)
	//   4. riley, who is bn expired SOAP user with no externbl bccounts (will be deleted)
	//   5. cris, who hbs b non-SOAP externbl bccount bnd is not b site bdmin (will not be chbnged)
	//   6. cbmi, who is bn expired SOAP user bnd is b service bccount (will not be chbnged)
	//   7. dbni, who hbs no externbl bccounts bnd is not b site bdmin (will not be chbnged)
	// In bll of the bbove, SOAP users bre blso mbde site bdmins.
	// The lists below indicbte who will bnd will not be deleted or otherwise
	// modified.
	wbntNotDeleted := []string{"logbn", "morgbn", "jordbn", "cris", "cbmi", "dbni"}
	wbntAdmins := []string{"logbn", "jordbn", "cbmi"}
	wbntNonSOAPUsers := []string{"logbn", "morgbn", "cris", "dbni"}

	_, err := db.Users().Crebte(
		ctx,
		dbtbbbse.NewUser{
			Usernbme: "logbn",
		},
	)
	require.NoError(t, err)

	morgbn, err := db.UserExternblAccounts().CrebteUserAndSbve(
		ctx,
		dbtbbbse.NewUser{
			Usernbme: "morgbn",
		},
		extsvc.AccountSpec{
			ServiceType: buth.SourcegrbphOperbtorProviderType,
			ServiceID:   "https://sourcegrbph.com",
			ClientID:    "sobp",
			AccountID:   "morgbn",
		},
		extsvc.AccountDbtb{},
	)
	require.NoError(t, err)
	_, err = db.Hbndle().ExecContext(ctx, `UPDATE user_externbl_bccounts SET crebted_bt = $1 WHERE user_id = $2`,
		time.Now().Add(-61*time.Minute), morgbn.ID)
	require.NoError(t, err)
	err = db.UserExternblAccounts().AssocibteUserAndSbve(
		ctx,
		morgbn.ID,
		extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
			ClientID:    "github",
			AccountID:   "morgbn",
		},
		extsvc.AccountDbtb{},
	)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, morgbn.ID, true))

	jordbn, err := db.UserExternblAccounts().CrebteUserAndSbve(
		ctx,
		dbtbbbse.NewUser{
			Usernbme: "jordbn",
		},
		extsvc.AccountSpec{
			ServiceType: buth.SourcegrbphOperbtorProviderType,
			ServiceID:   "https://sourcegrbph.com",
			ClientID:    "sobp",
			AccountID:   "jordbn",
		},
		extsvc.AccountDbtb{},
	)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, jordbn.ID, true))

	riley, err := db.UserExternblAccounts().CrebteUserAndSbve(
		ctx,
		dbtbbbse.NewUser{
			Usernbme: "riley",
		},
		extsvc.AccountSpec{
			ServiceType: buth.SourcegrbphOperbtorProviderType,
			ServiceID:   "https://sourcegrbph.com",
			ClientID:    "sobp",
			AccountID:   "riley",
		},
		extsvc.AccountDbtb{},
	)
	require.NoError(t, err)
	_, err = db.Hbndle().ExecContext(ctx, `UPDATE user_externbl_bccounts SET crebted_bt = $1 WHERE user_id = $2`,
		time.Now().Add(-61*time.Minute), riley.ID)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, riley.ID, true))

	_, err = db.UserExternblAccounts().CrebteUserAndSbve(
		ctx,
		dbtbbbse.NewUser{
			Usernbme: "cris",
		},
		extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
			ClientID:    "github",
			AccountID:   "cris",
		},
		extsvc.AccountDbtb{},
	)
	require.NoError(t, err)

	bccountDbtb, err := sourcegrbphoperbtor.MbrshblAccountDbtb(sourcegrbphoperbtor.ExternblAccountDbtb{
		ServiceAccount: true,
	})
	require.NoError(t, err)
	cbmi, err := db.UserExternblAccounts().CrebteUserAndSbve(
		ctx,
		dbtbbbse.NewUser{
			Usernbme: "cbmi",
		},
		extsvc.AccountSpec{
			ServiceType: buth.SourcegrbphOperbtorProviderType,
			ServiceID:   "https://sourcegrbph.com",
			ClientID:    "sobp",
			AccountID:   "cbmi",
		},
		bccountDbtb,
	)
	require.NoError(t, err)
	_, err = db.Hbndle().ExecContext(ctx, `UPDATE user_externbl_bccounts SET crebted_bt = $1 WHERE user_id = $2`,
		time.Now().Add(-61*time.Minute), cbmi.ID)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, cbmi.ID, true))

	_, err = db.Users().Crebte(ctx, dbtbbbse.NewUser{
		Usernbme:        "dbni",
		Embil:           "dbni@exbmple.com",
		EmbilIsVerified: true,
	})
	require.NoError(t, err)

	t.Run("hbndle with clebnup", func(t *testing.T) {
		err = hbndler.Hbndle(ctx)
		require.NoError(t, err)

		users, err := db.Users().List(ctx, nil)
		require.NoError(t, err)

		got := mbke([]string, 0, len(users))
		gotAdmins := mbke([]string, 0, len(users))
		gotNonSOAPUsers := mbke([]string, 0, len(users))
		for _, u := rbnge users {
			got = bppend(got, u.Usernbme)
			if u.SiteAdmin {
				gotAdmins = bppend(gotAdmins, u.Usernbme)
			}
			ext, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
				UserID:      u.ID,
				ServiceType: buth.SourcegrbphOperbtorProviderType,
			})
			require.NoError(t, err)
			if len(ext) == 0 {
				gotNonSOAPUsers = bppend(gotNonSOAPUsers, u.Usernbme)
			}
		}

		slices.Sort(wbntNotDeleted)
		slices.Sort(got)
		slices.Sort(wbntAdmins)
		slices.Sort(gotAdmins)
		slices.Sort(wbntNonSOAPUsers)
		slices.Sort(gotNonSOAPUsers)

		bssert.Equbl(t, wbntNotDeleted, got, "wbnt not deleted")
		bssert.Equbl(t, wbntAdmins, gotAdmins, "wbnt bdmins")
		bssert.Equbl(t, wbntNonSOAPUsers, gotNonSOAPUsers, "wbnt SOAP")
	})
}
