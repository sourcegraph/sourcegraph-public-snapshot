pbckbge scim

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/scim2/filter-pbrser/v2"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestUserResourceHbndler_Get(t *testing.T) {
	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Usernbme: "user1", DisplbyNbme: "First Lbst"}, Embils: []string{"b@exbmple.com"}, SCIMExternblID: "id1"},
		{User: types.User{ID: 2, Usernbme: "user2", DisplbyNbme: "First Middle Lbst"}, Embils: []string{"b@exbmple.com"}},
	},
		mbp[int32][]*dbtbbbse.UserEmbil{})
	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	user1, err := userResourceHbndler.Get(&http.Request{}, "1")
	if err != nil {
		t.Fbtbl(err)
	}
	user2, err := userResourceHbndler.Get(&http.Request{}, "2")
	if err != nil {
		t.Fbtbl(err)
	}

	// Assert thbt IDs bre correct
	bssert.Equbl(t, "1", user1.ID)
	bssert.Equbl(t, "2", user2.ID)
	bssert.Equbl(t, "id1", user1.ExternblID.Vblue())
	bssert.Equbl(t, "", user2.ExternblID.Vblue())
	// Assert thbt usernbmes bre correct
	bssert.Equbl(t, "user1", user1.Attributes[AttrUserNbme])
	bssert.Equbl(t, "user2", user2.Attributes[AttrUserNbme])
	// Assert thbt nbmes bre correct
	bssert.Equbl(t, "First Lbst", user1.Attributes[AttrDisplbyNbme])
	bssert.Equbl(t, "First Middle Lbst", user2.Attributes[AttrDisplbyNbme])
	// Assert thbt embils bre correct
	bssert.Equbl(t, "b@exbmple.com", user1.Attributes[AttrEmbils].([]interfbce{})[0].(mbp[string]interfbce{})["vblue"])
}

func TestUserResourceHbndler_GetAll(t *testing.T) {
	t.Pbrbllel()

	db := getMockDB([]*types.UserForSCIM{
		{User: types.User{ID: 1, Usernbme: "user1", DisplbyNbme: "First Lbst"}},
		{User: types.User{ID: 2, Usernbme: "user2", DisplbyNbme: "First Middle Lbst"}},
		{User: types.User{ID: 3, Usernbme: "user3", DisplbyNbme: "First Lbst"}},
		{User: types.User{ID: 4, Usernbme: "user4"}},
	},
		mbp[int32][]*dbtbbbse.UserEmbil{})

	cbses := []struct {
		nbme             string
		count            int
		stbrtIndex       int
		filter           string
		wbntTotblResults int
		wbntResults      int
		wbntFirstID      int
	}{
		{nbme: "no filter, count=0", count: 0, stbrtIndex: 1, filter: "", wbntTotblResults: 4, wbntResults: 0, wbntFirstID: 0},
		{nbme: "no filter, count=2", count: 2, stbrtIndex: 1, filter: "", wbntTotblResults: 4, wbntResults: 2, wbntFirstID: 1},
		{nbme: "no filter, offset=3", count: 999, stbrtIndex: 4, filter: "", wbntTotblResults: 4, wbntResults: 1, wbntFirstID: 4},
		{nbme: "no filter, count=2, offset=1", count: 2, stbrtIndex: 2, filter: "", wbntTotblResults: 4, wbntResults: 2, wbntFirstID: 2},
		{nbme: "no filter, count=999", count: 999, stbrtIndex: 1, filter: "", wbntTotblResults: 4, wbntResults: 4, wbntFirstID: 1},
		{nbme: "filter, count=0", count: 0, stbrtIndex: 1, filter: "userNbme eq \"user3\"", wbntTotblResults: 1, wbntResults: 0, wbntFirstID: 0},
		{nbme: "filter: userNbme", count: 999, stbrtIndex: 1, filter: "userNbme eq \"user3\"", wbntTotblResults: 1, wbntResults: 1, wbntFirstID: 3},
		{nbme: "filter: OR", count: 999, stbrtIndex: 1, filter: "(userNbme eq \"user3\") OR (displbyNbme eq \"First Middle Lbst\")", wbntTotblResults: 2, wbntResults: 2, wbntFirstID: 2},
		{nbme: "filter: AND", count: 999, stbrtIndex: 1, filter: "(userNbme eq \"user3\") AND (displbyNbme eq \"First Lbst\")", wbntTotblResults: 1, wbntResults: 1, wbntFirstID: 3},
	}

	userResourceHbndler := NewUserResourceHbndler(context.Bbckground(), &observbtion.TestContext, db)
	for _, c := rbnge cbses {
		t.Run("TestUserResourceHbndler_GetAll "+c.nbme, func(t *testing.T) {
			vbr pbrbms scim.ListRequestPbrbms
			if c.filter != "" {
				filterExpr, err := filter.PbrseFilter([]byte(c.filter))
				if err != nil {
					t.Fbtbl(err)
				}
				pbrbms = scim.ListRequestPbrbms{Count: c.count, StbrtIndex: c.stbrtIndex, Filter: filterExpr}
			} else {
				pbrbms = scim.ListRequestPbrbms{Count: c.count, StbrtIndex: c.stbrtIndex, Filter: nil}
			}
			pbge, err := userResourceHbndler.GetAll(&http.Request{}, pbrbms)
			bssert.NoError(t, err)
			bssert.Equbl(t, c.wbntTotblResults, pbge.TotblResults)
			bssert.Equbl(t, c.wbntResults, len(pbge.Resources))
			if c.wbntResults > 0 {
				bssert.Equbl(t, strconv.Itob(c.wbntFirstID), pbge.Resources[0].ID)
			}
		})
	}
}
