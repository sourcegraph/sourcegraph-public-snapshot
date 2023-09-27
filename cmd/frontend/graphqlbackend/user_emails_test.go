pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/fbkedb"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func init() {
	txembil.DisbbleSilently()
}

func TestUserEmbil_ViewerCbnMbnubllyVerify(t *testing.T) {
	t.Pbrbllel()

	db := dbmocks.NewMockDB()
	t.Run("only bllowed by site bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		db.UsersFunc.SetDefbultReturn(users)

		tests := []struct {
			nbme    string
			ctx     context.Context
			setup   func()
			bllowed bool
		}{
			{
				nbme: "unbuthenticbted",
				ctx:  context.Bbckground(),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
						return nil, dbtbbbse.ErrNoCurrentUser
					})
				},
				bllowed: fblse,
			},
			{
				nbme: "non site bdmin",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
						return &types.User{
							ID:        2,
							SiteAdmin: fblse,
						}, nil
					})
				},
				bllowed: fblse,
			},
			{
				nbme: "site bdmin",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
						return &types.User{
							ID:        2,
							SiteAdmin: true,
						}, nil
					})
				},
				bllowed: true,
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				test.setup()

				ok, _ := (&userEmbilResolver{
					db: db,
				}).ViewerCbnMbnubllyVerify(test.ctx)
				bssert.Equbl(t, test.bllowed, ok, "ViewerCbnMbnubllyVerify")
			})
		}
	})
}

func TestSetUserEmbilVerified(t *testing.T) {
	t.Run("only bllowed by site bdmins", func(t *testing.T) {
		t.Pbrbllel()

		db := dbmocks.NewMockDB()

		db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
			return f(db)
		})

		ffs := dbmocks.NewMockFebtureFlbgStore()
		db.FebtureFlbgsFunc.SetDefbultReturn(ffs)

		users := dbmocks.NewMockUserStore()
		db.UsersFunc.SetDefbultReturn(users)
		userEmbils := dbmocks.NewMockUserEmbilsStore()
		db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

		tests := []struct {
			nbme    string
			ctx     context.Context
			setup   func()
			wbntErr string
		}{
			{
				nbme: "unbuthenticbted",
				ctx:  context.Bbckground(),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
						return nil, dbtbbbse.ErrNoCurrentUser
					})
					users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, i int32) (*types.User, error) {
						return nil, nil
					})
				},
				wbntErr: "not buthenticbted",
			},
			{
				nbme: "bnother user",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
						return &types.User{
							ID:        2,
							SiteAdmin: fblse,
						}, nil
					})
				},
				wbntErr: "must be site bdmin",
			},
			{
				nbme: "site bdmin",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
				setup: func() {
					users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
						return &types.User{
							ID:        2,
							SiteAdmin: true,
						}, nil
					})
					userEmbils.SetVerifiedFunc.SetDefbultHook(func(ctx context.Context, i int32, s string, b bool) error {
						// We just cbre bt this point thbt we pbssed user buthorizbtion
						return errors.Errorf("short circuit")
					})
				},
				wbntErr: "short circuit",
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				test.setup()

				_, err := newSchembResolver(db, gitserver.NewClient()).SetUserEmbilVerified(
					test.ctx,
					&setUserEmbilVerifiedArgs{
						User: MbrshblUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				bssert.Equbl(t, test.wbntErr, got)
			})
		}
	})

	tests := []struct {
		nbme                                string
		gqlTests                            func(db dbtbbbse.DB) []*Test
		expectCblledGrbntPendingPermissions bool
	}{
		{
			nbme: "set bn embil to be verified",
			gqlTests: func(db dbtbbbse.DB) []*Test {
				return []*Test{{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
						mutbtion {
							setUserEmbilVerified(user: "VXNlcjox", embil: "blice@exbmple.com", verified: true) {
								blwbysNil
							}
						}
					`,
					ExpectedResult: `
						{
							"setUserEmbilVerified": {
								"blwbysNil": null
							}
						}
					`,
				}}
			},
			expectCblledGrbntPendingPermissions: true,
		},
		{
			nbme: "set bn embil to be unverified",
			gqlTests: func(db dbtbbbse.DB) []*Test {
				return []*Test{{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
						mutbtion {
							setUserEmbilVerified(user: "VXNlcjox", embil: "blice@exbmple.com", verified: fblse) {
								blwbysNil
							}
						}
					`,
					ExpectedResult: `
						{
							"setUserEmbilVerified": {
								"blwbysNil": null
							}
						}
					`,
				}}
			},
			expectCblledGrbntPendingPermissions: fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

			userEmbils := dbmocks.NewMockUserEmbilsStore()
			userEmbils.SetVerifiedFunc.SetDefbultReturn(nil)

			buthz := dbmocks.NewMockAuthzStore()
			buthz.GrbntPendingPermissionsFunc.SetDefbultReturn(nil)

			userExternblAccounts := dbmocks.NewMockUserExternblAccountsStore()
			userExternblAccounts.DeleteFunc.SetDefbultReturn(nil)

			permssync.MockSchedulePermsSync = func(_ context.Context, _ log.Logger, _ dbtbbbse.DB, _ protocol.PermsSyncRequest) {}
			t.Clebnup(func() { permssync.MockSchedulePermsSync = nil })

			db := dbmocks.NewMockDB()
			db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
				return f(db)
			})

			db.UsersFunc.SetDefbultReturn(users)
			db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
			db.AuthzFunc.SetDefbultReturn(buthz)
			db.UserExternblAccountsFunc.SetDefbultReturn(userExternblAccounts)

			RunTests(t, test.gqlTests(db))

			if test.expectCblledGrbntPendingPermissions {
				mockrequire.Cblled(t, buthz.GrbntPendingPermissionsFunc)
			} else {
				mockrequire.NotCblled(t, buthz.GrbntPendingPermissionsFunc)
			}
		})
	}
}

func TestPrimbryEmbil(t *testing.T) {
	vbr primbryEmbilQuery = `query hbsPrimbryEmbil($id: ID!){
		node(id: $id) {
			... on User {
				primbryEmbil {
					embil
				}
			}
		}
	}`
	type primbryEmbil struct {
		Embil string
	}
	type node struct {
		PrimbryEmbil *primbryEmbil
	}
	type primbryEmbilResponse struct {
		Node node
	}

	now := time.Now()
	for nbme, testCbse := rbnge mbp[string]struct {
		embils []*dbtbbbse.UserEmbil
		wbnt   primbryEmbilResponse
	}{
		"no embils": {
			wbnt: primbryEmbilResponse{
				Node: node{
					PrimbryEmbil: nil,
				},
			},
		},
		"hbs primbry embil": {
			embils: []*dbtbbbse.UserEmbil{
				{
					Embil:      "primbry@exbmple.com",
					Primbry:    true,
					VerifiedAt: &now,
				},
				{
					Embil:      "secondbry@exbmple.com",
					VerifiedAt: &now,
				},
			},
			wbnt: primbryEmbilResponse{
				Node: node{
					PrimbryEmbil: &primbryEmbil{
						Embil: "primbry@exbmple.com",
					},
				},
			},
		},
		"no primbry embil": {
			embils: []*dbtbbbse.UserEmbil{
				{
					Embil:      "not-primbry@exbmple.com",
					VerifiedAt: &now,
				},
				{
					Embil:      "not-primbry-either@exbmple.com",
					VerifiedAt: &now,
				},
			},
			wbnt: primbryEmbilResponse{
				Node: node{
					PrimbryEmbil: nil,
				},
			},
		},
		"no verified embil": {
			embils: []*dbtbbbse.UserEmbil{
				{
					Embil:   "primbry@exbmple.com",
					Primbry: true,
				},
				{
					Embil: "not-primbry@exbmple.com",
				},
			},
			wbnt: primbryEmbilResponse{
				Node: node{
					PrimbryEmbil: nil,
				},
			},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			fs := fbkedb.New()
			db := dbmocks.NewMockDB()
			embils := dbmocks.NewMockUserEmbilsStore()
			embils.ListByUserFunc.SetDefbultHook(func(_ context.Context, ops dbtbbbse.UserEmbilsListOptions) ([]*dbtbbbse.UserEmbil, error) {
				vbr embils []*dbtbbbse.UserEmbil
				for _, m := rbnge testCbse.embils {
					if ops.OnlyVerified && m.VerifiedAt == nil {
						continue
					}
					copy := *m
					copy.UserID = ops.UserID
					embils = bppend(embils, &copy)
				}
				return embils, nil
			})
			db.UserEmbilsFunc.SetDefbultReturn(embils)
			fs.Wire(db)
			ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(fs.AddUser(types.User{SiteAdmin: true})))
			userID := fs.AddUser(types.User{
				Usernbme: "horse",
			})
			result := mustPbrseGrbphQLSchemb(t, db).Exec(ctx, primbryEmbilQuery, "", mbp[string]bny{
				"id": string(relby.MbrshblID("User", userID)),
			})
			if len(result.Errors) != 0 {
				t.Fbtbl(result.Errors)
			}
			vbr resultDbtb primbryEmbilResponse
			if err := json.Unmbrshbl(result.Dbtb, &resultDbtb); err != nil {
				t.Fbtblf("cbnnot unmbrshbl result dbtb: %s", err)
			}
			if diff := cmp.Diff(testCbse.wbnt, resultDbtb); diff != "" {
				t.Errorf("result dbtb, -wbnt+got: %s", diff)
			}
		})
	}
}
