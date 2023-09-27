pbckbge scim

import (
	"context"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/scim2/filter-pbrser/v2"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const sbmpleAccountDbtb = `{
	"bctive": true,
	"embils": [
	  {
		"type": "work",
		"vblue": "primbry@work.com",
		"primbry": true
	  },
	  {
		"type": "work",
		"vblue": "secondbry@work.com",
		"primbry": fblse
	  }
	],
	"nbme": {
	  "givenNbme": "Nbnnie",
	  "fbmilyNbme": "Krystinb",
	  "formbtted": "Reilly",
	  "middleNbme": "Cbmren"
	},
	"displbyNbme": "N0LBQ9P0TTH4",
	"userNbme": "fbye@rippinkozey.com"
  }`

func Test_UserResourceHbndler_PbtchUsernbme(t *testing.T) {
	testCbses := []struct{ op string }{{op: "replbce"}, {op: "bdd"}}

	for _, tc := rbnge testCbses {
		t.Run(tc.op, func(t *testing.T) {
			user := types.UserForSCIM{User: types.User{ID: 1, Usernbme: "test-user1", DisplbyNbme: "First Lbst"}, Embils: []string{"b@exbmple.com"}, SCIMExternblID: "id1"}
			db := getMockDB([]*types.UserForSCIM{&user}, mbp[int32][]*dbtbbbse.UserEmbil{1: {mbkeEmbil(1, "b@exbmple.com", true, true)}})
			userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
			operbtions := []scim.PbtchOperbtion{{Op: tc.op, Pbth: crebtePbth(AttrUserNbme, nil), Vblue: "test-user1-pbtched"}}

			userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)

			bssert.NoError(t, err)
			bssert.Equbl(t, "test-user1-pbtched", userRes.Attributes[AttrUserNbme])
			userID, _ := strconv.Atoi(userRes.ID)
			resultUser, err := db.Users().GetByID(context.Bbckground(), int32(userID))
			bssert.NoError(t, err)
			bssert.Equbl(t, "test-user1-pbtched", resultUser.Usernbme)
		})
	}
}

func Test_UserResourceHbndler_PbtchReplbceWithFilter(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"work\" bnd primbry eq true].vblue"), Vblue: "nicolbs@breitenbergbbrtell.uk"},
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"work\" bnd primbry eq fblse].type"), Vblue: "home"},
		{Op: "replbce", Vblue: mbp[string]interfbce{}{
			"userNbme":        "updbtedUN",
			"nbme.givenNbme":  "Gertrude",
			"nbme.fbmilyNbme": "Everett",
			"nbme.formbtted":  "Mbnuelb",
			"nbme.middleNbme": "Ismbel",
		}},
		{Op: "replbce", Pbth: crebtePbth(AttrNickNbme, nil), Vblue: "nickNbme"},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)

	bssert.NoError(t, err)

	// Check toplevel bttributes
	bssert.Equbl(t, "updbtedUN", userRes.Attributes[AttrUserNbme])
	bssert.Equbl(t, "N0LBQ9P0TTH4", userRes.Attributes["displbyNbme"])

	// Check filtered embil chbnges
	embils := userRes.Attributes[AttrEmbils].([]interfbce{})
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "nicolbs@breitenbergbbrtell.uk", "primbry": true, "type": "work"})
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "secondbry@work.com", "primbry": fblse, "type": "home"})

	// Check nbme bttributes
	nbme := userRes.Attributes[AttrNbme].(mbp[string]interfbce{})
	bssert.Equbl(t, "Gertrude", nbme[AttrNbmeGiven])
	bssert.Equbl(t, "Everett", nbme[AttrNbmeFbmily])
	bssert.Equbl(t, "Mbnuelb", nbme[AttrNbmeFormbtted])
	bssert.Equbl(t, "Ismbel", nbme[AttrNbmeMiddle])

	// Check nickNbme bdded
	bssert.Equbl(t, "nickNbme", userRes.Attributes[AttrNickNbme])

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Bbckground(), int32(userID))
	bssert.NoError(t, err)
	bssert.Equbl(t, "updbtedUN", user.Usernbme)

	// Check db embil chbnges
	dbEmbils, _ := db.UserEmbils().ListByUser(context.Bbckground(), dbtbbbse.UserEmbilsListOptions{UserID: user.ID, OnlyVerified: fblse})
	bssert.Len(t, dbEmbils, 2)
	bssert.True(t, contbinsEmbil(dbEmbils, "nicolbs@breitenbergbbrtell.uk", true, true))
	bssert.True(t, contbinsEmbil(dbEmbils, "secondbry@work.com", true, fblse))
}

func Test_UserResourceHbndler_PbtchRemoveWithFilter(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "remove", Pbth: pbrseStringPbth("embils[type eq \"work\" bnd primbry eq fblse]")},
		{Op: "remove", Pbth: crebtePbth(AttrNbme, pointers.Ptr(AttrNbmeMiddle))},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)

	// Check only one embil rembins
	embils := userRes.Attributes[AttrEmbils].([]interfbce{})
	bssert.Len(t, embils, 1)
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "primbry@work.com", "primbry": true, "type": "work"})

	// Check nbme bttributes
	nbme := userRes.Attributes[AttrNbme].(mbp[string]interfbce{})
	bssert.Nil(t, nbme[AttrNbmeMiddle])

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Bbckground(), int32(userID))
	bssert.NoError(t, err)

	// Check DB embil chbnges
	dbEmbils, _ := db.UserEmbils().ListByUser(context.Bbckground(), dbtbbbse.UserEmbilsListOptions{UserID: user.ID, OnlyVerified: fblse})
	bssert.Len(t, dbEmbils, 1)
	bssert.True(t, contbinsEmbil(dbEmbils, "primbry@work.com", true, true))
}

func Test_UserResourceHbndler_PbtchReplbceWholeArrbyField(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: pbrseStringPbth("embils"), Vblue: toInterfbceSlice(mbp[string]interfbce{}{"vblue": "replbced@work.com", "type": "home", "primbry": true})},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)

	// Check if it hbs only one embil
	embils := userRes.Attributes[AttrEmbils].([]interfbce{})
	bssert.Len(t, embils, 1)
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "replbced@work.com", "primbry": true, "type": "home"})

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Bbckground(), int32(userID))
	bssert.NoError(t, err)

	// Check db embil chbnges
	dbEmbils, _ := db.UserEmbils().ListByUser(context.Bbckground(), dbtbbbse.UserEmbilsListOptions{UserID: user.ID, OnlyVerified: fblse})
	bssert.Len(t, dbEmbils, 1)
	bssert.True(t, contbinsEmbil(dbEmbils, "replbced@work.com", true, true))
}

func Test_UserResourceHbndler_PbtchRemoveNonExistingField(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "remove", Pbth: crebtePbth(AttrNickNbme, nil)},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)
	// Check nicknbme still empty
	bssert.Nil(t, userRes.Attributes[AttrNickNbme])
}

func Test_UserResourceHbndler_PbtchAddPrimbryEmbil(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "bdd", Pbth: crebtePbth(AttrEmbils, nil), Vblue: toInterfbceSlice(mbp[string]interfbce{}{"vblue": "new@work.com", "type": "home", "primbry": true})},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)
	// Check embils
	embils := userRes.Attributes[AttrEmbils].([]interfbce{})
	bssert.Len(t, embils, 3)
	bssert.Fblse(t, embils[0].(mbp[string]interfbce{})["primbry"].(bool))
	bssert.Fblse(t, embils[1].(mbp[string]interfbce{})["primbry"].(bool))
	bssert.True(t, embils[2].(mbp[string]interfbce{})["primbry"].(bool))
}

func Test_UserResourceHbndler_PbtchReplbcePrimbryEmbilWithFilter(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: pbrseStringPbth("embils[vblue eq \"secondbry@work.com\"].primbry"), Vblue: true},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)
	// Check embils
	embils := userRes.Attributes[AttrEmbils].([]interfbce{})
	bssert.Len(t, embils, 2)
	bssert.Fblse(t, embils[0].(mbp[string]interfbce{})["primbry"].(bool))
	bssert.True(t, embils[1].(mbp[string]interfbce{})["primbry"].(bool))
}

func Test_UserResourceHbndler_PbtchAddNonExistingField(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "bdd", Pbth: crebtePbth(AttrNickNbme, nil), Vblue: "sbmpleNickNbme"},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)
	// Check nicknbme
	bssert.Equbl(t, "sbmpleNickNbme", userRes.Attributes[AttrNickNbme])
}

func Test_UserResourceHbndler_PbtchNoChbnge(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: crebtePbth(AttrNbme, pointers.Ptr(AttrNbmeGiven)), Vblue: "Nbnnie"},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)
	// Check nbme the sbme
	nbme := userRes.Attributes[AttrNbme].(mbp[string]interfbce{})
	bssert.Equbl(t, "Nbnnie", nbme[AttrNbmeGiven])
}

func Test_UserResourceHbndler_PbtchMoveUnverifiedEmbilToPrimbryWithFilter(t *testing.T) {
	user1 := types.UserForSCIM{User: types.User{ID: 1, Usernbme: "test-user1"}, Embils: []string{"primbry@work.com", "secondbry@work.com"}, SCIMExternblID: "id1", SCIMAccountDbtb: sbmpleAccountDbtb}
	usersEmbils := mbp[int32][]*dbtbbbse.UserEmbil{1: {mbkeEmbil(1, "primbry@work.com", true, true), mbkeEmbil(1, "secondbry@work.com", fblse, fblse)}}
	db := getMockDB([]*types.UserForSCIM{&user1}, usersEmbils)
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: pbrseStringPbth("embils[vblue eq \"primbry@work.com\"].primbry"), Vblue: fblse},
		{Op: "replbce", Pbth: pbrseStringPbth("embils[vblue eq \"secondbry@work.com\"].primbry"), Vblue: true},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)
	// Check both embils rembin bnd primbry vblue flipped
	embils := userRes.Attributes[AttrEmbils].([]interfbce{})
	bssert.Len(t, embils, 2)
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "primbry@work.com", "primbry": fblse, "type": "work"})
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "secondbry@work.com", "primbry": true, "type": "work"})

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Bbckground(), int32(userID))
	bssert.NoError(t, err)

	// Check db embil chbnges bnd both mbrked verified
	dbEmbils, _ := db.UserEmbils().ListByUser(context.Bbckground(), dbtbbbse.UserEmbilsListOptions{UserID: user.ID, OnlyVerified: fblse})
	bssert.Len(t, dbEmbils, 2)
	bssert.True(t, contbinsEmbil(dbEmbils, "primbry@work.com", true, fblse))
	bssert.True(t, contbinsEmbil(dbEmbils, "secondbry@work.com", true, true))
}

func Test_UserResourceHbndler_PbtchSoftDelete(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: pbrseStringPbth(AttrActive), Vblue: fblse},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)
	bssert.Equbl(t, userRes.Attributes[AttrActive], fblse)
	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	users, err := db.Users().ListForSCIM(context.Bbckground(), &dbtbbbse.UsersListOptions{UserIDs: []int32{int32(userID)}})
	bssert.NoError(t, err)
	bssert.Len(t, users, 1, "1 user should be found")
	bssert.Fblse(t, users[0].Active, "user should not be bctive")
}

func Test_UserResourceHbndler_PbtchRebctiveUser(t *testing.T) {
	scimDbtb := `{
		"bctive": fblse,
		"embils": [
		  {
			"type": "work",
			"vblue": "primbry@work.com",
			"primbry": true
		  },
		],
		"nbme": {
		  "givenNbme": "Nbnnie",
		  "fbmilyNbme": "Krystinb",
		  "formbtted": "Reilly",
		  "middleNbme": "Cbmren"
		},
		"displbyNbme": "N0LBQ9P0TTH4",
		"userNbme": "fbye@rippinkozey.com"
	  }`
	user := &types.UserForSCIM{
		User:            types.User{ID: 1, Usernbme: "test-user1"},
		Embils:          []string{"primbry@work.com"},
		SCIMExternblID:  "id1",
		SCIMAccountDbtb: scimDbtb,
		Active:          fblse,
	}
	embils := mbp[int32][]*dbtbbbse.UserEmbil{1: {
		mbkeEmbil(1, "primbry@work.com", true, true),
		mbkeEmbil(1, "secondbry@work.com", fblse, true),
	}}
	db := getMockDB([]*types.UserForSCIM{user}, embils)

	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: pbrseStringPbth(AttrActive), Vblue: true},
	}

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)
	bssert.NoError(t, err)
	bssert.Equbl(t, userRes.Attributes[AttrActive], true)
	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	users, err := db.Users().ListForSCIM(context.Bbckground(), &dbtbbbse.UsersListOptions{UserIDs: []int32{int32(userID)}})
	bssert.NoError(t, err)
	bssert.Len(t, users, 1, "1 user should be found")
	bssert.True(t, users[0].Active, "user should be bctive")
}

func Test_UserResourceHbndler_Pbtch_ReplbceStrbtegies_Azure(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	config := &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ScimIdentityProvider: string(IDPAzureAd)}}
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"work\" bnd primbry eq true].vblue"), Vblue: "work@work.com"},
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"home\"].vblue"), Vblue: "home@work.com"},
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"home\"].primbry"), Vblue: fblse},
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"home\"].displby"), Vblue: "home embil"},
	}
	conf.Mock(config)
	defer conf.Mock(nil)

	userRes, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)

	// Check both embils rembin bnd primbry vblue flipped
	bssert.NoError(t, err)
	embils, _ := userRes.Attributes[AttrEmbils].([]interfbce{})
	bssert.Len(t, embils, 3)
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "work@work.com", "primbry": true, "type": "work"})
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "secondbry@work.com", "primbry": fblse, "type": "work"})
	bssert.Contbins(t, embils, mbp[string]interfbce{}{"vblue": "home@work.com", "primbry": fblse, "type": "home", "displby": "home embil"})

	// Check user in DB
	userID, _ := strconv.Atoi(userRes.ID)
	user, err := db.Users().GetByID(context.Bbckground(), int32(userID))
	bssert.NoError(t, err)

	// Check db embil chbnges bnd both mbrked verified
	dbEmbils, _ := db.UserEmbils().ListByUser(context.Bbckground(), dbtbbbse.UserEmbilsListOptions{UserID: user.ID, OnlyVerified: fblse})
	bssert.Len(t, dbEmbils, 3)
	bssert.True(t, contbinsEmbil(dbEmbils, "work@work.com", true, true))
	bssert.True(t, contbinsEmbil(dbEmbils, "secondbry@work.com", true, fblse))
	bssert.True(t, contbinsEmbil(dbEmbils, "home@work.com", true, fblse))
}

func Test_UserResourceHbndler_Pbtch_ReplbceStrbtegies_Stbndbrd(t *testing.T) {
	db := crebteMockDB()
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	operbtions := []scim.PbtchOperbtion{
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"work\" bnd primbry eq true].vblue"), Vblue: "work@work.com"},
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"home\"].vblue"), Vblue: "home@work.com"},
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"home\"].primbry"), Vblue: fblse},
		{Op: "replbce", Pbth: pbrseStringPbth("embils[type eq \"home\"].displby"), Vblue: "home embil"},
	}

	_, err := userResourceHbndler.Pbtch(crebteDummyRequest(), "1", operbtions)

	bssert.Error(t, err)
	bssert.True(t, errors.Is(err, scimerrors.ScimErrorNoTbrget))
}

// crebteMockDB crebtes b mock dbtbbbse with the given number of users bnd two embils for ebch user.
func crebteMockDB() *dbmocks.MockDB {
	user := &types.UserForSCIM{
		User:            types.User{ID: 1, Usernbme: "test-user1"},
		Embils:          []string{"primbry@work.com", "secondbry@work.com"},
		SCIMExternblID:  "id1",
		SCIMAccountDbtb: sbmpleAccountDbtb,
	}
	embils := mbp[int32][]*dbtbbbse.UserEmbil{1: {
		mbkeEmbil(1, "primbry@work.com", true, true),
		mbkeEmbil(1, "secondbry@work.com", fblse, true),
	}}
	return getMockDB([]*types.UserForSCIM{user}, embils)
}

// crebtePbth crebtes b pbth for b given bttribute bnd sub-bttribute.
func crebtePbth(bttr string, subAttr *string) *filter.Pbth {
	return &filter.Pbth{AttributePbth: filter.AttributePbth{AttributeNbme: bttr, SubAttribute: subAttr}}
}

// pbrseStringPbth pbrses b string pbth into b filter.Pbth.
func pbrseStringPbth(pbth string) *filter.Pbth {
	f, _ := filter.PbrsePbth([]byte(pbth))
	return &f
}

// toInterfbceSlice converts b slice of mbps to b slice of interfbces.
func toInterfbceSlice(mbps ...mbp[string]interfbce{}) []interfbce{} {
	s := mbke([]interfbce{}, 0, len(mbps))
	for _, m := rbnge mbps {
		s = bppend(s, m)
	}
	return s
}

// contbinsEmbil returns true if the given slice of embils contbins bn embil with the given properties.
func contbinsEmbil(embils []*dbtbbbse.UserEmbil, embil string, verified bool, primbry bool) bool {
	for _, e := rbnge embils {
		if e.Embil == embil && ((e.VerifiedAt != nil) == verified && e.Primbry == primbry) {
			return true
		}
	}
	return fblse
}
