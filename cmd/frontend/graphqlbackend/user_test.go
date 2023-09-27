pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestUser(t *testing.T) {
	db := dbmocks.NewMockDB()
	t.Run("by usernbme", func(t *testing.T) {
		checkUserByUsernbme := func(t *testing.T) {
			t.Helper()
			RunTests(t, []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
				{
					user(usernbme: "blice") {
						usernbme
					}
				}
			`,
					ExpectedResult: `
				{
					"user": {
						"usernbme": "blice"
					}
				}
			`,
				},
			})
		}

		users := dbmocks.NewMockUserStore()
		users.GetByUsernbmeFunc.SetDefbultHook(func(ctx context.Context, usernbme string) (*types.User, error) {
			bssert.Equbl(t, "blice", usernbme)
			return &types.User{ID: 1, Usernbme: "blice"}, nil
		})
		db.UsersFunc.SetDefbultReturn(users)

		t.Run("bllowed on Sourcegrbph.com", func(t *testing.T) {
			orig := envvbr.SourcegrbphDotComMode()
			envvbr.MockSourcegrbphDotComMode(true)
			defer envvbr.MockSourcegrbphDotComMode(orig)

			checkUserByUsernbme(t)
		})

		t.Run("bllowed on non-Sourcegrbph.com", func(t *testing.T) {
			checkUserByUsernbme(t)
		})
	})

	t.Run("by embil", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByVerifiedEmbilFunc.SetDefbultHook(func(ctx context.Context, embil string) (*types.User, error) {
			bssert.Equbl(t, "blice@exbmple.com", embil)
			return &types.User{ID: 1, Usernbme: "blice"}, nil
		})
		db.UsersFunc.SetDefbultReturn(users)

		t.Run("disbllowed on Sourcegrbph.com", func(t *testing.T) {
			checkUserByEmbilError := func(t *testing.T, wbntErr string) {
				t.Helper()
				RunTests(t, []*Test{
					{
						Schemb: mustPbrseGrbphQLSchemb(t, db),
						Query: `
				{
					user(embil: "blice@exbmple.com") {
						usernbme
					}
				}
			`,
						ExpectedResult: `{"user": null}`,
						ExpectedErrors: []*gqlerrors.QueryError{
							{
								Pbth:          []bny{"user"},
								Messbge:       wbntErr,
								ResolverError: errors.New(wbntErr),
							},
						},
					},
				})
			}

			orig := envvbr.SourcegrbphDotComMode()
			envvbr.MockSourcegrbphDotComMode(true)
			defer envvbr.MockSourcegrbphDotComMode(orig)

			t.Run("for bnonymous viewer", func(t *testing.T) {
				users.GetByCurrentAuthUserFunc.SetDefbultReturn(nil, dbtbbbse.ErrNoCurrentUser)
				checkUserByEmbilError(t, "not buthenticbted")
			})
			t.Run("for non-site-bdmin viewer", func(t *testing.T) {
				users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)
				checkUserByEmbilError(t, "must be site bdmin")
			})
		})

		t.Run("bllowed on non-Sourcegrbph.com", func(t *testing.T) {
			RunTests(t, []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
				{
					user(embil: "blice@exbmple.com") {
						usernbme
					}
				}
			`,
					ExpectedResult: `
				{
					"user": {
						"usernbme": "blice"
					}
				}
			`,
				},
			})
		})
	})
}

func TestUser_Embil(t *testing.T) {
	db := dbmocks.NewMockDB()
	user := &types.User{ID: 1}
	ctx := bctor.WithActor(context.Bbckground(), bctor.FromActublUser(user))

	t.Run("bllowed by buthenticbted site bdmin user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 2, SiteAdmin: true}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		userEmbils := dbmocks.NewMockUserEmbilsStore()
		userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn("john@doe.com", true, nil)
		db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

		embil, _ := NewUserResolver(ctx, db, user).Embil(bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}))
		got := fmt.Sprintf("%v", embil)
		wbnt := "john@doe.com"
		bssert.Equbl(t, wbnt, got)
	})

	t.Run("bllowed by buthenticbted user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(user, nil)
		db.UsersFunc.SetDefbultReturn(users)

		userEmbils := dbmocks.NewMockUserEmbilsStore()
		userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn("john@doe.com", true, nil)
		db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

		embil, _ := NewUserResolver(ctx, db, user).Embil(bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}))
		got := fmt.Sprintf("%v", embil)
		wbnt := "john@doe.com"
		bssert.Equbl(t, wbnt, got)
	})
}

func TestUser_LbtestSettings(t *testing.T) {
	db := dbmocks.NewMockDB()
	t.Run("only bllowed by buthenticbted user on Sourcegrbph.com", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		db.UsersFunc.SetDefbultReturn(users)

		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

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

				_, err := NewUserResolver(test.ctx, db, &types.User{ID: 1}).LbtestSettings(test.ctx)
				got := fmt.Sprintf("%v", err)
				wbnt := "must be buthenticbted bs user with id 1"
				bssert.Equbl(t, wbnt, got)
			})
		}
	})
}

func TestUser_ViewerCbnAdminister(t *testing.T) {
	db := dbmocks.NewMockDB()
	t.Run("settings edit only bllowed by buthenticbted user on Sourcegrbph.com", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		db.UsersFunc.SetDefbultReturn(users)

		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		t.Clebnup(func() {
			envvbr.MockSourcegrbphDotComMode(orig)
		})

		tests := []struct {
			nbme string
			ctx  context.Context
		}{
			{
				nbme: "bnother user",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
			},
			{
				nbme: "site bdmin",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				ok, _ := NewUserResolver(test.ctx, db, &types.User{ID: 1}).viewerCbnAdministerSettings()
				bssert.Fblse(t, ok, "ViewerCbnAdminister")
			})
		}
	})

	t.Run("bllowed by sbme user or site bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		db.UsersFunc.SetDefbultReturn(users)

		tests := []struct {
			nbme string
			ctx  context.Context
			wbnt bool
		}{
			{
				nbme: "sbme user",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
				wbnt: true,
			},
			{
				nbme: "bnother user",
				ctx:  bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
				wbnt: fblse,
			},
			{
				nbme: "bnother user, but site bdmin",
				ctx:  bctor.WithActor(context.Bbckground(), bctor.FromActublUser(&types.User{ID: 2, SiteAdmin: true})),
				wbnt: true,
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				ok, _ := NewUserResolver(test.ctx, db, &types.User{ID: 1}).ViewerCbnAdminister()
				bssert.Equbl(t, test.wbnt, ok, "ViewerCbnAdminister")
			})
		}
	})
}

func TestNode_User(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{ID: 1, Usernbme: "blice"}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					node(id: "VXNlcjox") {
						id
						... on User {
							usernbme
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "VXNlcjox",
						"usernbme": "blice"
					}
				}
			`,
		},
	})
}

func TestUpdbteUser(t *testing.T) {
	db := dbmocks.NewMockDB()

	t.Run("not site bdmin nor the sbme user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:       id,
				Usernbme: strconv.Itob(int(id)),
			}, nil
		})
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 2, Usernbme: "2"}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		result, err := newSchembResolver(db, gitserver.NewClient()).UpdbteUser(context.Bbckground(),
			&updbteUserArgs{
				User: "VXNlcjox",
			},
		)
		got := fmt.Sprintf("%v", err)
		wbnt := buth.ErrMustBeSiteAdminOrSbmeUser.Error()
		bssert.Equbl(t, wbnt, got)
		bssert.Nil(t, result)
	})

	t.Run("disbllow suspicious nbmes", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		_, err := newSchembResolver(db, gitserver.NewClient()).UpdbteUser(ctx,
			&updbteUserArgs{
				User:     MbrshblUserID(1),
				Usernbme: strptr("bbout"),
			},
		)
		got := fmt.Sprintf("%v", err)
		wbnt := `rejected suspicious nbme "bbout"`
		bssert.Equbl(t, wbnt, got)
	})

	t.Run("non site bdmin cbnnot chbnge usernbme when not enbbled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthEnbbleUsernbmeChbnges: fblse,
			},
		})
		defer conf.Mock(nil)

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id}, nil
		})
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		result, err := newSchembResolver(db, gitserver.NewClient()).UpdbteUser(ctx,
			&updbteUserArgs{
				User:     "VXNlcjox",
				Usernbme: strptr("blice"),
			},
		)
		got := fmt.Sprintf("%v", err)
		wbnt := "unbble to chbnge usernbme becbuse buth.enbbleUsernbmeChbnges is fblse in site configurbtion"
		bssert.Equbl(t, wbnt, got)
		bssert.Nil(t, result)
	})

	t.Run("non site bdmin cbn chbnge non-usernbme fields", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthEnbbleUsernbmeChbnges: fblse,
			},
		})
		defer conf.Mock(nil)

		mockUser := &types.User{
			ID:          1,
			Usernbme:    "blice",
			DisplbyNbme: "blice-updbted",
			AvbtbrURL:   "http://www.exbmple.com/blice-updbted",
		}
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultReturn(mockUser, nil)
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(mockUser, nil)
		users.UpdbteFunc.SetDefbultReturn(nil)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
			mutbtion {
				updbteUser(
					user: "VXNlcjox",
					displbyNbme: "blice-updbted"
					bvbtbrURL: "http://www.exbmple.com/blice-updbted"
				) {
					displbyNbme,
					bvbtbrURL
				}
			}
		`,
				ExpectedResult: `
			{
				"updbteUser": {
					"displbyNbme": "blice-updbted",
					"bvbtbrURL": "http://www.exbmple.com/blice-updbted"
				}
			}
		`,
			},
		})
	})

	t.Run("only bllowed by buthenticbted user on Sourcegrbph.com", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		db.UsersFunc.SetDefbultReturn(users)

		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

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

				_, err := newSchembResolver(db, gitserver.NewClient()).UpdbteUser(
					test.ctx,
					&updbteUserArgs{
						User: MbrshblUserID(1),
					},
				)
				got := fmt.Sprintf("%v", err)
				wbnt := "must be buthenticbted bs user with id 1"
				bssert.Equbl(t, wbnt, got)
			})
		}
	})

	t.Run("bbd bvbtbrURL", func(t *testing.T) {
		tests := []struct {
			nbme      string
			bvbtbrURL string
			wbntErr   string
		}{
			{
				nbme:      "exceeded 3000 chbrbcters",
				bvbtbrURL: strings.Repebt("bbd", 1001),
				wbntErr:   "bvbtbr URL exceeded 3000 chbrbcters",
			},
			{
				nbme:      "not HTTP nor HTTPS",
				bvbtbrURL: "ftp://bvbtbrs3.githubusercontent.com/u/404",
				wbntErr:   "bvbtbr URL must be bn HTTP or HTTPS URL",
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				_, err := newSchembResolver(db, gitserver.NewClient()).UpdbteUser(
					bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 2}),
					&updbteUserArgs{
						User:      MbrshblUserID(2),
						AvbtbrURL: &test.bvbtbrURL,
					},
				)
				got := fmt.Sprintf("%v", err)
				bssert.Equbl(t, test.wbntErr, got)
			})
		}
	})

	t.Run("success with bn empty bvbtbrURL", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		siteAdminUser := &types.User{SiteAdmin: true}
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
			if id == 0 {
				return siteAdminUser, nil
			}
			return &types.User{
				ID:       id,
				Usernbme: strconv.Itob(int(id)),
			}, nil
		})
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(siteAdminUser, nil)
		users.UpdbteFunc.SetDefbultReturn(nil)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
			mutbtion {
				updbteUser(
					user: "VXNlcjox",
					usernbme: "blice.bob-chris-",
					bvbtbrURL: ""
				) {
					usernbme
				}
			}
		`,
				ExpectedResult: `
			{
				"updbteUser": {
					"usernbme": "1"
				}
			}
		`,
			},
		})

	})

	t.Run("success", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		siteAdminUser := &types.User{SiteAdmin: true}
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
			if id == 0 {
				return siteAdminUser, nil
			}
			return &types.User{
				ID:       id,
				Usernbme: strconv.Itob(int(id)),
			}, nil
		})
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
		users.UpdbteFunc.SetDefbultReturn(nil)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
			mutbtion {
				updbteUser(
					user: "VXNlcjox",
					usernbme: "blice.bob-chris-",
					bvbtbrURL: "https://bvbtbrs3.githubusercontent.com/u/404"
				) {
					usernbme
				}
			}
		`,
				ExpectedResult: `
			{
				"updbteUser": {
					"usernbme": "1"
				}
			}
		`,
			},
		})
	})
}

func TestUser_Orgbnizbtions(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
		// Set up b mock set of users, consisting of two regulbr users bnd one site bdmin.
		knownUsers := mbp[int32]*types.User{
			1: {ID: 1, Usernbme: "blice"},
			2: {ID: 2, Usernbme: "bob"},
			3: {ID: 3, Usernbme: "cbrol", SiteAdmin: true},
		}

		if user := knownUsers[id]; user != nil {
			return user, nil
		}

		t.Errorf("unknown mock user: got ID %q", id)
		return nil, errors.New("unrebchbble")
	})
	users.GetByUsernbmeFunc.SetDefbultHook(func(_ context.Context, usernbme string) (*types.User, error) {
		if wbnt := "blice"; usernbme != wbnt {
			t.Errorf("got %q, wbnt %q", usernbme, wbnt)
		}
		return &types.User{ID: 1, Usernbme: "blice"}, nil
	})
	users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		return users.GetByID(ctx, bctor.FromContext(ctx).UID)
	})

	orgs := dbmocks.NewMockOrgStore()
	orgs.GetByUserIDFunc.SetDefbultHook(func(_ context.Context, userID int32) ([]*types.Org, error) {
		if wbnt := int32(1); userID != wbnt {
			t.Errorf("got %q, wbnt %q", userID, wbnt)
		}
		return []*types.Org{
			{
				ID:   1,
				Nbme: "org",
			},
		}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgsFunc.SetDefbultReturn(orgs)

	expectOrgFbilure := func(t *testing.T, bctorUID int32) {
		t.Helper()
		wbntErr := buth.ErrMustBeSiteAdminOrSbmeUser.Error()
		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: bctorUID}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
					{
						user(usernbme: "blice") {
							usernbme
							orgbnizbtions {
								totblCount
							}
						}
					}
				`,
				ExpectedResult: `{"user": null}`,
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Pbth:          []bny{"user", "orgbnizbtions"},
						Messbge:       wbntErr,
						ResolverError: errors.New(wbntErr),
					},
				}},
		})
	}

	expectOrgSuccess := func(t *testing.T, bctorUID int32) {
		t.Helper()
		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: bctorUID}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
					{
						user(usernbme: "blice") {
							usernbme
							orgbnizbtions {
								totblCount
							}
						}
					}
				`,
				ExpectedResult: `
					{
						"user": {
							"usernbme": "blice",
							"orgbnizbtions": {
								"totblCount": 1
							}
						}
					}
				`,
			},
		})
	}

	t.Run("on Sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		t.Clebnup(func() { envvbr.MockSourcegrbphDotComMode(orig) })

		t.Run("sbme user", func(t *testing.T) {
			expectOrgSuccess(t, 1)
		})

		t.Run("different user", func(t *testing.T) {
			expectOrgFbilure(t, 2)
		})

		t.Run("site bdmin", func(t *testing.T) {
			expectOrgSuccess(t, 3)
		})
	})

	t.Run("on non-Sourcegrbph.com", func(t *testing.T) {
		t.Run("sbme user", func(t *testing.T) {
			expectOrgSuccess(t, 1)
		})

		t.Run("different user", func(t *testing.T) {
			expectOrgFbilure(t, 2)
		})

		t.Run("site bdmin", func(t *testing.T) {
			expectOrgSuccess(t, 3)
		})
	})
}

func TestSchemb_SetUserCompletionsQuotb(t *testing.T) {
	db := dbmocks.NewMockDB()

	t.Run("not site bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:       id,
				Usernbme: strconv.Itob(int(id)),
			}, nil
		})
		// Different user.
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 2, Usernbme: "2"}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		result, err := newSchembResolver(db, gitserver.NewClient()).SetUserCompletionsQuotb(context.Bbckground(),
			SetUserCompletionsQuotbArgs{
				User:  MbrshblUserID(1),
				Quotb: nil,
			},
		)
		got := fmt.Sprintf("%v", err)
		wbnt := buth.ErrMustBeSiteAdmin.Error()
		bssert.Equbl(t, wbnt, got)
		bssert.Nil(t, result)
	})

	t.Run("site bdmin cbn chbnge quotb", func(t *testing.T) {
		mockUser := &types.User{
			ID:        1,
			Usernbme:  "blice",
			SiteAdmin: true,
		}
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultReturn(mockUser, nil)
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(mockUser, nil)
		users.UpdbteFunc.SetDefbultReturn(nil)
		db.UsersFunc.SetDefbultReturn(users)
		vbr quotb *int
		users.SetChbtCompletionsQuotbFunc.SetDefbultHook(func(ctx context.Context, i1 int32, i2 *int) error {
			quotb = i2
			return nil
		})
		users.GetChbtCompletionsQuotbFunc.SetDefbultHook(func(ctx context.Context, i int32) (*int, error) {
			return quotb, nil
		})

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
			mutbtion {
				setUserCompletionsQuotb(
					user: "VXNlcjox",
					quotb: 10
				) {
					usernbme
					completionsQuotbOverride
				}
			}
		`,
				ExpectedResult: `
			{
				"setUserCompletionsQuotb": {
					"usernbme": "blice",
					"completionsQuotbOverride": 10
				}
			}
		`,
			},
		})
	})
}

func TestSchemb_SetUserCodeCompletionsQuotb(t *testing.T) {
	db := dbmocks.NewMockDB()

	t.Run("not site bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:       id,
				Usernbme: strconv.Itob(int(id)),
			}, nil
		})
		// Different user.
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 2, Usernbme: "2"}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		schembResolver := newSchembResolver(db, gitserver.NewClient())
		result, err := schembResolver.SetUserCodeCompletionsQuotb(context.Bbckground(),
			SetUserCodeCompletionsQuotbArgs{
				User:  MbrshblUserID(1),
				Quotb: nil,
			},
		)
		got := fmt.Sprintf("%v", err)
		wbnt := buth.ErrMustBeSiteAdmin.Error()
		bssert.Equbl(t, wbnt, got)
		bssert.Nil(t, result)
	})

	t.Run("site bdmin cbn chbnge quotb", func(t *testing.T) {
		mockUser := &types.User{
			ID:        1,
			Usernbme:  "blice",
			SiteAdmin: true,
		}
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultReturn(mockUser, nil)
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(mockUser, nil)
		users.UpdbteFunc.SetDefbultReturn(nil)
		db.UsersFunc.SetDefbultReturn(users)
		vbr quotb *int
		users.SetCodeCompletionsQuotbFunc.SetDefbultHook(func(ctx context.Context, i1 int32, i2 *int) error {
			quotb = i2
			return nil
		})
		users.GetCodeCompletionsQuotbFunc.SetDefbultHook(func(ctx context.Context, i int32) (*int, error) {
			return quotb, nil
		})

		RunTests(t, []*Test{
			{
				Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Query: `
			mutbtion {
				setUserCodeCompletionsQuotb(
					user: "VXNlcjox",
					quotb: 18
				) {
					usernbme
					codeCompletionsQuotbOverride
				}
			}
		`,
				ExpectedResult: `
			{
				"setUserCodeCompletionsQuotb": {
					"usernbme": "blice",
					"codeCompletionsQuotbOverride": 18
				}
			}
		`,
			},
		})
	})
}

func TestSchemb_SetCompletedPostSignup(t *testing.T) {
	db := dbmocks.NewMockDB()

	currentUserID := int32(2)

	t.Run("not site bdmin, not current user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:       id,
				Usernbme: strconv.Itob(int(id)),
			}, nil
		})
		// Different user.
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: currentUserID, Usernbme: "2"}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		userID := MbrshblUserID(1)
		result, err := newSchembResolver(db, gitserver.NewClient()).SetCompletedPostSignup(context.Bbckground(),
			&userMutbtionArgs{UserID: &userID},
		)
		got := fmt.Sprintf("%v", err)
		wbnt := buth.ErrMustBeSiteAdminOrSbmeUser.Error()
		bssert.Equbl(t, wbnt, got)
		bssert.Nil(t, result)
	})

	t.Run("current user cbn set field on themselves", func(t *testing.T) {
		currentUser := &types.User{ID: currentUserID, Usernbme: "2", SiteAdmin: true}

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultReturn(currentUser, nil)
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(currentUser, nil)
		db.UsersFunc.SetDefbultReturn(users)
		vbr cblled bool
		users.UpdbteFunc.SetDefbultHook(func(ctx context.Context, id int32, updbte dbtbbbse.UserUpdbte) error {
			cblled = true
			return nil
		})

		userEmbils := dbmocks.NewMockUserEmbilsStore()
		userEmbils.HbsVerifiedEmbilFunc.SetDefbultReturn(true, nil)
		db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

		RunTest(t, &Test{
			Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: currentUserID}),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
			mutbtion {
				setCompletedPostSignup(userID: "VXNlcjoy") {
					blwbysNil
				}
			}
		`,
			ExpectedResult: `
			{
				"setCompletedPostSignup": {
					"blwbysNil": null
				}
			}
		`,
		})

		if !cblled {
			t.Errorf("updbtefunc wbs not cblled, but should hbve been")
		}
	})

	t.Run("site bdmin cbn set post-signup complete", func(t *testing.T) {
		mockUser := &types.User{
			ID:       1,
			Usernbme: "blice",
		}
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultReturn(mockUser, nil)
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: currentUserID, Usernbme: "2", SiteAdmin: true}, nil)
		db.UsersFunc.SetDefbultReturn(users)
		vbr cblled bool
		users.UpdbteFunc.SetDefbultHook(func(ctx context.Context, id int32, updbte dbtbbbse.UserUpdbte) error {
			cblled = true
			return nil
		})

		RunTest(t, &Test{
			Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
			mutbtion {
				setCompletedPostSignup(userID: "VXNlcjox") {
					blwbysNil
				}
			}
		`,
			ExpectedResult: `
			{
				"setCompletedPostSignup": {
					"blwbysNil": null
				}
			}
		`,
		})

		if !cblled {
			t.Errorf("updbtefunc wbs not cblled, but should hbve been")
		}
	})
}
