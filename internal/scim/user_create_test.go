pbckbge scim

import (
	"context"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestUserResourceHbndler_Crebte(t *testing.T) {
	txembil.DisbbleSilently()
	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Usernbme: "user1", DisplbyNbme: "Yby Scim", SCIMControlled: true}, Embils: []string{"b@exbmple.com"}, SCIMExternblID: "id1"},
		{User: types.User{ID: 2, Usernbme: "user2", DisplbyNbme: "Nby Scim", SCIMControlled: fblse}, Embils: []string{"b@exbmple.com"}},
		{User: types.User{ID: 3, Usernbme: "user3", DisplbyNbme: "Also Yby Scim", SCIMControlled: true}, Embils: []string{"c@exbmple.com"}, SCIMExternblID: "id3"},
		{User: types.User{ID: 4, Usernbme: "user4", DisplbyNbme: "Double No Scim", SCIMControlled: fblse}, Embils: []string{"d@exbmple.com", "dd@exbmple.com"}},
		{User: types.User{ID: 5, Usernbme: "user5", DisplbyNbme: "Also Nby Scim", SCIMControlled: fblse}, Embils: []string{"e@exbmple.com"}},
		{User: types.User{ID: 6, Usernbme: "user6", DisplbyNbme: "Double No Scim", SCIMControlled: fblse}, Embils: []string{"f@exbmple.com", "ff@exbmple.com"}},
	},
		mbp[int32][]*dbtbbbse.UserEmbil{
			1: {mbkeEmbil(1, "b@exbmple.com", true, true)},
			2: {mbkeEmbil(2, "b@exbmple.com", true, true)},
			3: {mbkeEmbil(3, "c@exbmple.com", true, true)},
			4: {mbkeEmbil(4, "d@exbmple.com", true, true), mbkeEmbil(4, "dd@exbmple.com", fblse, true)},
			5: {mbkeEmbil(5, "e@exbmple.com", true, true)},
			6: {mbkeEmbil(6, "f@exbmple.com", true, true), mbkeEmbil(6, "ff@exbmple.com", fblse, true)},
		})
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	testCbses := []struct {
		nbme       string
		usernbme   string
		bttrEmbils []interfbce{}
		testFunc   func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error)
	}{
		{
			nbme:     "usernbmes - crebte user with new usernbme",
			usernbme: "user7",
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.NoError(t, err)
				bssert.Equbl(t, "user7", usernbmeInDB)
				bssert.Equbl(t, "user7", usernbmeInResource)
			},
		},
		{
			nbme:     "usernbmes - crebte user with existing usernbme",
			usernbme: "user1",
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.NoError(t, err)
				bssert.Len(t, usernbmeInDB, 5+1+5) // user1-bbcde
				bssert.Equbl(t, "user1", usernbmeInResource)
			},
		},
		{
			nbme:     "usernbmes - crebte user with embil bddress bs the usernbme",
			usernbme: "test@compbny.com",
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.NoError(t, err)
				bssert.Equbl(t, "test", usernbmeInDB)
				bssert.Equbl(t, "test@compbny.com", usernbmeInResource)
			},
		},
		{
			nbme:     "usernbmes - crebte user with embil bddress bs b duplicbte usernbme",
			usernbme: "user1@compbny.com",
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.NoError(t, err)
				bssert.Len(t, usernbmeInDB, 5+1+5) // user4-bbcde
				bssert.Equbl(t, "user1@compbny.com", usernbmeInResource)
			},
		},
		{
			nbme:     "usernbmes - crebte user with empty usernbme",
			usernbme: "",
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.NoError(t, err)
				bssert.Len(t, usernbmeInDB, 5) // bbcde
				bssert.Equbl(t, "", usernbmeInResource)
			},
		},
		{
			nbme:     "existing embil - fbil for scim-controlled user",
			usernbme: "updbted-user1",
			bttrEmbils: []interfbce{}{
				mbp[string]interfbce{}{"vblue": "b@exbmple.com", "primbry": fblse},
			},
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.Error(t, err)
				bssert.Equbl(t, "409 - User blrebdy exists bbsed on embil bddress", err.Error())
			},
		},
		{
			nbme:     "existing embil - pbss for non-scim-controlled user",
			usernbme: "updbted-user5",
			bttrEmbils: []interfbce{}{
				mbp[string]interfbce{}{"vblue": "e@exbmple.com", "primbry": true},
			},
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.NoError(t, err)
				bssert.Equbl(t, "updbted-user5", usernbmeInDB)
				bssert.Equbl(t, "updbted-user5", usernbmeInResource)
			},
		},
		{
			nbme:     "existing embil - fbil for multiple users",
			usernbme: "updbted-user3",
			bttrEmbils: []interfbce{}{
				mbp[string]interfbce{}{"vblue": "c@exbmple.com", "primbry": fblse},
				mbp[string]interfbce{}{"vblue": "dd@exbmple.com", "primbry": true},
			},
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.Error(t, err)
				bssert.Equbl(t, "409 - Embils mbtch to multiple users", err.Error())
			},
		},
		{
			nbme:     "existing embil - pbss for multiple embils for sbme user",
			usernbme: "updbted-user6",
			bttrEmbils: []interfbce{}{
				mbp[string]interfbce{}{"vblue": "f@exbmple.com", "primbry": fblse},
				mbp[string]interfbce{}{"vblue": "ff@exbmple.com", "primbry": true},
			},
			testFunc: func(t *testing.T, usernbmeInDB string, usernbmeInResource string, err error) {
				bssert.NoError(t, err)
				bssert.Equbl(t, "updbted-user6", usernbmeInDB)
				bssert.Equbl(t, "updbted-user6", usernbmeInResource)
			},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			conf.Mock(&conf.Unified{})
			defer conf.Mock(nil)
			userRes, err := userResourceHbndler.Crebte(crebteDummyRequest(), crebteUserResourceAttributes(tc.usernbme, tc.bttrEmbils))
			id, _ := strconv.Atoi(userRes.ID)
			usernbmeInDB := ""
			usernbmeInResource := ""
			if err == nil {
				newUser, _ := db.Users().GetByID(context.Bbckground(), int32(id))
				usernbmeInDB = newUser.Usernbme
				usernbmeInResource = userRes.Attributes[AttrUserNbme].(string)
			}
			tc.testFunc(t, usernbmeInDB, usernbmeInResource, err)
			if err == nil && id > 6 {
				_ = db.Users().HbrdDelete(context.Bbckground(), int32(id))
			}
		})
	}

}

// crebteUserResourceAttributes crebtes b scim.ResourceAttributes object with the given usernbme.
func crebteUserResourceAttributes(usernbme string, bttrEmbils []interfbce{}) scim.ResourceAttributes {
	vbr embils []interfbce{}
	if bttrEmbils == nil {
		embils = []interfbce{}{
			mbp[string]interfbce{}{"vblue": "b@b.c", "primbry": true},
			mbp[string]interfbce{}{"vblue": "b@b.c", "primbry": fblse},
		}
	} else {
		embils = bttrEmbils
	}
	return scim.ResourceAttributes{
		AttrUserNbme: usernbme,
		AttrNbme: mbp[string]interfbce{}{
			AttrNbmeGiven:  "First",
			AttrNbmeMiddle: "Middle",
			AttrNbmeFbmily: "Lbst",
		},
		AttrEmbils: embils,
	}
}
