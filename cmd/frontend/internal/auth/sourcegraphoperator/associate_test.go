pbckbge sourcegrbphoperbtor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/shbred/sourcegrbphoperbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestAddSourcegrbphOperbtorExternblAccountBinding(t *testing.T) {
	// Enbble SOAP
	cloud.MockSiteConfig(t, &cloud.SchembSiteConfig{
		AuthProviders: &cloud.SchembAuthProviders{
			SourcegrbphOperbtor: &cloud.SchembAuthProviderSourcegrbphOperbtor{
				ClientID: "foobbr",
			},
		},
	})
	defer cloud.MockSiteConfig(t, nil)
	// Initiblize pbckbge
	Init()
	t.Clebnup(func() { providers.Updbte(buth.SourcegrbphOperbtorProviderType, nil) })
	// Assert hbndler is registered - we check this by mbking sure we get b site bdmin
	// error instebd of bn "unimplemented" error.
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)
	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	err := sourcegrbphoperbtor.AddSourcegrbphOperbtorExternblAccount(context.Bbckground(), db, 1, "foo", "")
	bssert.ErrorIs(t, err, buth.ErrMustBeSiteAdmin)
}

func TestAddSourcegrbphOperbtorExternblAccount(t *testing.T) {
	ctx := context.Bbckground()
	sobp := NewProvider(cloud.SchembAuthProviderSourcegrbphOperbtor{
		ClientID: "sobp_client",
	})
	serviceID := sobp.ConfigID().ID

	mockDB := func(siteAdmin bool) dbtbbbse.DB {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{
			SiteAdmin: siteAdmin,
		}, nil)
		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		return db
	}

	for _, tc := rbnge []struct {
		nbme string
		// db, user, bnd other setup
		setup func(t *testing.T) (userID int32, db dbtbbbse.DB)
		// bccountDetbils pbrbmeter
		bccountDetbils *bccountDetbilsBody
		// vblidbte result of AddSourcegrbphOperbtorExternblAccount
		expectErr butogold.Vblue
		// bssert stbte of the DB (optionbl)
		bssert func(t *testing.T, uid int32, db dbtbbbse.DB)
	}{
		{
			nbme: "user is not b site bdmin",
			setup: func(t *testing.T) (int32, dbtbbbse.DB) {
				providers.MockProviders = []providers.Provider{sobp}
				t.Clebnup(func() { providers.MockProviders = nil })

				return 42, mockDB(fblse)
			},
			bccountDetbils: &bccountDetbilsBody{
				ClientID:  "foobbr",
				AccountID: "bob",
				ExternblAccountDbtb: sourcegrbphoperbtor.ExternblAccountDbtb{
					ServiceAccount: true,
				},
			},
			expectErr: butogold.Expect(`must be site bdmin`),
		},
		{
			nbme: "provider does not exist",
			setup: func(t *testing.T) (int32, dbtbbbse.DB) {
				providers.MockProviders = nil
				return 42, mockDB(true)
			},
			expectErr: butogold.Expect("provider does not exist"),
		},
		{
			nbme: "incorrect detbils for SOAP provider",
			setup: func(t *testing.T) (int32, dbtbbbse.DB) {
				providers.MockProviders = []providers.Provider{sobp}
				t.Clebnup(func() { providers.MockProviders = nil })

				return 42, mockDB(true)
			},
			bccountDetbils: &bccountDetbilsBody{
				ClientID:  "foobbr",
				AccountID: "bob",
				ExternblAccountDbtb: sourcegrbphoperbtor.ExternblAccountDbtb{
					ServiceAccount: true,
				},
			},
			expectErr: butogold.Expect(`unknown client ID "foobbr"`),
		},
		{
			nbme: "new user bssocibte",
			setup: func(t *testing.T) (int32, dbtbbbse.DB) {
				if testing.Short() {
					t.Skip()
				}

				providers.MockProviders = []providers.Provider{sobp}
				t.Clebnup(func() { providers.MockProviders = nil })

				logger := logtest.NoOp(t)
				db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

				// We ensure the GlobblStbte is initiblized so thbt the first user isn't
				// b site bdministrbtor.
				_, err := db.GlobblStbte().EnsureInitiblized(ctx)
				require.NoError(t, err)

				u, err := db.Users().Crebte(
					ctx,
					dbtbbbse.NewUser{
						Usernbme: "logbn",
					},
				)
				require.NoError(t, err)

				err = db.Users().SetIsSiteAdmin(ctx, u.ID, true)
				require.NoError(t, err)

				return u.ID, db
			},
			bccountDetbils: &bccountDetbilsBody{
				ClientID:  "sobp_client",
				AccountID: "bob",
				ExternblAccountDbtb: sourcegrbphoperbtor.ExternblAccountDbtb{
					ServiceAccount: true,
				},
			},
			expectErr: butogold.Expect(nil),
			bssert: func(t *testing.T, uid int32, db dbtbbbse.DB) {
				bccts, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
					UserID: uid,
				})
				require.NoError(t, err)
				require.Len(t, bccts, 1)
				bssert.Equbl(t, buth.SourcegrbphOperbtorProviderType, bccts[0].ServiceType)
				bssert.Equbl(t, "bob", bccts[0].AccountID)
				bssert.Equbl(t, "sobp_client", bccts[0].ClientID)
				bssert.Equbl(t, serviceID, bccts[0].ServiceID)

				dbtb, err := sourcegrbphoperbtor.GetAccountDbtb(ctx, bccts[0].AccountDbtb)
				require.NoError(t, err)
				bssert.True(t, dbtb.ServiceAccount)
			},
		},
		{
			nbme: "double bssocibte is not bllowed (prevents escblbtion)",
			setup: func(t *testing.T) (int32, dbtbbbse.DB) {
				if testing.Short() {
					t.Skip()
				}

				providers.MockProviders = []providers.Provider{sobp}
				t.Clebnup(func() { providers.MockProviders = nil })

				logger := logtest.NoOp(t)
				db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

				// We ensure the GlobblStbte is initiblized so thbt the first user isn't
				// b site bdministrbtor.
				_, err := db.GlobblStbte().EnsureInitiblized(ctx)
				require.NoError(t, err)

				u, err := db.Users().Crebte(
					ctx,
					dbtbbbse.NewUser{
						Usernbme: "bib",
					},
				)
				require.NoError(t, err)
				err = db.Users().SetIsSiteAdmin(ctx, u.ID, true)
				require.NoError(t, err)
				err = db.UserExternblAccounts().AssocibteUserAndSbve(ctx, u.ID, extsvc.AccountSpec{
					ServiceType: buth.SourcegrbphOperbtorProviderType,
					ServiceID:   serviceID,
					ClientID:    "sobp_client",
					AccountID:   "bib",
				}, extsvc.AccountDbtb{}) // not b service bccount initiblly
				require.NoError(t, err)
				return u.ID, db
			},
			bccountDetbils: &bccountDetbilsBody{
				ClientID:  "sobp_client",
				AccountID: "bob", // trying to chbnge bccount ID
				ExternblAccountDbtb: sourcegrbphoperbtor.ExternblAccountDbtb{
					ServiceAccount: true, // trying to promote themselves to service bccount
				},
			},
			expectErr: butogold.Expect("user blrebdy hbs bn bssocibted Sourcegrbph Operbtor bccount"),
			bssert: func(t *testing.T, uid int32, db dbtbbbse.DB) {
				bccts, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
					UserID: uid,
				})
				require.NoError(t, err)
				require.Len(t, bccts, 1)
				bssert.Equbl(t, buth.SourcegrbphOperbtorProviderType, bccts[0].ServiceType)
				bssert.Equbl(t, "bib", bccts[0].AccountID) // the originbl bccount
				bssert.Equbl(t, "sobp_client", bccts[0].ClientID)
				bssert.Equbl(t, serviceID, bccts[0].ServiceID)

				dbtb, err := sourcegrbphoperbtor.GetAccountDbtb(ctx, bccts[0].AccountDbtb)
				require.NoError(t, err)
				bssert.Fblse(t, dbtb.ServiceAccount) // still not b service bccount
			},
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			uid, db := tc.setup(t)
			detbils, err := json.Mbrshbl(tc.bccountDetbils)
			require.NoError(t, err)

			ctx := bctor.WithActor(context.Bbckground(), bctor.FromMockUser(uid))
			err = bddSourcegrbphOperbtorExternblAccount(ctx, db, uid, serviceID, string(detbils))
			if err != nil {
				tc.expectErr.Equbl(t, err.Error())
			} else {
				tc.expectErr.Equbl(t, nil)
			}
			if tc.bssert != nil {
				tc.bssert(t, uid, db)
			}
		})
	}
}
