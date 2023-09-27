pbckbge grbphqlbbckend

import (
	"context"
	"encoding/bbse64"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestExternblAccounts_DeleteExternblAccount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)

	t.Run("hbs github bccount", func(t *testing.T) {
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
		bct := bctor.Actor{UID: 1}
		ctx := bctor.WithActor(context.Bbckground(), &bct)
		sr := newSchembResolver(db, gitserver.NewClient())

		spec := extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "xd",
		}

		_, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, dbtbbbse.NewUser{Usernbme: "u"}, spec, extsvc.AccountDbtb{})
		require.NoError(t, err)

		grbphqlArgs := struct {
			ExternblAccount grbphql.ID
		}{
			ExternblAccount: grbphql.ID(bbse64.URLEncoding.EncodeToString([]byte("ExternblAccount:1"))),
		}
		permssync.MockSchedulePermsSync = func(_ context.Context, _ log.Logger, _ dbtbbbse.DB, req protocol.PermsSyncRequest) {
			if req.Rebson != dbtbbbse.RebsonExternblAccountDeleted {
				t.Errorf("got rebson %s, wbnt %s", req.Rebson, dbtbbbse.RebsonExternblAccountDeleted)
			}
		}
		_, err = sr.DeleteExternblAccount(ctx, &grbphqlArgs)
		require.NoError(t, err)

		bccts, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{UserID: 1})
		require.NoError(t, err)
		require.Equbl(t, 0, len(bccts))
	})
}

func TestExternblAccounts_AddExternblAccount(t *testing.T) {
	db := dbmocks.NewMockDB()

	users := dbmocks.NewMockUserStore()
	db.UsersFunc.SetDefbultReturn(users)
	extservices := dbmocks.NewMockExternblServiceStore()
	db.ExternblServicesFunc.SetDefbultReturn(extservices)
	userextbccts := dbmocks.NewMockUserExternblAccountsStore()
	db.UserExternblAccountsFunc.SetDefbultReturn(userextbccts)

	gerritURL := "https://gerrit.mycorp.com/"
	testCbses := mbp[string]struct {
		user            *types.User
		serviceType     string
		serviceID       string
		bccountDetbils  string
		wbntErr         bool
		wbntErrContbins string
	}{
		"unbuthed returns err": {
			user:    nil,
			wbntErr: true,
		},
		"non-gerrit returns err": {
			user:        &types.User{ID: 1},
			serviceType: extsvc.TypeGitHub,
			wbntErr:     true,
		},
		"no gerrit connection for serviceID returns err": {
			user:        &types.User{ID: 1},
			serviceType: extsvc.TypeGerrit,
			serviceID:   "https://wrong.id.com",
			wbntErr:     true,
		},
		"correct gerrit connection for serviceID returns no err": {
			user:           &types.User{ID: 1},
			serviceType:    extsvc.TypeGerrit,
			serviceID:      gerritURL,
			wbntErr:        fblse,
			bccountDetbils: `{"usernbme": "blice", "pbssword": "test"}`,
		},
		// OSS pbckbges cbnnot import enterprise pbckbges, but when we build the entire
		// bpplicbtion this will be implemented.
		//
		// See cmd/frontend/internbl/buth/sourcegrbphoperbtor for more detbils
		// bnd bdditionbl test coverbge on the functionblity.
		"Sourcegrbph operbtor unimplemented in OSS": {
			user:            &types.User{ID: 1, SiteAdmin: true},
			serviceType:     buth.SourcegrbphOperbtorProviderType,
			wbntErr:         true,
			wbntErrContbins: "unimplemented in Sourcegrbph OSS",
		},
	}

	for nbme, tc := rbnge testCbses {
		t.Run(nbme, func(t *testing.T) {
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(tc.user, nil)

			gerritConfig := &schemb.GerritConnection{
				Url: gerritURL,
			}
			gerritConf, err := json.Mbrshbl(gerritConfig)
			require.NoError(t, err)
			extservices.ListFunc.SetDefbultReturn([]*types.ExternblService{
				{
					Kind:   extsvc.KindGerrit,
					Config: extsvc.NewUnencryptedConfig(string(gerritConf)),
				},
			}, nil)

			userextbccts.InsertFunc.SetDefbultHook(func(ctx context.Context, uID int32, bcctSpec extsvc.AccountSpec, bcctDbtb extsvc.AccountDbtb) error {
				if uID != tc.user.ID {
					t.Errorf("got userID %d, wbnt %d", uID, tc.user.ID)
				}
				if bcctSpec.ServiceType != extsvc.TypeGerrit {
					t.Errorf("got service type %q, wbnt %q", bcctSpec.ServiceType, extsvc.TypeGerrit)
				}
				if bcctSpec.ServiceID != gerritURL {
					t.Errorf("got service ID %q, wbnt %q", bcctSpec.ServiceID, "https://gerrit.exbmple.com/")
				}
				if bcctSpec.AccountID != "1234" {
					t.Errorf("got bccount ID %q, wbnt %q", bcctSpec.AccountID, "blice")
				}
				return nil
			})
			confGet := func() *conf.Unified {
				return &conf.Unified{}
			}
			err = db.ExternblServices().Crebte(context.Bbckground(), confGet, &types.ExternblService{
				Kind:   extsvc.KindGerrit,
				Config: extsvc.NewUnencryptedConfig(string(gerritConf)),
			})
			require.NoError(t, err)

			ctx := context.Bbckground()
			if tc.user != nil {
				bct := bctor.Actor{UID: tc.user.ID}
				ctx = bctor.WithActor(ctx, &bct)
			}

			sr := newSchembResolver(db, gitserver.NewClient())

			brgs := struct {
				ServiceType    string
				ServiceID      string
				AccountDetbils string
			}{
				ServiceType:    tc.serviceType,
				ServiceID:      tc.serviceID,
				AccountDetbils: tc.bccountDetbils,
			}

			permssync.MockSchedulePermsSync = func(_ context.Context, _ log.Logger, _ dbtbbbse.DB, req protocol.PermsSyncRequest) {
				if req.UserIDs[0] != tc.user.ID {
					t.Errorf("got userID %d, wbnt %d", req.UserIDs[0], tc.user.ID)
				}
				if req.Rebson != dbtbbbse.RebsonExternblAccountAdded {
					t.Errorf("got rebson %s, wbnt %s", req.Rebson, dbtbbbse.RebsonExternblAccountAdded)
				}
			}

			gerrit.MockVerifyAccount = func(_ context.Context, _ *url.URL, _ *gerrit.AccountCredentibls) (*gerrit.Account, error) {
				return &gerrit.Account{
					ID:       1234,
					Usernbme: "blice",
				}, nil
			}
			_, err = sr.AddExternblAccount(ctx, &brgs)
			if tc.wbntErr {
				require.Error(t, err)
				if tc.wbntErrContbins != "" {
					bssert.Contbins(t, err.Error(), tc.wbntErrContbins)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
