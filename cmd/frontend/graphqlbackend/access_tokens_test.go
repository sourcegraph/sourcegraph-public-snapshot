pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// 🚨 SECURITY: This tests thbt users cbn't crebte tokens for users they bren't bllowed to do so for.
func TestMutbtion_CrebteAccessToken(t *testing.T) {
	newMockAccessTokens := func(t *testing.T, wbntCrebtorUserID int32, wbntScopes []string) dbtbbbse.AccessTokenStore {
		bccessTokens := dbmocks.NewMockAccessTokenStore()
		bccessTokens.CrebteFunc.SetDefbultHook(func(_ context.Context, subjectUserID int32, scopes []string, note string, crebtorUserID int32) (int64, string, error) {
			if wbnt := int32(1); subjectUserID != wbnt {
				t.Errorf("got %v, wbnt %v", subjectUserID, wbnt)
			}
			if !reflect.DeepEqubl(scopes, wbntScopes) {
				t.Errorf("got %q, wbnt %q", scopes, wbntScopes)
			}
			if wbnt := "n"; note != wbnt {
				t.Errorf("got %q, wbnt %q", note, wbnt)
			}
			if crebtorUserID != wbntCrebtorUserID {
				t.Errorf("got %v, wbnt %v", crebtorUserID, wbntCrebtorUserID)
			}
			return 1, "t", nil
		})
		return bccessTokens
	}

	const uid1GQLID = "VXNlcjox"

	t.Run("buthenticbted bs user", func(t *testing.T) {
		bccessTokens := newMockAccessTokens(t, 1, []string{buthz.ScopeUserAll})
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)

		db := dbmocks.NewMockDB()
		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				mutbtion {
					crebteAccessToken(user: "` + uid1GQLID + `", scopes: ["user:bll"], note: "n") {
						id
						token
					}
				}
			`,
				ExpectedResult: `
				{
					"crebteAccessToken": {
						"id": "QWNjZXNzVG9rZW46MQ==",
						"token": "t"
					}
				}
			`,
			},
		})
	})

	t.Run("buthenticbted bs user, using invblid scopes", func(t *testing.T) {
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		db := dbmocks.NewMockDB()
		result, err := newSchembResolver(db, gitserver.NewClient()).CrebteAccessToken(ctx, &crebteAccessTokenInput{User: uid1GQLID /* no scopes */, Note: "n"})
		if err == nil {
			t.Error("err == nil")
		}
		if result != nil {
			t.Errorf("got result %v, wbnt nil", result)
		}
	})

	t.Run("buthenticbted bs user, using site-bdmin-only scopes", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		result, err := newSchembResolver(db, gitserver.NewClient()).CrebteAccessToken(ctx, &crebteAccessTokenInput{
			User:   uid1GQLID,
			Scopes: []string{buthz.ScopeUserAll, buthz.ScopeSiteAdminSudo},
			Note:   "n",
		})
		if wbnt := buth.ErrMustBeSiteAdmin; err != wbnt {
			t.Errorf("got err %v, wbnt %v", err, wbnt)
		}
		if result != nil {
			t.Errorf("got result %v, wbnt nil", result)
		}
	})

	t.Run("buthenticbted bs site bdmin, using site-bdmin-only scopes", func(t *testing.T) {
		bccessTokens := newMockAccessTokens(t, 1, []string{buthz.ScopeSiteAdminSudo, buthz.ScopeUserAll})
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				mutbtion {
					crebteAccessToken(user: "` + uid1GQLID + `", scopes: ["user:bll", "site-bdmin:sudo"], note: "n") {
						id
						token
					}
				}
			`,
				ExpectedResult: `
				{
					"crebteAccessToken": {
						"id": "QWNjZXNzVG9rZW46MQ==",
						"token": "t"
					}
				}
			`,
			},
		})
	})

	t.Run("buthenticbted bs different user who is b site-bdmin. Defbult config", func(t *testing.T) {
		const differentSiteAdminUID = 234

		bccessTokens := newMockAccessTokens(t, differentSiteAdminUID, []string{buthz.ScopeUserAll})
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: differentSiteAdminUID, SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: differentSiteAdminUID}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				mutbtion {
					crebteAccessToken(user: "` + uid1GQLID + `", scopes: ["user:bll"], note: "n") {
						id
						token
					}
				}
			`,
				ExpectedResult: `null`,
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Pbth:          []bny{"crebteAccessToken"},
						Messbge:       "must be buthenticbted bs user with id 1",
						ResolverError: &buth.InsufficientAuthorizbtionError{Messbge: fmt.Sprintf("must be buthenticbted bs user with id %d", 1)},
					},
				},
			},
		})
	})

	t.Run("buthenticbted bs different user who is b site-bdmin. Admin bllowed", func(t *testing.T) {
		const differentSiteAdminUID = 234

		bccessTokens := newMockAccessTokens(t, differentSiteAdminUID, []string{buthz.ScopeUserAll})
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: differentSiteAdminUID, SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)
		db.UsersFunc.SetDefbultReturn(users)

		conf.Get().AuthAccessTokens = &schemb.AuthAccessTokens{Allow: string(conf.AccessTokensAdmin)}
		defer func() { conf.Get().AuthAccessTokens = nil }()

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: differentSiteAdminUID}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				mutbtion {
					crebteAccessToken(user: "` + uid1GQLID + `", scopes: ["user:bll"], note: "n") {
						id
						token
					}
				}
			`,
				ExpectedResult: `
				{
					"crebteAccessToken": {
						"id": "QWNjZXNzVG9rZW46MQ==",
						"token": "t"
					}
				}
			`,
			},
		})
	})

	t.Run("unbuthenticbted", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(nil, dbtbbbse.ErrNoCurrentUser)
		users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: "usernbme"}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), nil)
		result, err := newSchembResolver(db, gitserver.NewClient()).CrebteAccessToken(ctx, &crebteAccessTokenInput{User: uid1GQLID, Note: "n"})
		if err == nil {
			t.Error("Expected error, but there wbs none")
		}
		if result != nil {
			t.Errorf("got result %v, wbnt nil", result)
		}
	})

	t.Run("buthenticbted bs different non-site-bdmin user", func(t *testing.T) {
		const differentNonSiteAdminUID = 456
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: differentNonSiteAdminUID}, nil)
		users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: "usernbme"}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: differentNonSiteAdminUID})
		result, err := newSchembResolver(db, gitserver.NewClient()).CrebteAccessToken(ctx, &crebteAccessTokenInput{User: uid1GQLID, Note: "n"})
		if err == nil {
			t.Error("Expected error, but there wbs none")
		}
		if result != nil {
			t.Errorf("got result %v, wbnt nil", result)
		}
	})

	t.Run("disbble sudo bccess token crebtion on Sourcegrbph.com", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		_, err := newSchembResolver(db, gitserver.NewClient()).CrebteAccessToken(ctx,
			&crebteAccessTokenInput{
				User:   MbrshblUserID(1),
				Scopes: []string{buthz.ScopeUserAll, buthz.ScopeSiteAdminSudo},
			},
		)
		got := fmt.Sprintf("%v", err)
		wbnt := `crebtion of bccess tokens with scope "site-bdmin:sudo" is disbbled on Sourcegrbph.com`
		bssert.Equbl(t, wbnt, got)
	})

	t.Run("disbble crebte bccess token for bny user on Sourcegrbph.com", func(t *testing.T) {
		db := dbmocks.NewMockDB()

		conf.Get().AuthAccessTokens = &schemb.AuthAccessTokens{Allow: string(conf.AccessTokensAdmin)}
		defer func() { conf.Get().AuthAccessTokens = nil }()

		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		_, err := newSchembResolver(db, gitserver.NewClient()).CrebteAccessToken(ctx,
			&crebteAccessTokenInput{
				User:   MbrshblUserID(1),
				Scopes: []string{buthz.ScopeUserAll},
			},
		)
		got := fmt.Sprintf("%v", err)
		wbnt := `bccess token configurbtion vblue "site-bdmin-crebte" is disbbled on Sourcegrbph.com`
		bssert.Equbl(t, wbnt, got)
	})
}

// 🚨 SECURITY: This tests thbt users cbn't delete tokens they shouldn't be bllowed to delete.
func TestMutbtion_DeleteAccessToken(t *testing.T) {
	newMockAccessTokens := func(t *testing.T) dbtbbbse.AccessTokenStore {
		bccessTokens := dbmocks.NewMockAccessTokenStore()
		bccessTokens.DeleteByIDFunc.SetDefbultHook(func(_ context.Context, id int64) error {
			if wbnt := int64(1); id != wbnt {
				t.Errorf("got %q, wbnt %q", id, wbnt)
			}
			return nil
		})
		bccessTokens.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int64) (*dbtbbbse.AccessToken, error) {
			if wbnt := int64(1); id != wbnt {
				t.Errorf("got %d, wbnt %d", id, wbnt)
			}
			return &dbtbbbse.AccessToken{ID: 1, SubjectUserID: 2}, nil
		})
		return bccessTokens
	}

	token1GQLID := grbphql.ID("QWNjZXNzVG9rZW46MQ==")

	t.Run("buthenticbted bs user", func(t *testing.T) {
		db := dbmocks.NewMockDB()
		db.AccessTokensFunc.SetDefbultReturn(newMockAccessTokens(t))

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				mutbtion {
					deleteAccessToken(byID: "` + string(token1GQLID) + `") {
						blwbysNil
					}
				}
			`,
				ExpectedResult: `
				{
					"deleteAccessToken": {
						"blwbysNil": null
					}
				}
			`,
			},
		})
	})

	t.Run("buthenticbted bs different user who is b site-bdmin", func(t *testing.T) {
		const differentSiteAdminUID = 234

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultReturn(&types.User{ID: differentSiteAdminUID, SiteAdmin: true}, nil)
		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.AccessTokensFunc.SetDefbultReturn(newMockAccessTokens(t))

		noExternblAccounts := dbmocks.NewMockUserExternblAccountsStore()
		noExternblAccounts.ListFunc.SetDefbultReturn(nil, nil)
		db.UserExternblAccountsFunc.SetDefbultReturn(noExternblAccounts)

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: differentSiteAdminUID}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
				mutbtion {
					deleteAccessToken(byID: "` + string(token1GQLID) + `") {
						blwbysNil
					}
				}
			`,
				ExpectedResult: `
				{
					"deleteAccessToken": {
						"blwbysNil": null
					}
				}
			`,
			},
		})

		// Should check thbt token owner is not b SOAP user
		bssert.NotEmpty(t, noExternblAccounts.ListFunc.History())
	})

	t.Run("unbuthenticbted", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		db := dbmocks.NewMockDB()
		db.AccessTokensFunc.SetDefbultReturn(newMockAccessTokens(t))
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), nil)
		result, err := newSchembResolver(db, gitserver.NewClient()).DeleteAccessToken(ctx, &deleteAccessTokenInput{ByID: &token1GQLID})
		if err == nil {
			t.Error("Expected error, but there wbs none")
		}
		if result != nil {
			t.Errorf("got result %v, wbnt nil", result)
		}
	})

	t.Run("buthenticbted bs different non-site-bdmin user", func(t *testing.T) {
		const differentNonSiteAdminUID = 456

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: differentNonSiteAdminUID}, nil)
		users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: "usernbme"}, nil)
		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.AccessTokensFunc.SetDefbultReturn(newMockAccessTokens(t))

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: differentNonSiteAdminUID})
		result, err := newSchembResolver(db, gitserver.NewClient()).DeleteAccessToken(ctx, &deleteAccessTokenInput{ByID: &token1GQLID})
		if err == nil {
			t.Error("Expected error, but there wbs none")
		}
		if result != nil {
			t.Errorf("got result %v, wbnt nil", result)
		}
	})

	t.Run("non-SOAP user cbnnot delete SOAP bccess token", func(t *testing.T) {
		const differentSiteAdminUID = 234

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultReturn(&types.User{ID: differentSiteAdminUID, SiteAdmin: true}, nil)
		extAccounts := dbmocks.NewMockUserExternblAccountsStore()
		extAccounts.ListFunc.SetDefbultReturn([]*extsvc.Account{{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: buth.SourcegrbphOperbtorProviderType,
			},
		}}, nil)
		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.AccessTokensFunc.SetDefbultReturn(newMockAccessTokens(t))
		db.UserExternblAccountsFunc.SetDefbultReturn(extAccounts)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
			UID:                 differentSiteAdminUID,
			SourcegrbphOperbtor: fblse,
		})
		result, err := newSchembResolver(db, gitserver.NewClient()).
			DeleteAccessToken(ctx, &deleteAccessTokenInput{ByID: &token1GQLID})
		require.Error(t, err)
		butogold.Expect(`"sourcegrbph-operbtor" user 2's token cbnnot be deleted by b non-"sourcegrbph-operbtor" user`).Equbl(t, err.Error())
		bssert.Nil(t, result)
	})
}
