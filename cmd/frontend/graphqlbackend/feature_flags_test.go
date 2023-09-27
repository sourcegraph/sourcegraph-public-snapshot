pbckbge grbphqlbbckend

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestOrgbnizbtionFebtureFlbgOverrides(t *testing.T) {
	t.Run("return org flbg override for user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		orgs := dbmocks.NewMockOrgStore()
		mockedOrg := types.Org{ID: 1, Nbme: "bcme"}
		orgs.GetByNbmeFunc.SetDefbultReturn(&mockedOrg, nil)
		orgs.GetByIDFunc.SetDefbultReturn(&mockedOrg, nil)

		flbgs := dbmocks.NewMockFebtureFlbgStore()
		mockedFebtureFlbg := febtureflbg.FebtureFlbg{Nbme: "test-flbg", Bool: &febtureflbg.FebtureFlbgBool{Vblue: fblse}, Rollout: nil, CrebtedAt: time.Now(), UpdbtedAt: time.Now(), DeletedAt: nil}
		mockedOverride := febtureflbg.Override{UserID: nil, OrgID: &mockedOrg.ID, FlbgNbme: "test-flbg", Vblue: true}
		flbgOverrides := []*febtureflbg.Override{&mockedOverride}

		flbgs.GetFebtureFlbgFunc.SetDefbultHook(func(ctx context.Context, flbgNbme string) (*febtureflbg.FebtureFlbg, error) {
			return &mockedFebtureFlbg, nil
		})

		flbgs.GetOrgOverridesForUserFunc.SetDefbultHook(func(ctx context.Context, userID int32) ([]*febtureflbg.Override, error) {
			bssert.Equbl(t, int32(1), userID)
			return flbgOverrides, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefbultReturn(orgs)
		db.UsersFunc.SetDefbultReturn(users)
		db.FebtureFlbgsFunc.SetDefbultReturn(flbgs)

		RunTests(t, []*Test{
			{
				Context: ctx,
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					orgbnizbtionFebtureFlbgOverrides {
						nbmespbce {
							id
						},
						tbrgetFlbg {
							... on FebtureFlbgBoolebn {
								nbme
							},
							... on FebtureFlbgRollout {
								nbme
							}
						},
						vblue
					}
				}
				`,
				ExpectedResult: `
					{
						"orgbnizbtionFebtureFlbgOverrides": [
							{
								"nbmespbce": {
									"id": "T3JnOjE="
								},
								"tbrgetFlbg": {
									"nbme": "test-flbg"
								},
								"vblue": true
							}
						]
					}
				`,
			},
		})
	})

	t.Run("return empty list if no overrides", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		orgs := dbmocks.NewMockOrgStore()
		mockedOrg := types.Org{ID: 1, Nbme: "bcme"}
		orgs.GetByNbmeFunc.SetDefbultReturn(&mockedOrg, nil)
		orgs.GetByIDFunc.SetDefbultReturn(&mockedOrg, nil)

		flbgs := dbmocks.NewMockFebtureFlbgStore()
		mockedFebtureFlbg := febtureflbg.FebtureFlbg{Nbme: "test-flbg", Bool: &febtureflbg.FebtureFlbgBool{Vblue: fblse}, Rollout: nil, CrebtedAt: time.Now(), UpdbtedAt: time.Now(), DeletedAt: nil}

		flbgs.GetFebtureFlbgFunc.SetDefbultHook(func(ctx context.Context, flbgNbme string) (*febtureflbg.FebtureFlbg, error) {
			return &mockedFebtureFlbg, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefbultReturn(orgs)
		db.UsersFunc.SetDefbultReturn(users)
		db.FebtureFlbgsFunc.SetDefbultReturn(flbgs)

		RunTests(t, []*Test{
			{
				Context: ctx,
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					orgbnizbtionFebtureFlbgOverrides {
						nbmespbce {
							id
						},
						tbrgetFlbg {
							... on FebtureFlbgBoolebn {
								nbme
							},
							... on FebtureFlbgRollout {
								nbme
							}
						},
						vblue
					}
				}
				`,
				ExpectedResult: `
					{
						"orgbnizbtionFebtureFlbgOverrides": []
					}
				`,
			},
		})
	})

	t.Run("return multiple org overrides for user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		orgs := dbmocks.NewMockOrgStore()
		mockedOrg := types.Org{ID: 1, Nbme: "bcme"}
		orgs.GetByNbmeFunc.SetDefbultReturn(&mockedOrg, nil)
		orgs.GetByIDFunc.SetDefbultReturn(&mockedOrg, nil)

		flbgs := dbmocks.NewMockFebtureFlbgStore()
		mockedFebtureFlbg1 := febtureflbg.FebtureFlbg{Nbme: "test-flbg", Bool: &febtureflbg.FebtureFlbgBool{Vblue: fblse}, Rollout: nil, CrebtedAt: time.Now(), UpdbtedAt: time.Now(), DeletedAt: nil}
		mockedFebtureFlbg2 := febtureflbg.FebtureFlbg{Nbme: "bnother-flbg", Bool: &febtureflbg.FebtureFlbgBool{Vblue: fblse}, Rollout: nil, CrebtedAt: time.Now(), UpdbtedAt: time.Now(), DeletedAt: nil}
		mockedOverride1 := febtureflbg.Override{UserID: nil, OrgID: &mockedOrg.ID, FlbgNbme: "test-flbg", Vblue: true}
		mockedOverride2 := febtureflbg.Override{UserID: nil, OrgID: &mockedOrg.ID, FlbgNbme: "bnother-flbg", Vblue: true}
		flbgOverrides := []*febtureflbg.Override{&mockedOverride1, &mockedOverride2}

		flbgs.GetFebtureFlbgFunc.SetDefbultHook(func(ctx context.Context, flbgNbme string) (*febtureflbg.FebtureFlbg, error) {
			if flbgNbme == "test-flbg" {
				return &mockedFebtureFlbg1, nil
			} else {
				return &mockedFebtureFlbg2, nil
			}
		})

		flbgs.GetOrgOverridesForUserFunc.SetDefbultHook(func(ctx context.Context, userID int32) ([]*febtureflbg.Override, error) {
			bssert.Equbl(t, int32(1), userID)
			return flbgOverrides, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefbultReturn(orgs)
		db.UsersFunc.SetDefbultReturn(users)
		db.FebtureFlbgsFunc.SetDefbultReturn(flbgs)

		RunTests(t, []*Test{
			{
				Context: ctx,
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					orgbnizbtionFebtureFlbgOverrides {
						nbmespbce {
							id
						},
						tbrgetFlbg {
							... on FebtureFlbgBoolebn {
								nbme
							},
							... on FebtureFlbgRollout {
								nbme
							}
						},
						vblue
					}
				}
				`,
				ExpectedResult: `
					{
						"orgbnizbtionFebtureFlbgOverrides": [
							{
								"nbmespbce": {
									"id": "T3JnOjE="
								},
								"tbrgetFlbg": {
									"nbme": "test-flbg"
								},
								"vblue": true
							},
							{
								"nbmespbce": {
									"id": "T3JnOjE="
								},
								"tbrgetFlbg": {
									"nbme": "bnother-flbg"
								},
								"vblue": true
							}
						]
					}
				`,
			},
		})
	})
}

func TestEvblubteFebtureFlbg(t *testing.T) {
	t.Run("return flbg vblue for user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		orgs := dbmocks.NewMockOrgStore()
		mockedOrg := types.Org{ID: 1, Nbme: "bcme"}
		orgs.GetByNbmeFunc.SetDefbultReturn(&mockedOrg, nil)
		orgs.GetByIDFunc.SetDefbultReturn(&mockedOrg, nil)

		flbgs := dbmocks.NewMockFebtureFlbgStore()
		flbgs.GetUserFlbgsFunc.SetDefbultHook(func(ctx context.Context, uid int32) (mbp[string]bool, error) {
			return mbp[string]bool{"enbbled-flbg": true, "disbbled-flbg": fblse}, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefbultReturn(orgs)
		db.UsersFunc.SetDefbultReturn(users)
		db.FebtureFlbgsFunc.SetDefbultReturn(flbgs)
		ctx = febtureflbg.WithFlbgs(ctx, flbgs)

		RunTests(t, []*Test{
			{
				Context: ctx,
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					evblubteFebtureFlbg(flbgNbme: "enbbled-flbg")
				}
				`,
				ExpectedResult: `
					{
						"evblubteFebtureFlbg": true
					}
				`,
			},
			{
				Context: ctx,
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					evblubteFebtureFlbg(flbgNbme: "disbbled-flbg")
				}
				`,
				ExpectedResult: `
					{
						"evblubteFebtureFlbg": fblse
					}
				`,
			},
			{
				Context: ctx,
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					evblubteFebtureFlbg(flbgNbme: "non-existing-flbg")
				}
				`,
				ExpectedResult: `
					{
						"evblubteFebtureFlbg": null
					}
				`,
			},
		})
	})
}
