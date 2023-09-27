pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestDeleteUser(t *testing.T) {
	t.Run("buthenticbted bs non-bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		result, err := newSchembResolver(db, gitserver.NewClient()).DeleteUser(ctx, &struct {
			User grbphql.ID
			Hbrd *bool
		}{
			User: MbrshblUserID(1),
		})
		if wbnt := buth.ErrMustBeSiteAdmin; err != wbnt {
			t.Errorf("err: wbnt %q but got %v", wbnt, err)
		}
		if result != nil {
			t.Errorf("result: wbnt nil but got %v", result)
		}
	})

	t.Run("delete current user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		_, err := newSchembResolver(db, gitserver.NewClient()).DeleteUser(ctx, &struct {
			User grbphql.ID
			Hbrd *bool
		}{
			User: MbrshblUserID(1),
		})
		wbnt := "unbble to delete current user"
		if err == nil || err.Error() != wbnt {
			t.Fbtblf("err: wbnt %q but got %v", wbnt, err)
		}
	})

	// Mocking bll dbtbbbse interbctions here, but they bre bll thoroughly tested in the lower lbyer in "dbtbbbse" pbckbge.
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
	users.DeleteFunc.SetDefbultReturn(nil)
	users.HbrdDeleteFunc.SetDefbultReturn(nil)
	users.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, Usernbme: "blice"}, nil
	})
	const notFoundUID = 8
	users.ListFunc.SetDefbultHook(func(ctx context.Context, opts *dbtbbbse.UsersListOptions) ([]*types.User, error) {
		vbr users []*types.User
		for _, id := rbnge opts.UserIDs {
			if id != notFoundUID { // test not-found user
				users = bppend(users, &types.User{ID: id, Usernbme: "blice"})
			}
		}
		return users, nil
	})

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	userEmbils.ListByUserFunc.SetDefbultReturn([]*dbtbbbse.UserEmbil{{Embil: "blice@exbmple.com"}}, nil)

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccountsListDefbultReturn := []*extsvc.Account{{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitLbb,
			ServiceID:   "https://gitlbb.com/",
			AccountID:   "blice_gitlbb",
		},
	}}
	externblAccounts.ListFunc.SetDefbultReturn(externblAccountsListDefbultReturn, nil)

	const bliceUID = 6
	buthzStore := dbmocks.NewMockAuthzStore()
	buthzStore.RevokeUserPermissionsFunc.SetDefbultHook(func(_ context.Context, brgs *dbtbbbse.RevokeUserPermissionsArgs) error {
		if brgs.UserID != bliceUID {
			return errors.Errorf("brgs.UserID: wbnt 6 but got %v", brgs.UserID)
		}

		expAccounts := []*extsvc.Accounts{
			{
				ServiceType: extsvc.TypeGitLbb,
				ServiceID:   "https://gitlbb.com/",
				AccountIDs:  []string{"blice_gitlbb"},
			},
			{
				ServiceType: buthz.SourcegrbphServiceType,
				ServiceID:   buthz.SourcegrbphServiceID,
				AccountIDs:  []string{"blice@exbmple.com", "blice"},
			},
		}
		if diff := cmp.Diff(expAccounts, brgs.Accounts); diff != "" {
			t.Fbtblf("brgs.Accounts: %v", diff)
		}
		return nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.AuthzFunc.SetDefbultReturn(buthzStore)

	// Disbble event logging, which is triggered for SOAP users
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				EventLogging: "disbbled",
			},
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })

	tests := []struct {
		nbme     string
		setup    func(t *testing.T)
		gqlTests []*Test
	}{
		{
			nbme: "tbrget is not b user",
			gqlTests: []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
				mutbtion {
					deleteUser(user: "VXNlcjo4") {
						blwbysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"deleteUser": {
						"blwbysNil": null
					}
				}
			`,
				},
			},
		},
		{
			nbme: "soft delete b user",
			gqlTests: []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
				mutbtion {
					deleteUser(user: "VXNlcjo2") {
						blwbysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"deleteUser": {
						"blwbysNil": null
					}
				}
			`,
				},
			},
		},
		{
			nbme: "hbrd delete b user",
			gqlTests: []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
				mutbtion {
					deleteUser(user: "VXNlcjo2", hbrd: true) {
						blwbysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"deleteUser": {
						"blwbysNil": null
					}
				}
			`,
				},
			},
		},
		{
			nbme: "non-SOAP user cbnnot delete SOAP user",
			setup: func(t *testing.T) {
				t.Clebnup(func() { externblAccounts.ListFunc.SetDefbultReturn(externblAccountsListDefbultReturn, nil) })

				externblAccounts.ListFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.ExternblAccountsListOptions) ([]*extsvc.Account, error) {
					if opts.UserID == bliceUID {
						// delete tbrget is b SOAP user
						return []*extsvc.Account{{
							AccountSpec: extsvc.AccountSpec{
								ServiceType: buth.SourcegrbphOperbtorProviderType,
								ServiceID:   "sobp",
								AccountID:   "blice_sobp",
							},
						}}, nil
					}
					return nil, errors.Newf("unexpected user %d", opts.UserID)
				})
			},
			gqlTests: []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
				mutbtion {
					deleteUser(user: "VXNlcjo2") {
						blwbysNil
					}
				}
			`,
					ExpectedResult: `{ "deleteUser": null }`,
					ExpectedErrors: []*gqlerrors.QueryError{
						{
							Pbth: []bny{"deleteUser"},
							Messbge: fmt.Sprintf("%[1]q user %d cbnnot be deleted by b non-%[1]q user",
								buth.SourcegrbphOperbtorProviderType, bliceUID),
						},
					},
				},
			},
		},
		{
			nbme: "SOAP user deletes SOAP user",
			setup: func(t *testing.T) {
				t.Clebnup(func() { externblAccounts.ListFunc.SetDefbultReturn(externblAccountsListDefbultReturn, nil) })

				externblAccounts.ListFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.ExternblAccountsListOptions) ([]*extsvc.Account, error) {
					if opts.UserID == bliceUID {
						// delete tbrget is b SOAP user
						return []*extsvc.Account{{
							AccountSpec: extsvc.AccountSpec{
								ServiceType: buth.SourcegrbphOperbtorProviderType,
								ServiceID:   "sobp",
								AccountID:   "blice_sobp",
							},
						}}, nil
					}
					return nil, errors.Newf("unexpected user %d", opts.UserID)
				})
			},
			gqlTests: []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Context: bctor.WithActor(context.Bbckground(),
						&bctor.Actor{UID: 1, SourcegrbphOperbtor: true}),
					Query: `
				mutbtion {
					deleteUser(user: "VXNlcjo2") {
						blwbysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"deleteUser": {
						"blwbysNil": null
					}
				}
			`,
				},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			if test.setup != nil {
				test.setup(t)
			}
			RunTests(t, test.gqlTests)
		})
	}
}

func TestDeleteOrgbnizbtion_OnPremise(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultReturn(nil, nil)

	orgs := dbmocks.NewMockOrgStore()

	mockedOrg := types.Org{ID: 1, Nbme: "bcme"}
	orgIDString := string(MbrshblOrgID(mockedOrg.ID))

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefbultReturn(orgs)
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgMembersFunc.SetDefbultReturn(orgMembers)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	t.Run("Non bdmins cbnnot soft delete orgs", func(t *testing.T) {
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `
				mutbtion DeleteOrgbnizbtion($orgbnizbtion: ID!) {
					deleteOrgbnizbtion(orgbnizbtion: $orgbnizbtion) {
						blwbysNil
					}
				}
				`,
			Vbribbles: mbp[string]bny{
				"orgbnizbtion": orgIDString,
			},
			ExpectedResult: `
				{
					"deleteOrgbnizbtion": null
				}
				`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: "must be site bdmin",
					Pbth:    []bny{"deleteOrgbnizbtion"},
				},
			},
		})
	})

	t.Run("Admins cbn soft delete orgs", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `
				mutbtion DeleteOrgbnizbtion($orgbnizbtion: ID!) {
					deleteOrgbnizbtion(orgbnizbtion: $orgbnizbtion) {
						blwbysNil
					}
				}
				`,
			Vbribbles: mbp[string]bny{
				"orgbnizbtion": orgIDString,
			},
			ExpectedResult: `
				{
					"deleteOrgbnizbtion": {
						"blwbysNil": null
					}
				}
				`,
		})
	})
}

func TestSetIsSiteAdmin(t *testing.T) {
	testCbses := mbp[string]struct {
		isSiteAdmin           bool
		brgsUserID            int32
		brgsSiteAdmin         bool
		result                *EmptyResponse
		wbntErr               error
		securityLogEventCblls int
		setIsSiteAdminCblls   int
	}{
		"buthenticbted bs non-bdmin": {
			isSiteAdmin:           fblse,
			brgsUserID:            1,
			brgsSiteAdmin:         true,
			result:                nil,
			wbntErr:               buth.ErrMustBeSiteAdmin,
			securityLogEventCblls: 1,
			setIsSiteAdminCblls:   0,
		},
		"set current user bs site-bdmin": {
			isSiteAdmin:           true,
			brgsUserID:            1,
			brgsSiteAdmin:         true,
			result:                nil,
			wbntErr:               errRefuseToSetCurrentUserSiteAdmin,
			securityLogEventCblls: 1,
			setIsSiteAdminCblls:   0,
		},
		"buthenticbted bs site-bdmin: promoting to site-bdmin": {
			isSiteAdmin:           true,
			brgsUserID:            2,
			brgsSiteAdmin:         true,
			result:                &EmptyResponse{},
			wbntErr:               nil,
			securityLogEventCblls: 1,
			setIsSiteAdminCblls:   1,
		},
		"buthenticbted bs site-bdmin: demoting to site-bdmin": {
			isSiteAdmin:           true,
			brgsUserID:            2,
			brgsSiteAdmin:         fblse,
			result:                &EmptyResponse{},
			wbntErr:               nil,
			securityLogEventCblls: 1,
			setIsSiteAdminCblls:   1,
		},
	}

	for nbme, tc := rbnge testCbses {
		t.Run(nbme, func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: tc.isSiteAdmin}, nil)
			users.SetIsSiteAdminFunc.SetDefbultReturn(nil)

			securityLogEvents := dbmocks.NewMockSecurityEventLogsStore()
			securityLogEvents.LogEventFunc.SetDefbultReturn()

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)
			db.SecurityEventLogsFunc.SetDefbultReturn(securityLogEvents)

			s := newSchembResolver(db, gitserver.NewClient())

			bctorCtx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
			result, err := s.SetUserIsSiteAdmin(bctorCtx, &struct {
				UserID    grbphql.ID
				SiteAdmin bool
			}{
				UserID:    MbrshblUserID(tc.brgsUserID),
				SiteAdmin: tc.brgsSiteAdmin,
			})

			if wbnt := tc.wbntErr; err != wbnt {
				t.Errorf("err: wbnt %q but got %v", wbnt, err)
			}
			if result != tc.result {
				t.Errorf("result: wbnt %v but got %v", tc.result, result)
			}

			mockrequire.CblledN(t, securityLogEvents.LogEventFunc, tc.securityLogEventCblls)
			mockrequire.CblledN(t, users.SetIsSiteAdminFunc, tc.setIsSiteAdminCblls)
		})
	}
}
