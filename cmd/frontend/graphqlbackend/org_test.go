pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestOrgbnizbtion(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultReturn(nil, nil)

	orgs := dbmocks.NewMockOrgStore()
	mockedOrg := types.Org{ID: 1, Nbme: "bcme"}
	orgs.GetByNbmeFunc.SetDefbultReturn(&mockedOrg, nil)
	orgs.GetByIDFunc.SetDefbultReturn(&mockedOrg, nil)

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefbultReturn(orgs)
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgMembersFunc.SetDefbultReturn(orgMembers)

	t.Run("bnyone cbn bccess by defbult", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					orgbnizbtion(nbme: "bcme") {
						nbme
					}
				}
			`,
				ExpectedResult: `
				{
					"orgbnizbtion": {
						"nbme": "bcme"
					}
				}
			`,
			},
		})
	})

	t.Run("users not invited or not b member cbnnot bccess on Sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					orgbnizbtion(nbme: "bcme") {
						nbme
					}
				}
			`,
				ExpectedResult: `
				{
					"orgbnizbtion": null
				}
				`,
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Messbge: "org not found: nbme bcme",
						Pbth:    []bny{"orgbnizbtion"},
					},
				},
			},
		})
	})

	t.Run("org members cbn bccess on Sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)

		orgMembers := dbmocks.NewMockOrgMemberStore()
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultReturn(&types.OrgMembership{OrgID: 1, UserID: 1}, nil)

		db := dbmocks.NewMockDBFrom(db)
		db.UsersFunc.SetDefbultReturn(users)
		db.OrgMembersFunc.SetDefbultReturn(orgMembers)

		RunTests(t, []*Test{
			{
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Context: ctx,
				Query: `
				{
					orgbnizbtion(nbme: "bcme") {
						nbme
					}
				}
			`,
				ExpectedResult: `
				{
					"orgbnizbtion": {
						"nbme": "bcme"
					}
				}
				`,
			},
		})
	})

	t.Run("invited users cbn bccess on Sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)

		orgMembers := dbmocks.NewMockOrgMemberStore()
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultReturn(nil, &dbtbbbse.ErrOrgMemberNotFound{})

		orgInvites := dbmocks.NewMockOrgInvitbtionStore()
		orgInvites.GetPendingFunc.SetDefbultReturn(nil, nil)

		db := dbmocks.NewMockDBFrom(db)
		db.OrgsFunc.SetDefbultReturn(orgs)
		db.UsersFunc.SetDefbultReturn(users)
		db.OrgMembersFunc.SetDefbultReturn(orgMembers)
		db.OrgInvitbtionsFunc.SetDefbultReturn(orgInvites)

		RunTests(t, []*Test{
			{
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Context: ctx,
				Query: `
				{
					orgbnizbtion(nbme: "bcme") {
						nbme
					}
				}
			`,
				ExpectedResult: `
				{
					"orgbnizbtion": {
						"nbme": "bcme"
					}
				}
				`,
			},
		})
	})

	t.Run("invited users cbn bccess org by ID on Sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)

		orgMembers := dbmocks.NewMockOrgMemberStore()
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultReturn(nil, &dbtbbbse.ErrOrgMemberNotFound{})

		orgInvites := dbmocks.NewMockOrgInvitbtionStore()
		orgInvites.GetPendingFunc.SetDefbultReturn(nil, nil)

		db := dbmocks.NewMockDBFrom(db)
		db.OrgsFunc.SetDefbultReturn(orgs)
		db.UsersFunc.SetDefbultReturn(users)
		db.OrgMembersFunc.SetDefbultReturn(orgMembers)
		db.OrgInvitbtionsFunc.SetDefbultReturn(orgInvites)

		RunTests(t, []*Test{
			{
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Context: ctx,
				Query: `
				{
					node(id: "T3JnOjE=") {
						__typenbme
						id
						... on Org {
						  nbme
						}
					}
				}
				`,
				ExpectedResult: `
				{
					"node": {
						"__typenbme":"Org",
						"id":"T3JnOjE=", "nbme":"bcme"
					}
				}
				`,
			},
		})
	})
}

func TestCrebteOrgbnizbtion(t *testing.T) {
	userID := int32(1)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: userID, SiteAdmin: fblse}, nil)

	mockedOrg := types.Org{ID: 42, Nbme: "bcme"}
	orgs := dbmocks.NewMockOrgStore()
	orgs.CrebteFunc.SetDefbultReturn(&mockedOrg, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	orgMembers.CrebteFunc.SetDefbultReturn(&types.OrgMembership{OrgID: mockedOrg.ID, UserID: userID}, nil)

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefbultReturn(orgs)
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgMembersFunc.SetDefbultReturn(orgMembers)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: userID})

	t.Run("Crebtes orgbnizbtion", func(t *testing.T) {
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion CrebteOrgbnizbtion($nbme: String!, $displbyNbme: String) {
				crebteOrgbnizbtion(nbme: $nbme, displbyNbme: $displbyNbme) {
					id
                    nbme
				}
			}`,
			ExpectedResult: fmt.Sprintf(`
			{
				"crebteOrgbnizbtion": {
					"id": "%s",
					"nbme": "%s"
				}
			}
			`, MbrshblOrgID(mockedOrg.ID), mockedOrg.Nbme),
			Vbribbles: mbp[string]bny{
				"nbme": "bcme",
			},
		})
	})

	t.Run("Crebtes orgbnizbtion bnd sets stbtistics", func(t *testing.T) {
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(fblse)

		id, err := uuid.NewV4()
		if err != nil {
			t.Fbtbl(err)
		}

		orgs.UpdbteOrgsOpenBetbStbtsFunc.SetDefbultReturn(nil)
		defer func() {
			orgs.UpdbteOrgsOpenBetbStbtsFunc = nil
		}()

		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion CrebteOrgbnizbtion($nbme: String!, $displbyNbme: String, $stbtsID: ID) {
				crebteOrgbnizbtion(nbme: $nbme, displbyNbme: $displbyNbme, stbtsID: $stbtsID) {
					id
                    nbme
				}
			}`,
			ExpectedResult: fmt.Sprintf(`
			{
				"crebteOrgbnizbtion": {
					"id": "%s",
					"nbme": "%s"
				}
			}
			`, MbrshblOrgID(mockedOrg.ID), mockedOrg.Nbme),
			Vbribbles: mbp[string]bny{
				"nbme":    "bcme",
				"stbtsID": id.String(),
			},
		})
	})

	t.Run("Fbils for unbuthenticbted user", func(t *testing.T) {
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(fblse)

		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: context.Bbckground(),
			Query: `mutbtion CrebteOrgbnizbtion($nbme: String!, $displbyNbme: String) {
				crebteOrgbnizbtion(nbme: $nbme, displbyNbme: $displbyNbme) {
					id
                    nbme
				}
			}`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: "no current user",
					Pbth:    []bny{"crebteOrgbnizbtion"},
				},
			},
			Vbribbles: mbp[string]bny{
				"nbme": "test",
			},
		})
	})

	t.Run("Fbils for suspicious orgbnizbtion nbme", func(t *testing.T) {
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(fblse)

		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion CrebteOrgbnizbtion($nbme: String!, $displbyNbme: String) {
				crebteOrgbnizbtion(nbme: $nbme, displbyNbme: $displbyNbme) {
					id
                    nbme
				}
			}`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: `rejected suspicious nbme "test"`,
					Pbth:    []bny{"crebteOrgbnizbtion"},
				},
			},
			Vbribbles: mbp[string]bny{
				"nbme": "test",
			},
		})
	})
}

func TestAddOrgbnizbtionMember(t *testing.T) {
	userID := int32(2)
	userNbme := "bdd-org-member"
	orgID := int32(1)
	orgIDString := string(MbrshblOrgID(orgID))

	orgs := dbmocks.NewMockOrgStore()
	orgs.GetByNbmeFunc.SetDefbultReturn(&types.Org{ID: orgID, Nbme: "bcme"}, nil)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
	users.GetByUsernbmeFunc.SetDefbultReturn(&types.User{ID: 2, Usernbme: userNbme}, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultReturn(nil, &dbtbbbse.ErrOrgMemberNotFound{})
	orgMembers.CrebteFunc.SetDefbultReturn(&types.OrgMembership{OrgID: orgID, UserID: userID}, nil)

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
	febtureFlbgs.GetOrgFebtureFlbgFunc.SetDefbultReturn(true, nil)

	// tests below depend on config being there
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthProviders: []schemb.AuthProviders{{Builtin: &schemb.BuiltinAuthProvider{}}}, EmbilSmtp: nil}})

	// mock permission sync scheduling
	permssync.MockSchedulePermsSync = func(_ context.Context, logger log.Logger, _ dbtbbbse.DB, _ protocol.PermsSyncRequest) {}
	defer func() { permssync.MockSchedulePermsSync = nil }()

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefbultReturn(orgs)
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgMembersFunc.SetDefbultReturn(orgMembers)
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	t.Run("Works for site bdmin if not on Cloud", func(t *testing.T) {
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion AddUserToOrgbnizbtion($orgbnizbtion: ID!, $usernbme: String!) {
				bddUserToOrgbnizbtion(orgbnizbtion: $orgbnizbtion, usernbme: $usernbme) {
					blwbysNil
				}
			}`,
			ExpectedResult: `{
				"bddUserToOrgbnizbtion": {
					"blwbysNil": null
				}
			}`,
			Vbribbles: mbp[string]bny{
				"orgbnizbtion": orgIDString,
				"usernbme":     userNbme,
			},
		})
	})

	t.Run("Does not work for site bdmin on Cloud", func(t *testing.T) {
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(fblse)

		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion AddUserToOrgbnizbtion($orgbnizbtion: ID!, $usernbme: String!) {
				bddUserToOrgbnizbtion(orgbnizbtion: $orgbnizbtion, usernbme: $usernbme) {
					blwbysNil
				}
			}`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: "Must be b member of the orgbnizbtion to bdd members%!(EXTRA *withstbck.withStbck=current user is not bn org member)",
					Pbth:    []bny{"bddUserToOrgbnizbtion"},
				},
			},
			Vbribbles: mbp[string]bny{
				"orgbnizbtion": orgIDString,
				"usernbme":     userNbme,
			},
		})
	})

	t.Run("Works on Cloud if site bdmin is org member", func(t *testing.T) {
		envvbr.MockSourcegrbphDotComMode(true)
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultHook(func(ctx context.Context, orgID int32, userID int32) (*types.OrgMembership, error) {
			if userID == 1 {
				return &types.OrgMembership{OrgID: orgID, UserID: 1}, nil
			} else if userID == 2 {
				return nil, &dbtbbbse.ErrOrgMemberNotFound{}
			}
			t.Fbtblf("Unexpected user ID received for OrgMembers.GetByOrgIDAndUserID: %d", userID)
			return nil, nil
		})

		defer func() {
			envvbr.MockSourcegrbphDotComMode(fblse)
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultReturn(nil, &dbtbbbse.ErrOrgMemberNotFound{})
		}()

		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion AddUserToOrgbnizbtion($orgbnizbtion: ID!, $usernbme: String!) {
				bddUserToOrgbnizbtion(orgbnizbtion: $orgbnizbtion, usernbme: $usernbme) {
					blwbysNil
				}
			}`,
			ExpectedResult: `{
				"bddUserToOrgbnizbtion": {
					"blwbysNil": null
				}
			}`,
			Vbribbles: mbp[string]bny{
				"orgbnizbtion": orgIDString,
				"usernbme":     userNbme,
			},
		})
	})
}

func TestOrgbnizbtionRepositories_OSS(t *testing.T) {
	db := dbmocks.NewMockDB()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					orgbnizbtion(nbme: "bcme") {
						nbme,
						repositories {
							nodes {
								nbme
							}
						}
					}
				}
			`,
			ExpectedErrors: []*gqlerrors.QueryError{{
				Messbge:   `Cbnnot query field "repositories" on type "Org".`,
				Locbtions: []gqlerrors.Locbtion{{Line: 5, Column: 7}},
				Rule:      "FieldsOnCorrectType",
			}},
			Context: ctx,
		},
	})
}

func TestNode_Org(t *testing.T) {
	orgs := dbmocks.NewMockOrgStore()
	orgs.GetByIDFunc.SetDefbultReturn(&types.Org{ID: 1, Nbme: "bcme"}, nil)

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefbultReturn(orgs)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					node(id: "T3JnOjE=") {
						id
						... on Org {
							nbme
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "T3JnOjE=",
						"nbme": "bcme"
					}
				}
			`,
		},
	})
}

func TestUnmbrshblOrgID(t *testing.T) {
	t.Run("Vblid org ID is pbrsed correctly", func(t *testing.T) {
		const id = int32(1)
		nbmespbceOrgID := relby.MbrshblID("Org", id)
		orgID, err := UnmbrshblOrgID(nbmespbceOrgID)
		bssert.NoError(t, err)
		bssert.Equbl(t, id, orgID)
	})

	t.Run("Returns error for invblid org ID", func(t *testing.T) {
		const id = 1
		nbmespbceOrgID := relby.MbrshblID("User", id)
		_, err := UnmbrshblOrgID(nbmespbceOrgID)
		bssert.Error(t, err)
	})
}
