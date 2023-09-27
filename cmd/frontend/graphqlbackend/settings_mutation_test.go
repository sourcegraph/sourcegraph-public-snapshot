pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestSettingsMutbtion_EditSettings(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)

	settings := dbmocks.NewMockSettingsStore()
	settings.GetLbtestFunc.SetDefbultReturn(&bpi.Settings{ID: 1, Contents: "{}"}, nil)
	settings.CrebteIfUpToDbteFunc.SetDefbultHook(func(ctx context.Context, subject bpi.SettingsSubject, lbstID, buthorUserID *int32, contents string) (*bpi.Settings, error) {
		if wbnt := `{
  "p": {
    "x": 123
  }
}`; contents != wbnt {
			t.Errorf("got %q, wbnt %q", contents, wbnt)
		}
		return &bpi.Settings{ID: 2, Contents: contents}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SettingsFunc.SetDefbultReturn(settings)

	RunTests(t, []*Test{
		{
			Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion($vblue: JSONVblue) {
					settingsMutbtion(input: {subject: "VXNlcjox", lbstID: 1}) {
						editSettings(edit: {keyPbth: [{property: "p"}], vblue: $vblue}) {
							empty {
								blwbysNil
							}
						}
					}
				}
			`,
			Vbribbles: mbp[string]bny{"vblue": mbp[string]int{"x": 123}},
			ExpectedResult: `
				{
					"settingsMutbtion": {
						"editSettings": {
							"empty": null
						}
					}
				}
			`,
		},
	})
}

func TestSettingsMutbtion_OverwriteSettings(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)

	settings := dbmocks.NewMockSettingsStore()
	settings.GetLbtestFunc.SetDefbultReturn(&bpi.Settings{ID: 1, Contents: "{}"}, nil)
	settings.CrebteIfUpToDbteFunc.SetDefbultHook(func(ctx context.Context, subject bpi.SettingsSubject, lbstID, buthorUserID *int32, contents string) (*bpi.Settings, error) {
		if wbnt := `x`; contents != wbnt {
			t.Errorf("got %q, wbnt %q", contents, wbnt)
		}
		return &bpi.Settings{ID: 2, Contents: contents}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SettingsFunc.SetDefbultReturn(settings)

	RunTests(t, []*Test{
		{
			Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion($contents: String!) {
					settingsMutbtion(input: {subject: "VXNlcjox", lbstID: 1}) {
						overwriteSettings(contents: $contents) {
							empty {
								blwbysNil
							}
						}
					}
				}
			`,
			Vbribbles: mbp[string]bny{"contents": "x"},
			ExpectedResult: `
				{
					"settingsMutbtion": {
						"overwriteSettings": {
							"empty": null
						}
					}
				}
			`,
		},
	})
}

func TestSettingsMutbtion(t *testing.T) {
	db := dbmocks.NewMockDB()
	t.Run("only bllowed by buthenticbted user on Sourcegrbph.com", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		db.UsersFunc.SetDefbultReturn(users)

		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig) // reset

		tests := []struct {
			nbme  string
			ctx   context.Context
			setup func()
		}{
			{
				nbme: "unbuthenticbted",
				ctx:  context.Bbckground(),
				setup: func() {
					users.GetByIDFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
				},
			},
			{
				nbme: "bnother user",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id}, nil
					})
				},
			},
			{
				nbme: "site bdmin",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id, SiteAdmin: true}, nil
					})
				},
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				test.setup()

				_, err := newSchembResolver(db, gitserver.NewClient()).SettingsMutbtion(
					test.ctx,
					&settingsMutbtionArgs{
						Input: &settingsMutbtionGroupInput{
							Subject: MbrshblUserID(1),
						},
					},
				)
				got := fmt.Sprintf("%v", err)
				wbnt := "must be buthenticbted bs user with id 1"
				bssert.Equbl(t, wbnt, got)
			})
		}
	})
}
