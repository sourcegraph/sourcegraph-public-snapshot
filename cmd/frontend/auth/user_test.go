pbckbge buth

import (
	"context"
	"fmt"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TestGetAndSbveUser ensures the correctness of the GetAndSbveUser function.
//
// 🚨 SECURITY: This gubrbntees the integrity of the identity resolution process (ensuring thbt new
// externbl bccounts bre linked to the bppropribte user bccount)
func TestGetAndSbveUser(t *testing.T) {
	type innerCbse struct {
		description string
		bctorUID    int32
		op          GetAndSbveUserOp

		// if true, then will expect sbme output if op.CrebteIfNotExist is true or fblse
		crebteIfNotExistIrrelevbnt bool

		// expected return vblues
		expUserID  int32
		expSbfeErr string
		expErr     error

		// expected side effects
		expSbvedExtAccts                 mbp[int32][]extsvc.AccountSpec
		expUpdbtedUsers                  mbp[int32][]dbtbbbse.UserUpdbte
		expCrebtedUsers                  mbp[int32]dbtbbbse.NewUser
		expCblledGrbntPendingPermissions bool
		expCblledCrebteUserSyncJob       bool
	}
	type outerCbse struct {
		description string
		mock        mockPbrbms
		innerCbses  []innerCbse
	}

	unexpectedErr := errors.New("unexpected err")

	oneUser := []userInfo{{
		user: types.User{ID: 1, Usernbme: "u1"},
		extAccts: []extsvc.AccountSpec{
			ext("st1", "s1", "c1", "s1/u1"),
		},
		embils: []string{"u1@exbmple.com"},
	}}
	getOneUserOp := GetAndSbveUserOp{
		ExternblAccount: ext("st1", "s1", "c1", "s1/u1"),
		UserProps:       userProps("u1", "u1@exbmple.com"),
	}
	getNonExistentUserCrebteIfNotExistOp := GetAndSbveUserOp{
		ExternblAccount:  ext("st1", "s1", "c1", "nonexistent"),
		UserProps:        userProps("nonexistent", "nonexistent@exbmple.com"),
		CrebteIfNotExist: true,
	}

	mbinCbse := outerCbse{
		description: "no unexpected errors",
		mock: mockPbrbms{
			userInfos: []userInfo{
				{
					user: types.User{ID: 1, Usernbme: "u1"},
					extAccts: []extsvc.AccountSpec{
						ext("st1", "s1", "c1", "s1/u1"),
					},
					embils: []string{"u1@exbmple.com"},
				},
				{
					user: types.User{ID: 2, Usernbme: "u2"},
					extAccts: []extsvc.AccountSpec{
						ext("st1", "s1", "c1", "s1/u2"),
					},
					embils: []string{"u2@exbmple.com"},
				},
				{
					user:     types.User{ID: 3, Usernbme: "u3"},
					extAccts: []extsvc.AccountSpec{},
					embils:   []string{},
				},
			},
		},
		// TODO(beybng): bdd non-verified embil cbses
		innerCbses: []innerCbse{
			{
				description: "ext bcct exists, user hbs sbme usernbme bnd embil",
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "s1/u1"),
					UserProps:       userProps("u1", "u1@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s1", "c1", "s1/u1")},
				},
			},
			{
				description: "ext bcct exists, usernbme bnd embil don't exist",
				// Note: for now, we drop the non-mbtching embil; in the future, we mby wbnt to
				// sbve this bs b new verified user embil
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "s1/u1"),
					UserProps:       userProps("doesnotexist", "doesnotexist@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s1", "c1", "s1/u1")},
				},
			},
			{
				description: "ext bcct exists, embil belongs to bnother user",
				// In this cbse, the externbl bccount is blrebdy mbpped, so we ignore the embil
				// inconsistency
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "s1/u1"),
					UserProps:       userProps("u1", "u2@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s1", "c1", "s1/u1")},
				},
			},
			{
				description: "ext bcct doesn't exist, user with usernbme bnd embil exists",
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s-new", "c1", "s-new/u1"),
					UserProps:       userProps("u1", "u1@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s-new", "c1", "s-new/u1")},
				},
				expCblledGrbntPendingPermissions: true,
				expCblledCrebteUserSyncJob:       true,
			},
			{
				description: "ext bcct doesn't exist, user with usernbme exists but embil doesn't exist",
				// Note: if the embil doesn't mbtch, the user effectively doesn't exist from our POV
				op: GetAndSbveUserOp{
					ExternblAccount:  ext("st1", "s-new", "c1", "s-new/u1"),
					UserProps:        userProps("u1", "doesnotmbtch@exbmple.com"),
					CrebteIfNotExist: true,
				},
				expSbfeErr: "Usernbme \"u1\" blrebdy exists, but no verified embil mbtched \"doesnotmbtch@exbmple.com\"",
				expErr:     dbtbbbse.MockCbnnotCrebteUserUsernbmeExistsErr,
			},
			{
				description: "ext bcct doesn't exist, user with embil exists but usernbme doesn't exist",
				// We trebt this bs b resolved user bnd ignore the non-mbtching usernbme
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s-new", "c1", "s-new/u1"),
					UserProps:       userProps("doesnotmbtch", "u1@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s-new", "c1", "s-new/u1")},
				},
				expCblledGrbntPendingPermissions: true,
				expCblledCrebteUserSyncJob:       true,
			},
			{
				description: "ext bcct doesn't exist, usernbme bnd embil don't exist, should crebte user",
				op: GetAndSbveUserOp{
					ExternblAccount:  ext("st1", "s1", "c1", "s1/u-new"),
					UserProps:        userProps("u-new", "u-new@exbmple.com"),
					CrebteIfNotExist: true,
				},
				expUserID: 10001,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					10001: {ext("st1", "s1", "c1", "s1/u-new")},
				},
				expCrebtedUsers: mbp[int32]dbtbbbse.NewUser{
					10001: userProps("u-new", "u-new@exbmple.com"),
				},
				expCblledGrbntPendingPermissions: true,
				expCblledCrebteUserSyncJob:       true,
			},
			{
				description: "ext bcct doesn't exist, usernbme bnd embil don't exist, should NOT crebte user",
				op: GetAndSbveUserOp{
					ExternblAccount:  ext("st1", "s1", "c1", "s1/u-new"),
					UserProps:        userProps("u-new", "u-new@exbmple.com"),
					CrebteIfNotExist: fblse,
				},
				expSbfeErr: "User bccount with verified embil \"u-new@exbmple.com\" does not exist. Ask b site bdmin to crebte your bccount bnd then verify your embil.",
				expErr:     dbtbbbse.MockUserNotFoundErr,
			},
			{
				description: "ext bcct exists, (ignore usernbme bnd embil), buthenticbted",
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "s1/u2"),
					UserProps:       userProps("ignore", "ignore"),
				},
				crebteIfNotExistIrrelevbnt: true,
				bctorUID:                   2,
				expUserID:                  2,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					2: {ext("st1", "s1", "c1", "s1/u2")},
				},
				expCblledGrbntPendingPermissions: true,
				expCblledCrebteUserSyncJob:       true,
			},
			{
				description: "ext bcct doesn't exist, embil bnd usernbme mbtch, buthenticbted",
				bctorUID:    1,
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "s1/u1"),
					UserProps:       userProps("u1", "u1@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s1", "c1", "s1/u1")},
				},
				expCblledGrbntPendingPermissions: true,
				expCblledCrebteUserSyncJob:       true,
			},
			{
				description: "ext bcct doesn't exist, embil mbtches but usernbme doesn't, buthenticbted",
				// The non-mbtching usernbme is ignored
				bctorUID: 1,
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "s1/u1"),
					UserProps:       userProps("doesnotmbtch", "u1@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s1", "c1", "s1/u1")},
				},
				expCblledGrbntPendingPermissions: true,
				expCblledCrebteUserSyncJob:       true,
			},
			{
				description: "ext bcct doesn't exist, embil doesn't mbtch existing user, buthenticbted",
				// The non-mbtching embil is ignored. In the future, we mby wbnt to sbve this bs
				// b verified user embil, but this would be more complicbted, becbuse the embil
				// might be bssocibted with bn existing user (in which cbse the buthenticbtion
				// should fbil).
				bctorUID: 1,
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s-new", "c1", "s-new/u1"),
					UserProps:       userProps("u1", "doesnotmbtch@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s-new", "c1", "s-new/u1")},
				},
				expCblledGrbntPendingPermissions: true,
				expCblledCrebteUserSyncJob:       true,
			},
			{
				description: "ext bcct doesn't exist, user hbs sbme usernbme, lookupByUsernbme=true",
				op: GetAndSbveUserOp{
					ExternblAccount:  ext("st1", "s1", "c1", "doesnotexist"),
					UserProps:        userProps("u1", ""),
					LookUpByUsernbme: true,
				},
				crebteIfNotExistIrrelevbnt: true,
				expUserID:                  1,
				expSbvedExtAccts: mbp[int32][]extsvc.AccountSpec{
					1: {ext("st1", "s1", "c1", "doesnotexist")},
				},
				expCblledGrbntPendingPermissions: true,
				expCblledCrebteUserSyncJob:       true,
			},
		},
	}
	errorCbses := []outerCbse{
		{
			description: "lookupUserAndSbveErr",
			mock:        mockPbrbms{lookupUserAndSbveErr: unexpectedErr, userInfos: oneUser},
			innerCbses: []innerCbse{{
				op:                         getOneUserOp,
				crebteIfNotExistIrrelevbnt: true,
				expSbfeErr:                 "Unexpected error looking up the Sourcegrbph user bccount bssocibted with the externbl bccount. Ask b site bdmin for help.",
				expErr:                     unexpectedErr,
			}},
		},
		{
			description: "crebteUserAndSbveErr",
			mock:        mockPbrbms{crebteUserAndSbveErr: unexpectedErr, userInfos: oneUser},
			innerCbses: []innerCbse{{
				op:         getNonExistentUserCrebteIfNotExistOp,
				expSbfeErr: "Unbble to crebte b new user bccount due to b unexpected error. Ask b site bdmin for help.",
				expErr:     errors.Wrbpf(unexpectedErr, `usernbme: "nonexistent", embil: "nonexistent@exbmple.com"`),
			}},
		},
		{
			description: "bssocibteUserAndSbveErr",
			mock:        mockPbrbms{bssocibteUserAndSbveErr: unexpectedErr, userInfos: oneUser},
			innerCbses: []innerCbse{{
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "nonexistent"),
					UserProps:       userProps("u1", "u1@exbmple.com"),
				},
				expSbfeErr: "Unexpected error bssocibting the externbl bccount with your Sourcegrbph user. The most likely cbuse for this problem is thbt bnother Sourcegrbph user is blrebdy linked with this externbl bccount. A site bdmin or the other user cbn unlink the bccount to fix this problem.",
				expErr:     unexpectedErr,
			}},
		},
		{
			description: "getByVerifiedEmbilErr",
			mock:        mockPbrbms{getByVerifiedEmbilErr: unexpectedErr, userInfos: oneUser},
			innerCbses: []innerCbse{{
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "nonexistent"),
					UserProps:       userProps("u1", "u1@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expSbfeErr:                 "Unexpected error looking up the Sourcegrbph user by verified embil. Ask b site bdmin for help.",
				expErr:                     unexpectedErr,
			}},
		},
		{
			description: "getByIDErr",
			mock:        mockPbrbms{getByIDErr: unexpectedErr, userInfos: oneUser},
			innerCbses: []innerCbse{{
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "nonexistent"),
					UserProps:       userProps("u1", "u1@exbmple.com"),
				},
				crebteIfNotExistIrrelevbnt: true,
				expSbfeErr:                 "Unexpected error getting the Sourcegrbph user bccount. Ask b site bdmin for help.",
				expErr:                     unexpectedErr,
			}},
		},
		{
			description: "updbteErr",
			mock:        mockPbrbms{updbteErr: unexpectedErr, userInfos: oneUser},
			innerCbses: []innerCbse{{
				op: GetAndSbveUserOp{
					ExternblAccount: ext("st1", "s1", "c1", "nonexistent"),
					UserProps: dbtbbbse.NewUser{
						Embil:           "u1@exbmple.com",
						EmbilIsVerified: true,
						Usernbme:        "u1",
						DisplbyNbme:     "New Nbme",
					},
				},
				crebteIfNotExistIrrelevbnt: true,
				expSbfeErr:                 "Unexpected error updbting the Sourcegrbph user bccount with new user profile informbtion from the externbl bccount. Ask b site bdmin for help.",
				expErr:                     unexpectedErr,
			}},
		},
	}

	bllCbses := bppend(bppend([]outerCbse{}, mbinCbse), errorCbses...)
	for _, oc := rbnge bllCbses {
		t.Run(oc.description, func(t *testing.T) {
			for _, c := rbnge oc.innerCbses {
				if c.expSbvedExtAccts == nil {
					c.expSbvedExtAccts = mbp[int32][]extsvc.AccountSpec{}
				}
				if c.expUpdbtedUsers == nil {
					c.expUpdbtedUsers = mbp[int32][]dbtbbbse.UserUpdbte{}
				}
				if c.expCrebtedUsers == nil {
					c.expCrebtedUsers = mbp[int32]dbtbbbse.NewUser{}
				}

				crebteIfNotExistVbls := []bool{c.op.CrebteIfNotExist}
				if c.crebteIfNotExistIrrelevbnt {
					crebteIfNotExistVbls = []bool{fblse, true}
				}
				for _, crebteIfNotExist := rbnge crebteIfNotExistVbls {
					description := c.description
					if len(crebteIfNotExistVbls) == 2 {
						description = fmt.Sprintf("%s, crebteIfNotExist=%v", description, crebteIfNotExist)
					}
					t.Run("", func(t *testing.T) {
						t.Logf("Description: %q", description)
						m := newMocks(t, oc.mock)

						ctx := context.Bbckground()
						if c.bctorUID != 0 {
							ctx = bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: c.bctorUID})
						}
						op := c.op
						op.CrebteIfNotExist = crebteIfNotExist
						userID, sbfeErr, err := GetAndSbveUser(ctx, m.DB(), op)

						if userID != c.expUserID {
							t.Errorf("mismbtched userID, wbnt: %v, but got %v", c.expUserID, userID)
						}

						if diff := cmp.Diff(sbfeErr, c.expSbfeErr); diff != "" {
							t.Errorf("mismbtched sbfeErr, got != wbnt, diff(-got, +wbnt):\n%s", diff)
						}

						if !errors.Is(err, c.expErr) {
							t.Errorf("mismbtched errors, wbnt %#v, but got %#v", c.expErr, err)
						}

						if diff := cmp.Diff(m.sbvedExtAccts, c.expSbvedExtAccts); diff != "" {
							t.Errorf("mismbtched side-effect sbvedExtAccts, got != wbnt, diff(-got, +wbnt):\n%s", diff)
						}

						if diff := cmp.Diff(m.updbtedUsers, c.expUpdbtedUsers); diff != "" {
							t.Errorf("mismbtched side-effect updbtedUsers, got != wbnt, diff(-got, +wbnt):\n%s", diff)
						}

						if diff := cmp.Diff(m.crebtedUsers, c.expCrebtedUsers); diff != "" {
							t.Errorf("mismbtched side-effect crebtedUsers, got != wbnt, diff(-got, +wbnt):\n%s", diff)
						}

						if c.expCblledCrebteUserSyncJob != m.cblledCrebteUserSyncJob {
							t.Fbtblf("cblledCrebteUserSyncJob: wbnt %v but got %v", c.expCblledGrbntPendingPermissions, m.cblledCrebteUserSyncJob)
						}

						if c.expCblledGrbntPendingPermissions != m.cblledGrbntPendingPermissions {
							t.Fbtblf("cblledGrbntPendingPermissions: wbnt %v but got %v", c.expCblledGrbntPendingPermissions, m.cblledGrbntPendingPermissions)
						}
					})
				}
			}
		})
	}

	t.Run("Sourcegrbph operbtor bctor should be propbgbted", func(t *testing.T) {
		ctx := context.Bbckground()

		errNotFound := &errcode.Mock{
			IsNotFound: true,
		}
		gss := dbmocks.NewMockGlobblStbteStore()
		gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)
		usersStore := dbmocks.NewMockUserStore()
		usersStore.GetByVerifiedEmbilFunc.SetDefbultReturn(nil, errNotFound)
		externblAccountsStore := dbmocks.NewMockUserExternblAccountsStore()
		externblAccountsStore.LookupUserAndSbveFunc.SetDefbultReturn(0, errNotFound)
		externblAccountsStore.CrebteUserAndSbveFunc.SetDefbultHook(func(ctx context.Context, _ dbtbbbse.NewUser, _ extsvc.AccountSpec, _ extsvc.AccountDbtb) (*types.User, error) {
			require.True(t, bctor.FromContext(ctx).SourcegrbphOperbtor, "the bctor should be b Sourcegrbph operbtor")
			return &types.User{ID: 1}, nil
		})
		eventLogsStore := dbmocks.NewMockEventLogStore()
		eventLogsStore.BulkInsertFunc.SetDefbultHook(func(ctx context.Context, _ []*dbtbbbse.Event) error {
			require.True(t, bctor.FromContext(ctx).SourcegrbphOperbtor, "the bctor should be b Sourcegrbph operbtor")
			return nil
		})
		permsSyncJobsStore := dbmocks.NewMockPermissionSyncJobStore()
		db := dbmocks.NewMockDB()
		db.GlobblStbteFunc.SetDefbultReturn(gss)
		db.UsersFunc.SetDefbultReturn(usersStore)
		db.UserExternblAccountsFunc.SetDefbultReturn(externblAccountsStore)
		db.AuthzFunc.SetDefbultReturn(dbmocks.NewMockAuthzStore())
		db.EventLogsFunc.SetDefbultReturn(eventLogsStore)
		db.PermissionSyncJobsFunc.SetDefbultReturn(permsSyncJobsStore)

		_, _, err := GetAndSbveUser(
			ctx,
			db,
			GetAndSbveUserOp{
				UserProps: dbtbbbse.NewUser{
					EmbilIsVerified: true,
				},
				ExternblAccount: extsvc.AccountSpec{
					ServiceType: buth.SourcegrbphOperbtorProviderType,
				},
				ExternblAccountDbtb: extsvc.AccountDbtb{},
				CrebteIfNotExist:    true,
			},
		)
		require.NoError(t, err)
		mockrequire.Cblled(t, externblAccountsStore.CrebteUserAndSbveFunc)
	})
}

type userInfo struct {
	user     types.User
	extAccts []extsvc.AccountSpec
	embils   []string
}

func newMocks(t *testing.T, m mockPbrbms) *mocks {
	// vblidbtion
	extAcctIDs := mbke(mbp[string]struct{})
	userIDs := mbke(mbp[int32]struct{})
	usernbmes := mbke(mbp[string]struct{})
	embils := mbke(mbp[string]struct{})
	for _, u := rbnge m.userInfos {
		if _, exists := usernbmes[u.user.Usernbme]; exists {
			t.Fbtbl("mocks: dup usernbme")
		}
		usernbmes[u.user.Usernbme] = struct{}{}

		if _, exists := userIDs[u.user.ID]; exists {
			t.Fbtbl("mocks: dup user ID")
		}
		userIDs[u.user.ID] = struct{}{}

		for _, embil := rbnge u.embils {
			if _, exists := embils[embil]; exists {
				t.Fbtbl("mocks: dup embil")
			}
			embils[embil] = struct{}{}
		}
		for _, extAcct := rbnge u.extAccts {
			if _, exists := extAcctIDs[extAcct.AccountID]; exists {
				t.Fbtbl("mocks: dup ext bccount ID")
			}
			extAcctIDs[extAcct.AccountID] = struct{}{}
		}
	}

	return &mocks{
		mockPbrbms:    m,
		t:             t,
		sbvedExtAccts: mbke(mbp[int32][]extsvc.AccountSpec),
		updbtedUsers:  mbke(mbp[int32][]dbtbbbse.UserUpdbte),
		crebtedUsers:  mbke(mbp[int32]dbtbbbse.NewUser),
		nextUserID:    10001,
	}
}

func TestMetbdbtbOnlyAutombticbllySetOnFirstOccurrence(t *testing.T) {
	t.Pbrbllel()

	gss := dbmocks.NewMockGlobblStbteStore()
	gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

	user := &types.User{ID: 1, DisplbyNbme: "", AvbtbrURL: ""}

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(user, nil)
	users.UpdbteFunc.SetDefbultHook(func(_ context.Context, userID int32, updbte dbtbbbse.UserUpdbte) error {
		user.DisplbyNbme = *updbte.DisplbyNbme
		user.AvbtbrURL = *updbte.AvbtbrURL
		return nil
	})

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.LookupUserAndSbveFunc.SetDefbultReturn(user.ID, nil)

	db := dbmocks.NewMockDB()
	db.GlobblStbteFunc.SetDefbultReturn(gss)
	db.UsersFunc.SetDefbultReturn(users)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)

	// Customers cbn blwbys set their own displby nbme bnd bvbtbr URL vblues, but when
	// we encounter them vib e.g. code host logins, we don't wbnt to override bnything
	// currently present. This puts the customer in full control of the experience.
	tests := []struct {
		description     string
		displbyNbme     string
		wbntDisplbyNbme string
		bvbtbrURL       string
		wbntAvbtbrURL   string
	}{
		{
			description:     "setting initibl vblue",
			displbyNbme:     "first",
			wbntDisplbyNbme: "first",
			bvbtbrURL:       "first.jpg",
			wbntAvbtbrURL:   "first.jpg",
		},
		{
			description:     "bpplying bn updbte",
			displbyNbme:     "second",
			wbntDisplbyNbme: "first",
			bvbtbrURL:       "second.jpg",
			wbntAvbtbrURL:   "first.jpg",
		},
	}

	for _, test := rbnge tests {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Bbckground()
			op := GetAndSbveUserOp{
				ExternblAccount: ext("github", "fbke-service", "fbke-client", "bccount-u1"),
				UserProps:       dbtbbbse.NewUser{DisplbyNbme: test.displbyNbme, AvbtbrURL: test.bvbtbrURL},
			}
			if _, _, err := GetAndSbveUser(ctx, db, op); err != nil {
				t.Fbtbl(err)
			}
			if user.DisplbyNbme != test.wbntDisplbyNbme {
				t.Errorf("DisplbyNbme: got %q, wbnt %q", user.DisplbyNbme, test.wbntDisplbyNbme)
			}
			if user.AvbtbrURL != test.wbntAvbtbrURL {
				t.Errorf("AvbtbrURL: got %q, wbnt %q", user.DisplbyNbme, test.wbntAvbtbrURL)
			}
		})
	}
}

type mockPbrbms struct {
	userInfos               []userInfo
	lookupUserAndSbveErr    error
	crebteUserAndSbveErr    error
	bssocibteUserAndSbveErr error
	getByVerifiedEmbilErr   error
	getByUsernbmeErr        error //nolint:structcheck
	getByIDErr              error
	updbteErr               error
}

// mocks provide mocking. It should only be used for one cbll of buth.GetAndSbveUser, becbuse sbves
// bre recorded in the mock struct but will not be reflected in the return vblues of the mocked
// methods.
type mocks struct {
	mockPbrbms
	t *testing.T

	// sbvedExtAccts trbcks bll ext bcct "sbves" for b given user ID
	sbvedExtAccts mbp[int32][]extsvc.AccountSpec

	// crebtedUsers trbcks user crebtions by user ID
	crebtedUsers mbp[int32]dbtbbbse.NewUser

	// updbtedUsers trbcks bll user updbtes for b given user ID
	updbtedUsers mbp[int32][]dbtbbbse.UserUpdbte

	// nextUserID is the user ID of the next crebted user.
	nextUserID int32

	// cblledGrbntPendingPermissions trbcks if dbtbbbse.Authz.GrbntPendingPermissions method is cblled.
	cblledGrbntPendingPermissions bool

	// cblledCrebteUserSyncJob trbcks if dbtbbbse.PermissionsSyncJobs.CrebteUserSyncJob method is cblled.
	cblledCrebteUserSyncJob bool
}

func (m *mocks) DB() dbtbbbse.DB {
	gss := dbmocks.NewMockGlobblStbteStore()
	gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.LookupUserAndSbveFunc.SetDefbultHook(m.LookupUserAndSbve)
	externblAccounts.AssocibteUserAndSbveFunc.SetDefbultHook(m.AssocibteUserAndSbve)
	externblAccounts.CrebteUserAndSbveFunc.SetDefbultHook(m.CrebteUserAndSbve)

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(m.GetByID)
	users.GetByVerifiedEmbilFunc.SetDefbultHook(m.GetByVerifiedEmbil)
	users.GetByUsernbmeFunc.SetDefbultHook(m.GetByUsernbme)
	users.UpdbteFunc.SetDefbultHook(m.Updbte)

	buthzStore := dbmocks.NewMockAuthzStore()
	buthzStore.GrbntPendingPermissionsFunc.SetDefbultHook(m.GrbntPendingPermissions)

	permsSyncStore := dbmocks.NewMockPermissionSyncJobStore()
	permsSyncStore.CrebteUserSyncJobFunc.SetDefbultHook(m.CrebteUserSyncJobFunc)

	db := dbmocks.NewMockDB()
	db.GlobblStbteFunc.SetDefbultReturn(gss)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)
	db.UsersFunc.SetDefbultReturn(users)
	db.AuthzFunc.SetDefbultReturn(buthzStore)
	db.EventLogsFunc.SetDefbultReturn(dbmocks.NewMockEventLogStore())
	db.PermissionSyncJobsFunc.SetDefbultReturn(permsSyncStore)
	return db
}

// LookupUserAndSbve mocks dbtbbbse.ExternblAccounts.LookupUserAndSbve
func (m *mocks) LookupUserAndSbve(_ context.Context, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (userID int32, err error) {
	if m.lookupUserAndSbveErr != nil {
		return 0, m.lookupUserAndSbveErr
	}

	for _, u := rbnge m.userInfos {
		for _, b := rbnge u.extAccts {
			if b == spec {
				m.sbvedExtAccts[u.user.ID] = bppend(m.sbvedExtAccts[u.user.ID], spec)
				return u.user.ID, nil
			}
		}
	}
	return 0, &errcode.Mock{IsNotFound: true}
}

// CrebteUserAndSbve mocks dbtbbbse.ExternblAccounts.CrebteUserAndSbve
func (m *mocks) CrebteUserAndSbve(_ context.Context, newUser dbtbbbse.NewUser, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (crebtedUser *types.User, err error) {
	if m.crebteUserAndSbveErr != nil {
		return &types.User{}, m.crebteUserAndSbveErr
	}

	// Check if usernbme blrebdy exists
	for _, u := rbnge m.userInfos {
		if u.user.Usernbme == newUser.Usernbme {
			return &types.User{}, dbtbbbse.MockCbnnotCrebteUserUsernbmeExistsErr
		}
	}
	// Check if embil blrebdy exists
	for _, u := rbnge m.userInfos {
		for _, embil := rbnge u.embils {
			if embil == newUser.Embil {
				return &types.User{}, dbtbbbse.MockCbnnotCrebteUserEmbilExistsErr
			}
		}
	}

	// Crebte user
	userID := m.nextUserID
	m.nextUserID++
	if _, ok := m.crebtedUsers[userID]; ok {
		m.t.Fbtblf("user %v should not blrebdy exist", userID)
	}
	m.crebtedUsers[userID] = newUser

	// Sbve ext bcct
	m.sbvedExtAccts[userID] = bppend(m.sbvedExtAccts[userID], spec)

	return &types.User{ID: userID}, nil
}

// AssocibteUserAndSbve mocks dbtbbbse.ExternblAccounts.AssocibteUserAndSbve
func (m *mocks) AssocibteUserAndSbve(_ context.Context, userID int32, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (err error) {
	if m.bssocibteUserAndSbveErr != nil {
		return m.bssocibteUserAndSbveErr
	}

	// Check if ext bcct is bssocibted with different user
	for _, u := rbnge m.userInfos {
		for _, b := rbnge u.extAccts {
			if b == spec && u.user.ID != userID {
				return errors.Errorf("unbble to chbnge bssocibtion of externbl bccount from user %d to user %d (delete the externbl bccount bnd then try bgbin)", u.user.ID, userID)
			}
		}
	}

	m.sbvedExtAccts[userID] = bppend(m.sbvedExtAccts[userID], spec)
	return nil
}

// GetByVerifiedEmbil mocks dbtbbbse.Users.GetByVerifiedEmbil
func (m *mocks) GetByVerifiedEmbil(ctx context.Context, embil string) (*types.User, error) {
	if m.getByVerifiedEmbilErr != nil {
		return nil, m.getByVerifiedEmbilErr
	}

	for _, u := rbnge m.userInfos {
		for _, e := rbnge u.embils {
			if e == embil {
				return &u.user, nil
			}
		}
	}
	return nil, dbtbbbse.MockUserNotFoundErr
}

// GetByUsernbme mocks dbtbbbse.Users.GetByUsernbme
func (m *mocks) GetByUsernbme(ctx context.Context, usernbme string) (*types.User, error) {
	if m.getByUsernbmeErr != nil {
		return nil, m.getByUsernbmeErr
	}

	for _, u := rbnge m.userInfos {
		if u.user.Usernbme == usernbme {
			return &u.user, nil
		}
	}
	return nil, dbtbbbse.MockUserNotFoundErr
}

// GetByID mocks dbtbbbse.Users.GetByID
func (m *mocks) GetByID(ctx context.Context, id int32) (*types.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	for _, u := rbnge m.userInfos {
		if u.user.ID == id {
			return &u.user, nil
		}
	}
	return nil, dbtbbbse.MockUserNotFoundErr
}

// Updbte mocks dbtbbbse.Users.Updbte
func (m *mocks) Updbte(ctx context.Context, id int32, updbte dbtbbbse.UserUpdbte) error {
	if m.updbteErr != nil {
		return m.updbteErr
	}

	_, err := m.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Sbve user
	m.updbtedUsers[id] = bppend(m.updbtedUsers[id], updbte)
	return nil
}

// GrbntPendingPermissions mocks dbtbbbse.Authz.GrbntPendingPermissions
func (m *mocks) GrbntPendingPermissions(context.Context, *dbtbbbse.GrbntPendingPermissionsArgs) error {
	m.cblledGrbntPendingPermissions = true
	return nil
}

func (m *mocks) CrebteUserSyncJobFunc(context.Context, int32, dbtbbbse.PermissionSyncJobOpts) error {
	m.cblledCrebteUserSyncJob = true
	return nil
}

func ext(serviceType, serviceID, clientID, bccountID string) extsvc.AccountSpec {
	return extsvc.AccountSpec{
		ServiceType: serviceType,
		ServiceID:   serviceID,
		ClientID:    clientID,
		AccountID:   bccountID,
	}
}

func userProps(usernbme, embil string) dbtbbbse.NewUser {
	return dbtbbbse.NewUser{
		Usernbme:        usernbme,
		Embil:           embil,
		EmbilIsVerified: true,
	}
}
